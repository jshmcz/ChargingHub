package app

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jshmcz/ChargingHub/internal/config"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  phone TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL DEFAULT '',
  password_hash TEXT NOT NULL,
  name TEXT NOT NULL DEFAULT '',
  role TEXT NOT NULL DEFAULT 'user',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS tenants (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS tenant_users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  role TEXT NOT NULL DEFAULT 'user',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL,
  UNIQUE(tenant_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_tenant_users_user ON tenant_users(user_id);
CREATE TABLE IF NOT EXISTS wallets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id INTEGER NOT NULL,
  tenant_user_id INTEGER NOT NULL,
  balance_cents INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'active',
  updated_at TEXT NOT NULL,
  UNIQUE(tenant_id, tenant_user_id)
);
CREATE TABLE IF NOT EXISTS wallet_ledgers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id INTEGER NOT NULL,
  wallet_id INTEGER NOT NULL,
  tenant_user_id INTEGER NOT NULL,
  kind TEXT NOT NULL,
  amount_cents INTEGER NOT NULL,
  balance_after INTEGER NOT NULL,
  method TEXT NOT NULL DEFAULT '',
  voucher_no TEXT NOT NULL DEFAULT '',
  remark TEXT NOT NULL DEFAULT '',
  operator_id INTEGER NOT NULL DEFAULT 0,
  biz_no TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ledgers_wallet ON wallet_ledgers(wallet_id, id);
CREATE TABLE IF NOT EXISTS stations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  address TEXT NOT NULL DEFAULT '',
  default_start_time TEXT NOT NULL DEFAULT '22:00',
  start_threshold_cents INTEGER NOT NULL DEFAULT 5000,
  stop_threshold_cents INTEGER NOT NULL DEFAULT 2000,
  warn_threshold_cents INTEGER NOT NULL DEFAULT 3000,
  settlement_timeout_hours INTEGER NOT NULL DEFAULT 24,
  price_cents_per_kwh INTEGER NOT NULL DEFAULT 150,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_stations_tenant ON stations(tenant_id);
CREATE TABLE IF NOT EXISTS piles (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id INTEGER NOT NULL,
  station_id INTEGER NOT NULL,
  device_no TEXT NOT NULL UNIQUE,
  secret TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  last_seen_at TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_piles_station ON piles(station_id);
CREATE TABLE IF NOT EXISTS guns (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id INTEGER NOT NULL,
  pile_id INTEGER NOT NULL,
  gun_no INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'idle',
  created_at TEXT NOT NULL,
  UNIQUE(pile_id, gun_no)
);
CREATE INDEX IF NOT EXISTS idx_guns_pile ON guns(pile_id);
CREATE TABLE IF NOT EXISTS reservations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id INTEGER NOT NULL,
  station_id INTEGER NOT NULL,
  gun_id INTEGER NOT NULL,
  tenant_user_id INTEGER NOT NULL,
  start_at TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  created_at TEXT NOT NULL,
  started_at TEXT NOT NULL DEFAULT '',
  closed_at TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_res_pending ON reservations(status, start_at);
CREATE TABLE IF NOT EXISTS orders (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id INTEGER NOT NULL,
  station_id INTEGER NOT NULL,
  gun_id INTEGER NOT NULL,
  tenant_user_id INTEGER NOT NULL,
  reservation_id INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'charging',
  start_at TEXT NOT NULL,
  end_at TEXT NOT NULL DEFAULT '',
  settle_at TEXT NOT NULL DEFAULT '',
  energy_wh INTEGER NOT NULL DEFAULT 0,
  amount_cents INTEGER NOT NULL DEFAULT 0,
  last_report_at TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status, end_at);
CREATE TABLE IF NOT EXISTS sessions (
  token TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL,
  tenant_id INTEGER NOT NULL DEFAULT 0,
  tenant_user_id INTEGER NOT NULL DEFAULT 0,
  role TEXT NOT NULL,
  device_id TEXT NOT NULL DEFAULT '',
  ip TEXT NOT NULL DEFAULT '',
  ua TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  last_active TEXT NOT NULL,
  expires_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE TABLE IF NOT EXISTS smtp_configs (
  tenant_id INTEGER PRIMARY KEY,
  host TEXT NOT NULL DEFAULT '',
  port INTEGER NOT NULL DEFAULT 587,
  username TEXT NOT NULL DEFAULT '',
  password_enc TEXT NOT NULL DEFAULT '',
  from_addr TEXT NOT NULL DEFAULT '',
  encryption TEXT NOT NULL DEFAULT 'starttls',
  enabled INTEGER NOT NULL DEFAULT 0,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS notifications (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tenant_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL DEFAULT 0,
  event TEXT NOT NULL,
  subject TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'queued',
  error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  sent_at TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_notify_tenant ON notifications(tenant_id, id);
CREATE TABLE IF NOT EXISTS device_commands (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  device_no TEXT NOT NULL,
  gun_no INTEGER NOT NULL DEFAULT 0,
  command TEXT NOT NULL,
  payload TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  created_at TEXT NOT NULL,
  acked_at TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_devcmd_device ON device_commands(device_no, status);
`

func OpenDB(cfg config.Config) (*sql.DB, error) {
	if dir := filepath.Dir(cfg.DBPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create data dir: %w", err)
		}
	}
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", filepath.ToSlash(cfg.DBPath))
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err = db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err = seedSuperAdmin(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func seedSuperAdmin(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE role='super'`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO users(phone,email,password_hash,name,role,status,created_at) VALUES(?,?,?,?,?,?,?)`,
		"admin", "", string(hash), "超级管理员", "super", "active", now())
	return err
}

func now() string {
	return time.Now().Format(time.RFC3339)
}
