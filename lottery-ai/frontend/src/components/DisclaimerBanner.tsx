import React from 'react';
import { Banner } from 'react-native-paper';
import { DISCLAIMER } from '../types';

export default function DisclaimerBanner() {
  return (
    <Banner visible icon="alert-outline" style={{ marginBottom: 4 }}>
      {DISCLAIMER}
    </Banner>
  );
}
