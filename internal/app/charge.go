package app

import (
	"net/http"
	"strconv"
	"time"
)

type chargeReq struct {
	StationID int64 `json:"stationId"`
	GunID     int64 `json:"gunId"`
	UserID    int64 `json:"userId"`
}

func (a *App) targetTenantUser(r *http.Request, userID int64) (int64, error) {
	s := current(r)
	if s.Role == "user" {
		return s.TenantUserID, nil
	}
	if userID <= 0 {
		return 0, fail(400, "请指定用户")
	}
	return a.tenantUserID(s.TenantID, userID)
}

func (a *App) gunBusy(gunID, excludeReservation int64) bool {
	var n int
	_ = a.DB.QueryRow(`SELECT (SELECT COUNT(*) FROM reservations WHERE gun_id=? AND status IN ('pending','active') AND id<>?)
		+ (SELECT COUNT(*) FROM orders WHERE gun_id=? AND status='charging')`, gunID, excludeReservation, gunID).Scan(&n)
	return n > 0
}

func (a *App) createReservation(w http.ResponseWriter, r *http.Request) {
	var req chargeReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	s := current(r)
	tuID, err := a.targetTenantUser(r, req.UserID)
	if err != nil {
		writeErr(w, err)
		return
	}
	st, err := a.getStation(req.StationID, s.TenantID)
	if err != nil || st.Status != "active" {
		writeErr(w, fail(404, "场站不存在或已停用"))
		return
	}
	g, err := a.getGun(req.GunID, s.TenantID)
	if err != nil || g.StationID != st.ID {
		writeErr(w, fail(404, "充电枪不存在"))
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.gunBusy(g.ID, 0) {
		writeErr(w, fail(409, "该充电枪已被预约或正在充电"))
		return
	}
	_, balance, err := a.walletID(s.TenantID, tuID)
	if err != nil {
		writeErr(w, fail(404, "钱包不存在"))
		return
	}
	if balance < st.StartThresholdCents {
		writeErr(w, fail(409, "钱包余额不足，启动需至少 "+moneyText(st.StartThresholdCents)))
		return
	}
	startAt := nextStartTime(st.DefaultStartTime)
	res, err := a.DB.Exec(`INSERT INTO reservations(tenant_id,station_id,gun_id,tenant_user_id,start_at,status,created_at)
		VALUES(?,?,?,?,?,?,?)`, s.TenantID, st.ID, g.ID, tuID, startAt.Format(time.RFC3339), "pending", now())
	if err != nil {
		writeErr(w, err)
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "startAt": startAt.Format(time.RFC3339), "status": "pending"})
	a.enqueueNotify(s.TenantID, a.userIDOfTenantUser(tuID), "reserved", "预约成功",
		"已预约场站「"+st.Name+"」，预计启动时间 "+startAt.Format("2006-01-02 15:04")+"。用户不可修改启动时间，当天过点自动顺延次日。")
}

func (a *App) userIDOfTenantUser(tuID int64) int64 {
	var uid int64
	_ = a.DB.QueryRow(`SELECT user_id FROM tenant_users WHERE id=?`, tuID).Scan(&uid)
	return uid
}

