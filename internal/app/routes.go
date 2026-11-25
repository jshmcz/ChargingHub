package app

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/login", a.login)
	mux.HandleFunc("GET /api/me", a.requireAuth(a.me))
	mux.HandleFunc("POST /api/logout", a.requireAuth(a.logout))

	tenantManage := requireRoles("tenant_admin", "operator")
	tenantAdmin := requireRoles("tenant_admin")

	mux.HandleFunc("GET /api/tenants", a.requireAuth(requireRoles("super")(a.listTenants)))
	mux.HandleFunc("POST /api/tenants", a.requireAuth(requireRoles("super")(a.createTenant)))
	mux.HandleFunc("PUT /api/tenants/{id}", a.requireAuth(requireRoles("super")(a.updateTenant)))

	userCreate := requireRoles("super", "tenant_admin", "operator")
	mux.HandleFunc("POST /api/users", a.requireAuth(userCreate(a.createUser)))
	mux.HandleFunc("GET /api/users", a.requireAuth(tenantManage(a.listUsers)))
	mux.HandleFunc("PUT /api/users/{id}", a.requireAuth(tenantManage(a.updateUser)))

	mux.HandleFunc("GET /api/stations", a.requireAuth(a.listStations))
	mux.HandleFunc("POST /api/stations", a.requireAuth(tenantManage(a.createStation)))
	mux.HandleFunc("PUT /api/stations/{id}", a.requireAuth(tenantManage(a.updateStation)))
	mux.HandleFunc("GET /api/piles", a.requireAuth(a.listPiles))
	mux.HandleFunc("POST /api/piles", a.requireAuth(tenantManage(a.createPile)))
	mux.HandleFunc("PUT /api/piles/{id}", a.requireAuth(tenantManage(a.updatePile)))
	mux.HandleFunc("POST /api/piles/{id}/regenerate-secret", a.requireAuth(tenantManage(a.regenerateSecret)))
	mux.HandleFunc("GET /api/guns", a.requireAuth(a.listGuns))
	mux.HandleFunc("GET /api/gun/by-token", a.requireAuth(a.gunByToken))
	mux.HandleFunc("GET /api/guns/{id}/qrcode", a.requireAuth(tenantManage(a.gunQrcode)))

	mux.HandleFunc("GET /api/platform/config", a.requireAuth(requireRoles("super")(a.getPlatformConfig)))
	mux.HandleFunc("PUT /api/platform/config", a.requireAuth(requireRoles("super")(a.putPlatformConfig)))

	mux.HandleFunc("GET /api/wallets", a.requireAuth(tenantManage(a.listWallets)))
	mux.HandleFunc("GET /api/ledgers", a.requireAuth(tenantManage(a.listLedgers)))
	mux.HandleFunc("POST /api/wallet/recharge", a.requireAuth(tenantAdmin(a.recharge)))
	mux.HandleFunc("POST /api/wallet/refund", a.requireAuth(tenantAdmin(a.refund)))
	mux.HandleFunc("GET /api/my/wallet", a.requireAuth(a.myWallet))
	mux.HandleFunc("GET /api/my/ledgers", a.requireAuth(a.myLedgers))
	mux.HandleFunc("GET /api/my/orders", a.requireAuth(a.listOrders))
	mux.HandleFunc("GET /api/my/reservations", a.requireAuth(a.listReservations))

	mux.HandleFunc("GET /api/reservations", a.requireAuth(a.listReservations))
	mux.HandleFunc("POST /api/reservations", a.requireAuth(a.createReservation))
	mux.HandleFunc("POST /api/reservations/{id}/cancel", a.requireAuth(a.cancelReservation))
	mux.HandleFunc("POST /api/charge/start", a.requireAuth(a.startCharge))
	mux.HandleFunc("POST /api/charge/stop", a.requireAuth(a.stopCharge))
	mux.HandleFunc("GET /api/orders", a.requireAuth(a.listOrders))
	mux.HandleFunc("POST /api/orders/{id}/settle", a.requireAuth(tenantManage(a.manualSettle)))

	mux.HandleFunc("GET /api/smtp", a.requireAuth(tenantAdmin(a.getSMTP)))
	mux.HandleFunc("PUT /api/smtp", a.requireAuth(tenantAdmin(a.putSMTP)))
	mux.HandleFunc("GET /api/notifications", a.requireAuth(a.listNotifications))

	mux.HandleFunc("/", a.static())

	return cors(mux)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) static() http.HandlerFunc {
	adminRoot := http.Dir(a.Cfg.StaticDir)
	mobileRoot := http.Dir(a.Cfg.MobileDir)
	adminFiles := http.FileServer(adminRoot)
	mobileFiles := http.StripPrefix("/m", http.FileServer(mobileRoot))
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/m" || strings.HasPrefix(r.URL.Path, "/m/") {
			sub := strings.TrimPrefix(r.URL.Path, "/m")
			if sub == "" {
				sub = "/"
			}
			if fileExists(a.Cfg.MobileDir, sub) {
				rr := r.Clone(r.Context())
				rr.URL.Path = sub
				mobileFiles.ServeHTTP(w, rr)
				return
			}
			if strings.Contains(filepath.Base(sub), ".") {
				http.NotFound(w, r)
				return
			}
			serveIndex(w, mobileRoot)
			return
		}
		if r.URL.Path == "/" || fileExists(a.Cfg.StaticDir, r.URL.Path) {
			adminFiles.ServeHTTP(w, r)
			return
		}
		if strings.Contains(filepath.Base(r.URL.Path), ".") {
			http.NotFound(w, r)
			return
		}
		serveIndex(w, adminRoot)
	}
}

func serveIndex(w http.ResponseWriter, root http.Dir) {
	f, err := root.Open("/index.html")
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, "<h1>ChargingHub</h1><p>前端尚未构建。请运行 npm run build。</p>")
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, f)
}

func fileExists(root, path string) bool {
	clean := filepath.Clean(strings.TrimPrefix(path, "/"))
	if clean == "." || clean == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(root, clean))
	return err == nil
}
