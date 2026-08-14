import React, { useState } from 'react';
import { useNavigation } from '@react-navigation/native';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Appbar, Button, Card, Text } from 'react-native-paper';
import { fetchLotteryTypes, fetchToday, runPredict } from '../api/client';
import LotteryChips from '../components/LotteryChips';
import NumberBalls from '../components/NumberBalls';
import QueryState from '../components/QueryState';
import Screen, { SectionTitle } from '../components/Screen';
import { LOTTERY_LABELS } from '../theme';

type Props = { onLogout: () => void };

export default function HomeScreen({ onLogout }: Props) {
  const nav = useNavigation<any>();
  const qc = useQueryClient();
  const [code, setCode] = useState('DLT');
  const typesQ = useQuery({ queryKey: ['types'], queryFn: fetchLotteryTypes });
  const todayQ = useQuery({ queryKey: ['today', code], queryFn: () => fetchToday(code) });
  const runM = useMutation({
    mutationFn: () => runPredict(code),
    onSuccess: (data) => {
      qc.setQueryData(['today', code], data);
      qc.invalidateQueries({ queryKey: ['acc', code] });
    },
  });

  const finalPred = todayQ.data?.final;
  const nums = finalPred?.predicted_numbers?.numbers || [];
  const back = finalPred?.predicted_numbers?.back_numbers || [];
  const types = typesQ.data?.length
    ? typesQ.data
    : Object.entries(LOTTERY_LABELS).map(([c, name]) => ({ code: c, name }));

  return (
    <Screen
      title="彩票 AI"
      subtitle="五模型投票"
      messageBell
      actions={<Appbar.Action icon="logout" accessibilityLabel="退出登录" onPress={onLogout} />}
    >
      <LotteryChips value={code} options={types} onChange={setCode} />
      <Button
        mode="contained"
        icon="creation"
        loading={runM.isPending}
        disabled={runM.isPending}
        onPress={() => runM.mutate()}
      >
        {runM.isPending ? '正在预测，约需一分钟' : '立即预测'}
      </Button>
      {runM.isError ? (
        <Text variant="bodyMedium" style={{ color: '#B3261E' }}>
          {runM.error instanceof Error ? runM.error.message : '预测失败'}
        </Text>
      ) : null}
      <QueryState loading={todayQ.isLoading} error={todayQ.error} onRetry={() => todayQ.refetch()} />
      {!todayQ.isLoading && !todayQ.isError ? (
        <Card mode="elevated" onPress={() => nav.navigate('PredictDetail', { lotteryCode: code })}>
          <Card.Title
            title={`${LOTTERY_LABELS[code] || code} · 最新预测`}
            subtitle={`期号 ${finalPred?.issue ?? '-'} · 置信度 ${finalPred?.confidence ?? '-'}`}
          />
          <Card.Content>
            <NumberBalls numbers={nums} />
            {back.length > 0 ? (
              <>
                <SectionTitle>后区</SectionTitle>
                <NumberBalls numbers={back} color="#1565C0" />
              </>
            ) : null}
            {!nums.length ? <Text variant="bodyMedium">暂无预测，点上方按钮生成</Text> : null}
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
