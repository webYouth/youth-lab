export type ApiResp<T> = {
  code: number;
  message: string;
  data: T;
  disclaimer?: string;
};

export type LotteryType = {
  code: string;
  name: string;
};

export type Prediction = {
  id: number;
  lottery_code: string;
  issue: string;
  model_code: string;
  predicted_numbers: any;
  confidence: number;
  reason?: string;
  final_flag: boolean;
  success: boolean;
  error_message?: string;
};

export type DrawResult = {
  id: number;
  lottery_code: string;
  issue: string;
  draw_date: string;
  result: any;
};

export type AccuracyStat = {
  lottery_code: string;
  model_code: string;
  total_predictions: number;
  avg_hit_rate: number;
  last_30_hit_rate: number;
};

export const DISCLAIMER = '预测结果仅供个人学习，不构成投注建议';
