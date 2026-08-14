// Package config 读取环境变量配置。
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	AdminToken    string
	APIToken      string
	HistoryN      int
	HTTPTimeout   time.Duration
	Disclaimer    string
	Timezone      string
}

func Load() Config {
	return Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8090"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://lottery:lottery@localhost:5432/lottery_ai?sslmode=disable"),
		AdminToken:  getenv("ADMIN_TOKEN", "change-me-admin"),
		APIToken:    getenv("API_TOKEN", ""),
		HistoryN:    getenvInt("HISTORY_N", 500),
		HTTPTimeout: time.Duration(getenvInt("HTTP_TIMEOUT_SEC", 20)) * time.Second,
		Timezone:    getenv("TZ", "Asia/Shanghai"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
