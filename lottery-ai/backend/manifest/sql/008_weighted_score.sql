-- 命中率改为按奖级/奖金加权（对数压缩）；score_version 用于历史重算。
ALTER TABLE prediction_results ADD COLUMN IF NOT EXISTS weight_yuan DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE prediction_results ADD COLUMN IF NOT EXISTS weight_score DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE prediction_results ADD COLUMN IF NOT EXISTS score_version INT NOT NULL DEFAULT 0;
