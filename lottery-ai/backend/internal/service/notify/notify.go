package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"youthlab/lottery-ai/internal/model"
	"youthlab/lottery-ai/internal/store"
)

type Service struct {
	Store *store.Store
	HTTP  *http.Client
}

func New(st *store.Store) *Service {
	return &Service{
		Store: st,
		HTTP:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *Service) Publish(ctx context.Context, n model.AppNotification) error {
	id, err := s.Store.InsertNotification(ctx, n)
	if err != nil {
		return err
	}
	n.ID = id
	tokens, err := s.Store.ListPushTokens(ctx)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return nil
	}
	msgs := make([]expoPush, 0, len(tokens))
	for _, t := range tokens {
		msgs = append(msgs, expoPush{
			To:    t,
			Title: n.Title,
			Body:  n.Body,
			Sound: "default",
			Data: map[string]any{
				"id":   n.ID,
				"type": n.Type,
			},
		})
	}
	b, _ := json.Marshal(msgs)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://exp.host/--/api/v2/push/send", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := s.HTTP.Do(req)
	if err != nil {
		log.Printf("[notify] expo push failed: %v", err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("[notify] expo push status=%d", resp.StatusCode)
	}
	return nil
}

func (s *Service) PublishBestEffort(ctx context.Context, typ, title, body string, payload any) {
	raw, _ := json.Marshal(payload)
	if len(raw) == 0 {
		raw = []byte("{}")
	}
	if err := s.Publish(ctx, model.AppNotification{
		Type:    typ,
		Title:   title,
		Body:    body,
		Payload: raw,
	}); err != nil {
		log.Printf("[notify] save failed: %v", err)
	}
}

type expoPush struct {
	To    string         `json:"to"`
	Title string         `json:"title"`
	Body  string         `json:"body"`
	Sound string         `json:"sound"`
	Data  map[string]any `json:"data"`
}

func Short(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return fmt.Sprintf("%s…", s[:n])
}
