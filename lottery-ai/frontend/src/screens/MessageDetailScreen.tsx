import React from 'react';
import { StyleSheet, View } from 'react-native';
import { useNavigation, useRoute } from '@react-navigation/native';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, Text } from 'react-native-paper';
import { markNotificationRead } from '../api/client';
import type { AppNotification } from '../types';
import Screen from '../components/Screen';

const TYPE_LABEL: Record<string, string> = {
  kl8: '快乐8',
  predict: '预测',
  evaluate: '评估',
};

function formatTime(iso: string) {
  try {
    const d = new Date(iso);
    return d.toLocaleString('zh-CN', { hour12: false });
  } catch {
    return iso;
  }
}

export default function MessageDetailScreen() {
  const nav = useNavigation<any>();
  const route = useRoute<any>();
  const qc = useQueryClient();
  const n = (route.params?.notification || {}) as AppNotification;

  const mark = useMutation({
    mutationFn: () => markNotificationRead(n.id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notifications'] });
      qc.invalidateQueries({ queryKey: ['notifications-unread'] });
    },
  });

  React.useEffect(() => {
    if (n?.id && !n.read) {
      mark.mutate();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [n?.id]);

  if (!n?.id) {
    return (
      <Screen title="消息详情" onBack={() => nav.goBack()}>
        <Text>消息不存在</Text>
      </Screen>
    );
  }

  return (
    <Screen title="消息详情" subtitle={TYPE_LABEL[n.type] || n.type} onBack={() => nav.goBack()}>
      <Card mode="elevated">
        <Card.Title title={n.title} subtitle={formatTime(n.created_at)} />
        <Card.Content>
          <Text variant="bodyLarge" style={styles.body}>
            {n.body}
          </Text>
          {n.payload ? (
            <View style={styles.payload}>
              {n.payload.period || n.payload.issue ? (
                <Text variant="bodyMedium">期号 {String(n.payload.period || n.payload.issue)}</Text>
              ) : null}
              {n.payload.numbers ? (
                <Text variant="bodyMedium">开奖号 {String(n.payload.numbers)}</Text>
              ) : null}
              {n.payload.stake != null ? (
                <Text variant="bodyMedium">
                  本期投入 {String(n.payload.stake)} 元 · 奖金 {String(n.payload.prize ?? n.payload.total ?? 0)} 元 · 盈亏 {String(n.payload.profit)} 元
                </Text>
              ) : n.payload.total != null ? (
                <Text variant="bodyMedium">奖金合计 {String(n.payload.total)} 元</Text>
              ) : null}
              {n.payload.chase_days != null ? (
                <Text variant="bodyMedium">
                  追号 {String(n.payload.chase_days)} 期累计投入 {String(n.payload.chase_stake)} 元 · 奖金 {String(n.payload.chase_prize)} 元 · 盈亏 {String(n.payload.chase_profit)} 元
                </Text>
              ) : null}
              {n.payload.lottery_code ? (
                <Text variant="bodyMedium">彩种 {String(n.payload.lottery_code)}</Text>
              ) : null}
              {n.payload.report ? (
                <Text variant="bodySmall" style={styles.report}>
                  {String(n.payload.report)}
                </Text>
              ) : null}
            </View>
          ) : null}
        </Card.Content>
      </Card>
    </Screen>
  );
}

const styles = StyleSheet.create({
  body: { lineHeight: 24 },
  payload: { marginTop: 16, gap: 8 },
  report: { marginTop: 8, opacity: 0.8, lineHeight: 20 },
});
