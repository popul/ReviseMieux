import { useState, useCallback, useLayoutEffect } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import {
  listerCours,
  genererQuiz,
  demarrerSession,
  repondreQuestion,
  terminerSession,
  type Cours,
  type Quiz as QuizType,
  type SessionQuiz,
} from '../services/api'
import ProcessingSection from '../components/ProcessingSection'
import ConfigurateurQuiz from '../components/ConfigurateurQuiz'
import QuestionQuiz from '../components/QuestionQuiz'
import FeedbackReponse from '../components/FeedbackReponse'
import ProgressionQuiz from '../components/ProgressionQuiz'
import ResultatsQuiz from '../components/ResultatsQuiz'

type EtatQuiz =
  | 'selection-cours'
  | 'configuration'
  | 'generation'
  | 'question'
  | 'feedback'
  | 'resultats'
  | 'erreur'

interface EtatFeedback {
  estCorrecte: boolean
  explication: string
}

// Hook pour charger les cours
function useChargementCours(coursId: string | null) {
  const [listeCours, setListeCours] = useState<Cours[]>([])
  const [chargement, setChargement] = useState(() => !coursId)
  const [erreur, setErreur] = useState<string | null>(null)

  useLayoutEffect(() => {
    if (coursId) return

    let cancelled = false

    listerCours(1, 50)
      .then((res) => {
        if (!cancelled) {
          setListeCours(res.cours)
          setChargement(false)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setErreur('Impossible de charger la liste des cours')
          setChargement(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [coursId])

  return { listeCours, chargement, erreur }
}

export default function Quiz() {
  const [searchParams, setSearchParams] = useSearchParams()
  const coursId = searchParams.get('cours')

  const { listeCours, chargement: chargementCours, erreur: erreurCours } = useChargementCours(coursId)

  // Etat du quiz
  const [etat, setEtat] = useState<EtatQuiz>(coursId ? 'configuration' : 'selection-cours')
  const [quiz, setQuiz] = useState<QuizType | null>(null)
  const [session, setSession] = useState<SessionQuiz | null>(null)
  const [questionIndex, setQuestionIndex] = useState(0)
  const [choixSelectionne, setChoixSelectionne] = useState<number | null>(null)
  const [feedback, setFeedback] = useState<EtatFeedback | null>(null)
  const [bonnesReponses, setBonnesReponses] = useState(0)
  const [erreur, setErreur] = useState<string | null>(null)

  // Configuration
  const [nombreQuestions, setNombreQuestions] = useState(10)
  const [difficulte, setDifficulte] = useState<'facile' | 'moyen' | 'difficile'>('moyen')
  const [chargementGeneration, setChargementGeneration] = useState(false)

  // Reinitialiser quand coursId change
  const [prevCoursId, setPrevCoursId] = useState(coursId)
  if (coursId !== prevCoursId) {
    setPrevCoursId(coursId)
    setEtat(coursId ? 'configuration' : 'selection-cours')
    setQuiz(null)
    setSession(null)
    setQuestionIndex(0)
    setChoixSelectionne(null)
    setFeedback(null)
    setBonnesReponses(0)
    setErreur(null)
  }

  // Lancer le quiz
  const lancerQuiz = useCallback(async () => {
    if (!coursId) return

    setChargementGeneration(true)
    setEtat('generation')
    setErreur(null)

    try {
      // Generer le quiz
      const resQuiz = await genererQuiz(coursId, {
        nombreQuestions,
        difficulte,
      })

      if (!resQuiz.succes || !resQuiz.quiz) {
        throw new Error(resQuiz.erreur?.message || 'Erreur lors de la generation du quiz')
      }

      setQuiz(resQuiz.quiz)

      // Demarrer la session
      const resSession = await demarrerSession(resQuiz.quiz.id)

      if (!resSession.succes || !resSession.session) {
        throw new Error(resSession.erreur?.message || 'Erreur lors du demarrage de la session')
      }

      setSession(resSession.session)
      setQuestionIndex(0)
      setChoixSelectionne(null)
      setBonnesReponses(0)
      setEtat('question')
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Une erreur est survenue')
      setEtat('erreur')
    } finally {
      setChargementGeneration(false)
    }
  }, [coursId, nombreQuestions, difficulte])

  // Valider une reponse
  const validerReponse = useCallback(async () => {
    if (!quiz || !session || choixSelectionne === null) return

    const question = quiz.questions[questionIndex]

    try {
      const res = await repondreQuestion(
        quiz.id,
        session.id,
        question.id,
        choixSelectionne
      )

      if (!res.succes) {
        throw new Error(res.erreur?.message || 'Erreur lors de la validation')
      }

      if (res.estCorrecte) {
        setBonnesReponses((prev) => prev + 1)
      }

      setFeedback({
        estCorrecte: res.estCorrecte,
        explication: res.explication || question.explication,
      })
      setEtat('feedback')
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Une erreur est survenue')
      setEtat('erreur')
    }
  }, [quiz, session, questionIndex, choixSelectionne])

  // Passer a la question suivante
  const questionSuivante = useCallback(async () => {
    if (!quiz || !session) return

    const estDerniere = questionIndex >= quiz.questions.length - 1

    if (estDerniere) {
      // Terminer le quiz
      try {
        const res = await terminerSession(quiz.id, session.id)

        if (!res.succes || !res.session) {
          throw new Error(res.erreur?.message || 'Erreur lors de la finalisation')
        }

        setSession(res.session)
        setEtat('resultats')
      } catch (err) {
        setErreur(err instanceof Error ? err.message : 'Une erreur est survenue')
        setEtat('erreur')
      }
    } else {
      // Passer a la question suivante
      setQuestionIndex((prev) => prev + 1)
      setChoixSelectionne(null)
      setFeedback(null)
      setEtat('question')
    }
  }, [quiz, session, questionIndex])

  // Refaire le quiz (meme quiz, nouvelle session)
  const refaireQuiz = useCallback(async () => {
    if (!quiz) return

    setEtat('generation')
    setErreur(null)

    try {
      const resSession = await demarrerSession(quiz.id)

      if (!resSession.succes || !resSession.session) {
        throw new Error(resSession.erreur?.message || 'Erreur lors du redemarrage')
      }

      setSession(resSession.session)
      setQuestionIndex(0)
      setChoixSelectionne(null)
      setFeedback(null)
      setBonnesReponses(0)
      setEtat('question')
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Une erreur est survenue')
      setEtat('erreur')
    }
  }, [quiz])

  // Nouveau quiz (retour a la configuration)
  const nouveauQuiz = useCallback(() => {
    setQuiz(null)
    setSession(null)
    setQuestionIndex(0)
    setChoixSelectionne(null)
    setFeedback(null)
    setBonnesReponses(0)
    setErreur(null)
    setEtat('configuration')
  }, [])

  // Selection d'un cours
  const selectionnerCours = useCallback((id: string) => {
    setSearchParams({ cours: id })
  }, [setSearchParams])

  // Page de selection de cours
  if (etat === 'selection-cours') {
    if (chargementCours) {
      return (
        <div className="flex items-center justify-center min-h-[400px]">
          <ProcessingSection message="Chargement des cours..." />
        </div>
      )
    }

    if (erreurCours || listeCours.length === 0) {
      return (
        <div className="text-center py-12">
          <div className="text-5xl mb-6">
            {String.fromCodePoint(0x1f4da)}
          </div>
          <h1 className="font-display text-2xl font-semibold text-ink mb-4">
            {erreurCours || 'Aucun cours disponible'}
          </h1>
          <p className="text-ink-light mb-8">
            Scannez d'abord un cours pour generer un quiz.
          </p>
          <Link
            to="/scanner"
            className="inline-flex items-center gap-4 px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
          >
            Scanner un cours
          </Link>
        </div>
      )
    }

    return (
      <div>
        <h1 className="font-display text-3xl font-semibold text-ink mb-8">
          Quiz interactif
        </h1>
        <p className="text-ink-light mb-8">
          Selectionnez un cours pour lancer un quiz.
        </p>

        <div className="grid gap-6">
          {listeCours.map((c) => (
            <button
              key={c.id}
              onClick={() => selectionnerCours(c.id)}
              className="bg-white rounded-lg p-8 shadow-sm hover:shadow-md transition-shadow border border-cream-dark text-left"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h2 className="font-display text-lg font-semibold text-ink mb-2">
                    {c.titre || 'Cours sans titre'}
                  </h2>
                  <p className="text-sm text-ink-muted">{c.matiere || 'Matiere non definie'}</p>
                </div>
                <span className="text-coral font-medium">Lancer un quiz →</span>
              </div>
            </button>
          ))}
        </div>
      </div>
    )
  }

  // Page d'erreur
  if (etat === 'erreur') {
    return (
      <div className="text-center py-12">
        <div className="text-5xl mb-6">
          {String.fromCodePoint(0x1f615)}
        </div>
        <h1 className="font-display text-2xl font-semibold text-ink mb-4">
          Une erreur est survenue
        </h1>
        <p className="text-ink-light mb-8">{erreur}</p>
        <div className="flex gap-6 justify-center">
          <button
            onClick={nouveauQuiz}
            className="px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
          >
            Reessayer
          </button>
          <Link
            to="/quiz"
            className="px-8 py-3 border-2 border-coral text-coral rounded-full font-medium hover:bg-coral/5 transition-colors"
          >
            Retour aux cours
          </Link>
        </div>
      </div>
    )
  }

  // Page de configuration
  if (etat === 'configuration') {
    return (
      <div>
        <div className="mb-8">
          <Link
            to="/quiz"
            className="flex items-center gap-2 text-ink-light hover:text-ink transition-colors"
          >
            ← Changer de cours
          </Link>
        </div>

        <h1 className="font-display text-3xl font-semibold text-ink mb-8 text-center">
          Preparez votre quiz
        </h1>

        <ConfigurateurQuiz
          nombreQuestions={nombreQuestions}
          difficulte={difficulte}
          onNombreChange={setNombreQuestions}
          onDifficulteChange={setDifficulte}
          onLancer={lancerQuiz}
          chargement={chargementGeneration}
        />
      </div>
    )
  }

  // Page de generation
  if (etat === 'generation') {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <ProcessingSection message="Generation du quiz en cours..." />
      </div>
    )
  }

  // Question courante
  if (etat === 'question' && quiz) {
    const question = quiz.questions[questionIndex]

    return (
      <div>
        <ProgressionQuiz
          questionActuelle={questionIndex}
          totalQuestions={quiz.questions.length}
          bonnesReponses={bonnesReponses}
        />

        <QuestionQuiz
          question={question}
          choixSelectionne={choixSelectionne}
          onSelectChoix={setChoixSelectionne}
          onValider={validerReponse}
        />
      </div>
    )
  }

  // Feedback
  if (etat === 'feedback' && quiz && feedback) {
    const question = quiz.questions[questionIndex]
    const estDerniere = questionIndex >= quiz.questions.length - 1

    return (
      <div>
        <ProgressionQuiz
          questionActuelle={questionIndex}
          totalQuestions={quiz.questions.length}
          bonnesReponses={bonnesReponses}
        />

        <FeedbackReponse
          question={question}
          choixSelectionne={choixSelectionne!}
          estCorrecte={feedback.estCorrecte}
          explication={feedback.explication}
          onSuivant={questionSuivante}
          estDerniere={estDerniere}
        />
      </div>
    )
  }

  // Resultats
  if (etat === 'resultats' && quiz && session && coursId) {
    return (
      <div>
        <h1 className="font-display text-3xl font-semibold text-ink mb-8 text-center">
          Quiz termine !
        </h1>

        <ResultatsQuiz
          score={session.score || 0}
          totalQuestions={quiz.questions.length}
          bonnesReponses={bonnesReponses}
          coursId={coursId}
          onRefaire={refaireQuiz}
          onNouveauQuiz={nouveauQuiz}
        />
      </div>
    )
  }

  // Fallback
  return null
}
