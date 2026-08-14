import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, Text } from 'react-native-paper';
import { fetchDraws } from '../api/client';
import type { DrawResult } from '../types';
import LotteryChips from '../components/LotteryChips';
import NumberBalls from '../components/NumberBalls';
import QueryState from '../components/QueryState';
import Screen from '../components/Screen';

export default function HistoryScreen() {
  const [code, setCode] = useState('DLT');
  const q = useQuery({ queryKey: ['draws', code], queryFn: (): Promise<{ list: DrawResult[]; total: number }> => fetchDraws(code, 1) });

  return (
    <Screen title="历史开奖" messageBell>
      <LotteryChips value={code} onChange={setCode} />
      <QueryState loading={q.isLoading} error={q.error} onRetry={() => q.refetch()} />
      {(q.data?.list || []).map((d: DrawResult) => {
        const nums = d.result?.numbers || d.result?.digits || d.result?.front || [];
        const back = d.result?.back || [];
        return (
          <Card key={d.id} mode="contained">
            <Card.Title title={`第 ${d.issue} 期`} subtitle={String(d.draw_date).slice(0, 10)} />
            <Card.Content>
              <NumberBalls numbers={nums} />
              {back.length ? <NumberBalls numbers={back} color="#1565C0" /> : null}
            </Card.Content>
          </Card>
        );
      })}
      {!q.isLoading && !q.isError && !(q.data?.list || []).length ? (
        <Text variant="bodyMedium">暂无开奖记录</Text>
      ) : null}
    </Screen>
  );
}
