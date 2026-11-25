package app

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jshmcz/ChargingHub/internal/config"
)

type CommandSender interface {
	SendCommand(deviceNo string, gunCode string, command string, payload string) bool
}

type App struct {
	DB       *sql.DB
	Cfg      config.Config
	mu       sync.Mutex
	notifyCh chan notifyJob
	Sender   CommandSender
	warned   map[int64]time.Time
}

func New(db *sql.DB, cfg config.Config) *App {
	return &App{
		DB:       db,
		Cfg:      cfg,
		notifyCh: make(chan notifyJob, 200),
		warned:   map[int64]time.Time{},
	}
}

func (a *App) SetSender(s CommandSender) {
	a.Sender = s
}

type ctxKey string

const ctxSession ctxKey = "session"

type Session struct {
	Token        string
	UserID       int64
	TenantID     int64
	TenantUserID int64
	Role         string
	DeviceID     string
}

type apiError struct {
	code int
	msg  string
}

func (e apiError) Error() string { return e.msg }

func fail(code int, msg string) error { return apiError{code: code, msg: msg} }

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "ok", "data": data})
}

func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := 500
	msg := err.Error()
	var ae apiError
	if errors.As(err, &ae) {
		code = ae.code
		msg = ae.msg
		if code == 401 || code == 403 || code == 404 || code == 409 {
			status = code
		} else {
			status = http.StatusBadRequest
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "message": msg, "data": nil})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := dec.Decode(dst); err != nil {
		return fail(400, "请求格式错误")
	}
	return nil
}

func current(r *http.Request) Session {
	v, _ := r.Context().Value(ctxSession).(Session)
	return v
}

func (a *App) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			writeErr(w, fail(401, "未登录"))
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		var s Session
		err := a.DB.QueryRow(`SELECT token,user_id,tenant_id,tenant_user_id,role,device_id FROM sessions WHERE token=? AND expires_at>?`,
			token, now()).Scan(&s.Token, &s.UserID, &s.TenantID, &s.TenantUserID, &s.Role, &s.DeviceID)
		if err != nil {
			writeErr(w, fail(401, "登录已失效，请重新登录"))
			return
		}
		_, _ = a.DB.Exec(`UPDATE sessions SET last_active=? WHERE token=?`, now(), token)
		next(w, r.WithContext(contextWith(r, s)))
	}
}

func contextWith(r *http.Request, s Session) context.Context {
	return context.WithValue(r.Context(), ctxSession, s)
}

func requireRoles(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			role := current(r).Role
			// 超管拥有全部权限;所见数据范围由各 list 接口按 role 自行放宽。
			if role == "super" {
				next(w, r)
				return
			}
			for _, x := range roles {
				if role == x {
					next(w, r)
					return
				}
			}
			writeErr(w, fail(403, "无权限"))
		}
	}
}

// scopeTenant 返回写操作应落到的租户 ID。
// 普通角色锁定在会话租户；超管需显式指定租户，避免误写入 tenant 0。
func scopeTenant(r *http.Request, requested int64) (int64, error) {
	s := current(r)
	if s.Role != "super" {
		return s.TenantID, nil
	}
	if requested <= 0 {
		return 0, fail(400, "超管操作需指定所属租户")
	}
	return requested, nil
}

func pathID(r *http.Request, name string) (int64, error) {
	v := r.PathValue(name)
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id <= 0 {
		return 0, fail(400, "ID 参数错误")
	}
	return id, nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func (a *App) cipher() (cipher.AEAD, error) {
	key := sha256.Sum256([]byte(a.Cfg.SecretKey))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (a *App) encrypt(plain string) string {
	if plain == "" {
		return ""
	}
	g, err := a.cipher()
	if err != nil {
		return ""
	}
	nonce := make([]byte, g.NonceSize())
	_, _ = rand.Read(nonce)
	out := g.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(out)
}

func (a *App) decrypt(encoded string) string {
	if encoded == "" {
		return ""
	}
	g, err := a.cipher()
	if err != nil {
		return ""
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) < g.NonceSize() {
		return ""
	}
	plain, err := g.Open(nil, raw[:g.NonceSize()], raw[g.NonceSize():], nil)
	if err != nil {
		return ""
	}
	return string(plain)
}

func nextStartTime(hhmm string) time.Time {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if loc == nil {
		loc = time.Local
	}
	parts := strings.Split(hhmm, ":")
	hour, min := 22, 0
	if len(parts) == 2 {
		if h, err := strconv.Atoi(parts[0]); err == nil && h >= 0 && h <= 23 {
			hour = h
		}
		if m, err := strconv.Atoi(parts[1]); err == nil && m >= 0 && m <= 59 {
			min = m
		}
	}
	now := time.Now().In(loc)
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)
	if !t.After(now) {
		t = t.Add(24 * time.Hour)
	}
	return t
}
