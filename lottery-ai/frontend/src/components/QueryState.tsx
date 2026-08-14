import React from 'react';
import { StyleSheet, View } from 'react-native';
import { ActivityIndicator, Button, Text } from 'react-native-paper';

type Props = {
  loading?: boolean;
  error?: unknown;
  onRetry?: () => void;
};

export default function QueryState({ loading, error, onRetry }: Props) {
  if (loading) {
    return (
      <View style={styles.box}>
        <ActivityIndicator animating />
        <Text variant="bodyMedium" style={styles.hint}>
          正在加载…
        </Text>
      </View>
    );
  }
  if (!error) return null;
  const msg = error instanceof Error ? error.message : '加载失败';
  return (
    <View style={styles.box}>
      <Text variant="titleMedium">无法连接服务器</Text>
      <Text variant="bodyMedium" style={styles.hint}>
        {msg}
      </Text>
      <Text variant="bodySmall" style={styles.hint}>
        请确认已登录，且服务器地址（LOTTERY_API_BASE）配置正确。
      </Text>
      {onRetry ? (
        <Button mode="contained" onPress={onRetry} style={styles.btn}>
          重试
        </Button>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  box: { paddingVertical: 24, alignItems: 'center', gap: 8 },
  hint: { textAlign: 'center', opacity: 0.75, paddingHorizontal: 16 },
  btn: { marginTop: 8 },
});
