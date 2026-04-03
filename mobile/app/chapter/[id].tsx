import { useEffect, useState } from 'react';
import { StyleSheet, ScrollView, Pressable, ActivityIndicator } from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useLocalSearchParams, router } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { useColors, Card, Button, ProgressBar } from '@/components/Themed';
import { MasteryBar, MasteryBadge } from '@/components/MasteryBar';
import { masteryColors } from '@/constants/Colors';
import { typography, spacing, radius } from '@/constants/Typography';
import { getLessonCard, getMasteries } from '@/services/api';
import type { LessonCardResponse, ItemResponse, NotionResponse, MasteryResponse, MasteryBreakdown } from '@/services/api';

export default function ChapterScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const colors = useColors();
  const insets = useSafeAreaInsets();
  const [expanded, setExpanded] = useState<string | null>(null);
  const [lessonCard, setLessonCard] = useState<LessonCardResponse | null>(null);
  const [masteries, setMasteries] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;

    Promise.all([
      getLessonCard(id).then(setLessonCard),
      getMasteries().then((list) => {
        const map: Record<string, string> = {};
        list.forEach((m) => {
          map[m.item_id] = m.state;
        });
        setMasteries(map);
      }),
    ])
      .catch(console.warn)
      .finally(() => setLoading(false));
  }, [id]);

  // Group items by notion. Items without notion go into a virtual "Tous les items" group.
  const allItems = lessonCard?.items ?? [];
  const allNotions = lessonCard?.notions ?? [];

  const notionsWithItems = (() => {
    if (allNotions.length > 0) {
      // Group items by their notion_id
      const groups = allNotions.map((n) => {
        const notionItems = allItems.filter((i) => i.notion_id === n.id);
        const breakdown: MasteryBreakdown = { unknown: 0, fragile: 0, ok: 0, solid: 0 };
        notionItems.forEach((item) => {
          const state = masteries[item.id] || 'unknown';
          if (state in breakdown) breakdown[state as keyof MasteryBreakdown]++;
          else breakdown.unknown++;
        });
        return { ...n, items: notionItems, breakdown };
      });
      // Add orphan items (no notion_id)
      const orphans = allItems.filter((i) => !i.notion_id);
      if (orphans.length > 0) {
        const breakdown: MasteryBreakdown = { unknown: 0, fragile: 0, ok: 0, solid: 0 };
        orphans.forEach((item) => {
          const state = masteries[item.id] || 'unknown';
          if (state in breakdown) breakdown[state as keyof MasteryBreakdown]++;
          else breakdown.unknown++;
        });
        groups.push({ id: '__orphans__', chapter_id: '', name: 'Autres items', sort_order: 999, items: orphans, breakdown });
      }
      return groups;
    }
    // No notions at all — show all items in a single group
    if (allItems.length > 0) {
      const breakdown: MasteryBreakdown = { unknown: 0, fragile: 0, ok: 0, solid: 0 };
      allItems.forEach((item) => {
        const state = masteries[item.id] || 'unknown';
        if (state in breakdown) breakdown[state as keyof MasteryBreakdown]++;
        else breakdown.unknown++;
      });
      return [{ id: '__all__', chapter_id: '', name: 'Items du chapitre', sort_order: 0, items: allItems, breakdown }];
    }
    return [];
  })();

  // Auto-expand first group once loaded
  useEffect(() => {
    if (notionsWithItems.length > 0 && expanded === null) {
      setExpanded(notionsWithItems[0].id);
    }
  }, [notionsWithItems.length]);

  const globalBreakdown = notionsWithItems.reduce(
    (acc, n) => ({
      unknown: acc.unknown + n.breakdown.unknown,
      fragile: acc.fragile + n.breakdown.fragile,
      ok: acc.ok + n.breakdown.ok,
      solid: acc.solid + n.breakdown.solid,
    }),
    { unknown: 0, fragile: 0, ok: 0, solid: 0 },
  );

  const toggleExpand = (nid: string) => {
    setExpanded((prev) => (prev === nid ? null : nid));
  };

  if (loading) {
    return (
      <RNView style={[styles.container, { backgroundColor: colors.backgroundSecondary, justifyContent: 'center', alignItems: 'center' }]}>
        <ActivityIndicator size="large" color={colors.tint} />
        <RNText style={[typography.body, { color: colors.textSecondary, marginTop: spacing.sm }]}>
          Chargement...
        </RNText>
      </RNView>
    );
  }

  return (
    <RNView style={[styles.container, { backgroundColor: colors.backgroundSecondary }]}>
      <ScrollView contentContainerStyle={{ paddingBottom: insets.bottom + 100 }}>
        {/* Chapter header */}
        <RNView style={[styles.chapterHeader, { backgroundColor: colors.background }]}>
          <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>
            {lessonCard?.chapter.subject ?? ''}
          </RNText>
          <RNText style={[typography.h2, { color: colors.text, marginTop: 2 }]}>
            {lessonCard?.chapter.name ?? ''}
          </RNText>
          <RNView style={{ marginTop: spacing.md }}>
            <MasteryBar breakdown={globalBreakdown} height={12} />
          </RNView>
        </RNView>

        {/* Notions */}
        <RNView style={styles.notions}>
          {notionsWithItems.map((notion) => {
            const isExpanded = expanded === notion.id;
            const total = notion.breakdown.unknown + notion.breakdown.fragile + notion.breakdown.ok + notion.breakdown.solid;
            const mastered = notion.breakdown.ok + notion.breakdown.solid;

            return (
              <Card key={notion.id} style={{ marginBottom: spacing.sm }}>
                <Pressable onPress={() => toggleExpand(notion.id)}>
                  <RNView style={styles.notionHeader}>
                    <RNText style={[typography.body, { color: colors.textSecondary }]}>
                      {isExpanded ? '▼' : '▶'}
                    </RNText>
                    <RNView style={{ flex: 1 }}>
                      <RNText style={[typography.bodyBold, { color: colors.text }]}>
                        {notion.name}
                      </RNText>
                    </RNView>
                    <RNText style={[typography.caption, { color: colors.textSecondary }]}>
                      {mastered}/{total}
                    </RNText>
                  </RNView>
                  <MasteryBar breakdown={notion.breakdown} height={6} showLabel={false} />
                </Pressable>

                {isExpanded && (
                  <RNView testID="chapter-item-list" style={styles.itemList}>
                    {notion.items.map((item) => (
                      <RNView key={item.id} style={styles.itemRow}>
                        <RNView style={{ flex: 1 }}>
                          <RNText style={[typography.body, { color: colors.text }]}>
                            {item.term ?? item.id}
                          </RNText>
                          <RNText style={[typography.small, { color: colors.textSecondary }]}>
                            {item.item_type}
                          </RNText>
                        </RNView>
                        <MasteryBadge state={(masteries[item.id] || 'unknown') as 'unknown' | 'fragile' | 'ok' | 'solid'} />
                      </RNView>
                    ))}
                  </RNView>
                )}
              </Card>
            );
          })}
        </RNView>

        {/* Empty state */}
        {notionsWithItems.length === 0 && (
          <RNView style={styles.notions}>
            <Card>
              <RNText style={[typography.body, { color: colors.textSecondary, textAlign: 'center' }]}>
                Aucune notion pour ce chapitre.
              </RNText>
            </Card>
          </RNView>
        )}
      </ScrollView>

      {/* Floating CTA */}
      <RNView style={[styles.floatingCTA, { backgroundColor: colors.background, borderTopColor: colors.border, paddingBottom: insets.bottom + spacing.sm }]}>
        <Button
          testID="chapter-revise-btn"
          title="🎯  Lancer une session"
          variant="primary"
          fullWidth
          onPress={() => router.push(`/session/${id}`)}
        />
      </RNView>
    </RNView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  chapterHeader: { padding: spacing.md },
  notions: { padding: spacing.md },
  notionHeader: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, marginBottom: spacing.sm },
  itemList: { marginTop: spacing.md, gap: spacing.sm },
  itemRow: { flexDirection: 'row', alignItems: 'center', paddingVertical: spacing.xs },
  floatingCTA: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    padding: spacing.md,
    borderTopWidth: 1,
  },
});
