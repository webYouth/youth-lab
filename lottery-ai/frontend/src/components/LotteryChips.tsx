import React from 'react';
import { StyleSheet, View } from 'react-native';
import { Chip } from 'react-native-paper';
import { LOTTERY_LABELS } from '../theme';

type Props = {
  value: string;
  options?: { code: string; name?: string }[];
  onChange: (code: string) => void;
};

const DEFAULTS = ['DLT', 'P3', 'KL8'].map((code) => ({ code, name: LOTTERY_LABELS[code] }));

export default function LotteryChips({ value, options = DEFAULTS, onChange }: Props) {
  return (
    <View style={styles.row}>
      {options.map((t) => (
        <Chip
          key={t.code}
          selected={value === t.code}
          showSelectedCheck={false}
          compact
          onPress={() => onChange(t.code)}
          style={styles.chip}
        >
          {t.name || LOTTERY_LABELS[t.code] || t.code}
        </Chip>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', flexWrap: 'wrap', gap: 8, marginBottom: 12 },
  chip: { borderRadius: 8 },
});
