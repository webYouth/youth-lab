import React, { useState } from 'react';
import { Dimensions } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { BarChart } from 'react-native-chart-kit';
import { Card, Text, useTheme } from 'react-native-paper';
import { fetchAccuracy } from '../api/client';
import type { AccuracyStat } from '../types';
import DisclaimerBanner from '../components/DisclaimerBanner';
import LotteryChips from '../components/LotteryChips';
import QueryState from '../components/QueryState';
import Screen from '../components/Screen';

export default function AccuracyScreen() {
  const theme = useTheme();
  const [code, setCode] = useState('DLT');
  const q = useQuery({ queryKey: ['acc', code], queryFn: (): Promise<{ list: AccuracyStat[] }> => fetchAccuracy(code) });
  const list: AccuracyStat[] = q.data?.list || [];

  return (
    <Screen title="命中率">
      <DisclaimerBanner />
      <LotteryChips value={code} onChange={setCode} />
      <QueryState loading={q.isLoading} error={q.error} onRetry={() => q.refetch()} />
      {list.length > 0 ? (
        <Card mode="contained">
          <Card.Content>
            <BarChart
              data={{
                labels: list.map((x) => x.model_code.slice(0, 6)),
                datasets: [{ data: list.map((x) => Number((x.avg_hit_rate * 100).toFixed(1)) || 0) }],
              }}
              width={Dimensions.get('window').width - 64}
              height={220}
              yAxisLabel=""
              yAxisSuffix="%"
              fromZero
              chartConfig={{
                backgroundGradientFrom: theme.colors.elevation.level1,
                backgroundGradientTo: theme.colors.elevation.level1,
                color: () => theme.colors.primary,
                labelColor: () => theme.colors.onSurfaceVariant,
                barPercentage: 0.55,
                decimalPlaces: 1,
              }}
              style={{ borderRadius: 12 }}
            />
          </Card.Content>
        </Card>
      ) : !q.isLoading && !q.isError ? (
        <Text variant="bodyMedium">暂无命中率数据，等待开奖评估后更新</Text>
      ) : null}
      {list.map((a) => (
        <Card key={a.model_code} mode="contained">
          <Card.Title title={a.model_code} subtitle={`总预测 ${a.total_predictions}`} />
          <Card.Content>
            <Text variant="bodyMedium">平均命中率 {(a.avg_hit_rate * 100).toFixed(1)}%</Text>
            <Text variant="bodyMedium">近 30 天 {(a.last_30_hit_rate * 100).toFixed(1)}%</Text>
          </Card.Content>
        </Card>
      ))}
    </Screen>
  );
}
