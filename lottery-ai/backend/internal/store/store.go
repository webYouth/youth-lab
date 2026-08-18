// Package store 封装 PostgreSQL 访问。
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	// 初始化 SQL 含多条语句，需使用 simple protocol。
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	var pool *pgxpool.Pool
	var lastErr error
	for i := 1; i <= 10; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = pool.Ping(pingCtx)
			cancel()
		}
		if err == nil {
			return &Store{pool: pool}, nil
		}
		lastErr = err
		if pool != nil {
			pool.Close()
		}
		time.Sleep(time.Duration(i) * time.Second)
	}
	return nil, fmt.Errorf("db connect failed after retries: %w", lastErr)
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) Pool() *pgxpool.Pool { return s.pool }

func (s *Store) Migrate(ctx context.Context, sqlPath string) error {
	files, err := sqlFiles(sqlPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		for _, stmt := range splitSQL(string(raw)) {
			if _, err := s.pool.Exec(ctx, stmt); err != nil {
				return fmt.Errorf("migrate %s failed: %w; sql=%s", filepath.Base(f), err, truncate(stmt, 180))
			}
		}
	}
	return nil
}

func sqlFiles(p string) ([]string, error) {
	st, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	dir := p
	if !st.IsDir() {
		dir = filepath.Dir(p)
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, filepath.Join(dir, e.Name()))
	}
	sort.Strings(files)
	if len(files) == 0 {
		return []string{p}, nil
	}
	return files, nil
}

func splitSQL(raw string) []string {
	parts := strings.Split(raw, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		stmt := strings.TrimSpace(p)
		if stmt == "" {
			continue
		}
		out = append(out, stmt)
	}
	return out
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
		d.LotteryCode, d.Issue, dateParam(d.DrawDate), d.Result, nullJSON(d.RawData))
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
		ON CONFLICT (lottery_code, issue, model_code) DO UPDATE SET
			predict_date=EXCLUDED.predict_date,
			predicted_numbers=EXCLUDED.predicted_numbers,
			confidence=EXCLUDED.confidence,
			reason=EXCLUDED.reason,
			raw_response=EXCLUDED.raw_response,
			final_flag=EXCLUDED.final_flag,
			success=EXCLUDED.success,
			error_message=EXCLUDED.error_message
		RETURNING id`,
		p.LotteryCode, p.Issue, dateParam(p.PredictDate), p.ModelCode, p.PredictedNumbers, p.Confidence, p.Reason, p.RawResponse, p.FinalFlag, p.Success, nullStr(p.ErrorMessage),
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

func (s *Store) TodayPredictions(ctx context.Context, lotteryCode string, _ time.Time) ([]model.Prediction, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, lottery_code, issue, predict_date, model_code, predicted_numbers, confidence, COALESCE(reason,''), COALESCE(raw_response,''), final_flag, success, COALESCE(error_message,''), created_at
		FROM predictions
		WHERE lottery_code=$1 AND issue=(
			SELECT issue FROM predictions WHERE lottery_code=$1 ORDER BY created_at DESC, id DESC LIMIT 1
		)
		ORDER BY final_flag DESC, id ASC`, lotteryCode)
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
		SELECT lottery_code, model_code, total_predictions, total_hits, total_wins, avg_hit_rate, last_30_hit_rate,
			total_stake, total_prize, total_profit, last_30_wins, last_30_profit, updated_at
		FROM prediction_accuracy WHERE lottery_code=$1 ORDER BY avg_hit_rate DESC, total_profit DESC`, lotteryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AccuracyStat
	for rows.Next() {
		var a model.AccuracyStat
		if err := rows.Scan(&a.LotteryCode, &a.ModelCode, &a.TotalPredictions, &a.TotalHits, &a.TotalWins, &a.AvgHitRate, &a.Last30HitRate,
			&a.TotalStake, &a.TotalPrize, &a.TotalProfit, &a.Last30Wins, &a.Last30Profit, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) ListAccuracyHistory(ctx context.Context, lotteryCode string, limit int) ([]model.AccuracyHistory, error) {
	if limit < 1 || limit > 100 {
		limit = 40
	}
	rows, err := s.pool.Query(ctx, `
		SELECT p.issue, p.model_code, p.final_flag, to_char(d.draw_date, 'YYYY-MM-DD'),
			pr.is_win, COALESCE(pr.level,''), pr.hit_count,
			COALESCE(pr.stake_yuan,0), COALESCE(pr.prize_yuan,0), COALESCE(pr.profit_yuan,0), COALESCE(pr.prize_floating,FALSE)
		FROM prediction_results pr
		JOIN predictions p ON p.id=pr.prediction_id
		JOIN draw_results d ON d.id=pr.draw_result_id
		WHERE p.lottery_code=$1 AND p.success=TRUE AND p.final_flag=TRUE
		ORDER BY d.draw_date DESC, p.issue DESC
		LIMIT $2`, lotteryCode, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AccuracyHistory
	for rows.Next() {
		var h model.AccuracyHistory
		if err := rows.Scan(&h.Issue, &h.ModelCode, &h.FinalFlag, &h.DrawDate, &h.IsWin, &h.Level, &h.HitCount,
			&h.StakeYuan, &h.PrizeYuan, &h.ProfitYuan, &h.PrizeFloating); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *Store) UpsertAccuracy(ctx context.Context, a model.AccuracyStat) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO prediction_accuracy (
			lottery_code, model_code, total_predictions, total_hits, total_wins, avg_hit_rate, last_30_hit_rate,
			total_stake, total_prize, total_profit, last_30_wins, last_30_profit, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW())
		ON CONFLICT (lottery_code, model_code) DO UPDATE SET
			total_predictions=EXCLUDED.total_predictions,
			total_hits=EXCLUDED.total_hits,
			total_wins=EXCLUDED.total_wins,
			avg_hit_rate=EXCLUDED.avg_hit_rate,
			last_30_hit_rate=EXCLUDED.last_30_hit_rate,
			total_stake=EXCLUDED.total_stake,
			total_prize=EXCLUDED.total_prize,
			total_profit=EXCLUDED.total_profit,
			last_30_wins=EXCLUDED.last_30_wins,
			last_30_profit=EXCLUDED.last_30_profit,
			updated_at=NOW()`,
		a.LotteryCode, a.ModelCode, a.TotalPredictions, a.TotalHits, a.TotalWins, a.AvgHitRate, a.Last30HitRate,
		a.TotalStake, a.TotalPrize, a.TotalProfit, a.Last30Wins, a.Last30Profit)
	return err
}

