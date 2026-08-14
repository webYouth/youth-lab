import React, { useState } from 'react';
import { ActivityIndicator, Dimensions, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { BarChart } from 'react-native-chart-kit';
import { fetchAccuracy } from '../api/client';
import DisclaimerBanner from '../components/DisclaimerBanner';

export default function AccuracyScreen() {
  const [code, setCode] = useState('DLT');
  const q = useQuery({ queryKey: ['acc', code], queryFn: () => fetchAccuracy(code) });
  const list = q.data?.list || [];

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.title}>命中率统计</Text>
      <DisclaimerBanner />
      <View style={styles.row}>
        {['DLT', 'P3', 'KL8'].map((c) => (
          <Pressable key={c} onPress={() => setCode(c)} style={[styles.chip, code === c && styles.chipActive]}>
            <Text style={styles.chipText}>{c}</Text>
          </Pressable>
        ))}
      </View>
      {q.isLoading ? <ActivityIndicator /> : null}
      {list.length > 0 ? (
        <BarChart
          data={{
            labels: list.map((x) => x.model_code.slice(0, 6)),
            datasets: [{ data: list.map((x) => Number((x.avg_hit_rate * 100).toFixed(1)) || 0) }],
          }}
          width={Dimensions.get('window').width - 32}
          height={220}
          yAxisSuffix="%"
          chartConfig={{
            backgroundGradientFrom: '#111827',
            backgroundGradientTo: '#111827',
            color: () => '#60a5fa',
            labelColor: () => '#cbd5e1',
          }}
          style={{ borderRadius: 12 }}
        />
      ) : (
        <Text style={styles.meta}>暂无命中率数据，等待开奖评估后更新</Text>
      )}
      {list.map((a) => (
        <View key={a.model_code} style={styles.card}>
          <Text style={styles.cardTitle}>{a.model_code}</Text>
          <Text style={styles.meta}>总预测 {a.total_predictions} · 平均命中率 {(a.avg_hit_rate * 100).toFixed(1)}%</Text>
          <Text style={styles.meta}>近30天命中率 {(a.last_30_hit_rate * 100).toFixed(1)}%</Text>
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { padding: 16, paddingTop: 56, backgroundColor: '#0f172a', minHeight: '100%', gap: 10 },
  title: { color: '#fff', fontSize: 24, fontWeight: '700' },
  row: { flexDirection: 'row', gap: 8 },
  chip: { backgroundColor: '#1e293b', paddingHorizontal: 12, paddingVertical: 8, borderRadius: 16 },
  chipActive: { backgroundColor: '#2563eb' },
  chipText: { color: '#fff' },
  card: { backgroundColor: '#111827', borderRadius: 12, padding: 12 },
  cardTitle: { color: '#e2e8f0', fontWeight: '600' },
  meta: { color: '#94a3b8' },
});
