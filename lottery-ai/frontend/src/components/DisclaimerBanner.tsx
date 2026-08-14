import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { DISCLAIMER } from '../types';

export default function DisclaimerBanner() {
  return (
    <View style={styles.box}>
      <Text style={styles.text}>{DISCLAIMER}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  box: {
    backgroundColor: '#fef3c7',
    borderColor: '#f59e0b',
    borderWidth: 1,
    borderRadius: 8,
    padding: 10,
    marginBottom: 12,
  },
  text: { color: '#92400e', fontSize: 12, lineHeight: 18 },
});
