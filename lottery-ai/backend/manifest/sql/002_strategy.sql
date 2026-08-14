-- 预测去重 + 模型策略快照：开奖评估后按命中率调整权重
CREATE UNIQUE INDEX IF NOT EXISTS uniq_predictions_lottery_issue_model
    ON predictions (lottery_code, issue, model_code);

CREATE TABLE IF NOT EXISTS model_strategies (
    id               BIGSERIAL PRIMARY KEY,
    lottery_code     VARCHAR(16) NOT NULL,
    model_code       VARCHAR(64) NOT NULL,
    weight           DOUBLE PRECISION NOT NULL,
    last_30_hit_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    notes            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_model_strategies_lookup
    ON model_strategies (lottery_code, model_code, created_at DESC);
