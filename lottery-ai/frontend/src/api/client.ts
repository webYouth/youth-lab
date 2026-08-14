import AsyncStorage from '@react-native-async-storage/async-storage';
import axios from 'axios';
import type { AccuracyStat, ApiResp, DrawResult, LotteryType, Prediction } from '../types';

const DEFAULT_BASE = 'http://127.0.0.1:8090';

export async function getBaseURL() {
  return (await AsyncStorage.getItem('server_url')) || DEFAULT_BASE;
}

export async function getToken() {
  return (await AsyncStorage.getItem('api_token')) || '';
}

async function client() {
  const baseURL = await getBaseURL();
  const token = await getToken();
  return axios.create({
    baseURL,
    timeout: 20000,
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

export async function fetchLotteryTypes() {
  const c = await client();
  const { data } = await c.get<ApiResp<LotteryType[]>>('/api/v1/lottery-types');
  if (data.code !== 0) throw new Error(data.message);
  return data.data;
}

export async function fetchToday(lotteryCode: string) {
  const c = await client();
  const { data } = await c.get<ApiResp<{ final: Prediction | null; models: Prediction[] }>>(
    `/api/v1/predictions/today`,
    { params: { lottery_code: lotteryCode } }
  );
  if (data.code !== 0) throw new Error(data.message);
  return data.data;
}

export async function fetchDraws(lotteryCode: string, page = 1) {
  const c = await client();
  const { data } = await c.get<ApiResp<{ list: DrawResult[]; total: number }>>('/api/v1/draw-results', {
    params: { lottery_code: lotteryCode, page, pageSize: 20 },
  });
  if (data.code !== 0) throw new Error(data.message);
  return data.data;
}

export async function fetchAccuracy(lotteryCode: string) {
  const c = await client();
  const { data } = await c.get<ApiResp<{ list: AccuracyStat[] }>>('/api/v1/accuracy', {
    params: { lottery_code: lotteryCode, days: 30 },
  });
  if (data.code !== 0) throw new Error(data.message);
  return data.data;
}
