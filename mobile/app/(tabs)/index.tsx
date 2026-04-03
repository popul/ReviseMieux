import { useCallback, useEffect, useState } from 'react';
import {
  ScrollView,
  StyleSheet,
  RefreshControl,
  Pressable,
  ActivityIndicator,
} from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { router } from 'expo-router';

import { useColors, Card, Button, ProgressBar } from '@/components/Themed';
import { MasteryBar } from '@/components/MasteryBar';
import { typography, spacing, radius } from '@/constants/Typography';
import { listChapters, seedDemo } from '@/services/api';
import type { Chapter, MasteryBreakdown } from '@/services/api';

export default function DashboardScreen() {
  const colors = useColors();
  const insets = useSafeAreaInsets();
  const [refreshing, setRefreshing] = useState(false);
  const [chapters, setChapters] = useState<Chapter[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    listChapters()
      .then(setChapters)
      .catch(console.warn)
      .finally(() => setLoading(false));
  }, []);

  const globalBreakdown = chapters.reduce(
    (acc, c) => {
      const bd = c.mastery_breakdown ?? { unknown: 0, fragile: 0, ok: 0, solid: 0 };
      return {
        unknown: acc.unknown + bd.unknown,
        fragile: acc.fragile + bd.fragile,
        ok: acc.ok + bd.ok,
        solid: acc.solid + bd.solid,
      };
    },
    { unknown: 0, fragile: 0, ok: 0, solid: 0 },
  );

  const onRefresh = useCallback(() => {
    setRefreshing(true);
    listChapters()
      .then(setChapters)
      .catch(console.warn)
      .finally(() => setRefreshing(false));
  }, []);

  const hasChapters = chapters.length > 0;
  const isEvening = new Date().getHours() >= 17;

  return (
    <ScrollView
      style={[styles.container, { backgroundColor: colors.backgroundSecondary }]}
      contentContainerStyle={{ paddingTop: insets.top + spacing.md, paddingBottom: insets.bottom + 100 }}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}
    >
      {/* Header */}
      <RNView style={styles.header}>
        <RNText style={[typography.h1, { color: colors.text }]}>
          {isEvening ? 'Bonsoir' : 'Bonjour'} Hugo !
        </RNText>
        {isEvening && hasChapters && (
          <RNText style={[typography.body, { color: colors.textSecondary, marginTop: spacing.xs }]}>
            Ce soir : {chapters.filter((c) => !c.is_demo).length} activités, ~15 min
          </RNText>
        )}
      </RNView>

      {/* Loading state */}
      {loading && (
        <RNView style={{ alignItems: 'center', marginTop: spacing.xl }}>
          <ActivityIndicator size="large" color={colors.tint} />
          <RNText style={[typography.body, { color: colors.textSecondary, marginTop: spacing.sm }]}>
            Chargement...
          </RNText>
        </RNView>
      )}

      {/* Empty state — seed demo */}
      {!loading && !hasChapters && (
        <Card style={[styles.section, { alignItems: 'center' as const }]}>
          <RNText style={{ fontSize: 48, marginBottom: spacing.md }}>🧪</RNText>
          <RNText style={[typography.h3, { color: colors.text, textAlign: 'center' }]}>
            Aucun chapitre
          </RNText>
          <RNText style={[typography.body, { color: colors.textSecondary, textAlign: 'center', marginTop: spacing.xs }]}>
            Essaie le chapitre démo pour découvrir l'app !
          </RNText>
          <Button
            title="Charger le chapitre démo"
            variant="primary"
            onPress={() => {
              setLoading(true);
              seedDemo()
                .then(() => listChapters())
                .then(setChapters)
                .catch(console.warn)
                .finally(() => setLoading(false));
            }}
            style={{ marginTop: spacing.md }}
          />
        </Card>
      )}

      {/* Global progress */}
      {hasChapters && (
        <Card style={styles.section}>
          <RNText style={[typography.captionBold, { color: colors.textSecondary, marginBottom: spacing.sm }]}>
            PROGRESSION GLOBALE
          </RNText>
          <MasteryBar breakdown={globalBreakdown} height={12} />
          <RNView style={[styles.statsRow, { marginTop: spacing.sm }]}>
            <StatPill label="SOLID" count={globalBreakdown.solid} color={colors.success} />
            <StatPill label="OK" count={globalBreakdown.ok} color={colors.tint} />
            <StatPill label="FRAGILE" count={globalBreakdown.fragile} color={colors.warning} />
          </RNView>
        </Card>
      )}

      {/* Chapters */}
      {hasChapters && (
        <RNView style={styles.section}>
          <RNText style={[typography.h3, { color: colors.text, marginBottom: spacing.md }]}>
            Mes chapitres
          </RNText>
          {chapters.map((chapter) => (
            <ChapterCard key={chapter.id} chapter={chapter} />
          ))}
        </RNView>
      )}

      {/* Capture CTA */}
      <RNView style={[styles.section, { alignItems: 'center' }]}>
        <Button
          testID="dashboard-capture-btn"
          title="📸  Capturer un cours"
          variant="primary"
          fullWidth
          onPress={() => router.push('/capture')}
        />
      </RNView>
    </ScrollView>
  );
}

function ChapterCard({ chapter }: { chapter: Chapter }) {
  const colors = useColors();
  const bd = chapter.mastery_breakdown ?? { unknown: 0, fragile: 0, ok: 0, solid: 0 };
  const mastered = bd.ok + bd.solid;
  const total = mastered + bd.fragile + bd.unknown;
  const pct = total > 0 ? Math.round((mastered / total) * 100) : 0;

  return (
    <Pressable testID="dashboard-chapter-card" onPress={() => router.push(`/chapter/${chapter.id}`)}>
      <Card style={{ marginBottom: spacing.sm }}>
        <RNView style={styles.chapterHeader}>
          <RNView style={{ flex: 1 }}>
            <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>
              {chapter.is_demo ? '🧪 ' : ''}{chapter.subject}
            </RNText>
            <RNText style={[typography.bodyBold, { color: colors.text, marginTop: 2 }]}>
              {chapter.name}
            </RNText>
          </RNView>
          <RNText style={[typography.h2, { color: colors.tint }]}>{pct}%</RNText>
        </RNView>

        <MasteryBar breakdown={bd} showLabel={false} />

        <RNView style={[styles.chapterFooter, { marginTop: spacing.sm }]}>
          <RNText style={[typography.small, { color: colors.textSecondary }]}>
            {chapter.item_count} items
          </RNText>
          <Button
            testID="dashboard-revise-btn"
            title={chapter.is_demo ? 'Essayer' : 'Réviser'}
            variant="secondary"
            onPress={() => router.push(`/chapter/${chapter.id}`)}
          />
        </RNView>
      </Card>
    </Pressable>
  );
}

function StatPill({ label, count, color }: { label: string; count: number; color: string }) {
  return (
    <RNView style={{ flexDirection: 'row', alignItems: 'center', gap: 4 }}>
      <RNView style={{ width: 8, height: 8, borderRadius: 4, backgroundColor: color }} />
      <RNText style={[typography.small, { color }]}>
        {count} {label}
      </RNText>
    </RNView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  header: { paddingHorizontal: spacing.md, marginBottom: spacing.lg },
  section: { paddingHorizontal: spacing.md, marginBottom: spacing.lg },
  statsRow: { flexDirection: 'row', gap: spacing.md },
  chapterHeader: { flexDirection: 'row', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: spacing.sm },
  chapterFooter: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
});
