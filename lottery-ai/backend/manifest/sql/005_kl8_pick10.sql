-- 快乐8 预测规则改为选十（10 个号码）
UPDATE lottery_types
SET rules = '{"numbers":{"min":1,"max":80,"count":10},"play":"选十"}'::jsonb
WHERE code = 'KL8';
