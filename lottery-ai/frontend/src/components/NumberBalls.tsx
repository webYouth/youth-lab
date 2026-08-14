import React from 'react';
import { StyleSheet, Text, View } from 'react-native';

type Props = { numbers: number[]; color?: string };

export default function NumberBalls({ numbers, color = '#2563eb' }: Props) {
  return (
    <View style={styles.wrap}>
      {(numbers || []).map((n, idx) => (
        <View key={`${n}-${idx}`} style={[styles.ball, { backgroundColor: color }]}>
          <Text style={styles.text}>{String(n).padStart(2, '0')}</Text>
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { flexDirection: 'row', flexWrap: 'wrap', gap: 8 },
  ball: {
    width: 36,
    height: 36,
    borderRadius: 18,
    alignItems: 'center',
    justifyContent: 'center',
  },
  text: { color: '#fff', fontWeight: '700', fontSize: 12 },
});
