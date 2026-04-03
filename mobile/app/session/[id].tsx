import { useState, useEffect } from 'react';
import {
  StyleSheet,
  ScrollView,
  TextInput,
  Pressable,
  KeyboardAvoidingView,
  Platform,
  ActivityIndicator,
} from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useLocalSearchParams, router } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { useColors, Card, Button, ProgressBar } from '@/components/Themed';
import { MasteryBadge } from '@/components/MasteryBar';
import { typography, spacing, radius } from '@/constants/Typography';
import {
  startDailySession,
  getQuestions,
  submitAnswer,
  getDebrief,
  getLessonCard,
} from '@/services/api';
import type { Question, SubmitAnswerResponse, DebriefResponse, ItemResponse } from '@/services/api';

// --- Types ---

type ViewState = 'loading' | 'error' | 'question' | 'self-score' | 'feedback' | 'debrief';

export default function SessionScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const colors = useColors();
  const insets = useSafeAreaInsets();

  const [sessionId, setSessionId] = useState<string | null>(null);
  const [questions, setQuestions] = useState<Question[]>([]);
  const [currentIdx, setCurrentIdx] = useState(0);
  const [view, setView] = useState<ViewState>('loading');
  const [error, setError] = useState<string | null>(null);
  const [answer, setAnswer] = useState('');
  const [score, setScore] = useState(0);
  const [startTime] = useState(Date.now());
  const [submitting, setSubmitting] = useState(false);
  const [lastFeedback, setLastFeedback] = useState<SubmitAnswerResponse | null>(null);
  const [debriefData, setDebriefData] = useState<DebriefResponse | null>(null);
  const [itemsMap, setItemsMap] = useState<Record<string, ItemResponse>>({});
  const [showLessonModal, setShowLessonModal] = useState(false);

  const question = questions[currentIdx];
  const total = questions.length;

  // Load session, questions, and lesson card items on mount
  useEffect(() => {
    if (!id) return;

    // Load lesson card items for "Voir cours" modal (Z1-AC21)
    getLessonCard(id)
      .then((lc) => {
        const map: Record<string, ItemResponse> = {};
        lc.items.forEach((item) => { map[item.id] = item; });
        setItemsMap(map);
      })
      .catch(() => {}); // non-blocking

    startDailySession(id)
      .then((session) => {
        setSessionId(session.id);
        return getQuestions(session.id);
      })
      .then((qs) => {
        if (qs.length === 0) {
          setError('Aucune question disponible pour ce chapitre.');
          setView('error');
        } else {
          setQuestions(qs);
          setView('question');
        }
      })
      .catch((err) => {
        setError(err.message ?? 'Impossible de lancer la session.');
        setView('error');
      });
  }, [id]);

  const handleSelfScore = (userScore: number) => {
    if (!sessionId || !question) return;
    setSubmitting(true);
    submitAnswer(sessionId, question.id, answer, userScore)
      .then((res) => {
        setLastFeedback(res);
        if (userScore >= 0.5) setScore((s) => s + 1);
        setView('feedback');
      })
      .catch((err) => {
        setError(err.message ?? 'Erreur lors de la soumission.');
        setView('error');
      })
      .finally(() => setSubmitting(false));
  };

  const handleNext = () => {
    if (currentIdx + 1 >= total) {
      // Fetch debrief
      if (!sessionId) return;
      setView('loading');
      getDebrief(sessionId)
        .then((data) => {
          setDebriefData(data);
          setView('debrief');
        })
        .catch(() => {
          // Fallback debrief from local data
          setDebriefData({
            score,
            total,
            percentage: Math.round((score / total) * 100),
            transitions: [],
          });
          setView('debrief');
        });
      return;
    }
    setCurrentIdx((i) => i + 1);
    setAnswer('');
    setLastFeedback(null);
    setView('question');
  };

  // --- Loading ---
  if (view === 'loading') {
    return (
      <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top, justifyContent: 'center', alignItems: 'center' }]}>
        <ActivityIndicator size="large" color={colors.tint} />
        <RNText style={[typography.body, { color: colors.textSecondary, marginTop: spacing.md }]}>
          Chargement de la session...
        </RNText>
      </RNView>
    );
  }

  // --- Error ---
  if (view === 'error') {
    return (
      <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top, justifyContent: 'center', alignItems: 'center', paddingHorizontal: spacing.lg }]}>
        <RNText style={{ fontSize: 48 }}>😕</RNText>
        <RNText style={[typography.h3, { color: colors.text, textAlign: 'center', marginTop: spacing.md }]}>
          {error}
        </RNText>
        <Button
          title="Retour"
          variant="outline"
          onPress={() => router.back()}
          style={{ marginTop: spacing.lg }}
        />
      </RNView>
    );
  }

  // --- Debrief ---
  if (view === 'debrief' && debriefData) {
    const elapsed = Math.round((Date.now() - startTime) / 1000);
    return (
      <DebriefView
        debrief={debriefData}
        elapsed={elapsed}
        colors={colors}
        insets={insets}
        chapterId={id!}
      />
    );
  }

  // --- Feedback ---
  if (view === 'feedback' && lastFeedback) {
    const wasCorrect = lastFeedback.score >= 0.5;
    return (
      <FeedbackView
        question={question}
        feedback={lastFeedback}
        isCorrect={wasCorrect}
        answer={answer}
        currentIdx={currentIdx}
        total={total}
        colors={colors}
        insets={insets}
        onNext={handleNext}
      />
    );
  }

  // --- Self-score (user typed answer, now self-assesses) ---
  if (view === 'self-score') {
    return (
      <KeyboardAvoidingView
        style={{ flex: 1 }}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      >
        <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top }]}>
          {/* Header */}
          <RNView style={styles.sessionHeader}>
            <Pressable testID="session-quit-btn" onPress={() => router.back()}>
              <RNText style={[typography.body, { color: colors.tint }]}>← Quitter</RNText>
            </Pressable>
            <RNText style={[typography.captionBold, { color: colors.text }]}>Session</RNText>
            <RNText style={[typography.caption, { color: colors.textSecondary }]}>
              Q {currentIdx + 1}/{total}
            </RNText>
          </RNView>
          <RNView style={{ paddingHorizontal: spacing.md }}>
            <ProgressBar progress={(currentIdx + 1) / total} height={4} />
          </RNView>

          <ScrollView style={styles.content} contentContainerStyle={{ paddingBottom: spacing.xxl }}>
            <RNText style={[typography.h3, { color: colors.text }]}>
              {question.rendered_prompt}
            </RNText>

            <RNView style={{ marginTop: spacing.lg }}>
              <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>Ta reponse :</RNText>
              <Card style={{ marginTop: spacing.xs }}>
                <RNText style={[typography.body, { color: colors.text }]}>{answer || '—'}</RNText>
              </Card>
            </RNView>

            {/* Show expected answer from item so student can self-assess */}
            {itemsMap[question.item_id] && (
              <RNView style={{ marginTop: spacing.md }}>
                <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>Reponse attendue :</RNText>
                <Card style={{ marginTop: spacing.xs, backgroundColor: colors.tintLight }}>
                  <RNText style={[typography.bodyBold, { color: colors.tint }]}>
                    {itemsMap[question.item_id].term}
                  </RNText>
                </Card>
              </RNView>
            )}

            <RNText style={[typography.bodyBold, { color: colors.text, textAlign: 'center', marginTop: spacing.xl }]}>
              Compare et juge : avais-tu bon ?
            </RNText>
          </ScrollView>

          {/* Self-score buttons */}
          <RNView style={[styles.bottomActions, { borderTopColor: colors.border, paddingBottom: insets.bottom + spacing.sm }]}>
            <RNView style={{ flexDirection: 'row', gap: spacing.sm }}>
              <RNView style={{ flex: 1 }}>
                <Button
                  title="J'avais faux"
                  variant="outline"
                  fullWidth
                  onPress={() => handleSelfScore(0.0)}
                  disabled={submitting}
                />
              </RNView>
              <RNView style={{ flex: 1 }}>
                <Button
                  title="J'avais bon"
                  variant="primary"
                  fullWidth
                  onPress={() => handleSelfScore(1.0)}
                  disabled={submitting}
                />
              </RNView>
            </RNView>
            {submitting && (
              <ActivityIndicator size="small" color={colors.tint} style={{ marginTop: spacing.sm }} />
            )}
          </RNView>
        </RNView>
      </KeyboardAvoidingView>
    );
  }

  // --- Question ---
  return (
    <KeyboardAvoidingView
      style={{ flex: 1 }}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top }]}>
        {/* Header */}
        <RNView style={styles.sessionHeader}>
          <Pressable testID="session-quit-btn" onPress={() => router.back()}>
            <RNText style={[typography.body, { color: colors.tint }]}>← Quitter</RNText>
          </Pressable>
          <RNText style={[typography.captionBold, { color: colors.text }]}>Session</RNText>
          <RNText style={[typography.caption, { color: colors.textSecondary }]}>
            Q {currentIdx + 1}/{total}
          </RNText>
        </RNView>
        <RNView style={{ paddingHorizontal: spacing.md }}>
          <ProgressBar testID="session-progress-bar" progress={(currentIdx + 1) / total} height={4} />
        </RNView>

        {/* Question content */}
        <ScrollView style={styles.content} contentContainerStyle={{ paddingBottom: spacing.xxl }}>
          <RNText testID="session-question-text" style={[typography.h3, { color: colors.text, marginTop: spacing.lg }]}>
            {question.rendered_prompt}
          </RNText>

          {/* Answer input — MCQ choices or text */}
          <RNView style={{ marginTop: spacing.xl }}>
            {question.choices && question.choices.length > 0 ? (
              <RNView style={{ gap: spacing.sm }}>
                {question.choices.map((choice, idx) => {
                  const isSelected = answer === choice;
                  return (
                    <Pressable
                      key={idx}
                      onPress={() => setAnswer(choice)}
                      style={[
                        styles.choiceBtn,
                        {
                          backgroundColor: isSelected ? colors.tintLight : colors.backgroundSecondary,
                          borderColor: isSelected ? colors.tint : colors.border,
                        },
                      ]}
                    >
                      <RNText style={[typography.bodyBold, { color: isSelected ? colors.tint : colors.text }]}>
                        {String.fromCharCode(65 + idx)}
                      </RNText>
                      <RNText style={[typography.body, { color: colors.text, flex: 1 }]}>
                        {choice}
                      </RNText>
                    </Pressable>
                  );
                })}
              </RNView>
            ) : (
              <TextInput
                testID="session-answer-input"
                style={[
                  styles.textInput,
                  {
                    backgroundColor: colors.backgroundSecondary,
                    borderColor: colors.border,
                    color: colors.text,
                  },
                ]}
                placeholder="Ta reponse..."
                placeholderTextColor={colors.textSecondary}
                value={answer}
                onChangeText={setAnswer}
                multiline
                autoFocus
              />
            )}
          </RNView>
        </ScrollView>

        {/* Bottom actions */}
        <RNView style={[styles.bottomActions, { borderTopColor: colors.border, paddingBottom: insets.bottom + spacing.sm }]}>
          <Button
            testID="session-validate-btn"
            title="Valider"
            variant="primary"
            fullWidth
            onPress={() => {
              if (question.choices && question.choices.length > 0) {
                // MCQ: auto-score by comparing answer to expected (choices[0] = correct)
                const autoScore = answer === question.choices[0] ? 1.0 : 0.0;
                handleSelfScore(autoScore);
              } else {
                // Text: backend will auto-score via LLM
                handleSelfScore(0.0);
              }
            }}
            disabled={answer.trim().length === 0}
          />
          <RNView style={styles.secondaryActions}>
            <Pressable onPress={() => setShowLessonModal(true)}>
              <RNText style={[typography.caption, { color: colors.tint }]}>Voir cours</RNText>
            </Pressable>
          </RNView>
        </RNView>
        {/* Z1-AC21: Lesson card modal */}
        {showLessonModal && question && (
          <LessonModal
            item={itemsMap[question.item_id]}
            colors={colors}
            insets={insets}
            onClose={() => setShowLessonModal(false)}
          />
        )}
      </RNView>
    </KeyboardAvoidingView>
  );
}

