import { useState, useRef, useEffect } from 'react';
import {
  StyleSheet,
  ScrollView,
  TextInput,
  Pressable,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useLocalSearchParams, router } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { useColors, Card, Button, ProgressBar, Badge } from '@/components/Themed';
import { MasteryBadge } from '@/components/MasteryBar';
import { masteryColors } from '@/constants/Colors';
import { typography, spacing, radius } from '@/constants/Typography';

// --- Types ---

type QuestionType = 'MCQ' | 'SHORT_ANSWER' | 'NUMERIC' | 'KEYWORDS' | 'CLOZE' | 'RUBRIC';

type MockQuestion = {
  id: string;
  type: QuestionType;
  itemType: string;
  prompt: string;
  choices?: string[];
  unit?: string;
  expectedAnswer: string;
  hint: string;
  masteryFrom?: string;
  masteryTo?: string;
};

type ViewState = 'question' | 'feedback' | 'debrief';

// --- Mock questions ---

const MOCK_QUESTIONS: MockQuestion[] = [
  {
    id: 'q1',
    type: 'SHORT_ANSWER',
    itemType: 'CONNAISSANCES',
    prompt: 'Quelle est la formule de la masse volumique ?',
    expectedAnswer: 'ρ = m / V',
    hint: 'Pense à "rho" comme "ratio masse/volume".',
    masteryFrom: 'fragile',
    masteryTo: 'ok',
  },
  {
    id: 'q2',
    type: 'MCQ',
    itemType: 'CONNAISSANCES',
    prompt: 'Quelle est l\'unité SI de la masse volumique ?',
    choices: ['g/L', 'kg/m³', 'g/cm³', 'kg/L'],
    expectedAnswer: 'kg/m³',
    hint: 'L\'unité SI utilise les mètres cubes, pas les litres.',
  },
  {
    id: 'q3',
    type: 'NUMERIC',
    itemType: 'PROCÉDURE',
    prompt: 'Un objet de 150g occupe un volume de 50 cm³. Calcule sa masse volumique.',
    unit: 'g/cm³',
    expectedAnswer: '3 g/cm³',
    hint: 'ρ = m/V = 150/50',
    masteryFrom: 'unknown',
    masteryTo: 'fragile',
  },
  {
    id: 'q4',
    type: 'KEYWORDS',
    itemType: 'CONNAISSANCES',
    prompt: 'Cite les conditions pour qu\'un objet flotte dans un liquide.',
    expectedAnswer: 'densité inférieure, masse volumique plus faible que le liquide',
    hint: 'Compare la densité de l\'objet à celle du liquide.',
  },
  {
    id: 'q5',
    type: 'SHORT_ANSWER',
    itemType: 'CONNAISSANCES',
    prompt: 'Quel instrument mesure précisément un volume de liquide au laboratoire ?',
    expectedAnswer: 'Éprouvette graduée',
    hint: 'C\'est un tube gradué en verre.',
    masteryFrom: 'ok',
    masteryTo: 'solid',
  },
  {
    id: 'q6',
    type: 'MCQ',
    itemType: 'ANALYSE',
    prompt: 'Un glaçon flotte dans l\'eau. Que peut-on en déduire ?',
    choices: [
      'La glace est plus dense que l\'eau',
      'La glace est moins dense que l\'eau',
      'La glace a la même densité que l\'eau',
      'On ne peut rien déduire',
    ],
    expectedAnswer: 'La glace est moins dense que l\'eau',
    hint: 'Si ça flotte, c\'est que la densité est plus faible.',
    masteryFrom: 'fragile',
    masteryTo: 'ok',
  },
  {
    id: 'q7',
    type: 'CLOZE',
    itemType: 'CONNAISSANCES',
    prompt: 'La poussée d\'Archimède est égale au _____ du fluide déplacé.',
    expectedAnswer: 'poids',
    hint: 'Principe d\'Archimède : tout corps plongé...',
  },
  {
    id: 'q8',
    type: 'SHORT_ANSWER',
    itemType: 'PROCÉDURE',
    prompt: 'Comment mesurer la masse volumique d\'un solide irrégulier ?',
    expectedAnswer: 'Peser le solide puis mesurer son volume par déplacement d\'eau dans une éprouvette',
    hint: 'Utilise deux instruments : une balance et une éprouvette.',
  },
  {
    id: 'q9',
    type: 'NUMERIC',
    itemType: 'PROCÉDURE',
    prompt: 'Convertis 2,5 L en cm³.',
    unit: 'cm³',
    expectedAnswer: '2500 cm³',
    hint: '1 L = 1000 cm³',
    masteryFrom: 'fragile',
    masteryTo: 'ok',
  },
  {
    id: 'q10',
    type: 'MCQ',
    itemType: 'CONNAISSANCES',
    prompt: 'La masse volumique de l\'eau pure est :',
    choices: ['0,5 g/cm³', '1 g/cm³', '10 g/cm³', '100 g/cm³'],
    expectedAnswer: '1 g/cm³',
    hint: 'C\'est la valeur de référence.',
    masteryFrom: 'ok',
    masteryTo: 'solid',
  },
];

