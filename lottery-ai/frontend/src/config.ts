import Constants from 'expo-constants';

/** 打包时由 LOTTERY_API_BASE / Secret 注入，见 app.config.js 与 .env.example */
export function getConfiguredApiBase(): string {
  const extra = Constants.expoConfig?.extra as { lotteryApiBase?: string } | undefined;
  const fromExtra = (extra?.lotteryApiBase || '').trim().replace(/\/$/, '');
  if (fromExtra) return fromExtra;
  const fromPublic = (process.env.EXPO_PUBLIC_LOTTERY_API_BASE || '').trim().replace(/\/$/, '');
  return fromPublic;
}
