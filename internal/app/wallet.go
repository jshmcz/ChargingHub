package app

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
)

type walletTxReq struct {
	TenantID   int64  `json:"tenantId"`
	UserID     int64  `json:"userId"`
	Amount     int64  `json:"amountCents"`
	Method     string `json:"method"`
	VoucherNo  string `json:"voucherNo"`
	Remark     string `json:"remark"`
	BizNo      string `json:"bizNo"`
}

func (a *App) tenantUserID(tenantID, userID int64) (int64, error) {
	var tuID int64
	err := a.DB.QueryRow(`SELECT id FROM tenant_users WHERE tenant_id=? AND user_id=?`, tenantID, userID).Scan(&tuID)
	return tuID, err
}

func (a *App) walletID(tenantID, tuID int64) (int64, int64, error) {
	var id, balance int64
	err := a.DB.QueryRow(`SELECT id,balance_cents FROM wallets WHERE tenant_id=? AND tenant_user_id=?`, tenantID, tuID).Scan(&id, &balance)
	return id, balance, err
}

func (a *App) listWallets(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	userID, _ := strconv.ParseInt(r.URL.Query().Get("userId"), 10, 64)
	query := `SELECT w.id,w.tenant_id,w.tenant_user_id,w.balance_cents,w.status,u.id,u.phone,u.name
		FROM wallets w JOIN tenant_users tu ON tu.id=w.tenant_user_id JOIN users u ON u.id=tu.user_id WHERE 1=1`
	args := []any{}
	if s.Role != "super" {
		query += ` AND w.tenant_id=?`
		args = append(args, s.TenantID)
	}
	if userID > 0 {
		query += ` AND u.id=?`
		args = append(args, userID)
	}
	query += ` ORDER BY w.id DESC`
	rows, err := a.DB.Query(query, args...)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, tenantID, tuID, balance, uid int64
		var status, phone, name string
		if err = rows.Scan(&id, &tenantID, &tuID, &balance, &status, &uid, &phone, &name); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "tenantId": tenantID, "tenantUserId": tuID, "userId": uid,
			"balanceCents": balance, "status": status, "phone": phone, "name": name})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) listLedgers(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	userID, _ := strconv.ParseInt(r.URL.Query().Get("userId"), 10, 64)
	query := `SELECT l.id,l.kind,l.amount_cents,l.balance_after,l.method,l.voucher_no,l.remark,l.biz_no,l.created_at,u.phone
		FROM wallet_ledgers l JOIN tenant_users tu ON tu.id=l.tenant_user_id JOIN users u ON u.id=tu.user_id WHERE 1=1`
	args := []any{}
	if s.Role != "super" {
		query += ` AND l.tenant_id=?`
		args = append(args, s.TenantID)
	}
	if userID > 0 {
		query += ` AND u.id=?`
		args = append(args, userID)
	}
	query += ` ORDER BY l.id DESC LIMIT 500`
	rows, err := a.DB.Query(query, args...)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id int64
		var kind, method, voucher, remark, bizNo, created, phone string
		var amount, after int64
		if err = rows.Scan(&id, &kind, &amount, &after, &method, &voucher, &remark, &bizNo, &created, &phone); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "kind": kind, "amountCents": amount, "balanceAfter": after,
			"method": method, "voucherNo": voucher, "remark": remark, "bizNo": bizNo, "createdAt": created, "phone": phone})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) myWallet(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	id, balance, err := a.walletID(s.TenantID, s.TenantUserID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "balanceCents": balance})
}

func (a *App) myLedgers(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	rows, err := a.DB.Query(`SELECT kind,amount_cents,balance_after,method,voucher_no,remark,biz_no,created_at
		FROM wallet_ledgers WHERE tenant_id=? AND tenant_user_id=? ORDER BY id DESC LIMIT 200`, s.TenantID, s.TenantUserID)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var kind, method, voucher, remark, bizNo, created string
		var amount, after int64
		if err = rows.Scan(&kind, &amount, &after, &method, &voucher, &remark, &bizNo, &created); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, map[string]any{"kind": kind, "amountCents": amount, "balanceAfter": after,
			"method": method, "voucherNo": voucher, "remark": remark, "bizNo": bizNo, "createdAt": created})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) recharge(w http.ResponseWriter, r *http.Request) {
	a.walletMutation(w, r, "recharge", 1)
}

