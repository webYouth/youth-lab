import React, { useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { fetchDraws } from '../api/client';
import DisclaimerBanner from '../components/DisclaimerBanner';
import NumberBalls from '../components/NumberBalls';

export default function HistoryScreen() {
  const [code, setCode] = useState('DLT');
  const q = useQuery({ queryKey: ['draws', code], queryFn: () => fetchDraws(code, 1) });

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.title}>历史开奖</Text>
      <DisclaimerBanner />
      <View style={styles.row}>
        {['DLT', 'P3', 'KL8'].map((c) => (
          <Pressable key={c} onPress={() => setCode(c)} style={[styles.chip, code === c && styles.chipActive]}>
            <Text style={styles.chipText}>{c}</Text>
          </Pressable>
        ))}
      </View>
      {q.isLoading ? <ActivityIndicator /> : null}
      {(q.data?.list || []).map((d) => {
        const nums = d.result?.numbers || d.result?.digits || d.result?.front || [];
        const back = d.result?.back || [];
        return (
          <View key={d.id} style={styles.card}>
            <Text style={styles.meta}>{d.issue} · {String(d.draw_date).slice(0, 10)}</Text>
            <NumberBalls numbers={nums} />
            {back.length ? <NumberBalls numbers={back} color="#dc2626" /> : null}
          </View>
        );
      })}
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
  card: { backgroundColor: '#111827', borderRadius: 12, padding: 12, gap: 8 },
  meta: { color: '#94a3b8' },
});
