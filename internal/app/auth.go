package app

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DeviceID string `json:"deviceId"`
	TenantID int64  `json:"tenantId"`
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeErr(w, fail(400, "用户名或密码不能为空"))
		return
	}
	if req.DeviceID == "" {
		req.DeviceID = "unknown"
	}

	var userID int64
	var hash, role, status, phone, email, name string
	err := a.DB.QueryRow(`SELECT id,password_hash,role,status,phone,email,name FROM users WHERE phone=? OR email=? LIMIT 1`,
		req.Username, req.Username).Scan(&userID, &hash, &role, &status, &phone, &email, &name)
	if err != nil {
		writeErr(w, fail(401, "账号或密码错误"))
		return
	}
	if status != "active" || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		writeErr(w, fail(401, "账号或密码错误"))
		return
	}

	tenantID, tenantUserID, effectiveRole := int64(0), int64(0), role
	if role != "super" {
		rows, qerr := a.DB.Query(`SELECT id,tenant_id,role FROM tenant_users WHERE user_id=? AND status='active'`, userID)
		if qerr != nil {
			writeErr(w, qerr)
			return
		}
		type link struct {
			id, tid int64
			role    string
		}
		var links []link
		for rows.Next() {
			var l link
			if err = rows.Scan(&l.id, &l.tid, &l.role); err != nil {
				break
			}
			links = append(links, l)
		}
		rows.Close()
		if err != nil {
			writeErr(w, err)
			return
		}
		if len(links) == 0 {
			writeErr(w, fail(403, "账号未关联任何租户"))
			return
		}
		var chosen *link
		if req.TenantID > 0 {
			for i := range links {
				if links[i].tid == req.TenantID {
					chosen = &links[i]
					break
				}
			}
			if chosen == nil {
				writeErr(w, fail(403, "账号不属于该租户"))
				return
			}
		} else if len(links) == 1 {
			chosen = &links[0]
		} else {
			writeErr(w, fail(400, "该账号属于多个租户，请指定 tenantId"))
			return
		}
		tenantID, tenantUserID, effectiveRole = chosen.tid, chosen.id, chosen.role
	}

	a.mu.Lock()
	// 同设备免踢：同设备重复登录复用原会话，其他设备全部踢掉。
	if _, err = a.DB.Exec(`DELETE FROM sessions WHERE user_id=? AND device_id<>?`, userID, req.DeviceID); err != nil {
		a.mu.Unlock()
		writeErr(w, err)
		return
	}
	var token string
	err = a.DB.QueryRow(`SELECT token FROM sessions WHERE user_id=? AND device_id=? LIMIT 1`, userID, req.DeviceID).Scan(&token)
	expires := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	if err == sql.ErrNoRows {
		token = randomHex(32)
		_, err = a.DB.Exec(`INSERT INTO sessions(token,user_id,tenant_id,tenant_user_id,role,device_id,ip,ua,created_at,last_active,expires_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?)`, token, userID, tenantID, tenantUserID, effectiveRole, req.DeviceID,
			r.RemoteAddr, r.UserAgent(), now(), now(), expires)
	} else if err == nil {
		_, err = a.DB.Exec(`UPDATE sessions SET tenant_id=?,tenant_user_id=?,role=?,last_active=?,expires_at=? WHERE token=?`,
			tenantID, tenantUserID, effectiveRole, now(), expires, token)
	}
	a.mu.Unlock()
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token, "role": effectiveRole, "tenantId": tenantID,
		"user": map[string]any{"id": userID, "phone": phone, "email": email, "name": name},
	})
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	var phone, email, name string
	if err := a.DB.QueryRow(`SELECT phone,email,name FROM users WHERE id=?`, s.UserID).Scan(&phone, &email, &name); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"role": s.Role, "tenantId": s.TenantID,
		"user": map[string]any{"id": s.UserID, "phone": phone, "email": email, "name": name},
	})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	_, _ = a.DB.Exec(`DELETE FROM sessions WHERE token=?`, current(r).Token)
	writeJSON(w, http.StatusOK, true)
}

