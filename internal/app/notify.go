package app

import (
	"crypto/tls"
	"mime"
	"net/smtp"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type notifyJob struct {
	ID       int64
	TenantID int64
	UserID   int64
	Event    string
	Subject  string
	Body     string
}

type smtpReq struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	From       string `json:"from"`
	Encryption string `json:"encryption"`
	Enabled    bool   `json:"enabled"`
}

func (a *App) getSMTP(w http.ResponseWriter, r *http.Request) {
	var host, username, passwordEnc, from, encryption string
	var port int
	var enabled int
	err := a.DB.QueryRow(`SELECT host,port,username,password_enc,from_addr,encryption,enabled FROM smtp_configs WHERE tenant_id=?`,
		current(r).TenantID).Scan(&host, &port, &username, &passwordEnc, &from, &encryption, &enabled)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"enabled": false, "port": 587, "encryption": "starttls"})
		return
	}
	masked := ""
	if passwordEnc != "" {
		masked = "******"
	}
	writeJSON(w, http.StatusOK, map[string]any{"host": host, "port": port, "username": username,
		"password": masked, "from": from, "encryption": encryption, "enabled": enabled == 1})
}

func (a *App) putSMTP(w http.ResponseWriter, r *http.Request) {
	var req smtpReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if req.Port <= 0 {
		req.Port = 587
	}
	if req.Encryption == "" {
		req.Encryption = "starttls"
	}
	if req.Encryption != "none" && req.Encryption != "ssl" && req.Encryption != "starttls" {
		writeErr(w, fail(400, "encryption 只能是 none / ssl / starttls"))
		return
	}
	enc := 0
	if req.Enabled {
		enc = 1
	}
	password := strings.TrimSpace(req.Password)
	if password == "******" || password == "" {
		_ = a.DB.QueryRow(`SELECT password_enc FROM smtp_configs WHERE tenant_id=?`, current(r).TenantID).Scan(&password)
	} else {
		password = a.encrypt(password)
	}
	_, err := a.DB.Exec(`INSERT INTO smtp_configs(tenant_id,host,port,username,password_enc,from_addr,encryption,enabled,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(tenant_id) DO UPDATE SET host=excluded.host,port=excluded.port,
		username=excluded.username,password_enc=excluded.password_enc,from_addr=excluded.from_addr,
		encryption=excluded.encryption,enabled=excluded.enabled,updated_at=excluded.updated_at`,
		current(r).TenantID, req.Host, req.Port, req.Username, password, req.From, req.Encryption, enc, now())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, true)
}

func (a *App) listNotifications(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	userID, _ := strconv.ParseInt(r.URL.Query().Get("userId"), 10, 64)
	query := `SELECT n.id,n.event,n.subject,n.body,n.status,n.error,n.created_at,n.sent_at,u.phone
		FROM notifications n LEFT JOIN users u ON u.id=n.user_id WHERE 1=1`
	args := []any{}
	if s.Role != "super" {
		query += ` AND n.tenant_id=?`
		args = append(args, s.TenantID)
	}
	if s.Role == "user" {
		query += ` AND n.user_id=?`
		args = append(args, s.UserID)
	} else if userID > 0 {
		query += ` AND n.user_id=?`
		args = append(args, userID)
	}
	query += ` ORDER BY n.id DESC LIMIT 300`
	rows, err := a.DB.Query(query, args...)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id int64
		var event, subject, body, status, errText, created, sent, phone string
		if err = rows.Scan(&id, &event, &subject, &body, &status, &errText, &created, &sent, &phone); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "event": event, "subject": subject, "body": body,
			"status": status, "error": errText, "createdAt": created, "sentAt": sent, "phone": phone})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) enqueueNotify(tenantID, userID int64, event, subject, body string) {
	if a.notifyCh == nil {
		return
	}
	res, err := a.DB.Exec(`INSERT INTO notifications(tenant_id,user_id,event,subject,body,status,created_at) VALUES(?,?,?,?,?,?,?)`,
		tenantID, userID, event, subject, body, "queued", now())
	if err != nil {
		return
	}
	id, _ := res.LastInsertId()
	job := notifyJob{ID: id, TenantID: tenantID, UserID: userID, Event: event, Subject: subject, Body: body}
	select {
	case a.notifyCh <- job:
	default:
		_, _ = a.DB.Exec(`UPDATE notifications SET status='dropped',error='队列已满' WHERE id=?`, id)
	}
}

func (a *App) RunNotifier(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case job := <-a.notifyCh:
			a.sendNotify(job)
		}
	}
}

func (a *App) sendNotify(job notifyJob) {
	var host, username, passwordEnc, from, encryption string
	var port, enabled int
	err := a.DB.QueryRow(`SELECT host,port,username,password_enc,from_addr,encryption,enabled FROM smtp_configs WHERE tenant_id=?`,
		job.TenantID).Scan(&host, &port, &username, &passwordEnc, &from, &encryption, &enabled)
	if err != nil || enabled == 0 {
		a.markNotify(job.ID, "disabled", "SMTP 未启用")
		return
	}
	var email string
	if err = a.DB.QueryRow(`SELECT email FROM users WHERE id=?`, job.UserID).Scan(&email); err != nil || email == "" {
		a.markNotify(job.ID, "no_email", "用户未配置邮箱")
		return
	}
	password := a.decrypt(passwordEnc)
	addr := host + ":" + strconv.Itoa(port)
	var client *smtp.Client
	switch encryption {
	case "ssl":
		conn, terr := tls.Dial("tcp", addr, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
		if terr != nil {
			a.markNotify(job.ID, "failed", terr.Error())
			return
		}
		client, err = smtp.NewClient(conn, host)
	case "starttls":
		client, err = smtp.Dial(addr)
		if err == nil {
			err = client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
		}
	default:
		client, err = smtp.Dial(addr)
	}
	if err != nil {
		a.markNotify(job.ID, "failed", err.Error())
		return
	}
	defer client.Close()
	if username != "" {
		if err = client.Auth(smtp.PlainAuth("", username, password, host)); err != nil {
			a.markNotify(job.ID, "failed", err.Error())
			return
		}
	}
	if err = client.Mail(from); err != nil {
		a.markNotify(job.ID, "failed", err.Error())
		return
	}
	if err = client.Rcpt(email); err != nil {
		a.markNotify(job.ID, "failed", err.Error())
		return
	}
	wc, err := client.Data()
	if err != nil {
		a.markNotify(job.ID, "failed", err.Error())
		return
	}
	msg := "From: " + from + "\r\nTo: " + email + "\r\nSubject: " + mime.QEncoding.Encode("utf-8", job.Subject) +
		"\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + job.Body + "\r\n"
	if _, err = wc.Write([]byte(msg)); err != nil {
		_ = wc.Close()
		a.markNotify(job.ID, "failed", err.Error())
		return
	}
	if err = wc.Close(); err != nil {
		a.markNotify(job.ID, "failed", err.Error())
		return
	}
	_ = client.Quit()
	a.markNotify(job.ID, "sent", "")
}

func (a *App) markNotify(id int64, status, errText string) {
	_, _ = a.DB.Exec(`UPDATE notifications SET status=?,error=?,sent_at=? WHERE id=?`,
		status, errText, time.Now().Format(time.RFC3339), id)
}
