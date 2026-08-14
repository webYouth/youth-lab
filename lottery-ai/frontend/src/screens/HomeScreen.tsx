import React, { useState } from 'react';
import { useNavigation } from '@react-navigation/native';
import { useQuery } from '@tanstack/react-query';
import { Button, Card, Text } from 'react-native-paper';
import { fetchLotteryTypes, fetchToday } from '../api/client';
import DisclaimerBanner from '../components/DisclaimerBanner';
import LotteryChips from '../components/LotteryChips';
import NumberBalls from '../components/NumberBalls';
import QueryState from '../components/QueryState';
import Screen, { SectionTitle } from '../components/Screen';
import { LOTTERY_LABELS } from '../theme';

export default function HomeScreen() {
  const nav = useNavigation<any>();
  const [code, setCode] = useState('DLT');
  const typesQ = useQuery({ queryKey: ['types'], queryFn: fetchLotteryTypes });
  const todayQ = useQuery({ queryKey: ['today', code], queryFn: () => fetchToday(code) });

  const finalPred = todayQ.data?.final;
  const nums = finalPred?.predicted_numbers?.numbers || [];
  const back = finalPred?.predicted_numbers?.back_numbers || [];
  const types = typesQ.data?.length
    ? typesQ.data
    : Object.entries(LOTTERY_LABELS).map(([c, name]) => ({ code: c, name }));

  return (
    <Screen title="彩票 AI" subtitle="学习研究">
      <DisclaimerBanner />
      <LotteryChips value={code} options={types} onChange={setCode} />
      <QueryState loading={todayQ.isLoading} error={todayQ.error} onRetry={() => todayQ.refetch()} />
      {!todayQ.isLoading && !todayQ.isError ? (
        <Card mode="elevated" onPress={() => nav.navigate('PredictDetail', { lotteryCode: code })}>
          <Card.Title title={`${LOTTERY_LABELS[code] || code} · 今日预测`} subtitle={`置信度 ${finalPred?.confidence ?? '-'}`} />
          <Card.Content>
            <NumberBalls numbers={nums} />
            {back.length > 0 ? (
              <>
                <SectionTitle>后区</SectionTitle>
                <NumberBalls numbers={back} color="#1565C0" />
              </>
            ) : null}
            {!nums.length ? <Text variant="bodyMedium">暂无预测，等待模型生成</Text> : null}
          </Card.Content>
          <Card.Actions>
            <Button mode="text" onPress={() => nav.navigate('PredictDetail', { lotteryCode: code })}>
              查看各模型
            </Button>
          </Card.Actions>
        </Card>
      ) : null}
    </Screen>
  );
}
