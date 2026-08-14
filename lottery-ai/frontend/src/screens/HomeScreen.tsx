import React, { useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { useNavigation } from '@react-navigation/native';
import { fetchLotteryTypes, fetchToday } from '../api/client';
import DisclaimerBanner from '../components/DisclaimerBanner';
import NumberBalls from '../components/NumberBalls';

export default function HomeScreen() {
  const nav = useNavigation<any>();
  const [code, setCode] = useState('DLT');
  const typesQ = useQuery({ queryKey: ['types'], queryFn: fetchLotteryTypes });
  const todayQ = useQuery({ queryKey: ['today', code], queryFn: () => fetchToday(code) });

  const finalPred = todayQ.data?.final;
  const nums = finalPred?.predicted_numbers?.numbers || [];
  const back = finalPred?.predicted_numbers?.back_numbers || [];

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.title}>彩票 AI 预测</Text>
      <DisclaimerBanner />
      <View style={styles.row}>
        {(typesQ.data || [{ code: 'DLT', name: '大乐透' }, { code: 'P3', name: '排列三' }, { code: 'KL8', name: '快乐8' }]).map((t) => (
          <Pressable key={t.code} onPress={() => setCode(t.code)} style={[styles.chip, code === t.code && styles.chipActive]}>
            <Text style={styles.chipText}>{t.name || t.code}</Text>
          </Pressable>
        ))}
      </View>
      {todayQ.isLoading ? <ActivityIndicator /> : null}
      {todayQ.isError ? <Text style={styles.err}>加载失败，请下拉重试或检查服务器地址</Text> : null}
      <Pressable style={styles.card} onPress={() => nav.navigate('PredictDetail', { lotteryCode: code })}>
        <Text style={styles.cardTitle}>{code} 今日最终预测</Text>
        <Text style={styles.meta}>置信度：{finalPred?.confidence ?? '-'}</Text>
        <NumberBalls numbers={nums} />
        {back.length > 0 ? (
          <>
            <Text style={styles.meta}>后区</Text>
            <NumberBalls numbers={back} color="#dc2626" />
          </>
        ) : null}
        <Text style={styles.link}>查看各模型详情 →</Text>
      </Pressable>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { padding: 16, paddingTop: 56, backgroundColor: '#0f172a', minHeight: '100%' },
  title: { color: '#fff', fontSize: 24, fontWeight: '700', marginBottom: 12 },
  row: { flexDirection: 'row', gap: 8, marginBottom: 12 },
  chip: { backgroundColor: '#1e293b', paddingHorizontal: 12, paddingVertical: 8, borderRadius: 16 },
  chipActive: { backgroundColor: '#2563eb' },
  chipText: { color: '#fff' },
  card: { backgroundColor: '#111827', borderRadius: 12, padding: 16, gap: 8 },
  cardTitle: { color: '#e2e8f0', fontSize: 18, fontWeight: '600' },
  meta: { color: '#94a3b8' },
  link: { color: '#60a5fa', marginTop: 8 },
  err: { color: '#f87171', marginBottom: 8 },
});
