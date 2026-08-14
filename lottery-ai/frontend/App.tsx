// 彩票 AI 预测 App 入口。预测结果仅供个人学习，不构成投注建议。
import React from 'react';
import { NavigationContainer, DarkTheme as NavDark, DefaultTheme as NavLight } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { StatusBar } from 'expo-status-bar';
import { useColorScheme } from 'react-native';
import { PaperProvider } from 'react-native-paper';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import MaterialCommunityIcons from '@expo/vector-icons/MaterialCommunityIcons';
import HomeScreen from './src/screens/HomeScreen';
import HistoryScreen from './src/screens/HistoryScreen';
import AccuracyScreen from './src/screens/AccuracyScreen';
import SettingsScreen from './src/screens/SettingsScreen';
import PredictDetailScreen from './src/screens/PredictDetailScreen';
import { darkTheme, lightTheme } from './src/theme';

const Tab = createBottomTabNavigator();
const Stack = createNativeStackNavigator();
const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: false } },
});

function Tabs() {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerShown: false,
        animation: 'fade',
        tabBarIcon: ({ color, size }) => {
          const names: Record<string, React.ComponentProps<typeof MaterialCommunityIcons>['name']> = {
            首页: 'home-variant',
            开奖: 'history',
            命中率: 'chart-bar',
            设置: 'cog',
          };
          return <MaterialCommunityIcons name={names[route.name] || 'circle'} color={color} size={size} />;
        },
      })}
    >
      <Tab.Screen name="首页" component={HomeScreen} />
      <Tab.Screen name="开奖" component={HistoryScreen} />
      <Tab.Screen name="命中率" component={AccuracyScreen} />
      <Tab.Screen name="设置" component={SettingsScreen} />
    </Tab.Navigator>
  );
}

export default function App() {
  const scheme = useColorScheme();
  const theme = scheme === 'dark' ? darkTheme : lightTheme;
  const navTheme = {
    ...(scheme === 'dark' ? NavDark : NavLight),
    colors: {
      ...(scheme === 'dark' ? NavDark.colors : NavLight.colors),
      primary: theme.colors.primary,
      background: theme.colors.background,
      card: theme.colors.elevation.level2,
      text: theme.colors.onSurface,
      border: theme.colors.outline,
      notification: theme.colors.primary,
    },
  };

  return (
    <SafeAreaProvider>
      <PaperProvider
        theme={theme}
        settings={{
          icon: (props) => <MaterialCommunityIcons {...props} />,
        }}
      >
        <QueryClientProvider client={queryClient}>
          <NavigationContainer theme={navTheme}>
            <StatusBar style={scheme === 'dark' ? 'light' : 'dark'} />
            <Stack.Navigator
              screenOptions={{
                animation: 'slide_from_right',
                animationDuration: 280,
                headerShown: false,
                gestureEnabled: true,
              }}
            >
              <Stack.Screen name="Main" component={Tabs} />
              <Stack.Screen name="PredictDetail" component={PredictDetailScreen} />
            </Stack.Navigator>
          </NavigationContainer>
        </QueryClientProvider>
      </PaperProvider>
    </SafeAreaProvider>
  );
}