func (a *App) listReservations(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	userID, _ := strconv.ParseInt(r.URL.Query().Get("userId"), 10, 64)
	query := `SELECT rs.id,rs.gun_id,rs.start_at,rs.status,st.name,u.phone,u.name,rs.created_at
		FROM reservations rs JOIN stations st ON st.id=rs.station_id
		JOIN tenant_users tu ON tu.id=rs.tenant_user_id JOIN users u ON u.id=tu.user_id WHERE rs.tenant_id=?`
	args := []any{s.TenantID}
	if s.Role == "user" {
		query += ` AND rs.tenant_user_id=?`
		args = append(args, s.TenantUserID)
	} else if userID > 0 {
		query += ` AND u.id=?`
		args = append(args, userID)
	}
	query += ` ORDER BY rs.id DESC LIMIT 300`
	rows, err := a.DB.Query(query, args...)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, gunID int64
		var startAt, status, stationName, phone, name, created string
		if err = rows.Scan(&id, &gunID, &startAt, &status, &stationName, &phone, &name, &created); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "gunId": gunID, "startAt": startAt, "status": status,
			"stationName": stationName, "phone": phone, "name": name, "createdAt": created})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) cancelReservation(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	s := current(r)
	var tuID, uid int64
	var stName, startAt string
	err = a.DB.QueryRow(`SELECT rs.tenant_user_id,st.name,u.id,rs.start_at FROM reservations rs
		JOIN stations st ON st.id=rs.station_id JOIN tenant_users tu ON tu.id=rs.tenant_user_id JOIN users u ON u.id=tu.user_id
		WHERE rs.id=? AND rs.tenant_id=?`, id, s.TenantID).Scan(&tuID, &stName, &uid, &startAt)
	if err != nil {
		writeErr(w, fail(404, "预约不存在"))
		return
	}
	statusText := ""
	_ = a.DB.QueryRow(`SELECT status FROM reservations WHERE id=?`, id).Scan(&statusText)
	if s.Role == "user" && tuID != s.TenantUserID {
		writeErr(w, fail(403, "不能取消他人的预约"))
		return
	}
	if statusText != "pending" {
		writeErr(w, fail(409, "预约已启动或已关闭，不能取消"))
		return
	}
	_, err = a.DB.Exec(`UPDATE reservations SET status='cancelled',closed_at=? WHERE id=?`, now(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, true)
	a.enqueueNotify(s.TenantID, uid, "reservation_cancelled", "预约已取消", "场站「"+stName+"」的预约已取消，原启动时间 "+startAt+"。")
}

func (a *App) startCharge(w http.ResponseWriter, r *http.Request) {
	var req chargeReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	s := current(r)
	tuID, err := a.targetTenantUser(r, req.UserID)
	if err != nil {
		writeErr(w, err)
		return
	}
	g, err := a.getGun(req.GunID, s.TenantID)
	if err != nil {
		writeErr(w, fail(404, "充电枪不存在"))
		return
	}
	st, err := a.getStation(g.StationID, s.TenantID)
	if err != nil {
		writeErr(w, fail(404, "场站不存在"))
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.gunBusy(g.ID, 0) {
		writeErr(w, fail(409, "该充电枪已被预约或正在充电"))
		return
	}
	_, balance, err := a.walletID(s.TenantID, tuID)
	if err != nil {
		writeErr(w, fail(404, "钱包不存在"))
		return
	}
	if balance < st.StartThresholdCents {
		writeErr(w, fail(409, "余额需不少于 "+moneyText(st.StartThresholdCents)+" 才能启动"))
		return
	}
	orderID, err := a.insertOrder(s.TenantID, st.ID, g.ID, tuID, 0)
	if err != nil {
		writeErr(w, err)
		return
	}
	_, _ = a.DB.Exec(`UPDATE guns SET status='charging' WHERE id=?`, g.ID)
	if a.Sender != nil {
		a.Sender.SendCommand(g.DeviceNo, g.GunNo, "START", strconv.FormatInt(orderID, 10))
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": orderID, "orderId": orderID, "status": "charging"})
	a.enqueueNotify(s.TenantID, a.userIDOfTenantUser(tuID), "charge_started", "充电已启动", "场站「"+st.Name+"」充电已启动，订单号 "+strconv.FormatInt(orderID, 10)+"。")
}

func (a *App) stopCharge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GunID int64 `json:"gunId"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	s := current(r)
	g, err := a.getGun(req.GunID, s.TenantID)
	if err != nil {
		writeErr(w, fail(404, "充电枪不存在"))
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var orderID, tuID int64
	err = a.DB.QueryRow(`SELECT id,tenant_user_id FROM orders WHERE gun_id=? AND status='charging' ORDER BY id DESC LIMIT 1`, g.ID).Scan(&orderID, &tuID)
	if err != nil {
		writeErr(w, fail(404, "没有正在充电的订单"))
		return
	}
	if s.Role == "user" && tuID != s.TenantUserID {
		writeErr(w, fail(403, "不能停止他人的充电"))
		return
	}
	if a.Sender != nil {
		a.Sender.SendCommand(g.DeviceNo, g.GunNo, "STOP", strconv.FormatInt(orderID, 10))
	}
	_, _ = a.DB.Exec(`UPDATE orders SET status='pending_settle',end_at=? WHERE id=?`, now(), orderID)
	_, _ = a.DB.Exec(`UPDATE guns SET status='idle' WHERE id=?`, g.ID)
	_, _ = a.DB.Exec(`UPDATE reservations SET status='completed',closed_at=? WHERE gun_id=? AND status='active'`, now(), g.ID)
	writeJSON(w, http.StatusOK, true)
}

func (a *App) insertOrder(tenantID, stationID, gunID, tuID, reservationID int64) (int64, error) {
	res, err := a.DB.Exec(`INSERT INTO orders(tenant_id,station_id,gun_id,tenant_user_id,reservation_id,status,start_at,created_at)
		VALUES(?,?,?,?,?,?,?,?)`, tenantID, stationID, gunID, tuID, reservationID, "charging", now(), now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *App) listOrders(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	userID, _ := strconv.ParseInt(r.URL.Query().Get("userId"), 10, 64)
	query := `SELECT o.id,o.status,o.start_at,o.end_at,o.settle_at,o.energy_wh,o.amount_cents,st.name,u.phone,u.name
		FROM orders o JOIN stations st ON st.id=o.station_id JOIN tenant_users tu ON tu.id=o.tenant_user_id JOIN users u ON u.id=tu.user_id
		WHERE o.tenant_id=?`
	args := []any{s.TenantID}
	if s.Role == "user" {
		query += ` AND o.tenant_user_id=?`
		args = append(args, s.TenantUserID)
	} else if userID > 0 {
		query += ` AND u.id=?`
		args = append(args, userID)
	}
	query += ` ORDER BY o.id DESC LIMIT 500`
	rows, err := a.DB.Query(query, args...)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id int64
		var status, startAt, endAt, settleAt, stationName, phone, name string
		var energyWh, amount int64
		if err = rows.Scan(&id, &status, &startAt, &endAt, &settleAt, &energyWh, &amount, &stationName, &phone, &name); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "status": status, "startAt": startAt, "endAt": endAt,
			"settleAt": settleAt, "energyWh": energyWh, "amountCents": amount, "stationName": stationName,
			"phone": phone, "name": name})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) manualSettle(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	s := current(r)
	var status string
	if err = a.DB.QueryRow(`SELECT status FROM orders WHERE id=? AND tenant_id=?`, id, s.TenantID).Scan(&status); err != nil {
		writeErr(w, fail(404, "订单不存在"))
		return
	}
	if status != "pending_settle" {
		writeErr(w, fail(409, "订单不在待结算状态"))
		return
	}
	a.mu.Lock()
	err = a.settleOrder(id)
	a.mu.Unlock()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, true)
}

func (a *App) settleOrder(orderID int64) error {
	var tenantID, tuID, stationID, energyWh, price int64
	err := a.DB.QueryRow(`SELECT o.tenant_id,o.tenant_user_id,o.station_id,o.energy_wh,s.price_cents_per_kwh
		FROM orders o JOIN stations s ON s.id=o.station_id WHERE o.id=?`, orderID).
		Scan(&tenantID, &tuID, &stationID, &energyWh, &price)
	if err != nil {
		return err
	}
	amount := int64(float64(energyWh) / 1000.0 * float64(price))
	wid, balance, err := a.walletID(tenantID, tuID)
	if err != nil {
		return err
	}
	after := balance - amount
	bizNo := "CS" + strconv.FormatInt(orderID, 10)
	var exists int64
	if err = a.DB.QueryRow(`SELECT id FROM wallet_ledgers WHERE biz_no=?`, bizNo).Scan(&exists); err == nil {
		_, _ = a.DB.Exec(`UPDATE orders SET status='settled',amount_cents=?,settle_at=? WHERE id=?`, amount, now(), orderID)
		return nil
	}
	_, err = a.DB.Exec(`INSERT INTO wallet_ledgers(tenant_id,wallet_id,tenant_user_id,kind,amount_cents,balance_after,method,voucher_no,remark,operator_id,biz_no,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, tenantID, wid, tuID, "consume", -amount, after, "system", "", "充电结算", 0, bizNo, now())
	if err != nil {
		return err
	}
	_, err = a.DB.Exec(`UPDATE wallets SET balance_cents=?,updated_at=? WHERE id=?`, after, now(), wid)
	if err != nil {
		return err
	}
	_, err = a.DB.Exec(`UPDATE orders SET status='settled',amount_cents=?,settle_at=? WHERE id=?`, amount, now(), orderID)
	if err == nil {
		a.enqueueNotify(tenantID, a.userIDOfTenantUser(tuID), "settled", "充电已结算",
			"订单 "+strconv.FormatInt(orderID, 10)+" 已结算，用电 " + strconv.FormatFloat(float64(energyWh)/1000, 'f', 2, 64)+
				" 度，金额 "+moneyText(amount)+"，余额 "+moneyText(after)+"。")
	}
	return err
}

func (a *App) RunScheduler(stop <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			a.schedulerTick()
		}
	}
}

