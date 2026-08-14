import React from 'react';
import { Image, ScrollView, StyleSheet, View } from 'react-native';
import { Appbar, Text, useTheme } from 'react-native-paper';

type Props = {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
  scroll?: boolean;
  onBack?: () => void;
};

export default function Screen({ title, subtitle, children, scroll = true, onBack }: Props) {
  const theme = useTheme();
  const body = scroll ? (
    <ScrollView contentContainerStyle={styles.content}>{children}</ScrollView>
  ) : (
    <View style={[styles.content, styles.flex]}>{children}</View>
  );

  return (
    <View style={[styles.root, { backgroundColor: theme.colors.background }]}>
      <Appbar.Header elevated mode="small">
        {onBack ? (
          <Appbar.BackAction onPress={onBack} />
        ) : (
          <Image source={require('../../assets/icon.png')} style={styles.logo} />
        )}
        <Appbar.Content title={title} subtitle={subtitle} />
      </Appbar.Header>
      {body}
    </View>
  );
}

export function SectionTitle({ children }: { children: string }) {
  return (
    <Text variant="titleMedium" style={styles.section}>
      {children}
    </Text>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1 },
  flex: { flex: 1 },
  content: { padding: 16, paddingBottom: 32, gap: 12 },
  logo: { width: 32, height: 32, borderRadius: 8, marginLeft: 8, marginRight: 4 },
  section: { marginTop: 4 },
});
