-- 开奖日说明：大乐透周一/三/六；排列三、快乐8每天
UPDATE lottery_types SET draw_days = '[1,3,6]'::jsonb WHERE code = 'DLT';
UPDATE lottery_types SET draw_days = '[]'::jsonb WHERE code IN ('P3', 'KL8');
UPDATE lottery_types
SET rules = jsonb_set(COALESCE(rules, '{}'::jsonb), '{draw_time}', '"21:25"'::jsonb)
WHERE code = 'DLT';
UPDATE lottery_types
SET rules = jsonb_set(COALESCE(rules, '{}'::jsonb), '{draw_time}', '"21:30"'::jsonb)
WHERE code = 'P3';
UPDATE lottery_types
SET rules = jsonb_set(COALESCE(rules, '{}'::jsonb), '{draw_time}', '"21:15"'::jsonb)
WHERE code = 'KL8';