func (a *App) schedulerTick() {
	a.mu.Lock()
	defer a.mu.Unlock()

	rows, err := a.DB.Query(`SELECT id,tenant_id,station_id,gun_id,tenant_user_id FROM reservations WHERE status='pending' AND start_at<=? LIMIT 50`, now())
	if err != nil {
		return
	}
	type due struct{ id, tid, sid, gid, tu int64 }
	var dues []due
	for rows.Next() {
		var d due
		if rows.Scan(&d.id, &d.tid, &d.sid, &d.gid, &d.tu) == nil {
			dues = append(dues, d)
		}
	}
	rows.Close()
	for _, d := range dues {
		st, err := a.getStation(d.sid, d.tid)
		if err != nil || a.gunBusy(d.gid, d.id) {
			a.closeReservation(d.id, "failed")
			continue
		}
		_, balance, err := a.walletID(d.tid, d.tu)
		if err != nil || balance < st.StartThresholdCents {
			a.closeReservation(d.id, "failed")
			a.enqueueNotify(d.tid, a.userIDOfTenantUser(d.tu), "reservation_failed", "预约启动失败",
				"余额不足或钱包异常，预约已取消，请充值后重新预约。")
			continue
		}
		var gun gunInfo
		if err = a.DB.QueryRow(`SELECT g.id,g.tenant_id,g.gun_no,g.status,p.device_no,p.station_id FROM guns g
			JOIN piles p ON p.id=g.pile_id WHERE g.id=?`, d.gid).Scan(&gun.ID, &gun.TenantID, &gun.GunNo, &gun.Status, &gun.DeviceNo, &gun.StationID); err != nil {
			a.closeReservation(d.id, "failed")
			continue
		}
		orderID, err := a.insertOrder(d.tid, d.sid, d.gid, d.tu, d.id)
		if err != nil {
			continue
		}
		_, _ = a.DB.Exec(`UPDATE reservations SET status='active',started_at=? WHERE id=?`, now(), d.id)
		_, _ = a.DB.Exec(`UPDATE guns SET status='charging' WHERE id=?`, d.gid)
		if a.Sender != nil {
			a.Sender.SendCommand(gun.DeviceNo, gun.GunNo, "START", strconv.FormatInt(orderID, 10))
		}
		a.enqueueNotify(d.tid, a.userIDOfTenantUser(d.tu), "charge_started", "预约充电已启动",
			"场站「"+st.Name+"」预约充电已启动，订单号 "+strconv.FormatInt(orderID, 10)+"。")
	}

	rows, err = a.DB.Query(`SELECT o.id,o.tenant_id,o.gun_id,o.tenant_user_id,s.stop_threshold_cents,s.warn_threshold_cents
		FROM orders o JOIN stations s ON s.id=o.station_id WHERE o.status='charging'`)
	if err != nil {
		return
	}
	type live struct {
		id, tid, gid, tu, stop, warn int64
	}
	var lives []live
	for rows.Next() {
		var v live
		if rows.Scan(&v.id, &v.tid, &v.gid, &v.tu, &v.stop, &v.warn) == nil {
			lives = append(lives, v)
		}
	}
	rows.Close()
	for _, v := range lives {
		_, balance, err := a.walletID(v.tid, v.tu)
		if err != nil {
			continue
		}
		if balance < v.stop {
			a.stopOrderByScheduler(v.id, v.gid, "余额低于断电阈值，已自动停止")
			continue
		}
		if balance < v.warn {
			if last, ok := a.warned[v.id]; !ok || time.Since(last) > time.Hour {
				a.warned[v.id] = time.Now()
				a.enqueueNotify(v.tid, a.userIDOfTenantUser(v.tu), "low_balance", "余额预警",
					"当前余额 "+moneyText(balance)+"，低于预警值，请及时充值，否则将自动断电。")
			}
		}
	}

	rows, err = a.DB.Query(`SELECT o.id,s.settlement_timeout_hours,o.end_at FROM orders o JOIN stations s ON s.id=o.station_id
		WHERE o.status='pending_settle' AND o.end_at<>''`)
	if err != nil {
		return
	}
	type pending struct {
		id     int64
		hours  int64
		endAt  string
	}
	var pendings []pending
	for rows.Next() {
		var p pending
		if rows.Scan(&p.id, &p.hours, &p.endAt) == nil {
			pendings = append(pendings, p)
		}
	}
	rows.Close()
	for _, p := range pendings {
		end, err := time.Parse(time.RFC3339, p.endAt)
		if err != nil {
			continue
		}
		if time.Since(end) >= time.Duration(p.hours)*time.Hour {
			_ = a.settleOrder(p.id)
		}
	}
}

