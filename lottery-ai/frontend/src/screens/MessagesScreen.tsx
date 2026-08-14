import React, { useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Appbar, Button, Card, Checkbox, Text, useTheme } from 'react-native-paper';
import {
  batchSetNotificationsRead,
  fetchNotifications,
  markAllNotificationsRead,
} from '../api/client';
import type { AppNotification } from '../types';
import QueryState from '../components/QueryState';
import Screen from '../components/Screen';

const TYPE_LABEL: Record<string, string> = {
  kl8: '快乐8',
  predict: '预测',
  evaluate: '评估',
};

function formatTime(iso: string) {
  try {
    const d = new Date(iso);
    return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
  } catch {
    return iso;
  }
}

export default function MessagesScreen() {
  const theme = useTheme();
  const nav = useNavigation<any>();
  const qc = useQueryClient();
  const [selected, setSelected] = useState<number[]>([]);
  const q = useQuery({
    queryKey: ['notifications'],
    queryFn: () => fetchNotifications(1),
    refetchInterval: 45000,
  });
  const invalidate = () => {
    qc.invalidateQueries({ queryKey: ['notifications'] });
    qc.invalidateQueries({ queryKey: ['notifications-unread'] });
  };
  const batchM = useMutation({
    mutationFn: ({ ids, read }: { ids: number[]; read: boolean }) => batchSetNotificationsRead(ids, read),
    onSuccess: () => {
      setSelected([]);
      invalidate();
    },
  });
  const markAll = useMutation({
    mutationFn: markAllNotificationsRead,
    onSuccess: () => {
      setSelected([]);
      invalidate();
    },
  });

  const list: AppNotification[] = q.data?.list || [];
  const unread = q.data?.unread || 0;
  const allIds = list.map((n: AppNotification) => n.id);
  const allSelected = allIds.length > 0 && selected.length === allIds.length;

  const toggle = (id: number) => {
    setSelected((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  };

  return (
    <Screen
      title="消息"
      subtitle={unread > 0 ? `${unread} 条未读` : '全部已读'}
      onBack={() => nav.goBack()}
      actions={
        <>
          <Appbar.Action
            icon={allSelected ? 'checkbox-multiple-marked' : 'checkbox-multiple-blank-outline'}
            accessibilityLabel="全选"
            onPress={() => setSelected(allSelected ? [] : allIds)}
          />
          <Appbar.Action
            icon="email-open-outline"
            accessibilityLabel="全部已读"
            onPress={() => markAll.mutate()}
          />
        </>
      }
    >
      <QueryState loading={q.isLoading} error={q.error} onRetry={() => q.refetch()} />
      {selected.length > 0 ? (
        <View style={styles.batchBar}>
          <Text variant="labelLarge">已选 {selected.length}</Text>
          <Button
            mode="contained-tonal"
            compact
            loading={batchM.isPending}
            onPress={() => batchM.mutate({ ids: selected, read: true })}
          >
            标为已读
          </Button>
          <Button
            mode="outlined"
            compact
            loading={batchM.isPending}
            onPress={() => batchM.mutate({ ids: selected, read: false })}
          >
            标为未读
          </Button>
        </View>
      ) : null}
      <View style={styles.list}>
        {(list as AppNotification[]).map((n) => {
          const checked = selected.includes(n.id);
          return (
            <Card
              key={n.id}
              mode={n.read ? 'outlined' : 'elevated'}
              style={!n.read ? { borderLeftWidth: 3, borderLeftColor: theme.colors.primary } : undefined}
              onPress={() => toggle(n.id)}
              onLongPress={() => toggle(n.id)}
            >
              <Card.Title
                title={n.title}
                subtitle={`${TYPE_LABEL[n.type] || n.type} · ${formatTime(n.created_at)}`}
                left={() => (
                  <Checkbox
                    status={checked ? 'checked' : 'unchecked'}
                    onPress={() => toggle(n.id)}
                  />
                )}
                right={() =>
                  n.read ? (
                    <Text variant="labelSmall" style={styles.tag}>
                      已读
                    </Text>
                  ) : (
                    <Text variant="labelSmall" style={[styles.tag, { color: theme.colors.primary }]}>
                      未读
                    </Text>
                  )
                }
              />
              <Card.Content>
                <Text variant="bodyMedium">{n.body}</Text>
                {n.type === 'kl8' && n.payload?.numbers ? (
                  <Text variant="bodySmall" style={styles.meta}>
                    开奖号 {String(n.payload.numbers)}
                  </Text>
                ) : null}
              </Card.Content>
            </Card>
          );
        })}
      </View>
      {!q.isLoading && !q.isError && !list.length ? (
        <Text variant="bodyMedium">暂无消息。快乐8 每晚查奖、定时预测/评估会推送到这里。</Text>
      ) : null}
    </Screen>
  );
}

const styles = StyleSheet.create({
  list: { gap: 12 },
  batchBar: { flexDirection: 'row', alignItems: 'center', gap: 8, flexWrap: 'wrap' },
  tag: { marginRight: 16, fontWeight: '700' },
  meta: { marginTop: 8, opacity: 0.75 },
});
