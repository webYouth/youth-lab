import React, { useEffect, useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { Button, HelperText, Text, TextInput } from 'react-native-paper';
import { getConfiguredApiBase } from '../config';
import { login, register } from '../api/client';
import Screen from '../components/Screen';

type Props = {
  onAuthed: () => void;
};

export default function AuthScreen({ onAuthed }: Props) {
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const apiBase = getConfiguredApiBase();

  useEffect(() => {
    setError('');
  }, [mode]);

  const submit = async () => {
    setBusy(true);
    setError('');
    try {
      if (mode === 'login') {
        await login(username.trim(), password);
      } else {
        await register(username.trim(), password);
      }
      onAuthed();
    } catch (e) {
      setError(e instanceof Error ? e.message : '失败');
    } finally {
      setBusy(false);
    }
  };

  return (
    <Screen title={mode === 'login' ? '登录' : '注册'} subtitle="彩票 AI">
      {!apiBase ? (
        <HelperText type="error">
          未配置 API 地址。请在 lottery-ai/frontend/.env 设置 LOTTERY_API_BASE（与 GitHub Secret 同名）。
        </HelperText>
      ) : (
        <HelperText type="info">服务器：{apiBase}</HelperText>
      )}
      <TextInput
        label="用户名"
        mode="outlined"
        value={username}
        onChangeText={setUsername}
        autoCapitalize="none"
        autoCorrect={false}
      />
      <TextInput
        label="密码"
        mode="outlined"
        value={password}
        onChangeText={setPassword}
        secureTextEntry
      />
      {error ? (
        <Text variant="bodyMedium" style={styles.err}>
          {error}
        </Text>
      ) : null}
      <Button mode="contained" loading={busy} disabled={busy || !apiBase} onPress={submit}>
        {mode === 'login' ? '登录' : '注册'}
      </Button>
      <View style={styles.switch}>
        <Button
          mode="text"
          onPress={() => setMode(mode === 'login' ? 'register' : 'login')}
          disabled={busy}
        >
          {mode === 'login' ? '没有账号？去注册' : '已有账号？去登录'}
        </Button>
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  err: { color: '#B3261E' },
  switch: { alignItems: 'center' },
});
