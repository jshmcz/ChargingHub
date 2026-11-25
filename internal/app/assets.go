package app

import (
	"net/http"
	"strconv"
	"strings"
)

type Station struct {
	ID                    int64
	TenantID              int64
	Name                  string
	Address               string
	DefaultStartTime      string
	StartThresholdCents   int64
	StopThresholdCents    int64
	WarnThresholdCents    int64
	SettlementTimeoutHrs  int64
	PriceCentsPerKWh      int64
	Status                string
}

type stationReq struct {
	Name                  string `json:"name"`
	Address               string `json:"address"`
	DefaultStartTime      string `json:"defaultStartTime"`
	StartThresholdCents   int64  `json:"startThresholdCents"`
	StopThresholdCents    int64  `json:"stopThresholdCents"`
	WarnThresholdCents    int64  `json:"warnThresholdCents"`
	SettlementTimeoutHrs  int64  `json:"settlementTimeoutHours"`
	PriceCentsPerKWh      int64  `json:"priceCentsPerKwh"`
	Status                string `json:"status"`
}

func normalizeStation(req *stationReq) {
	req.Name = strings.TrimSpace(req.Name)
	if req.DefaultStartTime == "" {
		req.DefaultStartTime = "22:00"
	}
	if req.StartThresholdCents <= 0 {
		req.StartThresholdCents = 5000
	}
	if req.StopThresholdCents <= 0 {
		req.StopThresholdCents = 2000
	}
	if req.WarnThresholdCents <= 0 {
		req.WarnThresholdCents = 3000
	}
	if req.SettlementTimeoutHrs <= 0 {
		req.SettlementTimeoutHrs = 24
	}
	if req.PriceCentsPerKWh <= 0 {
		req.PriceCentsPerKWh = 150
	}
	if req.Status == "" {
		req.Status = "active"
	}
}

func (a *App) getStation(id, tenantID int64) (Station, error) {
	var st Station
	err := a.DB.QueryRow(`SELECT id,tenant_id,name,address,default_start_time,start_threshold_cents,stop_threshold_cents,
		warn_threshold_cents,settlement_timeout_hours,price_cents_per_kwh,status FROM stations WHERE id=? AND tenant_id=?`, id, tenantID).
		Scan(&st.ID, &st.TenantID, &st.Name, &st.Address, &st.DefaultStartTime, &st.StartThresholdCents,
			&st.StopThresholdCents, &st.WarnThresholdCents, &st.SettlementTimeoutHrs, &st.PriceCentsPerKWh, &st.Status)
	return st, err
}

func (a *App) listStations(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	query := `SELECT id,tenant_id,name,address,default_start_time,start_threshold_cents,stop_threshold_cents,
		warn_threshold_cents,settlement_timeout_hours,price_cents_per_kwh,status FROM stations`
	args := []any{}
	if s.Role != "super" {
		query += ` WHERE tenant_id=?`
		args = append(args, s.TenantID)
	}
	query += ` ORDER BY id DESC`
	rows, err := a.DB.Query(query, args...)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var st Station
		if err = rows.Scan(&st.ID, &st.TenantID, &st.Name, &st.Address, &st.DefaultStartTime, &st.StartThresholdCents,
			&st.StopThresholdCents, &st.WarnThresholdCents, &st.SettlementTimeoutHrs, &st.PriceCentsPerKWh, &st.Status); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, stationMap(st))
	}
	writeJSON(w, http.StatusOK, out)
}

func stationMap(st Station) map[string]any {
	return map[string]any{
		"id": st.ID, "tenantId": st.TenantID, "name": st.Name, "address": st.Address,
		"defaultStartTime": st.DefaultStartTime, "startThresholdCents": st.StartThresholdCents,
		"stopThresholdCents": st.StopThresholdCents, "warnThresholdCents": st.WarnThresholdCents,
		"settlementTimeoutHours": st.SettlementTimeoutHrs, "priceCentsPerKwh": st.PriceCentsPerKWh,
		"status": st.Status,
	}
}

func (a *App) createStation(w http.ResponseWriter, r *http.Request) {
	var req stationReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	normalizeStation(&req)
	if req.Name == "" {
		writeErr(w, fail(400, "场站名称不能为空"))
		return
	}
	res, err := a.DB.Exec(`INSERT INTO stations(tenant_id,name,address,default_start_time,start_threshold_cents,
		stop_threshold_cents,warn_threshold_cents,settlement_timeout_hours,price_cents_per_kwh,status,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`, current(r).TenantID, req.Name, req.Address, req.DefaultStartTime,
		req.StartThresholdCents, req.StopThresholdCents, req.WarnThresholdCents, req.SettlementTimeoutHrs,
		req.PriceCentsPerKWh, req.Status, now())
	if err != nil {
		writeErr(w, err)
		return
	}
	id, _ := res.LastInsertId()
	req.Name = ""
	writeJSON(w, http.StatusOK, map[string]any{"id": id})
}

func (a *App) updateStation(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	var req stationReq
	if err = decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	normalizeStation(&req)
	_, err = a.DB.Exec(`UPDATE stations SET name=?,address=?,default_start_time=?,start_threshold_cents=?,
		stop_threshold_cents=?,warn_threshold_cents=?,settlement_timeout_hours=?,price_cents_per_kwh=?,status=?
		WHERE id=? AND tenant_id=?`, req.Name, req.Address, req.DefaultStartTime, req.StartThresholdCents,
		req.StopThresholdCents, req.WarnThresholdCents, req.SettlementTimeoutHrs, req.PriceCentsPerKWh,
		req.Status, id, current(r).TenantID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, true)
}

