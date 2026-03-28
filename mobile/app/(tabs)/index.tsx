import { useCallback, useState } from 'react';
import {
  ScrollView,
  StyleSheet,
  RefreshControl,
  Pressable,
} from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { router } from 'expo-router';

import { useColors, Card, Button, ProgressBar } from '@/components/Themed';
import { MasteryBar } from '@/components/MasteryBar';
import { typography, spacing, radius } from '@/constants/Typography';
import type { Chapter } from '@/services/api';

// Mock data until API is connected
const MOCK_CHAPTERS: Chapter[] = [
  {
    id: '1',
    subject: 'Physique',
    title: 'Densité et masse volumique',
    is_demo: false,
    mastery_breakdown: { unknown: 1, fragile: 1, ok: 3, solid: 3 },
    item_count: 8,
    last_revised_at: new Date(Date.now() - 2 * 3600_000).toISOString(),
    created_at: new Date().toISOString(),
  },
  {
    id: '2',
    subject: 'Maths',
    title: 'Fractions et proportionnalité',
    is_demo: false,
    mastery_breakdown: { unknown: 4, fragile: 4, ok: 3, solid: 1 },
    item_count: 12,
    last_revised_at: new Date(Date.now() - 24 * 3600_000).toISOString(),
    created_at: new Date().toISOString(),
  },
  {
    id: 'demo',
    subject: 'SVT',
    title: 'La cellule (DEMO)',
    is_demo: true,
    mastery_breakdown: { unknown: 5, fragile: 1, ok: 0, solid: 0 },
    item_count: 6,
    last_revised_at: null,
    created_at: new Date().toISOString(),
  },
];

function timeAgo(date: string | null): string {
  if (!date) return 'Jamais révisé';
  const diff = Date.now() - new Date(date).getTime();
  const hours = Math.floor(diff / 3600_000);
  if (hours < 1) return 'Révisé à l\'instant';
  if (hours < 24) return `Révisé il y a ${hours}h`;
  const days = Math.floor(hours / 24);
  return `Révisé il y a ${days}j`;
}

export default function DashboardScreen() {
  const colors = useColors();
  const insets = useSafeAreaInsets();
  const [refreshing, setRefreshing] = useState(false);
  const [chapters] = useState<Chapter[]>(MOCK_CHAPTERS);

  const globalBreakdown = chapters.reduce(
    (acc, c) => ({
      unknown: acc.unknown + c.mastery_breakdown.unknown,
      fragile: acc.fragile + c.mastery_breakdown.fragile,
      ok: acc.ok + c.mastery_breakdown.ok,
      solid: acc.solid + c.mastery_breakdown.solid,
    }),
    { unknown: 0, fragile: 0, ok: 0, solid: 0 },
  );

  const onRefresh = useCallback(() => {
    setRefreshing(true);
    setTimeout(() => setRefreshing(false), 800);
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
      <RNView style={styles.section}>
        <RNText style={[typography.h3, { color: colors.text, marginBottom: spacing.md }]}>
          Mes chapitres
        </RNText>
        {chapters.map((chapter) => (
          <ChapterCard key={chapter.id} chapter={chapter} />
        ))}
      </RNView>

      {/* Capture CTA */}
      <RNView style={[styles.section, { alignItems: 'center' }]}>
        <Button
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
  const mastered =
    chapter.mastery_breakdown.ok + chapter.mastery_breakdown.solid;
  const total =
    mastered +
    chapter.mastery_breakdown.fragile +
    chapter.mastery_breakdown.unknown;
  const pct = total > 0 ? Math.round((mastered / total) * 100) : 0;

  return (
    <Pressable onPress={() => router.push(`/chapter/${chapter.id}`)}>
      <Card style={{ marginBottom: spacing.sm }}>
        <RNView style={styles.chapterHeader}>
          <RNView style={{ flex: 1 }}>
            <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>
              {chapter.is_demo ? '🧪 ' : ''}{chapter.subject}
            </RNText>
            <RNText style={[typography.bodyBold, { color: colors.text, marginTop: 2 }]}>
              {chapter.title}
            </RNText>
          </RNView>
          <RNText style={[typography.h2, { color: colors.tint }]}>{pct}%</RNText>
        </RNView>

        <MasteryBar breakdown={chapter.mastery_breakdown} showLabel={false} />

        <RNView style={[styles.chapterFooter, { marginTop: spacing.sm }]}>
          <RNText style={[typography.small, { color: colors.textSecondary }]}>
            {chapter.item_count} items · {timeAgo(chapter.last_revised_at)}
          </RNText>
          <Button
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
