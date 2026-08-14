import { MD3DarkTheme, MD3LightTheme, type MD3Theme } from 'react-native-paper';

const seed = {
  primary: '#B3261E',
  secondary: '#1565C0',
};

export const lightTheme: MD3Theme = {
  ...MD3LightTheme,
  colors: {
    ...MD3LightTheme.colors,
    primary: seed.primary,
    onPrimary: '#FFFFFF',
    primaryContainer: '#FFDAD6',
    onPrimaryContainer: '#410002',
    secondary: seed.secondary,
    onSecondary: '#FFFFFF',
    secondaryContainer: '#D6E3FF',
    onSecondaryContainer: '#001B3D',
    tertiary: '#7C5800',
    background: '#FFFBFF',
    surface: '#FFFBFF',
    surfaceVariant: '#F5DDDA',
    onSurface: '#201A19',
    onSurfaceVariant: '#534341',
    outline: '#857370',
    elevation: {
      ...MD3LightTheme.colors.elevation,
      level1: '#FCEEEE',
      level2: '#F8E8E6',
      level3: '#F4E2E0',
    },
  },
};

export const darkTheme: MD3Theme = {
  ...MD3DarkTheme,
  colors: {
    ...MD3DarkTheme.colors,
    primary: '#FFB4AB',
    onPrimary: '#690005',
    primaryContainer: '#93000A',
    onPrimaryContainer: '#FFDAD6',
    secondary: '#A8C8FF',
    onSecondary: '#003063',
    secondaryContainer: '#00468C',
    onSecondaryContainer: '#D6E3FF',
    background: '#1C1B1F',
    surface: '#1C1B1F',
    surfaceVariant: '#534341',
    onSurface: '#EDE0DE',
    onSurfaceVariant: '#D8C2BE',
    outline: '#A08C89',
    elevation: {
      ...MD3DarkTheme.colors.elevation,
      level1: '#251F23',
      level2: '#2B2328',
      level3: '#32282E',
    },
  },
};

export const LOTTERY_LABELS: Record<string, string> = {
  DLT: '大乐透',
  P3: '排列三',
  KL8: '快乐8',
};
