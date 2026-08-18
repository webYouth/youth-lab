import React, { useMemo, useRef, useState } from 'react';
import { Dimensions, Modal, Pressable, StyleSheet, View } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import SvgChart, { SVGRenderer } from '@wuba/react-native-echarts/svgChart';
import * as echarts from 'echarts/core';
import { BarChart } from 'echarts/charts';
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components';
import { Card, IconButton, Text, useTheme } from 'react-native-paper';
import { fetchAccuracy } from '../api/client';
import type { AccuracyHistory, AccuracyStat, ModelStrategy } from '../types';
import { formatDrawDate, formatYuan } from '../lib/dates';
import LotteryChips from '../components/LotteryChips';
import QueryState from '../components/QueryState';
import Screen from '../components/Screen';

echarts.use([SVGRenderer, BarChart, GridComponent, TooltipComponent, LegendComponent]);

function modelLabel(code: string) {
  return code === 'FINAL' ? '推荐号' : code;
}

function AccuracyChart({
  list,
  width,
  height,
  dark,
}: {
  list: AccuracyStat[];
  width: number;
  height: number;
  dark: boolean;
}) {
  const chartRef = useRef<any>(null);
  const option = useMemo(
    () => ({
      backgroundColor: 'transparent',
      tooltip: { trigger: 'axis' },
      legend: {
        data: ['加权准确率', '近30天'],
        textStyle: { color: dark ? '#EDE0DE' : '#201A19' },
      },
      grid: { left: 40, right: 16, top: 40, bottom: 40 },
      xAxis: {
        type: 'category',
        data: list.map((x) => modelLabel(x.model_code)),
        axisLabel: { color: dark ? '#D8C2BE' : '#534341', rotate: 20, fontSize: 10 },
      },
      yAxis: {
        type: 'value',
        axisLabel: { formatter: '{value}%', color: dark ? '#D8C2BE' : '#534341' },
        splitLine: { lineStyle: { color: dark ? '#534341' : '#E7E0EC' } },
      },
      series: [
        {
          name: '加权准确率',
          type: 'bar',
          data: list.map((x) => Number((x.avg_hit_rate * 100).toFixed(1))),
          itemStyle: { color: '#B3261E', borderRadius: [6, 6, 0, 0] },
        },
        {
          name: '近30天',
          type: 'bar',
          data: list.map((x) => Number((x.last_30_hit_rate * 100).toFixed(1))),
          itemStyle: { color: '#1565C0', borderRadius: [6, 6, 0, 0] },
        },
      ],
    }),
    [list, dark]
  );

  React.useEffect(() => {
    let chart: echarts.ECharts | undefined;
    if (chartRef.current) {
      chart = echarts.init(chartRef.current, dark ? 'dark' : 'light', {
        renderer: 'svg',
        width,
        height,
      });
      chart.setOption(option);
    }
    return () => chart?.dispose();
  }, [option, width, height, dark]);

  return <SvgChart ref={chartRef} handleGesture={false} />;
}

