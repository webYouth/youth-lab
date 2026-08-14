import React, { useEffect, useState } from 'react';
import { HelperText, TextInput, Button, Snackbar } from 'react-native-paper';
import AsyncStorage from '@react-native-async-storage/async-storage';
import DisclaimerBanner from '../components/DisclaimerBanner';
import Screen from '../components/Screen';
import { DEFAULT_BASE } from '../api/client';

export default function SettingsScreen() {
  const [server, setServer] = useState(DEFAULT_BASE);
  const [token, setToken] = useState('');
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    (async () => {
      setServer((await AsyncStorage.getItem('server_url')) || DEFAULT_BASE);
      setToken((await AsyncStorage.getItem('api_token')) || '');
    })();
  }, []);

  const save = async () => {
    await AsyncStorage.setItem('server_url', server.trim());
    await AsyncStorage.setItem('api_token', token.trim());
    setSaved(true);
  };

  return (
    <Screen title="设置" scroll>
      <DisclaimerBanner />
      <TextInput
        label="服务器地址"
        mode="outlined"
        value={server}
        onChangeText={setServer}
        autoCapitalize="none"
        autoCorrect={false}
      />
      <HelperText type="info">
        默认连 ECS 的 8090 端口。阿里云安全组请添加「自定义 TCP / 8090 / 0.0.0.0/0」，不要只开 HTTP 80。
      </HelperText>
      <TextInput
        label="API Token（可选）"
        mode="outlined"
        value={token}
        onChangeText={setToken}
        autoCapitalize="none"
        autoCorrect={false}
      />
      <Button mode="contained" onPress={save}>
        保存
      </Button>
      <Snackbar visible={saved} onDismiss={() => setSaved(false)} duration={2000}>
        已保存
      </Snackbar>
    </Screen>
  );
}
