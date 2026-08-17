import { MD3DarkTheme, MD3LightTheme, type MD3Theme } from 'react-native-paper';

/** 知乎主题蓝 */
const seed = {
  primary: '#0084FF',
  secondary: '#056DE8',
};

export const lightTheme: MD3Theme = {
  ...MD3LightTheme,
  colors: {
    ...MD3LightTheme.colors,
    primary: seed.primary,
    onPrimary: '#FFFFFF',
    primaryContainer: '#D6E9FF',
    onPrimaryContainer: '#001D36',
    secondary: seed.secondary,
    onSecondary: '#FFFFFF',
    secondaryContainer: '#D3E4FF',
    onSecondaryContainer: '#001C38',
    tertiary: '#00639A',
    background: '#F6F6F6',
    surface: '#FFFFFF',
    surfaceVariant: '#E7EEF7',
    onSurface: '#1A1A1A',
    onSurfaceVariant: '#3C4A5A',
    outline: '#8A96A3',
    elevation: {
      ...MD3LightTheme.colors.elevation,
      level1: '#FFFFFF',
      level2: '#F3F7FC',
      level3: '#EAF2FB',
    },
  },
};

export const darkTheme: MD3Theme = {
  ...MD3DarkTheme,
  colors: {
    ...MD3DarkTheme.colors,
    primary: '#8AB4FF',
    onPrimary: '#003258',
    primaryContainer: '#00497D',
    onPrimaryContainer: '#D6E9FF',
    secondary: '#A8C8FF',
    onSecondary: '#003063',
    secondaryContainer: '#004A77',
    onSecondaryContainer: '#D3E4FF',
    tertiary: '#8DCDFF',
    background: '#121212',
    surface: '#1E1E1E',
    surfaceVariant: '#2A3440',
    onSurface: '#E8EAED',
    onSurfaceVariant: '#C2CBD6',
    outline: '#8A96A3',
    elevation: {
      ...MD3DarkTheme.colors.elevation,
      level1: '#1E1E1E',
      level2: '#252525',
      level3: '#2C2C2C',
    },
  },
};

export const LOTTERY_LABELS: Record<string, string> = {
  DLT: '大乐透',
  P3: '排列三',
  KL8: '快乐8',
};

export const LOTTERY_DRAW_HINT: Record<string, string> = {
  DLT: '周一/三/六 约 21:25',
  P3: '每天 21:30',
  KL8: '每天 21:15',
};
