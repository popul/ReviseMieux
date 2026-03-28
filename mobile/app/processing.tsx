import { useEffect, useState } from 'react';
import { StyleSheet, Animated } from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { router } from 'expo-router';

import { useColors, Button, ProgressBar } from '@/components/Themed';
import { typography, spacing, radius } from '@/constants/Typography';

type Phase = 1 | 2 | 3;

const PHASE_MESSAGES: Record<Phase, string> = {
  1: 'Lecture de tes pages...',
  2: 'Création des questions...',
  3: 'Presque fini...',
};

const PHASE_ICONS: Record<Phase, string> = {
  1: '📄',
  2: '🔍',
  3: '✨',
};

export default function ProcessingScreen() {
  const colors = useColors();
  const insets = useSafeAreaInsets();
  const [phase, setPhase] = useState<Phase>(1);
  const [progress, setProgress] = useState(0);
  const [showClose, setShowClose] = useState(false);
  const [failed, setFailed] = useState(false);

  // Simulate pipeline progress
  useEffect(() => {
    const interval = setInterval(() => {
      setProgress((prev) => {
        if (prev >= 1) {
          clearInterval(interval);
          // Simulate completion — navigate to chapter
          setTimeout(() => router.replace('/(tabs)'), 500);
          return 1;
        }
        const next = prev + 0.02;
        if (next > 0.3 && phase === 1) setPhase(2);
        if (next > 0.7 && phase === 2) setPhase(3);
        return next;
      });
    }, 200);

    const closeTimer = setTimeout(() => setShowClose(true), 15_000);

    return () => {
      clearInterval(interval);
      clearTimeout(closeTimer);
    };
  }, [phase]);

  if (failed) {
    return <RecoveryScreen colors={colors} insets={insets} onRetry={() => setFailed(false)} />;
  }

  return (
    <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top + spacing.xxl }]}>
      <RNView style={styles.center}>
        <RNText style={{ fontSize: 64 }}>{PHASE_ICONS[phase]}</RNText>
        <RNText style={[typography.h2, { color: colors.text, marginTop: spacing.lg, textAlign: 'center' }]}>
          {PHASE_MESSAGES[phase]}
        </RNText>

        <RNView style={{ width: '80%', marginTop: spacing.xl }}>
          <ProgressBar progress={progress} height={10} />
          <RNText style={[typography.caption, { color: colors.textSecondary, textAlign: 'center', marginTop: spacing.sm }]}>
            Phase {phase}/3 · {Math.round(progress * 100)}%
          </RNText>
        </RNView>
      </RNView>

      {showClose && (
        <RNView style={[styles.bottomHint, { backgroundColor: colors.tintLight, borderColor: colors.tint }]}>
          <RNText style={[typography.body, { color: colors.tint, textAlign: 'center' }]}>
            Tu peux fermer l'app, on te prévient quand c'est prêt ! 🔔
          </RNText>
        </RNView>
      )}
    </RNView>
  );
}

function RecoveryScreen({
  colors,
  insets,
  onRetry,
}: {
  colors: any;
  insets: { top: number; bottom: number };
  onRetry: () => void;
}) {
  return (
    <RNView
      style={[
        styles.container,
        { backgroundColor: colors.background, paddingTop: insets.top + spacing.xl, paddingHorizontal: spacing.lg },
      ]}
    >
      <RNText style={[typography.h2, { color: colors.text, textAlign: 'center' }]}>
        Les photos sont un peu{'\n'}difficiles à lire
      </RNText>
      <RNText style={[typography.body, { color: colors.textSecondary, textAlign: 'center', marginTop: spacing.sm }]}>
        Pas de panique, ça arrive souvent au début ! 😊
      </RNText>

      <RNView style={styles.tips}>
        <TipRow emoji="☀️" good="Bonne lumière" bad="Pas d'ombre" colors={colors} />
        <TipRow emoji="📄" good="Page entière" bad="Pas coupée" colors={colors} />
        <TipRow emoji="🔍" good="Texte lisible" bad="Pas trop petit" colors={colors} />
      </RNView>

      <RNView style={[styles.recoveryActions, { paddingBottom: insets.bottom + spacing.lg }]}>
        <Button title="📸  Reprendre les photos" variant="primary" fullWidth onPress={onRetry} />
        <Button
          title="Essayer quand même"
          variant="outline"
          fullWidth
          onPress={() => router.replace('/(tabs)')}
          style={{ marginTop: spacing.sm }}
        />
      </RNView>
    </RNView>
  );
}

function TipRow({
  emoji,
  good,
  bad,
  colors,
}: {
  emoji: string;
  good: string;
  bad: string;
  colors: any;
}) {
  return (
    <RNView style={styles.tipRow}>
      <RNText style={{ fontSize: 32 }}>{emoji}</RNText>
      <RNView style={{ flex: 1 }}>
        <RNText style={[typography.body, { color: colors.success }]}>✅ {good}</RNText>
        <RNText style={[typography.body, { color: colors.error }]}>❌ {bad}</RNText>
      </RNView>
    </RNView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  bottomHint: {
    marginHorizontal: spacing.lg,
    marginBottom: spacing.xl,
    padding: spacing.md,
    borderRadius: radius.md,
    borderWidth: 1,
  },
  tips: { marginTop: spacing.xl, gap: spacing.lg },
  tipRow: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  recoveryActions: { marginTop: 'auto' as any },
});
