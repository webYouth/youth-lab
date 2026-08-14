import React, { useEffect, useState } from 'react';
import { Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import DisclaimerBanner from '../components/DisclaimerBanner';

export default function SettingsScreen() {
  const [server, setServer] = useState('http://127.0.0.1:8090');
  const [token, setToken] = useState('');
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    (async () => {
      setServer((await AsyncStorage.getItem('server_url')) || 'http://127.0.0.1:8090');
      setToken((await AsyncStorage.getItem('api_token')) || '');
    })();
  }, []);

  const save = async () => {
    await AsyncStorage.setItem('server_url', server.trim());
    await AsyncStorage.setItem('api_token', token.trim());
    setSaved(true);
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>设置</Text>
      <DisclaimerBanner />
      <Text style={styles.label}>服务器地址</Text>
      <TextInput style={styles.input} value={server} onChangeText={setServer} autoCapitalize="none" />
      <Text style={styles.label}>API Token（可选）</Text>
      <TextInput style={styles.input} value={token} onChangeText={setToken} autoCapitalize="none" />
      <Pressable style={styles.btn} onPress={save}>
        <Text style={styles.btnText}>{saved ? '已保存' : '保存'}</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, padding: 16, paddingTop: 56, backgroundColor: '#0f172a', gap: 10 },
  title: { color: '#fff', fontSize: 24, fontWeight: '700' },
  label: { color: '#cbd5e1' },
  input: { backgroundColor: '#1e293b', color: '#fff', borderRadius: 8, padding: 12 },
  btn: { backgroundColor: '#2563eb', padding: 14, borderRadius: 10, alignItems: 'center' },
  btnText: { color: '#fff', fontWeight: '700' },
});
