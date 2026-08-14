// 彩票 AI 预测 App 入口。预测结果仅供个人学习，不构成投注建议。
import React from 'react';
import { NavigationContainer, DarkTheme, DefaultTheme } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { StatusBar } from 'expo-status-bar';
import { useColorScheme } from 'react-native';
import HomeScreen from './src/screens/HomeScreen';
import HistoryScreen from './src/screens/HistoryScreen';
import AccuracyScreen from './src/screens/AccuracyScreen';
import SettingsScreen from './src/screens/SettingsScreen';
import PredictDetailScreen from './src/screens/PredictDetailScreen';
import { createNativeStackNavigator } from '@react-navigation/native-stack';

const Tab = createBottomTabNavigator();
const Stack = createNativeStackNavigator();
const queryClient = new QueryClient();

function Tabs() {
  return (
    <Tab.Navigator screenOptions={{ headerShown: false }}>
      <Tab.Screen name="首页" component={HomeScreen} />
      <Tab.Screen name="开奖" component={HistoryScreen} />
      <Tab.Screen name="命中率" component={AccuracyScreen} />
      <Tab.Screen name="设置" component={SettingsScreen} />
    </Tab.Navigator>
  );
}

export default function App() {
  const scheme = useColorScheme();
  return (
    <QueryClientProvider client={queryClient}>
      <NavigationContainer theme={scheme === 'dark' ? DarkTheme : DefaultTheme}>
        <StatusBar style="auto" />
        <Stack.Navigator>
          <Stack.Screen name="Main" component={Tabs} options={{ headerShown: false }} />
          <Stack.Screen name="PredictDetail" component={PredictDetailScreen} options={{ title: '预测详情' }} />
        </Stack.Navigator>
      </NavigationContainer>
    </QueryClientProvider>
  );
}
