import { useState, useEffect, useCallback } from 'react';
import { StyleSheet, FlatList, ActivityIndicator, Pressable, Alert } from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { router } from 'expo-router';

import { useColors, Card, Button } from '@/components/Themed';
import { typography, spacing, radius } from '@/constants/Typography';
import { listValidations, resolveValidation, ValidationTask } from '@/services/api';

type ViewState = 'loading' | 'list' | 'detail' | 'empty' | 'error';

export default function ValidationScreen() {
  const colors = useColors();
  const insets = useSafeAreaInsets();

  const [view, setView] = useState<ViewState>('loading');
  const [tasks, setTasks] = useState<ValidationTask[]>([]);
  const [selected, setSelected] = useState<ValidationTask | null>(null);
  const [resolving, setResolving] = useState(false);

  const load = useCallback(() => {
    setView('loading');
    listValidations()
      .then((data) => {
        const pending = data.filter((t) => t.status === 'PENDING');
        setTasks(pending);
        setView(pending.length === 0 ? 'empty' : 'list');
      })
      .catch(() => setView('error'));
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleResolve = (action: string) => {
    if (!selected) return;
    setResolving(true);
    resolveValidation(selected.id, action)
      .then(() => {
        setTasks((prev) => prev.filter((t) => t.id !== selected.id));
        setSelected(null);
        setView(tasks.length <= 1 ? 'empty' : 'list');
      })
      .catch(() => Alert.alert('Erreur', 'Impossible de resoudre cette tache'))
      .finally(() => setResolving(false));
  };

  // --- Loading ---
  if (view === 'loading') {
    return (
      <RNView style={[styles.center, { backgroundColor: colors.background }]}>
        <ActivityIndicator size="large" color={colors.tint} />
      </RNView>
    );
  }

  // --- Error ---
  if (view === 'error') {
    return (
      <RNView style={[styles.center, { backgroundColor: colors.background }]}>
        <RNText style={[typography.body, { color: colors.error }]}>Erreur de chargement</RNText>
        <Button title="Reessayer" onPress={load} variant="outline" />
      </RNView>
    );
  }

  // --- Empty ---
  if (view === 'empty') {
    return (
      <RNView style={[styles.center, { backgroundColor: colors.background, paddingTop: insets.top }]}>
        <RNText style={[typography.h2, { color: colors.text, textAlign: 'center' }]}>
          Aucune validation en attente
        </RNText>
        <RNText style={[typography.caption, { color: colors.textSecondary, marginTop: spacing.sm }]}>
          Les items a faible confiance apparaitront ici
        </RNText>
        <RNView style={{ marginTop: spacing.lg }}>
          <Button title="Retour" onPress={() => router.back()} variant="outline" />
        </RNView>
      </RNView>
    );
  }

  // --- Detail ---
  if (view === 'detail' && selected) {
    const confidenceColor =
      selected.priority >= 7 ? colors.success :
      selected.priority >= 4 ? colors.warning : colors.error;

    return (
      <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top }]}>
        <RNView style={styles.header}>
          <Pressable onPress={() => { setSelected(null); setView('list'); }}>
            <RNText style={[typography.body, { color: colors.tint }]}>← Retour</RNText>
          </Pressable>
          <RNText style={[typography.captionBold, { color: colors.text }]}>Detail</RNText>
          <RNView style={{ width: 60 }} />
        </RNView>

        <RNView style={{ padding: spacing.md }}>
          <Card>
            <RNText style={[typography.h3, { color: colors.text }]}>
              {selected.item_term || 'Item sans terme'}
            </RNText>
            <RNView style={{ flexDirection: 'row', gap: spacing.sm, marginTop: spacing.sm }}>
              <RNView style={[styles.badge, { backgroundColor: colors.tintLight }]}>
                <RNText style={[typography.caption, { color: colors.tint }]}>{selected.source}</RNText>
              </RNView>
              <RNView style={[styles.badge, { backgroundColor: confidenceColor + '20' }]}>
                <RNText style={[typography.caption, { color: confidenceColor }]}>
                  Priorite {selected.priority}
                </RNText>
              </RNView>
            </RNView>
            {selected.suggestion && (
              <RNView style={{ marginTop: spacing.md }}>
                <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>Suggestion</RNText>
                <RNText style={[typography.body, { color: colors.text, marginTop: spacing.xs }]}>
                  {selected.suggestion}
                </RNText>
              </RNView>
            )}
          </Card>

          <RNView style={{ marginTop: spacing.lg, gap: spacing.sm }}>
            <Button
              testID="validation-confirm-btn"
              title="Confirmer"
              variant="primary"
              fullWidth
              onPress={() => handleResolve('confirm')}
              disabled={resolving}
            />
            <Button
              testID="validation-correct-btn"
              title="Corriger"
              variant="outline"
              fullWidth
              onPress={() => handleResolve('correct')}
              disabled={resolving}
            />
            <RNView style={{ flexDirection: 'row', gap: spacing.sm }}>
              <RNView style={{ flex: 1 }}>
                <Button
                  testID="validation-ignore-btn"
                  title="Ignorer"
                  variant="outline"
                  fullWidth
                  onPress={() => handleResolve('ignore')}
                  disabled={resolving}
                />
              </RNView>
              <RNView style={{ flex: 1 }}>
                <Button
                  testID="validation-skip-btn"
                  title="NSP"
                  variant="outline"
                  fullWidth
                  onPress={() => handleResolve('unknown')}
                  disabled={resolving}
                />
              </RNView>
            </RNView>
          </RNView>
        </RNView>
      </RNView>
    );
  }

  // --- List ---
  return (
    <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top }]}>
      <RNView style={styles.header}>
        <Pressable onPress={() => router.back()}>
          <RNText style={[typography.body, { color: colors.tint }]}>← Retour</RNText>
        </Pressable>
        <RNText style={[typography.h3, { color: colors.text }]}>Validation ({tasks.length})</RNText>
        <RNView style={{ width: 60 }} />
      </RNView>

      <FlatList
        testID="validation-task-list"
        data={tasks}
        keyExtractor={(t) => t.id}
        contentContainerStyle={{ padding: spacing.md, gap: spacing.sm }}
        renderItem={({ item }) => {
          const confColor =
            item.priority >= 7 ? colors.success :
            item.priority >= 4 ? colors.warning : colors.error;

          return (
            <Pressable onPress={() => { setSelected(item); setView('detail'); }}>
              <Card>
                <RNView style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
                  <RNView style={{ flex: 1 }}>
                    <RNText style={[typography.bodyBold, { color: colors.text }]}>
                      {item.item_term || 'Item #' + item.item_id.slice(0, 8)}
                    </RNText>
                    <RNView style={{ flexDirection: 'row', gap: spacing.xs, marginTop: spacing.xs }}>
                      <RNView style={[styles.badge, { backgroundColor: colors.tintLight }]}>
                        <RNText style={[typography.small, { color: colors.tint }]}>{item.source}</RNText>
                      </RNView>
                    </RNView>
                  </RNView>
                  <RNView style={[styles.priorityDot, { backgroundColor: confColor }]} />
                </RNView>
              </Card>
            </Pressable>
          );
        }}
      />
    </RNView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center', padding: spacing.lg },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
  },
  badge: {
    paddingHorizontal: spacing.sm,
    paddingVertical: 2,
    borderRadius: radius.sm,
  },
  priorityDot: {
    width: 12,
    height: 12,
    borderRadius: 6,
  },
});
