// Package store 封装 PostgreSQL 访问。
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"youthlab/lottery-ai/internal/model"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) Pool() *pgxpool.Pool { return s.pool }

func (s *Store) Migrate(ctx context.Context, sqlPath string) error {
	raw, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, string(raw))
	return err
}

func (s *Store) ListLotteryTypes(ctx context.Context) ([]model.LotteryType, error) {
	rows, err := s.pool.Query(ctx, `SELECT code, name, rules, draw_days FROM lottery_types ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.LotteryType
	for rows.Next() {
		var item model.LotteryType
		if err := rows.Scan(&item.Code, &item.Name, &item.Rules, &item.DrawDays); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) LatestDraw(ctx context.Context, lotteryCode string) (*model.DrawResult, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, lottery_code, issue, draw_date, result, COALESCE(raw_data, '{}'::jsonb)
		FROM draw_results WHERE lottery_code=$1
		ORDER BY draw_date DESC, issue DESC LIMIT 1`, lotteryCode)
	var d model.DrawResult
	if err := row.Scan(&d.ID, &d.LotteryCode, &d.Issue, &d.DrawDate, &d.Result, &d.RawData); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (s *Store) UpsertDraw(ctx context.Context, d model.DrawResult) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO draw_results (lottery_code, issue, draw_date, result, raw_data)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (lottery_code, issue) DO UPDATE
		SET draw_date=EXCLUDED.draw_date, result=EXCLUDED.result, raw_data=EXCLUDED.raw_data`,
		d.LotteryCode, d.Issue, d.DrawDate, d.Result, nullJSON(d.RawData))
	return err
}

func (s *Store) ListDraws(ctx context.Context, lotteryCode string, page, pageSize int) ([]model.DrawResult, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM draw_results WHERE lottery_code=$1`, lotteryCode).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, lottery_code, issue, draw_date, result
		FROM draw_results WHERE lottery_code=$1
		ORDER BY draw_date DESC, issue DESC
		LIMIT $2 OFFSET $3`, lotteryCode, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.DrawResult
	for rows.Next() {
		var d model.DrawResult
		if err := rows.Scan(&d.ID, &d.LotteryCode, &d.Issue, &d.DrawDate, &d.Result); err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, total, rows.Err()
}

func (s *Store) ListRecentDraws(ctx context.Context, lotteryCode string, limit int) ([]model.DrawResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, lottery_code, issue, draw_date, result
		FROM draw_results WHERE lottery_code=$1
		ORDER BY draw_date DESC, issue DESC LIMIT $2`, lotteryCode, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.DrawResult
	for rows.Next() {
		var d model.DrawResult
		if err := rows.Scan(&d.ID, &d.LotteryCode, &d.Issue, &d.DrawDate, &d.Result); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) SaveRawLog(ctx context.Context, lotteryCode, url string, status int, body string) {
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO sync_raw_logs (lottery_code, source_url, status_code, body)
		VALUES ($1,$2,$3,$4)`, lotteryCode, url, status, truncate(body, 200000))
}

func (s *Store) ListEnabledModels(ctx context.Context) ([]model.ModelConfig, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, model_code, display_name, provider, base_url, model_name, api_key_env, weight, enabled, timeout_sec
		FROM model_configs WHERE enabled=TRUE ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ModelConfig
	for rows.Next() {
		var m model.ModelConfig
		if err := rows.Scan(&m.ID, &m.ModelCode, &m.DisplayName, &m.Provider, &m.BaseURL, &m.ModelName, &m.APIKeyEnv, &m.Weight, &m.Enabled, &m.TimeoutSec); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) InsertPrediction(ctx context.Context, p model.Prediction) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO predictions
		(lottery_code, issue, predict_date, model_code, predicted_numbers, confidence, reason, raw_response, final_flag, success, error_message)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id`,
		p.LotteryCode, p.Issue, p.PredictDate, p.ModelCode, p.PredictedNumbers, p.Confidence, p.Reason, p.RawResponse, p.FinalFlag, p.Success, nullStr(p.ErrorMessage),
	).Scan(&id)
	return id, err
}

