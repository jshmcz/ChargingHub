package app

import (
	"net/http"
	"strconv"

	qrcode "github.com/skip2/go-qrcode"
)

// gunByToken 供手机扫码后按令牌定位枪。需登录，自动锁定会话租户。
func (a *App) gunByToken(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	token := r.URL.Query().Get("token")
	if token == "" {
		writeErr(w, fail(400, "缺少枪令牌"))
		return
	}
	var gunID, pileID, stationID, gunNo int64
	var gunCode, status, deviceNo, stationName, stationAddr string
	err := a.DB.QueryRow(`SELECT g.id,g.pile_id,g.gun_no,g.gun_code,g.status,p.device_no,p.station_id,
		COALESCE(st.name,''),COALESCE(st.address,'')
		FROM guns g JOIN piles p ON p.id=g.pile_id LEFT JOIN stations st ON st.id=p.station_id
		WHERE g.qr_token=? AND g.tenant_id=?`, token, s.TenantID).
		Scan(&gunID, &pileID, &gunNo, &gunCode, &status, &deviceNo, &stationID, &stationName, &stationAddr)
	if err != nil {
		writeErr(w, fail(404, "二维码无效或与当前租户不匹配"))
		return
	}
	// 当前用户在该租户的钱包余额（供扫码页展示）
	balance := int64(0)
	_, b, berr := a.walletID(s.TenantID, s.TenantUserID)
	if berr == nil {
		balance = b
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"gunId": gunID, "pileId": pileID, "gunNo": gunNo, "gunCode": gunCode,
		"status": status, "deviceNo": deviceNo, "stationId": stationID,
		"stationName": stationName, "stationAddress": stationAddr,
		"balanceCents": balance,
	})
}

// gunQrcode 生成单把枪的二维码 PNG。运营端（租户管理员/超管）下载。
func (a *App) gunQrcode(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	s := current(r)
	var token string
	query := `SELECT qr_token FROM guns WHERE id=?`
	args := []any{id}
	if s.Role != "super" {
		query += ` AND tenant_id=?`
		args = append(args, s.TenantID)
	}
	if err = a.DB.QueryRow(query, args...).Scan(&token); err != nil || token == "" {
		writeErr(w, fail(404, "枪不存在或未生成二维码令牌"))
		return
	}
	base := a.getConfig(configBaseURL)
	if base == "" {
		writeErr(w, fail(409, "请先用平台管理员账号配置对外域名"))
		return
	}
	content := base + "/m/#/charge?token=" + token
	png, err := qrcode.Encode(content, qrcode.Medium, 320)
	if err != nil {
		writeErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=\"gun-"+strconv.FormatInt(id, 10)+".png\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}