func (s *Store) InsertPredictionResult(ctx context.Context, r model.PredictionResult) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO prediction_results (
			prediction_id, draw_result_id, matched_numbers, hit_count, hit_rate, level, is_win,
			stake_yuan, prize_yuan, profit_yuan, prize_floating, weight_yuan, weight_score, score_version
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (prediction_id, draw_result_id) DO UPDATE SET
			matched_numbers=EXCLUDED.matched_numbers,
			hit_count=EXCLUDED.hit_count,
			hit_rate=EXCLUDED.hit_rate,
			level=EXCLUDED.level,
			is_win=EXCLUDED.is_win,
			stake_yuan=EXCLUDED.stake_yuan,
			prize_yuan=EXCLUDED.prize_yuan,
			profit_yuan=EXCLUDED.profit_yuan,
			prize_floating=EXCLUDED.prize_floating,
			weight_yuan=EXCLUDED.weight_yuan,
			weight_score=EXCLUDED.weight_score,
			score_version=EXCLUDED.score_version`,
		r.PredictionID, r.DrawResultID, r.MatchedNumbers, r.HitCount, r.HitRate, r.Level, r.IsWin,
		r.StakeYuan, r.PrizeYuan, r.ProfitYuan, r.PrizeFloating, r.WeightYuan, r.WeightScore, r.ScoreVersion)
	return err
}

func (s *Store) UnevaluatedPredictions(ctx context.Context, lotteryCode string, minScoreVersion int) ([]model.Prediction, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT p.id, p.lottery_code, p.issue, p.predict_date, p.model_code, p.predicted_numbers, p.confidence,
			COALESCE(p.reason,''), COALESCE(p.raw_response,''), p.final_flag, p.success, COALESCE(p.error_message,''), p.created_at
		FROM predictions p
		JOIN draw_results d ON d.lottery_code=p.lottery_code AND d.issue=p.issue
		LEFT JOIN prediction_results pr ON pr.prediction_id=p.id
		WHERE p.lottery_code=$1 AND p.success=TRUE AND (pr.id IS NULL OR COALESCE(pr.score_version,0) < $2)
		ORDER BY p.id ASC LIMIT 500`, lotteryCode, minScoreVersion)
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

func (s *Store) UpdateModelWeight(ctx context.Context, modelCode string, weight float64) error {
	_, err := s.pool.Exec(ctx, `UPDATE model_configs SET weight=$2, updated_at=NOW() WHERE model_code=$1`, modelCode, weight)
	return err
}