export default function SessionScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const colors = useColors();
  const insets = useSafeAreaInsets();

  const [currentIdx, setCurrentIdx] = useState(0);
  const [view, setView] = useState<ViewState>('question');
  const [answer, setAnswer] = useState('');
  const [selectedChoice, setSelectedChoice] = useState<string | null>(null);
  const [isCorrect, setIsCorrect] = useState(false);
  const [skipsLeft, setSkipsLeft] = useState(2);
  const [score, setScore] = useState(0);
  const [startTime] = useState(Date.now());

  const question = MOCK_QUESTIONS[currentIdx];
  const total = MOCK_QUESTIONS.length;

  const handleSubmit = () => {
    // Simple scoring simulation
    const correct = Math.random() > 0.3; // 70% success rate for demo
    setIsCorrect(correct);
    if (correct) setScore((s) => s + 1);
    setView('feedback');
  };

  const handleNext = () => {
    if (currentIdx + 1 >= total) {
      setView('debrief');
      return;
    }
    setCurrentIdx((i) => i + 1);
    setAnswer('');
    setSelectedChoice(null);
    setView('question');
  };

  const handleSkip = () => {
    if (skipsLeft <= 0) return;
    setSkipsLeft((s) => s - 1);
    handleNext();
  };

  if (view === 'debrief') {
    const elapsed = Math.round((Date.now() - startTime) / 1000);
    return (
      <DebriefView
        score={score}
        total={total}
        elapsed={elapsed}
        colors={colors}
        insets={insets}
        chapterId={id!}
      />
    );
  }

  if (view === 'feedback') {
    return (
      <FeedbackView
        question={question}
        isCorrect={isCorrect}
        answer={question.type === 'MCQ' ? selectedChoice ?? '' : answer}
        currentIdx={currentIdx}
        total={total}
        colors={colors}
        insets={insets}
        onNext={handleNext}
      />
    );
  }

  return (
    <KeyboardAvoidingView
      style={{ flex: 1 }}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top }]}>
        {/* Header */}
        <RNView style={styles.sessionHeader}>
          <Pressable onPress={() => router.back()}>
            <RNText style={[typography.body, { color: colors.tint }]}>← Quitter</RNText>
          </Pressable>
          <RNText style={[typography.captionBold, { color: colors.text }]}>Physique</RNText>
          <RNText style={[typography.caption, { color: colors.textSecondary }]}>
            Q {currentIdx + 1}/{total}
          </RNText>
        </RNView>
        <RNView style={{ paddingHorizontal: spacing.md }}>
          <ProgressBar progress={(currentIdx + 1) / total} height={4} />
        </RNView>

        {/* Question content */}
        <ScrollView style={styles.content} contentContainerStyle={{ paddingBottom: spacing.xxl }}>
          <Badge label={question.itemType} />

          <RNText style={[typography.h3, { color: colors.text, marginTop: spacing.lg }]}>
            {question.prompt}
          </RNText>

          {/* Input area based on question type */}
          <RNView style={{ marginTop: spacing.xl }}>
            {question.type === 'MCQ' && question.choices && (
              <RNView style={{ gap: spacing.sm }}>
                {question.choices.map((choice, idx) => {
                  const isSelected = selectedChoice === choice;
                  return (
                    <Pressable
                      key={idx}
                      onPress={() => setSelectedChoice(choice)}
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
            )}

            {(question.type === 'SHORT_ANSWER' || question.type === 'CLOZE' || question.type === 'KEYWORDS' || question.type === 'RUBRIC') && (
              <TextInput
                style={[
                  styles.textInput,
                  {
                    backgroundColor: colors.backgroundSecondary,
                    borderColor: colors.border,
                    color: colors.text,
                    minHeight: question.type === 'RUBRIC' ? 120 : 50,
                  },
                ]}
                placeholder="Ta réponse..."
                placeholderTextColor={colors.textSecondary}
                value={answer}
                onChangeText={setAnswer}
                multiline={question.type === 'RUBRIC'}
                autoFocus
              />
            )}

            {question.type === 'NUMERIC' && (
              <RNView>
                <TextInput
                  style={[
                    styles.textInput,
                    { backgroundColor: colors.backgroundSecondary, borderColor: colors.border, color: colors.text },
                  ]}
                  placeholder="Ta réponse..."
                  placeholderTextColor={colors.textSecondary}
                  value={answer}
                  onChangeText={setAnswer}
                  keyboardType="numeric"
                  autoFocus
                />
                {question.unit && (
                  <RNText style={[typography.caption, { color: colors.textSecondary, marginTop: spacing.xs }]}>
                    Unité : {question.unit}
                  </RNText>
                )}
              </RNView>
            )}
          </RNView>
        </ScrollView>

        {/* Bottom actions */}
        <RNView style={[styles.bottomActions, { borderTopColor: colors.border, paddingBottom: insets.bottom + spacing.sm }]}>
          <Button
            title="Valider ✓"
            variant="primary"
            fullWidth
            onPress={handleSubmit}
          />
          <RNView style={styles.secondaryActions}>
            <Pressable onPress={handleSkip} disabled={skipsLeft === 0}>
              <RNText style={[typography.caption, { color: skipsLeft > 0 ? colors.textSecondary : colors.border }]}>
                Passer ({skipsLeft}/2)
              </RNText>
            </Pressable>
            <Pressable onPress={() => router.push(`/chapter/${id}`)}>
              <RNText style={[typography.caption, { color: colors.tint }]}>Voir cours</RNText>
            </Pressable>
          </RNView>
        </RNView>
      </RNView>
    </KeyboardAvoidingView>
  );
}

// --- Feedback View ---

function FeedbackView({
  question,
  isCorrect,
  answer,
  currentIdx,
  total,
  colors,
  insets,
  onNext,
}: {
  question: MockQuestion;
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
        <RNText style={[typography.captionBold, { color: colors.text }]}>Physique</RNText>
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
          <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>Ta réponse :</RNText>
          <Card style={{ marginTop: spacing.xs }}>
            <RNText style={[typography.body, { color: colors.text }]}>{answer || '—'}</RNText>
          </Card>
        </RNView>

        {/* Expected answer */}
        <RNView style={{ marginTop: spacing.md }}>
          <RNText style={[typography.captionBold, { color: colors.textSecondary }]}>Réponse attendue :</RNText>
          <RNText style={[typography.bodyBold, { color: colors.text, marginTop: spacing.xs }]}>
            {question.expectedAnswer}
          </RNText>
        </RNView>

        {/* Hint */}
        <Card style={{ marginTop: spacing.lg, backgroundColor: colors.tintLight }}>
          <RNText style={[typography.body, { color: colors.tint }]}>
            💡 {question.hint}
          </RNText>
        </Card>

        {/* Mastery transition */}
        {question.masteryFrom && question.masteryTo && isCorrect && (
          <RNView style={[styles.transitionRow, { marginTop: spacing.lg }]}>
            <RNText style={{ fontSize: 16 }}>📈</RNText>
            <MasteryBadge state={question.masteryFrom as any} />
            <RNText style={[typography.body, { color: colors.textSecondary }]}>→</RNText>
            <MasteryBadge state={question.masteryTo as any} />
            <RNText style={[typography.caption, { color: colors.success, flex: 1 }]}>
              Tu progresses !
            </RNText>
          </RNView>
        )}
      </ScrollView>

      {/* Next button */}
      <RNView style={[styles.bottomActions, { borderTopColor: colors.border, paddingBottom: insets.bottom + spacing.sm }]}>
        <Button
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
  score,
  total,
  elapsed,
  colors,
  insets,
  chapterId,
}: {
  score: number;
  total: number;
  elapsed: number;
  colors: any;
  insets: { top: number; bottom: number };
  chapterId: string;
}) {
  const pct = Math.round((score / total) * 100);
  const minutes = Math.floor(elapsed / 60);
  const seconds = elapsed % 60;

  const transitions = [
    { term: 'Formule ρ=m/V', from: 'fragile', to: 'ok' },
    { term: 'Condition flotte/coule', from: 'ok', to: 'solid' },
    { term: 'Conversion L↔cm³', from: 'ok', to: 'fragile' },
  ];

  const toConsolidate = ['Poussée d\'Archimède', 'Protocole mesure volume'];

  return (
    <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top + spacing.lg }]}>
      <ScrollView contentContainerStyle={{ paddingBottom: insets.bottom + 100 }}>
        {/* Celebration */}
        <RNView style={{ alignItems: 'center', paddingHorizontal: spacing.lg }}>
          <RNText style={{ fontSize: 48 }}>🎉</RNText>
          <RNText style={[typography.h1, { color: colors.text, marginTop: spacing.md }]}>
            Bravo Hugo !
          </RNText>
        </RNView>

        {/* Score card */}
        <Card style={[styles.scoreCard, { marginTop: spacing.lg }]}>
          <RNText style={[typography.h1, { color: colors.tint, textAlign: 'center' }]}>
            {score} / {total}
          </RNText>
          <RNText style={[typography.h3, { color: colors.text, textAlign: 'center' }]}>
            {pct}%
          </RNText>
          <RNText style={[typography.caption, { color: colors.textSecondary, textAlign: 'center', marginTop: spacing.xs }]}>
            ⏱ {minutes} min {seconds.toString().padStart(2, '0')}s
          </RNText>
        </Card>

        {/* Transitions */}
        <RNView style={styles.debriefSection}>
          <RNText style={[typography.captionBold, { color: colors.textSecondary, marginBottom: spacing.sm }]}>
            PROGRESSIONS
          </RNText>
          {transitions.map((t, idx) => {
            const isUp = ['fragile', 'unknown'].includes(t.from) && ['ok', 'solid', 'fragile'].includes(t.to) && t.from !== t.to;
            const icon = t.to === 'fragile' && t.from === 'ok' ? '📉' : '📈';
            return (
              <RNView key={idx} style={styles.transitionRow}>
                <RNText>{icon}</RNText>
                <RNText style={[typography.body, { color: colors.text, flex: 1 }]}>{t.term}</RNText>
                <MasteryBadge state={t.from as any} />
                <RNText style={[typography.caption, { color: colors.textSecondary }]}>→</RNText>
                <MasteryBadge state={t.to as any} />
              </RNView>
            );
          })}
        </RNView>

        {/* To consolidate */}
        <RNView style={styles.debriefSection}>
          <RNText style={[typography.captionBold, { color: colors.textSecondary, marginBottom: spacing.sm }]}>
            À CONSOLIDER
          </RNText>
          {toConsolidate.map((item, idx) => (
            <RNView key={idx} style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm, marginBottom: spacing.xs }}>
              <RNText style={[typography.body, { color: colors.warning }]}>·</RNText>
              <RNText style={[typography.body, { color: colors.text }]}>{item}</RNText>
            </RNView>
          ))}
        </RNView>
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
            // Reset session — in real app, would call API
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
    flexDirection: 'row',
    alignItems: 'center',
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
});
