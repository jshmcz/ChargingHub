package app

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// 平台级全局配置：目前只有对外域名 base_url，供二维码生成使用。
const configBaseURL = "base_url"

func (a *App) getConfig(key string) string {
	var v string
	if err := a.DB.QueryRow(`SELECT value FROM platform_configs WHERE key=?`, key).Scan(&v); err != nil {
		return ""
	}
	return v
}

func (a *App) setConfig(key, value string) error {
	_, err := a.DB.Exec(`INSERT INTO platform_configs(key,value,updated_at) VALUES(?,?,?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value,updated_at=excluded.updated_at`,
		key, value, time.Now().Format(time.RFC3339))
	return err
}

func (a *App) getPlatformConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"baseUrl": a.getConfig(configBaseURL),
	})
}

func (a *App) putPlatformConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BaseURL string `json:"baseUrl"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	base := strings.TrimSpace(req.BaseURL)
	base = strings.TrimRight(base, "/")
	if base != "" && !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		writeErr(w, fail(400, "域名需以 http:// 或 https:// 开头"))
		return
	}
	if err := a.setConfig(configBaseURL, base); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"baseUrl": base})
}

// nextPileSeq 生成全局递增桩流水（10 位，不足补 0）。用事务内计数表保证并发唯一。
func (a *App) nextPileSeq() (int64, error) {
	var seq int64
	err := a.DB.QueryRow(`SELECT value FROM platform_configs WHERE key='pile_seq'`).Scan(&seq)
	if err == sql.ErrNoRows {
		seq = 0
	} else if err != nil {
		return 0, err
	}
	seq++
	if err = a.setConfig("pile_seq", fmt.Sprintf("%d", seq)); err != nil {
		return 0, err
	}
	return seq, nil
}

// makePileNo 按场站区划码生成 16 位桩编号：6 位区划码 + 10 位流水。
func makePileNo(areaCode string, seq int64) string {
	code := areaCode
	if len(code) != 6 {
		code = "000000"
	}
	return code + fmt.Sprintf("%010d", seq)
}

// makeGunCode 生成 18 位枪编号：16 位桩编号 + 2 位枪序号（从 01 开始）。
func makeGunCode(pileNo string, gunNo int) string {
	return fmt.Sprintf("%s%02d", pileNo, gunNo)
}