func (s *Store) InsertStrategy(ctx context.Context, lotteryCode, modelCode string, weight, hitRate float64, notes string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO model_strategies (lottery_code, model_code, weight, last_30_hit_rate, notes)
		VALUES ($1,$2,$3,$4,$5)`, lotteryCode, modelCode, weight, hitRate, notes)
	return err
}

func (s *Store) LatestStrategies(ctx context.Context, lotteryCode string) ([]model.ModelStrategy, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ON (model_code) id, lottery_code, model_code, weight, last_30_hit_rate, COALESCE(notes,''), created_at
		FROM model_strategies
		WHERE lottery_code=$1
		ORDER BY model_code, created_at DESC`, lotteryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ModelStrategy
	for rows.Next() {
		var st model.ModelStrategy
		if err := rows.Scan(&st.ID, &st.LotteryCode, &st.ModelCode, &st.Weight, &st.Last30HitRate, &st.Notes, &st.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *Store) RecentFinalHits(ctx context.Context, lotteryCode string, limit int) ([]model.PredictionResult, error) {
	if limit < 1 {
		limit = 5
	}
	rows, err := s.pool.Query(ctx, `
		SELECT pr.id, pr.prediction_id, pr.draw_result_id, pr.matched_numbers, pr.hit_count, pr.hit_rate, COALESCE(pr.level,''), pr.is_win,
			COALESCE(pr.stake_yuan,0), COALESCE(pr.prize_yuan,0), COALESCE(pr.profit_yuan,0), COALESCE(pr.prize_floating,FALSE)
		FROM prediction_results pr
		JOIN predictions p ON p.id=pr.prediction_id
		WHERE p.lottery_code=$1 AND p.final_flag=TRUE
		ORDER BY pr.id DESC LIMIT $2`, lotteryCode, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.PredictionResult
	for rows.Next() {
		var r model.PredictionResult
		if err := rows.Scan(&r.ID, &r.PredictionID, &r.DrawResultID, &r.MatchedNumbers, &r.HitCount, &r.HitRate, &r.Level, &r.IsWin,
			&r.StakeYuan, &r.PrizeYuan, &r.ProfitYuan, &r.PrizeFloating); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) InsertNotification(ctx context.Context, n model.AppNotification) (int64, error) {
	payload := n.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO app_notifications (type, title, body, payload)
		VALUES ($1,$2,$3,$4) RETURNING id`, n.Type, n.Title, n.Body, payload).Scan(&id)
	return id, err
}

func (s *Store) ListNotifications(ctx context.Context, page, pageSize int) ([]model.AppNotification, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 30
	}
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM app_notifications`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, type, title, body, payload, read, created_at
		FROM app_notifications ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.AppNotification
	for rows.Next() {
		var n model.AppNotification
		if err := rows.Scan(&n.ID, &n.Type, &n.Title, &n.Body, &n.Payload, &n.Read, &n.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, n)
	}
	return out, total, rows.Err()
}

func (s *Store) UnreadCount(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM app_notifications WHERE read=FALSE`).Scan(&n)
	return n, err
}

func (s *Store) MarkNotificationRead(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `UPDATE app_notifications SET read=TRUE WHERE id=$1`, id)
	return err
}

func (s *Store) MarkAllNotificationsRead(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `UPDATE app_notifications SET read=TRUE WHERE read=FALSE`)
	return err
}

func (s *Store) SetNotificationsRead(ctx context.Context, ids []int64, read bool) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.pool.Exec(ctx, `UPDATE app_notifications SET read=$1 WHERE id = ANY($2)`, read, ids)
	return err
}

func (s *Store) UpsertPushDevice(ctx context.Context, token, platform string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO push_devices (token, platform, updated_at)
		VALUES ($1,$2,NOW())
		ON CONFLICT (token) DO UPDATE SET platform=EXCLUDED.platform, updated_at=NOW()`, token, platform)
	return err
}

func (s *Store) ListPushTokens(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT token FROM push_devices ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string) (*model.AppUser, error) {
	var u model.AppUser
	err := s.pool.QueryRow(ctx, `
		INSERT INTO app_users (username, password_hash)
		VALUES ($1,$2)
		RETURNING id, username, password_hash, created_at`, username, passwordHash).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*model.AppUser, error) {
	var u model.AppUser
	err := s.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, created_at FROM app_users WHERE username=$1`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func dateParam(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	if t.IsZero() {
		return time.Now().In(loc).Format("2006-01-02")
	}
	if t.Location() == time.UTC && t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return t.Format("2006-01-02")
	}
	return t.In(loc).Format("2006-01-02")
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