export default function AccuracyScreen() {
  const theme = useTheme();
  const dark = theme.dark;
  const [code, setCode] = useState('DLT');
  const [fullscreen, setFullscreen] = useState(false);
  const q = useQuery({ queryKey: ['acc', code], queryFn: () => fetchAccuracy(code) });
  const list: AccuracyStat[] = q.data?.list || [];
  const strategies: ModelStrategy[] = q.data?.strategies || [];
  const history: AccuracyHistory[] = q.data?.history || [];
  const stake = q.data?.stake_yuan ?? 2;
  const chartW = Dimensions.get('window').width - 64;
  const fullW = Dimensions.get('window').width - 24;
  const fullH = Dimensions.get('window').height - 120;

  return (
    <Screen title="中奖盈亏" subtitle={`按单注 ${stake} 元；命中率按奖级奖金加权`} messageBell>
      <LotteryChips value={code} onChange={setCode} />
      <QueryState loading={q.isLoading} error={q.error} onRetry={() => q.refetch()} />
      {list.length > 0 ? (
        <Card mode="contained" onPress={() => setFullscreen(true)}>
          <Card.Title
            title="加权准确率"
            subtitle="官方单注奖金对数加权；浮动奖按封顶计分、不计入盈亏"
            right={(props) => <IconButton {...props} icon="fullscreen" onPress={() => setFullscreen(true)} />}
          />
          <Card.Content>
            <AccuracyChart list={list} width={chartW} height={240} dark={dark} />
          </Card.Content>
        </Card>
      ) : !q.isLoading && !q.isError ? (
        <Text variant="bodyMedium">暂无评估数据，等待开奖后更新</Text>
      ) : null}
      {list.map((a) => (
        <Card key={a.model_code} mode="contained">
          <Card.Title title={modelLabel(a.model_code)} subtitle={`已评估 ${a.total_predictions} 期`} />
          <Card.Content>
            <Text variant="bodyMedium">
              加权准确率 {(a.avg_hit_rate * 100).toFixed(1)}% · 中奖 {a.total_wins ?? 0} 次
            </Text>
            <Text variant="bodyMedium">
              投入 {formatYuan(a.total_stake)} · 奖金 {formatYuan(a.total_prize)} · 盈亏 {formatYuan(a.total_profit, true)}
            </Text>
            <Text variant="bodyMedium">
              近 30 天加权 {(a.last_30_hit_rate * 100).toFixed(1)}% · 中奖 {a.last_30_wins ?? 0} 次 · 盈亏 {formatYuan(a.last_30_profit, true)}
            </Text>
          </Card.Content>
        </Card>
      ))}
      {history.length > 0 ? (
        <Card mode="contained">
          <Card.Title title="推荐号历史中奖" subtitle="每期按 1 注 2 元计；浮动奖只记中奖不记金额" />
          <Card.Content>
            {history.map((h) => (
              <Text key={`${h.issue}-${h.model_code}`} variant="bodyMedium" style={{ marginBottom: 6 }}>
                {formatDrawDate(h.draw_date)} 第{h.issue}期 · {h.is_win ? h.level : '未中'} ·{' '}
                {h.prize_floating ? '浮动未计入' : formatYuan(h.prize_yuan)} / 盈亏 {formatYuan(h.profit_yuan, true)}
              </Text>
            ))}
          </Card.Content>
        </Card>
      ) : null}
      {strategies.length > 0 ? (
        <Card mode="contained">
          <Card.Title title="自动策略" subtitle="按近30天加权准确率调整投票权重" />
          <Card.Content>
            {strategies.map((s) => (
              <Text key={s.model_code} variant="bodyMedium">
                {modelLabel(s.model_code)} · 权重 {s.weight.toFixed(2)} · 近30天加权 {(s.last_30_hit_rate * 100).toFixed(1)}%
              </Text>
            ))}
          </Card.Content>
        </Card>
      ) : null}

      <Modal visible={fullscreen} animationType="fade" onRequestClose={() => setFullscreen(false)}>
        <View style={[styles.full, { backgroundColor: theme.colors.background }]}>
          <View style={styles.fullBar}>
            <Text variant="titleMedium">加权准确率 · 全屏</Text>
            <IconButton icon="close" onPress={() => setFullscreen(false)} />
          </View>
          <Pressable style={styles.fullChart} onPress={() => setFullscreen(false)}>
            {list.length ? <AccuracyChart list={list} width={fullW} height={fullH} dark={dark} /> : null}
          </Pressable>
        </View>
      </Modal>
    </Screen>
  );
}

const styles = StyleSheet.create({
  full: { flex: 1, paddingTop: 40 },
  fullBar: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', paddingHorizontal: 12 },
  fullChart: { flex: 1, alignItems: 'center', justifyContent: 'center' },
});