// --- Lesson Modal (Z1-AC21) ---

function LessonModal({
  item,
  colors,
  insets,
  onClose,
}: {
  item?: ItemResponse;
  colors: any;
  insets: { top: number; bottom: number };
  onClose: () => void;
}) {
  return (
    <RNView style={[styles.modalOverlay, { paddingTop: insets.top + spacing.xl, paddingBottom: insets.bottom + spacing.lg }]}>
      <RNView style={[styles.modalContent, { backgroundColor: colors.background }]}>
        <RNView style={styles.modalHeader}>
          <RNText style={[typography.h3, { color: colors.text, flex: 1 }]}>Extrait du cours</RNText>
          <Pressable onPress={onClose}>
            <RNText style={[typography.h3, { color: colors.textSecondary }]}>✕</RNText>
          </Pressable>
        </RNView>

        <ScrollView style={{ marginTop: spacing.md }}>
          {item ? (
            <>
              <Card style={{ marginBottom: spacing.md }}>
                <RNText style={[typography.captionBold, { color: colors.textSecondary, marginBottom: spacing.xs }]}>
                  {item.item_type}
                </RNText>
                <RNText style={[typography.body, { color: colors.text }]}>
                  {item.term ?? 'Contenu non disponible'}
                </RNText>
              </Card>

              {item.keywords && item.keywords.length > 0 && (
                <RNView>
                  <RNText style={[typography.captionBold, { color: colors.textSecondary, marginBottom: spacing.xs }]}>
                    Mots-cles
                  </RNText>
                  <RNView style={{ flexDirection: 'row', flexWrap: 'wrap', gap: spacing.xs }}>
                    {item.keywords.map((kw, i) => (
                      <RNView key={i} style={[styles.keywordChip, { backgroundColor: colors.tintLight }]}>
                        <RNText style={[typography.small, { color: colors.tint }]}>{kw}</RNText>
                      </RNView>
                    ))}
                  </RNView>
                </RNView>
              )}
            </>
          ) : (
            <RNText style={[typography.body, { color: colors.textSecondary, textAlign: 'center' }]}>
              Contenu non disponible pour cette question.
            </RNText>
          )}
        </ScrollView>

        <Button
          title="Fermer"
          variant="outline"
          fullWidth
          onPress={onClose}
          style={{ marginTop: spacing.lg }}
        />
      </RNView>
    </RNView>
  );
}