type userReq struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	TenantID int64  `json:"tenantId"`
	Status   string `json:"status"`
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var req userReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Phone == "" {
		writeErr(w, fail(400, "手机号不能为空"))
		return
	}
	s := current(r)
	tenantID := s.TenantID
	role := "user"
	if s.Role == "super" {
		if req.TenantID <= 0 {
			writeErr(w, fail(400, "请指定租户"))
			return
		}
		tenantID = req.TenantID
		if req.Role == "tenant_admin" || req.Role == "operator" || req.Role == "user" {
			role = req.Role
		}
	}

	var userID int64
	err := a.DB.QueryRow(`SELECT id FROM users WHERE phone=?`, req.Phone).Scan(&userID)
	if err == sql.ErrNoRows {
		if req.Password == "" {
			writeErr(w, fail(400, "首次创建用户必须设置密码"))
			return
		}
		hash, herr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if herr != nil {
			writeErr(w, herr)
			return
		}
		res, ierr := a.DB.Exec(`INSERT INTO users(phone,email,password_hash,name,role,status,created_at) VALUES(?,?,?,?,?,?,?)`,
			req.Phone, req.Email, string(hash), req.Name, "user", "active", now())
		if ierr != nil {
			writeErr(w, ierr)
			return
		}
		userID, _ = res.LastInsertId()
	} else if err != nil {
		writeErr(w, err)
		return
	} else if req.Email != "" {
		_, _ = a.DB.Exec(`UPDATE users SET email=? WHERE id=?`, req.Email, userID)
	}

	if _, err = a.DB.Exec(`INSERT OR IGNORE INTO tenant_users(tenant_id,user_id,role,status,created_at) VALUES(?,?,?,?,?)`,
		tenantID, userID, role, "active", now()); err != nil {
		writeErr(w, err)
		return
	}
	var tuID int64
	if err = a.DB.QueryRow(`SELECT id FROM tenant_users WHERE tenant_id=? AND user_id=?`, tenantID, userID).Scan(&tuID); err != nil {
		writeErr(w, err)
		return
	}
	if role == "user" {
		if _, err = a.DB.Exec(`INSERT OR IGNORE INTO wallets(tenant_id,tenant_user_id,balance_cents,updated_at) VALUES(?,?,0,?)`,
			tenantID, tuID, now()); err != nil {
			writeErr(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": userID, "tenantUserId": tuID, "phone": req.Phone, "email": req.Email, "name": req.Name, "role": role})
}

func (a *App) listUsers(w http.ResponseWriter, r *http.Request) {
	s := current(r)
	keyword := "%" + r.URL.Query().Get("keyword") + "%"
	query := `SELECT u.id,u.phone,u.email,u.name,u.status,tu.role,tu.id,tu.tenant_id,COALESCE(t.name,''),COALESCE(w.balance_cents,0)
		FROM tenant_users tu JOIN users u ON u.id=tu.user_id
		LEFT JOIN tenants t ON t.id=tu.tenant_id
		LEFT JOIN wallets w ON w.tenant_user_id=tu.id AND w.tenant_id=tu.tenant_id
		WHERE (u.phone LIKE ? OR u.name LIKE ?)`
	args := []any{keyword, keyword}
	if s.Role != "super" {
		query += ` AND tu.tenant_id=?`
		args = append(args, s.TenantID)
	}
	query += ` ORDER BY tu.id DESC`
	rows, err := a.DB.Query(query, args...)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, tuID, tenantID, balance int64
		var phone, email, name, status, role, tenantName string
		if err = rows.Scan(&id, &phone, &email, &name, &status, &role, &tuID, &tenantID, &tenantName, &balance); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "tenantUserId": tuID, "phone": phone, "email": email,
			"name": name, "status": status, "role": role, "balanceCents": balance,
			"tenantId": tenantID, "tenantName": tenantName})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	var req userReq
	if err = decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if _, err = a.DB.Exec(`UPDATE users SET email=?,name=?,status=? WHERE id=?`, req.Email, req.Name, req.Status, id); err != nil {
		writeErr(w, err)
		return
	}
	if current(r).Role == "super" {
		_, err = a.DB.Exec(`UPDATE tenant_users SET status=? WHERE user_id=?`, req.Status, id)
	} else {
		_, err = a.DB.Exec(`UPDATE tenant_users SET status=? WHERE user_id=? AND tenant_id=?`, req.Status, id, current(r).TenantID)
	}
	if err != nil {
		writeErr(w, err)
		return
	}
	if req.Status != "active" {
		_, _ = a.DB.Exec(`DELETE FROM sessions WHERE user_id=?`, id)
	}
	writeJSON(w, http.StatusOK, true)
}

