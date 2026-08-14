const path = require('path');
require('dotenv').config({ path: path.join(__dirname, '.env') });

const apiBase = (
  process.env.LOTTERY_API_BASE ||
  process.env.EXPO_PUBLIC_LOTTERY_API_BASE ||
  ''
).replace(/\/$/, '');

/** @type {import('expo/config').ExpoConfig} */
module.exports = {
  name: '彩票 AI',
  slug: 'lottery-ai',
  version: '1.0.4',
  orientation: 'portrait',
  userInterfaceStyle: 'automatic',
  icon: './assets/icon.png',
  splash: {
    image: './assets/icon.png',
    resizeMode: 'contain',
    backgroundColor: '#1C1B1F',
  },
  ios: {
    supportsTablet: true,
    bundleIdentifier: 'top.webyouth.lotteryai',
  },
  android: {
    package: 'top.webyouth.lotteryai',
    adaptiveIcon: {
      foregroundImage: './assets/adaptive-icon.png',
      backgroundColor: '#6750A4',
    },
    usesCleartextTraffic: true,
    permissions: [
      'android.permission.POST_NOTIFICATIONS',
      'android.permission.RECEIVE_BOOT_COMPLETED',
      'android.permission.VIBRATE',
    ],
  },
  plugins: [
    [
      'expo-notifications',
      {
        color: '#B3261E',
        defaultChannel: 'default',
      },
    ],
  ],
  extra: {
    lotteryApiBase: apiBase,
  },
};