// --- Feedback View ---

function FeedbackView({
  question,
  feedback,
  isCorrect,
  answer,
  currentIdx,
  total,
  colors,
  insets,
  onNext,
}: {
  question: Question;
  feedback: SubmitAnswerResponse;
  isCorrect: boolean;
  answer: string;
  currentIdx: number;
  total: number;
  colors: any;
  insets: { top: number; bottom: number };
  onNext: () => void;
}) {
  return (
    <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top }]}>
      {/* Header */}
      <RNView style={styles.sessionHeader}>
        <RNView />
        <RNText style={[typography.captionBold, { color: colors.text }]}>Session</RNText>
        <RNText style={[typography.caption, { color: colors.textSecondary }]}>
          Q {currentIdx + 1}/{total}
        </RNText>
      </RNView>
      <RNView style={{ paddingHorizontal: spacing.md }}>
        <ProgressBar progress={(currentIdx + 1) / total} height={4} />
      </RNView>

      <ScrollView style={styles.content} contentContainerStyle={{ paddingBottom: spacing.xxl }}>
        {/* Result indicator */}
        <RNView style={[styles.resultBanner, { backgroundColor: isCorrect ? colors.successLight : colors.errorLight }]}>
          <RNText style={{ fontSize: 32 }}>{isCorrect ? '✅' : '❌'}</RNText>
          <RNText style={[typography.h2, { color: isCorrect ? colors.success : colors.error }]}>
            {isCorrect ? 'Correct !' : 'Pas tout à fait...'}
          </RNText>
        </RNView>

        {/* User answer */}
        <RNView style={{ marginTop: spacing.lg }}>
          <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>Ta reponse :</RNText>
          <Card style={{ marginTop: spacing.xs }}>
            <RNText style={[typography.body, { color: colors.text }]}>{answer || '—'}</RNText>
          </Card>
        </RNView>

        {/* Expected answer from feedback */}
        {feedback.feedback?.correct_answer && (
          <RNView style={{ marginTop: spacing.md }}>
            <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>Reponse attendue :</RNText>
            <RNText style={[typography.bodyBold, { color: colors.text, marginTop: spacing.xs }]}>
              {feedback.feedback.correct_answer}
            </RNText>
          </RNView>
        )}

        {/* What was missing */}
        {feedback.feedback?.what_was_missing && !isCorrect && (
          <RNView style={{ marginTop: spacing.md }}>
            <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>Ce qui manquait :</RNText>
            <RNText style={[typography.body, { color: colors.text, marginTop: spacing.xs }]}>
              {feedback.feedback.what_was_missing}
            </RNText>
          </RNView>
        )}

        {/* Hint */}
        {feedback.feedback?.hint && (
          <Card style={{ marginTop: spacing.lg, backgroundColor: colors.tintLight }}>
            <RNText style={[typography.body, { color: colors.tint }]}>
              💡 {feedback.feedback.hint}
            </RNText>
          </Card>
        )}
      </ScrollView>

      {/* Next button */}
      <RNView style={[styles.bottomActions, { borderTopColor: colors.border, paddingBottom: insets.bottom + spacing.sm }]}>
        <Button
          testID="session-next-btn"
          title={currentIdx + 1 >= total ? 'Voir le bilan' : 'Suivant →'}
          variant="primary"
          fullWidth
          onPress={onNext}
        />
      </RNView>
    </RNView>
  );
}

