import { useState } from 'react';
import { StyleSheet } from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { router } from 'expo-router';

import { useColors, Card, Button } from '@/components/Themed';
import { typography, spacing, radius } from '@/constants/Typography';
import { seedDemo } from '@/services/api';

export default function OnboardingScreen() {
  const colors = useColors();
  const insets = useSafeAreaInsets();
  const [seeding, setSeeding] = useState(false);

  const handleSeedDemo = () => {
    setSeeding(true);
    seedDemo()
      .then(() => router.replace('/(tabs)'))
      .catch((err) => {
        console.warn('Seed failed:', err.message);
        router.replace('/(tabs)'); // go anyway
      })
      .finally(() => setSeeding(false));
  };

  return (
    <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top + spacing.xl }]}>
      <RNText style={[typography.h1, { color: colors.text, textAlign: 'center' }]}>
        Bienvenue Hugo ! 🎉
      </RNText>

      <RNText style={[typography.body, { color: colors.textSecondary, textAlign: 'center', marginTop: spacing.sm }]}>
        Prêt en 3 étapes :
      </RNText>

      {/* Steps */}
      <RNView style={styles.steps}>
        {/* Step 1 - Done */}
        <StepRow
          number="1"
          title="Compte créé"
          subtitle=""
          done
          colors={colors}
        />

        {/* Step 2 - Active */}
        <StepRow
          number="2"
          title="Photographie ton premier cours"
          subtitle="30 secondes chrono !"
          active
          colors={colors}
        />
        <Button
          testID="onboarding-start-btn"
          title="📸  Capturer"
          variant="primary"
          fullWidth
          onPress={() => router.replace('/(tabs)/capture')}
          style={{ marginBottom: spacing.lg }}
        />

        {/* Step 3 - Future */}
        <StepRow
          number="3"
          title="Ta première révision"
          subtitle="~5 min après la capture"
          colors={colors}
        />
      </RNView>

      {/* Demo CTA */}
      <RNView style={styles.demoSection}>
        <RNText style={[typography.caption, { color: colors.textSecondary, textAlign: 'center', marginBottom: spacing.md }]}>
          — ou essaie d'abord —
        </RNText>
        <Card style={{ alignItems: 'center' as const }}>
          <RNText style={{ fontSize: 32, marginBottom: spacing.sm }}>🧪</RNText>
          <RNText style={[typography.bodyBold, { color: colors.text }]}>
            Chapitre démo : Densité
          </RNText>
          <RNText style={[typography.caption, { color: colors.textSecondary, marginTop: spacing.xs }]}>
            8 questions prêtes · 3 min
          </RNText>
          <Button
            testID="onboarding-seed-btn"
            title={seeding ? 'Chargement...' : 'Essayer →'}
            variant="secondary"
            onPress={handleSeedDemo}
            disabled={seeding}
            style={{ marginTop: spacing.md }}
          />
        </Card>
      </RNView>
    </RNView>
  );
}

function StepRow({
  number,
  title,
  subtitle,
  done,
  active,
  colors,
}: {
  number: string;
  title: string;
  subtitle: string;
  done?: boolean;
  active?: boolean;
  colors: any;
}) {
  const circleColor = done
    ? colors.success
    : active
      ? colors.warning
      : colors.border;
  const icon = done ? '✓' : number;

  return (
    <RNView style={styles.stepRow}>
      <RNView
        style={[
          styles.stepCircle,
          { backgroundColor: circleColor },
        ]}
      >
        <RNText style={[typography.bodyBold, { color: '#FFF' }]}>{icon}</RNText>
      </RNView>
      <RNView style={{ flex: 1 }}>
        <RNText style={[typography.bodyBold, { color: done ? colors.textSecondary : colors.text }]}>
          {title}
        </RNText>
        {subtitle !== '' && (
          <RNText style={[typography.small, { color: colors.textSecondary }]}>{subtitle}</RNText>
        )}
      </RNView>
    </RNView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, paddingHorizontal: spacing.lg },
  steps: { marginTop: spacing.xl },
  stepRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
    marginBottom: spacing.md,
  },
  stepCircle: {
    width: 36,
    height: 36,
    borderRadius: 18,
    alignItems: 'center',
    justifyContent: 'center',
  },
  demoSection: { marginTop: 'auto' as any, paddingBottom: spacing.xl },
});
