// SMTP email sender for daily lottery report.
package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
)

type MailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	To       string
}

func loadMailConfigFromEnv() (MailConfig, error) {
	cfg := MailConfig{
		Host:     getenv("SMTP_HOST", "smtp.qq.com"),
		Port:     getenv("SMTP_PORT", "465"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     getenv("MAIL_FROM", os.Getenv("SMTP_USER")),
		To:       getenv("MAIL_TO", "webyouth@qq.com"),
	}
	if cfg.Username == "" || cfg.Password == "" {
		return cfg, fmt.Errorf("SMTP_USER/SMTP_PASS are required")
	}
	if cfg.From == "" {
		cfg.From = cfg.Username
	}
	if cfg.To == "" {
		return cfg, fmt.Errorf("MAIL_TO is required")
	}
	return cfg, nil
}

func sendMail(cfg MailConfig, subject, body string) error {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	msg := buildMessage(cfg.From, cfg.To, subject, body)

	// QQ mail commonly uses SMTPS on 465.
	if cfg.Port == "465" {
		return sendSMTPS(addr, cfg, msg)
	}
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	return smtp.SendMail(addr, auth, cfg.From, splitAddrs(cfg.To), msg)
}

func sendSMTPS(addr string, cfg MailConfig, msg []byte) error {
	tlsCfg := &tls.Config{ServerName: cfg.Host}
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(cfg.From); err != nil {
		return err
	}
	for _, to := range splitAddrs(cfg.To) {
		if err := client.Rcpt(to); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

func splitAddrs(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