// --- Debrief View ---

function DebriefView({
  debrief,
  elapsed,
  colors,
  insets,
  chapterId,
}: {
  debrief: DebriefResponse;
  elapsed: number;
  colors: any;
  insets: { top: number; bottom: number };
  chapterId: string;
}) {
  const minutes = Math.floor(elapsed / 60);
  const seconds = elapsed % 60;

  return (
    <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top + spacing.lg }]}>
      <ScrollView contentContainerStyle={{ paddingBottom: insets.bottom + 100 }}>
        {/* Celebration */}
        <RNView style={{ alignItems: 'center', paddingHorizontal: spacing.lg }}>
          <RNText style={{ fontSize: 48 }}>🎉</RNText>
          <RNText style={[typography.h1, { color: colors.text, marginTop: spacing.md }]}>
            Bravo !
          </RNText>
        </RNView>

        {/* Score card */}
        <Card style={[styles.scoreCard, { marginTop: spacing.lg }]}>
          <RNText style={[typography.h1, { color: colors.tint, textAlign: 'center' }]}>
            {debrief.score} / {debrief.total}
          </RNText>
          <RNText style={[typography.h3, { color: colors.text, textAlign: 'center' }]}>
            {debrief.percentage}%
          </RNText>
          <RNText style={[typography.caption, { color: colors.textSecondary, textAlign: 'center', marginTop: spacing.xs }]}>
            ⏱ {minutes} min {seconds.toString().padStart(2, '0')}s
          </RNText>
        </Card>

        {/* Transitions */}
        {debrief.transitions.length > 0 && (
          <RNView style={styles.debriefSection}>
            <RNText style={[typography.captionBold, { color: colors.textSecondary, marginBottom: spacing.sm }]}>
              PROGRESSIONS
            </RNText>
            {debrief.transitions.map((t, idx) => {
              const isRegression =
                (t.from === 'solid' && t.to !== 'solid') ||
                (t.from === 'ok' && (t.to === 'fragile' || t.to === 'unknown'));
              const icon = isRegression ? '📉' : '📈';
              return (
                <RNView key={idx} style={styles.transitionRow}>
                  <RNText>{icon}</RNText>
                  <RNText style={[typography.body, { color: colors.text, flex: 1 }]}>
                    {t.item_term || `Item ${t.item_id.slice(0, 8)}`}
                  </RNText>
                  <MasteryBadge state={t.from as any} />
                  <RNText style={[typography.caption, { color: colors.textSecondary }]}>→</RNText>
                  <MasteryBadge state={t.to as any} />
                </RNView>
              );
            })}
          </RNView>
        )}
      </ScrollView>

      {/* Bottom actions */}
      <RNView style={[styles.bottomActions, { borderTopColor: colors.border, paddingBottom: insets.bottom + spacing.sm }]}>
        <Button
          title="Retour au chapitre"
          variant="outline"
          fullWidth
          onPress={() => router.replace(`/chapter/${chapterId}`)}
        />
        <Button
          title="Encore une session"
          variant="primary"
          fullWidth
          onPress={() => {
            router.replace(`/session/${chapterId}`);
          }}
          style={{ marginTop: spacing.sm }}
        />
      </RNView>
    </RNView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  sessionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
  },
  content: { flex: 1, paddingHorizontal: spacing.md, paddingTop: spacing.lg },
  choiceBtn: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    gap: spacing.md,
    padding: spacing.md,
    borderRadius: radius.md,
    borderWidth: 1.5,
  },
  textInput: {
    borderRadius: radius.md,
    borderWidth: 1,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    fontSize: 16,
    textAlignVertical: 'top',
    minHeight: 50,
  },
  bottomActions: {
    paddingHorizontal: spacing.md,
    paddingTop: spacing.md,
    borderTopWidth: 1,
  },
  secondaryActions: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginTop: spacing.sm,
    paddingHorizontal: spacing.xs,
  },
  resultBanner: {
    alignItems: 'center',
    padding: spacing.lg,
    borderRadius: radius.lg,
    gap: spacing.sm,
  },
  transitionRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    marginBottom: spacing.xs,
  },
  scoreCard: { marginHorizontal: spacing.md, alignItems: 'center' as const, paddingVertical: spacing.lg },
  debriefSection: { paddingHorizontal: spacing.md, marginTop: spacing.lg },
  modalOverlay: {
    position: 'absolute' as const,
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'center' as const,
    paddingHorizontal: spacing.md,
  },
  modalContent: {
    borderRadius: radius.lg,
    padding: spacing.lg,
    maxHeight: '80%' as any,
  },
  modalHeader: {
    flexDirection: 'row' as const,
    alignItems: 'center' as const,
    justifyContent: 'space-between' as const,
  },
  keywordChip: {
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
    borderRadius: radius.sm,
  },
});
