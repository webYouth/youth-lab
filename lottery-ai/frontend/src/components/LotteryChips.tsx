import React from 'react';
import { StyleSheet, View } from 'react-native';
import { Chip, useTheme } from 'react-native-paper';
import { LOTTERY_LABELS } from '../theme';

type Props = {
  value: string;
  options?: { code: string; name?: string }[];
  onChange: (code: string) => void;
};

const DEFAULTS = ['DLT', 'P3', 'KL8'].map((code) => ({ code, name: LOTTERY_LABELS[code] }));

export default function LotteryChips({ value, options = DEFAULTS, onChange }: Props) {
  const theme = useTheme();
  return (
    <View style={styles.row}>
      {options.map((t) => {
        const active = value === t.code;
        return (
          <Chip
            key={t.code}
            selected={active}
            showSelectedCheck={false}
            compact
            mode={active ? 'flat' : 'outlined'}
            onPress={() => onChange(t.code)}
            style={[
              styles.chip,
              active
                ? { backgroundColor: theme.colors.primary }
                : {
                    backgroundColor: 'transparent',
                    borderColor: theme.colors.outline,
                  },
            ]}
            textStyle={{
              color: active ? theme.colors.onPrimary : theme.colors.onSurfaceVariant,
              fontWeight: active ? '700' : '500',
            }}
          >
            {t.name || LOTTERY_LABELS[t.code] || t.code}
          </Chip>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', flexWrap: 'wrap', gap: 8, marginBottom: 12 },
  chip: { borderRadius: 8 },
});