func (a *App) refund(w http.ResponseWriter, r *http.Request) {
	a.walletMutation(w, r, "refund", -1)
}

func (a *App) walletMutation(w http.ResponseWriter, r *http.Request, kind string, sign int64) {
	var req walletTxReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if req.Amount <= 0 {
		writeErr(w, fail(400, "金额必须大于 0"))
		return
	}
	s := current(r)
	tenantID, err := scopeTenant(r, req.TenantID)
	if err != nil {
		writeErr(w, err)
		return
	}
	tuID, err := a.tenantUserID(tenantID, req.UserID)
	if err != nil {
		writeErr(w, fail(404, "用户不属于当前租户"))
		return
	}
	amount := sign * req.Amount
	bizNo := strings.TrimSpace(req.BizNo)
	if bizNo == "" {
		prefix := "RC"
		if kind == "refund" {
			prefix = "RF"
		}
		bizNo = prefix + randomHex(12)
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	var existing int64
	if err = a.DB.QueryRow(`SELECT id FROM wallet_ledgers WHERE biz_no=?`, bizNo).Scan(&existing); err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"id": existing, "duplicate": true})
		return
	}
	wid, balance, err := a.walletID(tenantID, tuID)
	if err == sql.ErrNoRows && kind == "recharge" {
		// 充值时按当前租户自动创建钱包（建用户时不建钱包）。
		res, ierr := a.DB.Exec(`INSERT OR IGNORE INTO wallets(tenant_id,tenant_user_id,balance_cents,updated_at) VALUES(?,?,0,?)`, tenantID, tuID, now())
		if ierr != nil {
			writeErr(w, ierr)
			return
		}
		_ = res
		wid, balance, err = a.walletID(tenantID, tuID)
	}
	if err != nil {
		writeErr(w, fail(404, "钱包不存在"))
		return
	}
	if kind == "refund" {
		if balance < req.Amount {
			writeErr(w, fail(409, "退款金额超过当前余额"))
			return
		}
		var openOrders, openRes int
		if err = a.DB.QueryRow(`SELECT COUNT(*) FROM orders WHERE tenant_user_id=? AND status IN ('charging','pending_settle')`, tuID).Scan(&openOrders); err != nil {
			writeErr(w, err)
			return
		}
		if err = a.DB.QueryRow(`SELECT COUNT(*) FROM reservations WHERE tenant_user_id=? AND status IN ('pending','active')`, tuID).Scan(&openRes); err != nil {
			writeErr(w, err)
			return
		}
		if openOrders > 0 || openRes > 0 {
			writeErr(w, fail(409, "存在未关闭的订单或预约单，不能退款"))
			return
		}
	}
	after := balance + amount
	if after < 0 {
		writeErr(w, fail(409, "余额不足"))
		return
	}
	_, err = a.DB.Exec(`INSERT INTO wallet_ledgers(tenant_id,wallet_id,tenant_user_id,kind,amount_cents,balance_after,method,voucher_no,remark,operator_id,biz_no,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, tenantID, wid, tuID, kind, amount, after, req.Method, req.VoucherNo,
		req.Remark, s.UserID, bizNo, now())
	if err != nil {
		writeErr(w, err)
		return
	}
	if _, err = a.DB.Exec(`UPDATE wallets SET balance_cents=?,updated_at=? WHERE id=?`, after, now(), wid); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"balanceCents": after, "bizNo": bizNo})

	var subject string
	if kind == "recharge" {
		subject = "充值到账"
	} else {
		subject = "退款到账"
	}
	a.enqueueNotify(tenantID, req.UserID, kind, subject,
		subject+"，金额 "+moneyText(req.Amount)+"，当前余额 "+moneyText(after)+"。")
}