func (a *App) closeReservation(id int64, status string) {
	_, _ = a.DB.Exec(`UPDATE reservations SET status=?,closed_at=? WHERE id=?`, status, now(), id)
}

func (a *App) stopOrderByScheduler(orderID, gunID int64, reason string) {
	var tuID, tid int64
	_ = a.DB.QueryRow(`SELECT tenant_user_id,tenant_id FROM orders WHERE id=?`, orderID).Scan(&tuID, &tid)
	var deviceNo string
	var gunNo int
	_ = a.DB.QueryRow(`SELECT p.device_no,g.gun_no FROM guns g JOIN piles p ON p.id=g.pile_id WHERE g.id=?`, gunID).Scan(&deviceNo, &gunNo)
	if a.Sender != nil {
		a.Sender.SendCommand(deviceNo, gunNo, "STOP", strconv.FormatInt(orderID, 10))
	}
	_, _ = a.DB.Exec(`UPDATE orders SET status='pending_settle',end_at=? WHERE id=?`, now(), orderID)
	_, _ = a.DB.Exec(`UPDATE guns SET status='idle' WHERE id=?`, gunID)
	_, _ = a.DB.Exec(`UPDATE reservations SET status='completed',closed_at=? WHERE gun_id=? AND status='active'`, now(), gunID)
	a.enqueueNotify(tid, a.userIDOfTenantUser(tuID), "auto_stopped", "充电已自动停止", reason+"，订单 "+strconv.FormatInt(orderID, 10)+" 待结算。")
}

func moneyText(cents int64) string {
	return strconv.FormatFloat(float64(cents)/100.0, 'f', 2, 64) + " 元"
}
