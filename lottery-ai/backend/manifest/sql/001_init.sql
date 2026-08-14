-- 彩票 AI 预测系统初始化 SQL
-- 免责声明：预测结果仅供个人学习研究，不构成投注建议。

CREATE TABLE IF NOT EXISTS lottery_types (
    code        VARCHAR(16) PRIMARY KEY,
    name        VARCHAR(64) NOT NULL,
    rules       JSONB NOT NULL DEFAULT '{}',
    draw_days   JSONB NOT NULL DEFAULT '[]', -- 0=周日 ... 6=周六；空数组表示每天
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS draw_results (
    id           BIGSERIAL PRIMARY KEY,
    lottery_code VARCHAR(16) NOT NULL REFERENCES lottery_types(code),
    issue        VARCHAR(32) NOT NULL,
    draw_date    DATE NOT NULL,
    result       JSONB NOT NULL,
    raw_data     JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (lottery_code, issue)
);

CREATE INDEX IF NOT EXISTS idx_draw_results_lottery_date
    ON draw_results (lottery_code, draw_date DESC);

CREATE TABLE IF NOT EXISTS model_configs (
    id            BIGSERIAL PRIMARY KEY,
    model_code    VARCHAR(64) NOT NULL UNIQUE,
    display_name  VARCHAR(128) NOT NULL,
    provider      VARCHAR(64) NOT NULL,
    base_url      VARCHAR(512) NOT NULL,
    model_name    VARCHAR(128) NOT NULL,
    api_key_env   VARCHAR(128) NOT NULL, -- 仅存环境变量名，不存明文
    weight        DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    enabled       BOOLEAN NOT NULL DEFAULT TRUE,
    timeout_sec   INT NOT NULL DEFAULT 60,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS predictions (
    id                 BIGSERIAL PRIMARY KEY,
    lottery_code       VARCHAR(16) NOT NULL REFERENCES lottery_types(code),
    issue              VARCHAR(32) NOT NULL,
    predict_date       DATE NOT NULL,
    model_code         VARCHAR(64) NOT NULL, -- FINAL 表示聚合结果
    predicted_numbers  JSONB NOT NULL,
    confidence         DOUBLE PRECISION NOT NULL DEFAULT 0,
    reason             TEXT,
    raw_response       TEXT,
    final_flag         BOOLEAN NOT NULL DEFAULT FALSE,
    success            BOOLEAN NOT NULL DEFAULT TRUE,
    error_message      TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_predictions_lottery_date
    ON predictions (lottery_code, predict_date DESC);

CREATE TABLE IF NOT EXISTS prediction_results (
    id              BIGSERIAL PRIMARY KEY,
    prediction_id   BIGINT NOT NULL REFERENCES predictions(id) ON DELETE CASCADE,
    draw_result_id  BIGINT NOT NULL REFERENCES draw_results(id) ON DELETE CASCADE,
    matched_numbers JSONB NOT NULL DEFAULT '[]',
    hit_count       INT NOT NULL DEFAULT 0,
    hit_rate        DOUBLE PRECISION NOT NULL DEFAULT 0,
    level           VARCHAR(32),
    is_win          BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (prediction_id, draw_result_id)
);

CREATE TABLE IF NOT EXISTS prediction_accuracy (
    id                 BIGSERIAL PRIMARY KEY,
    lottery_code       VARCHAR(16) NOT NULL,
    model_code         VARCHAR(64) NOT NULL,
    total_predictions  INT NOT NULL DEFAULT 0,
    total_hits         INT NOT NULL DEFAULT 0,
    avg_hit_rate       DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_30_hit_rate   DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (lottery_code, model_code)
);

CREATE TABLE IF NOT EXISTS sync_raw_logs (
    id           BIGSERIAL PRIMARY KEY,
    lottery_code VARCHAR(16) NOT NULL,
    source_url   TEXT NOT NULL,
    status_code  INT,
    body         TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 初始彩种
INSERT INTO lottery_types (code, name, rules, draw_days) VALUES
('DLT', '超级大乐透', '{"front":{"min":1,"max":35,"count":5},"back":{"min":1,"max":12,"count":2}}', '[1,3,6]'),
('P3',  '排列三',     '{"digits":{"min":0,"max":9,"count":3,"ordered":true}}', '[]'),
('KL8', '快乐8',      '{"numbers":{"min":1,"max":80,"count":10},"play":"选十"}', '[]')
ON CONFLICT (code) DO NOTHING;

-- 五个大模型配置（API Key 仅从环境变量读取）
INSERT INTO model_configs (model_code, display_name, provider, base_url, model_name, api_key_env, weight, enabled) VALUES
('deepseek', 'DeepSeek', 'deepseek', 'https://api.deepseek.com/v1', 'deepseek-chat', 'DEEP_SEEK_API_KEY', 1.0, TRUE),
('qwen',     '通义千问', 'qwen',     'https://dashscope.aliyuncs.com/compatible-mode/v1', 'qwen-plus', 'QWEN_API_KEY', 1.0, TRUE),
('kimi',     'Kimi',     'moonshot', 'https://api.moonshot.cn/v1', 'moonshot-v1-8k', 'KIMI_API_KEY', 1.0, TRUE),
('glm',      '智谱GLM',  'zhipu',    'https://open.bigmodel.cn/api/paas/v4', 'glm-4-flash', 'GLM_API_KEY', 1.0, TRUE),
('minimax',  'MiniMax',  'minimax',  'https://api.minimax.chat/v1', 'MiniMax-Text-01', 'MINIMAX_API_KEY', 1.0, TRUE)
ON CONFLICT (model_code) DO NOTHING;