type tenantReq struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	AdminName     string `json:"adminName"`
	AdminPhone    string `json:"adminPhone"`
	AdminEmail    string `json:"adminEmail"`
	AdminPassword string `json:"adminPassword"`
}

func (a *App) listTenants(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.Query(`SELECT id,name,status,created_at FROM tenants ORDER BY id DESC`)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id int64
		var name, status, created string
		if err = rows.Scan(&id, &name, &status, &created); err != nil {
			writeErr(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "status": status, "createdAt": created})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *App) createTenant(w http.ResponseWriter, r *http.Request) {
	var req tenantReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.AdminPhone = strings.TrimSpace(req.AdminPhone)
	if req.Name == "" {
		writeErr(w, fail(400, "租户名称不能为空"))
		return
	}
	if req.AdminPhone == "" {
		writeErr(w, fail(400, "请填写管理员手机号（用于登录该租户）"))
		return
	}
	if strings.TrimSpace(req.AdminPassword) == "" {
		writeErr(w, fail(400, "请填写管理员初始密码"))
		return
	}

	hash, herr := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if herr != nil {
		writeErr(w, herr)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	tx, err := a.DB.Begin()
	if err != nil {
		writeErr(w, err)
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO tenants(name,status,created_at) VALUES(?,?,?)`, req.Name, "active", now())
	if err != nil {
		writeErr(w, err)
		return
	}
	tenantID, _ := res.LastInsertId()

	// 管理员账号全局按手机号唯一；已存在则复用用户主档，仅在该租户下新建关联。
	var adminID int64
	err = tx.QueryRow(`SELECT id FROM users WHERE phone=?`, req.AdminPhone).Scan(&adminID)
	if err == sql.ErrNoRows {
		res, err = tx.Exec(`INSERT INTO users(phone,email,password_hash,name,role,status,created_at) VALUES(?,?,?,?,?,?,?)`,
			req.AdminPhone, req.AdminEmail, string(hash), req.AdminName, "user", "active", now())
		if err != nil {
			writeErr(w, err)
			return
		}
		adminID, _ = res.LastInsertId()
	} else if err != nil {
		writeErr(w, err)
		return
	} else if req.AdminEmail != "" {
		_, _ = tx.Exec(`UPDATE users SET email=? WHERE id=?`, req.AdminEmail, adminID)
	}

	if _, err = tx.Exec(`INSERT INTO tenant_users(tenant_id,user_id,role,status,created_at) VALUES(?,?,?,?,?)`,
		tenantID, adminID, "tenant_admin", "active", now()); err != nil {
		writeErr(w, err)
		return
	}
	if err = tx.Commit(); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": tenantID, "name": req.Name, "adminPhone": req.AdminPhone, "adminUserId": adminID,
	})
}

func (a *App) updateTenant(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, err)
		return
	}
	var req tenantReq
	if err = decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if _, err = a.DB.Exec(`UPDATE tenants SET name=?,status=? WHERE id=?`, req.Name, req.Status, id); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, true)
}