type pileReq struct {
	StationID int64  `json:"stationId"`
	DeviceNo  string `json:"deviceNo"`
	GunCount  int    `json:"gunCount"`
}

func (a *App) listPiles(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	rows, err := a.DB.Query(`SELECT p.id,p.station_id,p.device_no,p.status,p.last_seen_at,COALESCE(s.name,''),
		(SELECT COUNT(*) FROM guns g WHERE g.pile_id=p.id) FROM piles p LEFT JOIN stations s ON s.id=p.station_id
		WHERE p.tenant_id=? ORDER BY p.id DESC`, s.TenantID)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, stationID int64
		var deviceNo, status, lastSeen, stationName string
		var gunCount int
		if err = rows.Scan(&id, &stationID, &deviceNo, &status, &lastSeen, &stationName, &gunCount); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "stationId": stationID, "deviceNo": deviceNo, "status": status,
			"lastSeenAt": lastSeen, "stationName": stationName, "gunCount": gunCount})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) createPile(w http.ResponseWriter, r *http.Request) {
	var req pileReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if req.StationID <= 0 {
		writeErr(w, fail(400, "请选择场站"))
		return
	}
	if _, err := a.getStation(req.StationID, current(r).TenantID); err != nil {
		writeErr(w, fail(404, "场站不存在"))
		return
	}
	if req.GunCount <= 0 {
		req.GunCount = 1
	}
	if req.GunCount > 32 {
		req.GunCount = 32
	}
	req.DeviceNo = strings.ToUpper(strings.TrimSpace(req.DeviceNo))
	if req.DeviceNo == "" {
		req.DeviceNo = "CH" + strings.ToUpper(randomHex(6))
	}
	secret := randomHex(16)
	tenantID := current(r).TenantID
	tx, err := a.DB.Begin()
	if err != nil {
		writeErr(w, err)
		return
	}
	res, err := tx.Exec(`INSERT INTO piles(tenant_id,station_id,device_no,secret,status,created_at) VALUES(?,?,?,?,?,?)`,
		tenantID, req.StationID, req.DeviceNo, secret, "pending", now())
	if err != nil {
		tx.Rollback()
		writeErr(w, err)
		return
	}
	pileID, _ := res.LastInsertId()
	for i := 1; i <= req.GunCount; i++ {
		if _, err = tx.Exec(`INSERT INTO guns(tenant_id,pile_id,gun_no,status,created_at) VALUES(?,?,?,?,?)`,
			tenantID, pileID, i, "idle", now()); err != nil {
			tx.Rollback()
			writeErr(w, err)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": pileID, "deviceNo": req.DeviceNo, "secret": secret, "gunCount": req.GunCount})
}

func (a *App) updatePile(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err = decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if _, err = a.DB.Exec(`UPDATE piles SET status=? WHERE id=? AND tenant_id=?`, req.Status, id, current(r).TenantID); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, true)
}

func (a *App) regenerateSecret(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	secret := randomHex(16)
	res, err := a.DB.Exec(`UPDATE piles SET secret=? WHERE id=? AND tenant_id=?`, secret, id, current(r).TenantID)
	if err != nil {
		writeErr(w, err)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeErr(w, fail(404, "桩不存在"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"secret": secret})
}

func (a *App) listGuns(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	stationID, _ := strconv.ParseInt(r.URL.Query().Get("stationId"), 10, 64)
	query := `SELECT g.id,g.pile_id,g.gun_no,g.status,p.device_no,COALESCE(st.name,''),p.station_id
		FROM guns g JOIN piles p ON p.id=g.pile_id LEFT JOIN stations st ON st.id=p.station_id WHERE g.tenant_id=?`
	args := []any{s.TenantID}
	if stationID > 0 {
		query += ` AND p.station_id=?`
		args = append(args, stationID)
	}
	query += ` ORDER BY p.id,g.gun_no`
	rows, err := a.DB.Query(query, args...)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, pileID, station int64
		var status, deviceNo, stationName string
		var gunNo int
		if err = rows.Scan(&id, &pileID, &gunNo, &status, &deviceNo, &stationName, &station); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "pileId": pileID, "gunNo": gunNo, "status": status,
			"deviceNo": deviceNo, "stationName": stationName, "stationId": station})
	}
	writeJSON(w, http.StatusOK, out)
}

type gunInfo struct {
	ID        int64
	TenantID  int64
	GunNo     int
	Status    string
	DeviceNo  string
	StationID int64
}

func (a *App) getGun(id, tenantID int64) (gunInfo, error) {
	var g gunInfo
	err := a.DB.QueryRow(`SELECT g.id,g.tenant_id,g.gun_no,g.status,p.device_no,p.station_id FROM guns g
		JOIN piles p ON p.id=g.pile_id WHERE g.id=? AND g.tenant_id=?`, id, tenantID).
		Scan(&g.ID, &g.TenantID, &g.GunNo, &g.Status, &g.DeviceNo, &g.StationID)
	return g, err
}
