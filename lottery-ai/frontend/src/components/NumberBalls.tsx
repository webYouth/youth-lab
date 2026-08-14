import React from 'react';
import { StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';

type Props = { numbers: number[]; color?: string };

export default function NumberBalls({ numbers, color = '#B3261E' }: Props) {
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
    width: 40,
    height: 40,
    borderRadius: 20,
    alignItems: 'center',
    justifyContent: 'center',
    elevation: 2,
    shadowColor: '#000',
    shadowOpacity: 0.18,
    shadowRadius: 2,
    shadowOffset: { width: 0, height: 1 },
  },
  text: { color: '#fff', fontWeight: '700', fontSize: 13, lineHeight: 16 },
});
