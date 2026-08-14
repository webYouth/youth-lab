import AsyncStorage from '@react-native-async-storage/async-storage';
import axios, { AxiosError } from 'axios';
import { getConfiguredApiBase } from '../config';
import type { AccuracyStat, ApiResp, AppNotification, DrawResult, LotteryType, ModelStrategy, Prediction } from '../types';

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'auth_user';

export async function getBaseURL() {
  const base = getConfiguredApiBase();
  if (!base) {
    throw new Error('未配置 LOTTERY_API_BASE。请在 frontend/.env 或 GitHub Secret 中设置公网 API 地址。');
  }
  return base;
}

export async function getToken() {
  return (await AsyncStorage.getItem(TOKEN_KEY)) || '';
}

export async function setSession(token: string, user: { id: number; username: string }) {
  await AsyncStorage.setItem(TOKEN_KEY, token);
  await AsyncStorage.setItem(USER_KEY, JSON.stringify(user));
}

export async function clearSession() {
  await AsyncStorage.multiRemove([TOKEN_KEY, USER_KEY]);
}

export async function getStoredUser(): Promise<{ id: number; username: string } | null> {
  const raw = await AsyncStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function explain(err: unknown): never {
  const ax = err as AxiosError<ApiResp<unknown>>;
  if (ax?.response?.status === 401) {
    throw new Error('登录已失效，请重新登录');
  }
  const msg = ax?.response?.data?.message;
  if (typeof msg === 'string' && msg) {
    throw new Error(msg);
  }
  if (ax?.code === 'ECONNABORTED') {
    throw new Error('连接超时。请确认服务器在线且端口已放行。');
  }
  if (ax?.message?.includes('Network Error') || ax?.code === 'ERR_NETWORK') {
    throw new Error('网络不可达。请确认服务器在线且端口已放行。');
  }
  throw err instanceof Error ? err : new Error('请求失败');
}

async function client(timeout = 8000) {
  const baseURL = await getBaseURL();
  const token = await getToken();
  return axios.create({
    baseURL,
    timeout,
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
}

export async function register(username: string, password: string) {
  try {
    const c = await client();
    const { data } = await c.post<ApiResp<{ token: string; user: { id: number; username: string } }>>(
      '/api/v1/auth/register',
      { username, password }
    );
    if (data.code !== 0) throw new Error(data.message);
    await setSession(data.data.token, data.data.user);
    return data.data;
  } catch (e) {
    return explain(e);
  }
}

export async function login(username: string, password: string) {
  try {
    const c = await client();
    const { data } = await c.post<ApiResp<{ token: string; user: { id: number; username: string } }>>(
      '/api/v1/auth/login',
      { username, password }
    );
    if (data.code !== 0) throw new Error(data.message);
    await setSession(data.data.token, data.data.user);
    return data.data;
  } catch (e) {
    return explain(e);
  }
}

export async function fetchMe() {
  try {
    const c = await client();
    const { data } = await c.get<ApiResp<{ id: number; username: string }>>('/api/v1/auth/me');
    if (data.code !== 0) throw new Error(data.message);
    return data.data;
  } catch (e) {
    return explain(e);
  }
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

export async function fetchAccuracy(lotteryCode: string): Promise<{ list: AccuracyStat[]; strategies?: ModelStrategy[] }> {
  try {
    const c = await client();
    const { data } = await c.get<ApiResp<{ list: AccuracyStat[]; strategies?: ModelStrategy[] }>>('/api/v1/accuracy', {
      params: { lottery_code: lotteryCode, days: 30 },
    });
    if (data.code !== 0) throw new Error(data.message);
    return data.data;
  } catch (e) {
    return explain(e);
  }
}

export async function runPredict(lotteryCode: string): Promise<{
  final: Prediction | null;
  models: Prediction[];
  strategies?: ModelStrategy[];
}> {
  try {
    const c = await client(180000);
    const { data } = await c.post<ApiResp<{ final: Prediction | null; models: Prediction[]; strategies?: ModelStrategy[] }>>(
      '/api/v1/predictions/run',
      null,
      { params: { lottery_code: lotteryCode } }
    );
    if (data.code !== 0) throw new Error(data.message);
    return data.data;
  } catch (e) {
    return explain(e);
  }
}

export async function fetchNotifications(page = 1): Promise<{ list: AppNotification[]; total: number; unread: number }> {
  try {
    const c = await client();
    const { data } = await c.get<ApiResp<{ list: AppNotification[]; total: number; unread: number }>>('/api/v1/notifications', {
      params: { page, pageSize: 40 },
    });
    if (data.code !== 0) throw new Error(data.message);
    return data.data;
  } catch (e) {
    return explain(e);
  }
}

export async function fetchUnreadCount(): Promise<number> {
  try {
    const c = await client();
    const { data } = await c.get<ApiResp<{ unread: number }>>('/api/v1/notifications/unread');
    if (data.code !== 0) throw new Error(data.message);
    return data.data.unread;
  } catch (e) {
    return explain(e);
  }
}

export async function markNotificationRead(id: number): Promise<void> {
  try {
    const c = await client();
    const { data } = await c.post<ApiResp<{ ok: boolean }>>(`/api/v1/notifications/${id}/read`);
    if (data.code !== 0) throw new Error(data.message);
  } catch (e) {
    return explain(e);
  }
}

export async function markNotificationUnread(id: number): Promise<void> {
  try {
    const c = await client();
    const { data } = await c.post<ApiResp<{ ok: boolean }>>(`/api/v1/notifications/${id}/unread`);
    if (data.code !== 0) throw new Error(data.message);
  } catch (e) {
    return explain(e);
  }
}

export async function markAllNotificationsRead(): Promise<void> {
  try {
    const c = await client();
    const { data } = await c.post<ApiResp<{ ok: boolean }>>('/api/v1/notifications/read-all');
    if (data.code !== 0) throw new Error(data.message);
  } catch (e) {
    return explain(e);
  }
}

export async function batchSetNotificationsRead(ids: number[], read: boolean): Promise<void> {
  try {
    const c = await client();
    const { data } = await c.post<ApiResp<{ ok: boolean }>>('/api/v1/notifications/batch-read', { ids, read });
    if (data.code !== 0) throw new Error(data.message);
  } catch (e) {
    return explain(e);
  }
}

export async function registerDevice(token: string, platform: string): Promise<void> {
  try {
    const c = await client();
    const { data } = await c.post<ApiResp<{ ok: boolean }>>('/api/v1/devices', { token, platform });
    if (data.code !== 0) throw new Error(data.message);
  } catch (e) {
    return explain(e);
  }
}
