// 彩票 AI 预测 App 入口。
import React, { useCallback, useEffect, useState } from 'react';
import { NavigationContainer, DarkTheme as NavDark, DefaultTheme as NavLight } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { QueryClient, QueryClientProvider, useQueryClient } from '@tanstack/react-query';
import { StatusBar } from 'expo-status-bar';
import { useColorScheme } from 'react-native';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { PaperProvider } from 'react-native-paper';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import MaterialCommunityIcons from '@expo/vector-icons/MaterialCommunityIcons';
import HomeScreen from './src/screens/HomeScreen';
import HistoryScreen from './src/screens/HistoryScreen';
import AccuracyScreen from './src/screens/AccuracyScreen';
import AuthScreen from './src/screens/AuthScreen';
import MessagesScreen from './src/screens/MessagesScreen';
import MessageDetailScreen from './src/screens/MessageDetailScreen';
import PredictDetailScreen from './src/screens/PredictDetailScreen';
import { clearSession, fetchMe, getToken } from './src/api/client';
import { startNotificationWatch } from './src/notifications';
import { darkTheme, lightTheme } from './src/theme';

const Tab = createBottomTabNavigator();
const Stack = createNativeStackNavigator();
const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: false } },
});

function NotificationBootstrap({ enabled }: { enabled: boolean }) {
  const qc = useQueryClient();
  useEffect(() => {
    if (!enabled) return;
    let cleanup: (() => void) | undefined;
    void startNotificationWatch(() => {
      qc.invalidateQueries({ queryKey: ['notifications'] });
      qc.invalidateQueries({ queryKey: ['notifications-unread'] });
    }).then((fn) => {
      cleanup = fn;
    });
    return () => cleanup?.();
  }, [qc, enabled]);
  return null;
}

function Tabs({ onLogout }: { onLogout: () => void }) {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerShown: false,
        tabBarIcon: ({ color, size }) => {
          const names: Record<string, React.ComponentProps<typeof MaterialCommunityIcons>['name']> = {
            首页: 'home-variant',
            开奖: 'history',
            命中率: 'chart-bar',
          };
          return <MaterialCommunityIcons name={names[route.name] || 'circle'} color={color} size={size} />;
        },
      })}
    >
      <Tab.Screen name="首页">
        {() => <HomeScreen onLogout={onLogout} />}
      </Tab.Screen>
      <Tab.Screen name="开奖" component={HistoryScreen} />
      <Tab.Screen name="命中率" component={AccuracyScreen} />
    </Tab.Navigator>
  );
}

export default function App() {
  const scheme = useColorScheme();
  const theme = scheme === 'dark' ? darkTheme : lightTheme;
  const [ready, setReady] = useState(false);
  const [authed, setAuthed] = useState(false);

  const refreshAuth = useCallback(async () => {
    const token = await getToken();
    if (!token) {
      setAuthed(false);
      setReady(true);
      return;
    }
    try {
      await fetchMe();
      setAuthed(true);
    } catch {
      await clearSession();
      setAuthed(false);
    } finally {
      setReady(true);
    }
  }, []);

  useEffect(() => {
    void refreshAuth();
  }, [refreshAuth]);

  const onLogout = async () => {
    await clearSession();
    queryClient.clear();
    setAuthed(false);
  };

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

  if (!ready) {
    return (
      <SafeAreaProvider>
        <PaperProvider theme={theme}>
          <StatusBar style={scheme === 'dark' ? 'light' : 'dark'} />
        </PaperProvider>
      </SafeAreaProvider>
    );
  }

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
    <SafeAreaProvider>
      <PaperProvider
        theme={theme}
        settings={{
          icon: (props) => <MaterialCommunityIcons {...props} />,
        }}
      >
        <QueryClientProvider client={queryClient}>
          <NotificationBootstrap enabled={authed} />
          <NavigationContainer theme={navTheme}>
            <StatusBar style={scheme === 'dark' ? 'light' : 'dark'} />
            {!authed ? (
              <AuthScreen onAuthed={() => setAuthed(true)} />
            ) : (
              <Stack.Navigator
                screenOptions={{
                  animation: 'slide_from_right',
                  animationDuration: 280,
                  headerShown: false,
                  gestureEnabled: true,
                  fullScreenGestureEnabled: true,
                  gestureDirection: 'horizontal',
                }}
              >
                <Stack.Screen name="Main">{() => <Tabs onLogout={onLogout} />}</Stack.Screen>
                <Stack.Screen name="PredictDetail" component={PredictDetailScreen} />
                <Stack.Screen name="Messages" component={MessagesScreen} />
                <Stack.Screen name="MessageDetail" component={MessageDetailScreen} />
              </Stack.Navigator>
            )}
          </NavigationContainer>
        </QueryClientProvider>
      </PaperProvider>
    </SafeAreaProvider>
    </GestureHandlerRootView>
  );
}
