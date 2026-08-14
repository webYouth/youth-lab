import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Device from 'expo-device';
import * as Notifications from 'expo-notifications';
import { AppState, Platform } from 'react-native';
import { fetchNotifications, fetchUnreadCount, registerDevice } from './api/client';

const LAST_ID_KEY = 'notify_last_seen_id';

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: true,
    shouldShowBanner: true,
    shouldShowList: true,
  }),
});

export async function ensureNotificationPermission(): Promise<boolean> {
  if (Platform.OS === 'android') {
    await Notifications.setNotificationChannelAsync('default', {
      name: '彩票消息',
      importance: Notifications.AndroidImportance.MAX,
      vibrationPattern: [0, 250, 250, 250],
      lightColor: '#B3261E',
    });
  }
  const current = await Notifications.getPermissionsAsync();
  let status = current.status;
  if (status !== 'granted') {
    const next = await Notifications.requestPermissionsAsync();
    status = next.status;
  }
  return status === 'granted';
}

async function tryRegisterPushToken() {
  if (!Device.isDevice) return;
  try {
    const token = await Notifications.getExpoPushTokenAsync();
    await registerDevice(token.data, Platform.OS);
  } catch {
    // 无 EAS projectId / FCM 时忽略；前台轮询 + 本地通知仍可用
  }
}

async function seedLastSeenIfNeeded() {
  const existing = await AsyncStorage.getItem(LAST_ID_KEY);
  if (existing) return;
  try {
    const { list } = await fetchNotifications(1);
    if (list[0]) {
      await AsyncStorage.setItem(LAST_ID_KEY, String(list[0].id));
    }
  } catch {
    // ignore
  }
}

async function pollOnce(onChange?: () => void) {
  try {
    const [{ list }, unread] = await Promise.all([fetchNotifications(1), fetchUnreadCount()]);
    onChange?.();
    const last = Number((await AsyncStorage.getItem(LAST_ID_KEY)) || '0');
    const newest = list[0];
    if (!newest) return;
    if (newest.id <= last) return;

    const fresh = list.filter((n) => n.id > last);
    if (fresh.length > 0) {
      const top = fresh[0];
      const extra = fresh.length > 1 ? `（另有 ${fresh.length - 1} 条）` : '';
      await Notifications.scheduleNotificationAsync({
        content: {
          title: top.title,
          body: `${top.body || ''}${extra}`.trim() || `未读 ${unread} 条`,
          data: { id: top.id, type: top.type },
          sound: true,
        },
        trigger: null,
      });
    }
    await AsyncStorage.setItem(LAST_ID_KEY, String(newest.id));
  } catch {
    // 轮询失败不影响 App
  }
}

/** 前台轮询消息中心，有新消息时弹出系统通知。返回清理函数。 */
export async function startNotificationWatch(onChange?: () => void): Promise<() => void> {
  const granted = await ensureNotificationPermission();
  if (granted) {
    await tryRegisterPushToken();
  }
  await seedLastSeenIfNeeded();
  await pollOnce(onChange);

  const timer = setInterval(() => {
    void pollOnce(onChange);
  }, 45000);
  const sub = AppState.addEventListener('change', (state) => {
    if (state === 'active') void pollOnce(onChange);
  });

  return () => {
    clearInterval(timer);
    sub.remove();
  };
}
