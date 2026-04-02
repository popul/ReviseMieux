import { StyleSheet, ScrollView, Pressable } from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { useColors, Card, Separator } from '@/components/Themed';
import { typography, spacing, radius } from '@/constants/Typography';

type SettingsRow = {
  icon: string;
  label: string;
  detail?: string;
  onPress?: () => void;
};

const ROWS: SettingsRow[] = [
  { icon: '👤', label: 'Profil', detail: 'Hugo · Zone A' },
  { icon: '📅', label: 'Emploi du temps', detail: '4 matières' },
  { icon: '👨‍👩‍👦', label: 'Inviter un parent', detail: 'Non lié' },
  { icon: '🔔', label: 'Notifications', detail: 'Activées' },
  { icon: '📊', label: 'Exporter mes données' },
  { icon: '❓', label: 'Aide & feedback' },
  { icon: '📖', label: 'À propos', detail: 'v0.1.0' },
];

export default function SettingsScreen() {
  const colors = useColors();
  const insets = useSafeAreaInsets();

  return (
    <ScrollView
      style={[styles.container, { backgroundColor: colors.backgroundSecondary }]}
      contentContainerStyle={{ paddingBottom: insets.bottom + 40 }}
    >
      <Card style={styles.card}>
        {ROWS.map((row, idx) => (
          <Pressable key={row.label} onPress={row.onPress}>
            <RNView style={styles.row}>
              <RNText style={{ fontSize: 20 }}>{row.icon}</RNText>
              <RNView style={{ flex: 1 }}>
                <RNText style={[typography.body, { color: colors.text }]}>{row.label}</RNText>
                {row.detail && (
                  <RNText style={[typography.small, { color: colors.textSecondary }]}>{row.detail}</RNText>
                )}
              </RNView>
              <RNText style={[typography.body, { color: colors.textSecondary }]}>›</RNText>
            </RNView>
            {idx < ROWS.length - 1 && (
              <RNView style={{ height: 1, backgroundColor: colors.border, marginLeft: 44 }} />
            )}
          </Pressable>
        ))}
      </Card>

      <Pressable style={[styles.dangerButton, { borderColor: colors.error }]}>
        <RNText style={[typography.body, { color: colors.error, textAlign: 'center' }]}>
          Supprimer mon compte
        </RNText>
      </Pressable>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  card: { margin: spacing.md },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
    paddingVertical: spacing.md,
  },
  dangerButton: {
    marginHorizontal: spacing.md,
    marginTop: spacing.lg,
    paddingVertical: spacing.md,
    borderRadius: radius.md,
    borderWidth: 1,
  },
});
