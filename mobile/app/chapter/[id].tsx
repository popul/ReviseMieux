import { useState } from 'react';
import { StyleSheet, ScrollView, Pressable } from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useLocalSearchParams, router } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { useColors, Card, Button, ProgressBar } from '@/components/Themed';
import { MasteryBar, MasteryBadge } from '@/components/MasteryBar';
import { masteryColors } from '@/constants/Colors';
import { typography, spacing, radius } from '@/constants/Typography';

// Mock data
const MOCK_NOTIONS = [
  {
    id: 'n1',
    name: 'Masse volumique (ρ)',
    breakdown: { unknown: 0, fragile: 1, ok: 2, solid: 2 },
    items: [
      { id: 'i1', term: 'Définition ρ', type: 'KNOWLEDGE', state: 'solid' as const },
      { id: 'i2', term: 'Formule ρ=m/V', type: 'KNOWLEDGE', state: 'ok' as const },
      { id: 'i3', term: 'Unité g/cm³', type: 'KNOWLEDGE', state: 'ok' as const },
      { id: 'i4', term: 'Conversion L↔cm³', type: 'PROCEDURE', state: 'fragile' as const },
      { id: 'i5', term: 'Application densité', type: 'ANALYSIS', state: 'unknown' as const },
    ],
  },
  {
    id: 'n2',
    name: 'Flottabilité',
    breakdown: { unknown: 2, fragile: 1, ok: 1, solid: 0 },
    items: [
      { id: 'i6', term: 'Condition flotte/coule', type: 'KNOWLEDGE', state: 'ok' as const },
      { id: 'i7', term: 'Poussée d\'Archimède', type: 'KNOWLEDGE', state: 'fragile' as const },
      { id: 'i8', term: 'Calcul poussée', type: 'PROCEDURE', state: 'unknown' as const },
      { id: 'i9', term: 'Expérience flottabilité', type: 'ANALYSIS', state: 'unknown' as const },
    ],
  },
  {
    id: 'n3',
    name: 'Mesures & instruments',
    breakdown: { unknown: 2, fragile: 0, ok: 1, solid: 0 },
    items: [
      { id: 'i10', term: 'Éprouvette graduée', type: 'KNOWLEDGE', state: 'ok' as const },
      { id: 'i11', term: 'Balance précision', type: 'KNOWLEDGE', state: 'unknown' as const },
      { id: 'i12', term: 'Protocole mesure', type: 'PROCEDURE', state: 'unknown' as const },
    ],
  },
];

const IGNORED_ITEMS = [
  { id: 'ig1', term: 'Volume molaire' },
  { id: 'ig2', term: 'Pression atmosphérique' },
];

export default function ChapterScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const colors = useColors();
  const insets = useSafeAreaInsets();
  const [expanded, setExpanded] = useState<string | null>('n1');

  const globalBreakdown = MOCK_NOTIONS.reduce(
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

  return (
    <RNView style={[styles.container, { backgroundColor: colors.backgroundSecondary }]}>
      <ScrollView contentContainerStyle={{ paddingBottom: insets.bottom + 100 }}>
        {/* Chapter header */}
        <RNView style={[styles.chapterHeader, { backgroundColor: colors.background }]}>
          <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>Physique</RNText>
          <RNText style={[typography.h2, { color: colors.text, marginTop: 2 }]}>Densité et masse volumique</RNText>
          <RNView style={{ marginTop: spacing.md }}>
            <MasteryBar breakdown={globalBreakdown} height={12} />
          </RNView>
        </RNView>

        {/* Notions */}
        <RNView style={styles.notions}>
          {MOCK_NOTIONS.map((notion) => {
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
                  <RNView style={styles.itemList}>
                    {notion.items.map((item) => (
                      <RNView key={item.id} style={styles.itemRow}>
                        <RNView style={{ flex: 1 }}>
                          <RNText style={[typography.body, { color: colors.text }]}>{item.term}</RNText>
                          <RNText style={[typography.small, { color: colors.textSecondary }]}>{item.type}</RNText>
                        </RNView>
                        <MasteryBadge state={item.state} />
                      </RNView>
                    ))}
                  </RNView>
                )}
              </Card>
            );
          })}
        </RNView>

        {/* Ignored items */}
        {IGNORED_ITEMS.length > 0 && (
          <RNView style={styles.ignoredSection}>
            <RNText style={[typography.captionBold, { color: colors.warning, marginBottom: spacing.sm }]}>
              Points non vérifiés ({IGNORED_ITEMS.length})
            </RNText>
            {IGNORED_ITEMS.map((item) => (
              <RNView key={item.id} style={[styles.ignoredRow, { borderColor: colors.border }]}>
                <RNText style={[typography.body, { color: colors.textSecondary, flex: 1 }]}>
                  {item.term}
                </RNText>
                <Button title="Réactiver" variant="ghost" onPress={() => {}} />
              </RNView>
            ))}
          </RNView>
        )}
      </ScrollView>

      {/* Floating CTA */}
      <RNView style={[styles.floatingCTA, { backgroundColor: colors.background, borderTopColor: colors.border, paddingBottom: insets.bottom + spacing.sm }]}>
        <Button
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
  ignoredSection: { paddingHorizontal: spacing.md, marginTop: spacing.md },
  ignoredRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.sm,
    borderBottomWidth: 1,
  },
  floatingCTA: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    padding: spacing.md,
    borderTopWidth: 1,
  },
});
