import AsyncStorage from '@react-native-async-storage/async-storage';
import axios, { AxiosError } from 'axios';
import type { AccuracyStat, ApiResp, DrawResult, LotteryType, Prediction } from '../types';

export const DEFAULT_BASE = 'http://47.106.178.183:8090';

export async function getBaseURL() {
  return (await AsyncStorage.getItem('server_url')) || DEFAULT_BASE;
}

export async function getToken() {
  return (await AsyncStorage.getItem('api_token')) || '';
}

function explain(err: unknown): never {
  const ax = err as AxiosError;
  if (ax?.code === 'ECONNABORTED') {
    throw new Error('连接超时。安全组需放行自定义 TCP 8090。');
  }
  if (ax?.message?.includes('Network Error') || ax?.code === 'ERR_NETWORK') {
    throw new Error('网络不可达。请确认服务器在线，且安全组已放行 8090。');
  }
  throw err instanceof Error ? err : new Error('请求失败');
}

async function client() {
  const baseURL = await getBaseURL();
  const token = await getToken();
  return axios.create({
    baseURL,
    timeout: 8000,
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

export async function fetchLotteryTypes(): Promise<LotteryType[]> {
  try {
    const c = await client();
    const { data } = await c.get<ApiResp<LotteryType[]>>('/api/v1/lottery-types');
    if (data.code !== 0) throw new Error(data.message);
    return data.data;
  } catch (e) {
    return explain(e);
  }
}

export async function fetchToday(lotteryCode: string): Promise<{ final: Prediction | null; models: Prediction[] }> {
  try {
    const c = await client();
    const { data } = await c.get<ApiResp<{ final: Prediction | null; models: Prediction[] }>>(
      `/api/v1/predictions/today`,
      { params: { lottery_code: lotteryCode } }
    );
    if (data.code !== 0) throw new Error(data.message);
    return data.data;
  } catch (e) {
    return explain(e);
  }
}

export async function fetchDraws(lotteryCode: string, page = 1): Promise<{ list: DrawResult[]; total: number }> {
  try {
    const c = await client();
    const { data } = await c.get<ApiResp<{ list: DrawResult[]; total: number }>>('/api/v1/draw-results', {
      params: { lottery_code: lotteryCode, page, pageSize: 20 },
    });
    if (data.code !== 0) throw new Error(data.message);
    return data.data;
  } catch (e) {
    return explain(e);
  }
}

export async function fetchAccuracy(lotteryCode: string): Promise<{ list: AccuracyStat[] }> {
  try {
    const c = await client();
    const { data } = await c.get<ApiResp<{ list: AccuracyStat[] }>>('/api/v1/accuracy', {
      params: { lottery_code: lotteryCode, days: 30 },
    });
    if (data.code !== 0) throw new Error(data.message);
    return data.data;
  } catch (e) {
    return explain(e);
  }
}
