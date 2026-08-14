import React from 'react';
import { StyleSheet, View } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useQuery } from '@tanstack/react-query';
import { Appbar, Badge } from 'react-native-paper';
import { fetchUnreadCount } from '../api/client';

/** 右上角消息入口 */
export default function MessageBell() {
  const nav = useNavigation<any>();
  const q = useQuery({
    queryKey: ['notifications-unread'],
    queryFn: fetchUnreadCount,
    refetchInterval: 45000,
  });
  const unread = q.data || 0;

  return (
    <View style={styles.wrap}>
      <Appbar.Action icon="bell-outline" accessibilityLabel="消息" onPress={() => nav.navigate('Messages')} />
      {unread > 0 ? (
        <Badge style={styles.badge} size={16}>
          {unread > 99 ? '99+' : unread}
        </Badge>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { justifyContent: 'center' },
  badge: { position: 'absolute', top: 6, right: 6 },
});
