import React from 'react';
import { ActivityIndicator, ScrollView, StyleSheet, Text, View } from 'react-native';
import { useRoute } from '@react-navigation/native';
import { useQuery } from '@tanstack/react-query';
import { fetchToday } from '../api/client';
import DisclaimerBanner from '../components/DisclaimerBanner';
import NumberBalls from '../components/NumberBalls';

export default function PredictDetailScreen() {
  const route = useRoute<any>();
  const lotteryCode = route.params?.lotteryCode || 'DLT';
  const q = useQuery({ queryKey: ['today', lotteryCode], queryFn: () => fetchToday(lotteryCode) });
  const finalPred = q.data?.final;
  const models = q.data?.models || [];

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <DisclaimerBanner />
      {q.isLoading ? <ActivityIndicator /> : null}
      <Text style={styles.title}>最终推荐</Text>
      <NumberBalls numbers={finalPred?.predicted_numbers?.numbers || []} />
      <Text style={styles.reason}>{finalPred?.reason || '暂无理由'}</Text>
      <Text style={styles.title}>各模型预测</Text>
      {models.map((m) => (
        <View key={m.id} style={styles.card}>
          <Text style={styles.cardTitle}>{m.model_code} · 置信度 {m.confidence}</Text>
          <NumberBalls numbers={m.predicted_numbers?.numbers || []} color="#059669" />
          <Text style={styles.reason}>{m.reason || (m.success ? '' : m.error_message)}</Text>
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { padding: 16, backgroundColor: '#0f172a', minHeight: '100%', gap: 10 },
  title: { color: '#fff', fontSize: 18, fontWeight: '700', marginTop: 8 },
  reason: { color: '#94a3b8', lineHeight: 20 },
  card: { backgroundColor: '#111827', borderRadius: 12, padding: 12, gap: 8 },
  cardTitle: { color: '#e2e8f0', fontWeight: '600' },
});
