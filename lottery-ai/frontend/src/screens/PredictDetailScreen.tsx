import React from 'react';
import { useNavigation, useRoute } from '@react-navigation/native';
import { useQuery } from '@tanstack/react-query';
import { Card, Text } from 'react-native-paper';
import { fetchToday } from '../api/client';
import type { Prediction } from '../types';
import NumberBalls from '../components/NumberBalls';
import QueryState from '../components/QueryState';
import Screen, { SectionTitle } from '../components/Screen';

function StatsLine({ pn }: { pn: any }) {
  if (!pn) return null;
  const parts: string[] = [];
  if (pn.sum != null) parts.push(`和值 ${pn.sum}`);
  if (pn.span != null) parts.push(`跨度 ${pn.span}`);
  if (pn.back_sum != null) parts.push(`后区和值 ${pn.back_sum}`);
  if (pn.back_span != null) parts.push(`后区跨度 ${pn.back_span}`);
  if (!parts.length) return null;
  return (
    <Text variant="bodyMedium" style={{ marginTop: 8, opacity: 0.85 }}>
      {parts.join(' · ')}
    </Text>
  );
}

export default function PredictDetailScreen() {
  const nav = useNavigation<any>();
  const route = useRoute<any>();
  const lotteryCode = route.params?.lotteryCode || 'DLT';
  const q = useQuery({
    queryKey: ['today', lotteryCode],
    queryFn: (): Promise<{ final: Prediction | null; models: Prediction[] }> => fetchToday(lotteryCode),
  });
  const finalPred = q.data?.final;
  const models = q.data?.models || [];

  return (
    <Screen title="预测详情" onBack={() => nav.goBack()}>
      <QueryState loading={q.isLoading} error={q.error} onRetry={() => q.refetch()} />
      {!q.isLoading && !q.isError ? (
        <>
          <SectionTitle>最终推荐</SectionTitle>
          <Card mode="elevated">
            <Card.Content>
              <NumberBalls numbers={finalPred?.predicted_numbers?.numbers || []} />
              {(finalPred?.predicted_numbers?.back_numbers || []).length ? (
                <NumberBalls numbers={finalPred?.predicted_numbers?.back_numbers || []} color="#056DE8" />
              ) : null}
              <StatsLine pn={finalPred?.predicted_numbers} />
              <Text variant="bodyMedium" style={{ marginTop: 8 }}>
                {finalPred?.reason || '暂无理由'}
              </Text>
            </Card.Content>
          </Card>
          <SectionTitle>各模型预测</SectionTitle>
          {models.map((m: Prediction) => (
            <Card key={m.id} mode="contained">
              <Card.Title title={m.model_code} subtitle={`置信度 ${m.confidence}`} />
              <Card.Content>
                <NumberBalls numbers={m.predicted_numbers?.numbers || []} color="#056DE8" />
                {(m.predicted_numbers?.back_numbers || []).length ? (
                  <NumberBalls numbers={m.predicted_numbers?.back_numbers || []} color="#0084FF" />
                ) : null}
                <StatsLine pn={m.predicted_numbers} />
                <Text variant="bodyMedium" style={{ marginTop: 8 }}>
                  {m.reason || (m.success ? '' : m.error_message)}
                </Text>
              </Card.Content>
            </Card>
          ))}
        </>
      ) : null}
    </Screen>
  );
}