func (s *Store) ListPredictions(ctx context.Context, lotteryCode string, date *time.Time, page, pageSize int) ([]model.Prediction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	where := []string{"lottery_code=$1"}
	args := []any{lotteryCode}
	if date != nil {
		args = append(args, date.Format("2006-01-02"))
		where = append(where, fmt.Sprintf("predict_date=$%d", len(args)))
	}
	w := strings.Join(where, " AND ")
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM predictions WHERE "+w, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, (page-1)*pageSize)
	q := fmt.Sprintf(`SELECT id, lottery_code, issue, predict_date, model_code, predicted_numbers, confidence, COALESCE(reason,''), COALESCE(raw_response,''), final_flag, success, COALESCE(error_message,''), created_at
		FROM predictions WHERE %s ORDER BY predict_date DESC, id DESC LIMIT $%d OFFSET $%d`, w, len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.Prediction
	for rows.Next() {
		var p model.Prediction
		if err := rows.Scan(&p.ID, &p.LotteryCode, &p.Issue, &p.PredictDate, &p.ModelCode, &p.PredictedNumbers, &p.Confidence, &p.Reason, &p.RawResponse, &p.FinalFlag, &p.Success, &p.ErrorMessage, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

func (s *Store) TodayPredictions(ctx context.Context, lotteryCode string, day time.Time) ([]model.Prediction, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, lottery_code, issue, predict_date, model_code, predicted_numbers, confidence, COALESCE(reason,''), COALESCE(raw_response,''), final_flag, success, COALESCE(error_message,''), created_at
		FROM predictions WHERE lottery_code=$1 AND predict_date=$2
		ORDER BY final_flag DESC, id ASC`, lotteryCode, day.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Prediction
	for rows.Next() {
		var p model.Prediction
		if err := rows.Scan(&p.ID, &p.LotteryCode, &p.Issue, &p.PredictDate, &p.ModelCode, &p.PredictedNumbers, &p.Confidence, &p.Reason, &p.RawResponse, &p.FinalFlag, &p.Success, &p.ErrorMessage, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetAccuracy(ctx context.Context, lotteryCode string) ([]model.AccuracyStat, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT lottery_code, model_code, total_predictions, total_hits, avg_hit_rate, last_30_hit_rate, updated_at
		FROM prediction_accuracy WHERE lottery_code=$1 ORDER BY avg_hit_rate DESC`, lotteryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AccuracyStat
	for rows.Next() {
		var a model.AccuracyStat
		if err := rows.Scan(&a.LotteryCode, &a.ModelCode, &a.TotalPredictions, &a.TotalHits, &a.AvgHitRate, &a.Last30HitRate, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) UpsertAccuracy(ctx context.Context, a model.AccuracyStat) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO prediction_accuracy (lottery_code, model_code, total_predictions, total_hits, avg_hit_rate, last_30_hit_rate, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,NOW())
		ON CONFLICT (lottery_code, model_code) DO UPDATE SET
			total_predictions=EXCLUDED.total_predictions,
			total_hits=EXCLUDED.total_hits,
			avg_hit_rate=EXCLUDED.avg_hit_rate,
			last_30_hit_rate=EXCLUDED.last_30_hit_rate,
			updated_at=NOW()`,
		a.LotteryCode, a.ModelCode, a.TotalPredictions, a.TotalHits, a.AvgHitRate, a.Last30HitRate)
	return err
}

func (s *Store) InsertPredictionResult(ctx context.Context, r model.PredictionResult) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO prediction_results (prediction_id, draw_result_id, matched_numbers, hit_count, hit_rate, level, is_win)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (prediction_id, draw_result_id) DO UPDATE SET
			matched_numbers=EXCLUDED.matched_numbers,
			hit_count=EXCLUDED.hit_count,
			hit_rate=EXCLUDED.hit_rate,
			level=EXCLUDED.level,
			is_win=EXCLUDED.is_win`,
		r.PredictionID, r.DrawResultID, r.MatchedNumbers, r.HitCount, r.HitRate, r.Level, r.IsWin)
	return err
}

func (s *Store) UnevaluatedPredictions(ctx context.Context, lotteryCode string) ([]model.Prediction, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.lottery_code, p.issue, p.predict_date, p.model_code, p.predicted_numbers, p.confidence,
			COALESCE(p.reason,''), COALESCE(p.raw_response,''), p.final_flag, p.success, COALESCE(p.error_message,''), p.created_at
		FROM predictions p
		LEFT JOIN prediction_results pr ON pr.prediction_id=p.id
		WHERE p.lottery_code=$1 AND p.success=TRUE AND pr.id IS NULL
		ORDER BY p.id ASC LIMIT 500`, lotteryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Prediction
	for rows.Next() {
		var p model.Prediction
		if err := rows.Scan(&p.ID, &p.LotteryCode, &p.Issue, &p.PredictDate, &p.ModelCode, &p.PredictedNumbers, &p.Confidence, &p.Reason, &p.RawResponse, &p.FinalFlag, &p.Success, &p.ErrorMessage, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetDrawByIssue(ctx context.Context, lotteryCode, issue string) (*model.DrawResult, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, lottery_code, issue, draw_date, result FROM draw_results
		WHERE lottery_code=$1 AND issue=$2`, lotteryCode, issue)
	var d model.DrawResult
	if err := row.Scan(&d.ID, &d.LotteryCode, &d.Issue, &d.DrawDate, &d.Result); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (s *Store) ModelHitRate(ctx context.Context, lotteryCode, modelCode string) float64 {
	var rate float64
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(avg_hit_rate, 0) FROM prediction_accuracy
		WHERE lottery_code=$1 AND model_code=$2`, lotteryCode, modelCode).Scan(&rate)
	return rate
}

func nullJSON(b json.RawMessage) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
