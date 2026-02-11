import { useState, useMemo, useCallback, useLayoutEffect } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import {
  listerCours,
  getTermesLexique,
  extraireTermesLexique,
  mettreAJourMaitrise,
  genererQuizVocabulaire,
  type Cours,
  type TermeLexique,
  type QuizVocabulaire,
  type QuestionVocabulaire,
} from '../services/api'
import ProcessingSection from '../components/ProcessingSection'

type ModeVue = 'lexique' | 'quiz' | 'flashcards'
type ModeQuiz = 'terme_vers_definition' | 'definition_vers_terme'
type TriOption = 'alphabetique' | 'maitrise' | 'categorie'

// ===== Composant Indicateur de Maitrise =====
function IndicateurMaitrise({ niveau, onChange }: { niveau: number; onChange?: (n: number) => void }) {
  return (
    <div className="flex gap-1.5" role="group" aria-label={`Maitrise: ${niveau} sur 5`}>
      {[1, 2, 3, 4, 5].map((i) => (
        <button
          key={i}
          onClick={() => onChange?.(i === niveau ? i - 1 : i)}
          className={`w-2.5 h-2.5 rounded-full transition-all ${
            i <= niveau ? 'bg-teal' : 'bg-gray-200'
          } ${onChange ? 'cursor-pointer hover:scale-125' : 'cursor-default'}`}
          aria-label={`Niveau ${i}`}
          title={`Maitrise ${i}/5`}
        />
      ))}
    </div>
  )
}

// ===== Composant Carte Terme =====
function CarteTerme({
  terme,
  onMaitriseChange,
}: {
  terme: TermeLexique
  onMaitriseChange: (id: string, maitrise: number) => void
}) {
  return (
    <div className="bg-white rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow border border-cream-dark">
      <div className="flex items-start justify-between mb-3">
        <h3 className="font-display text-lg font-semibold text-ink">{terme.terme}</h3>
        <IndicateurMaitrise niveau={terme.maitrise} onChange={(n) => onMaitriseChange(terme.id, n)} />
      </div>

      <p className="text-ink-light leading-relaxed mb-3">{terme.definition}</p>

      {terme.contexte && (
        <div className="text-sm text-ink-muted mb-2">
          <span className="font-medium">Contexte :</span> {terme.contexte}
        </div>
      )}

      {terme.exemple && (
        <div className="text-sm text-ink-muted italic mb-2">
          <span className="font-medium not-italic">Exemple :</span> {terme.exemple}
        </div>
      )}

      {terme.categorie && (
        <span className="inline-block text-xs font-semibold uppercase tracking-wide px-2.5 py-1 rounded-full bg-teal/10 text-teal">
          {terme.categorie}
        </span>
      )}
    </div>
  )
}

// ===== Vue Liste des Termes =====
function VueListeTermes({
  termes,
  recherche,
  setRecherche,
  tri,
  setTri,
  onMaitriseChange,
  onExtraire,
  extractionEnCours,
}: {
  termes: TermeLexique[]
  recherche: string
  setRecherche: (v: string) => void
  tri: TriOption
  setTri: (v: TriOption) => void
  onMaitriseChange: (id: string, maitrise: number) => void
  onExtraire: () => void
  extractionEnCours: boolean
}) {
  const termesFiltres = useMemo(() => {
    let result = [...termes]

    // Filtrage par recherche
    if (recherche) {
      const lower = recherche.toLowerCase()
      result = result.filter(
        (t) =>
          t.terme.toLowerCase().includes(lower) ||
          t.definition.toLowerCase().includes(lower) ||
          (t.categorie && t.categorie.toLowerCase().includes(lower))
      )
    }

    // Tri
    switch (tri) {
      case 'alphabetique':
        result.sort((a, b) => a.terme.localeCompare(b.terme))
        break
      case 'maitrise':
        result.sort((a, b) => a.maitrise - b.maitrise)
        break
      case 'categorie':
        result.sort((a, b) => (a.categorie || '').localeCompare(b.categorie || ''))
        break
    }

    return result
  }, [termes, recherche, tri])

  if (termes.length === 0 && !extractionEnCours) {
    return (
      <div className="text-center py-12">
        <div className="text-5xl mb-6">{"\u{1F4D6}"}</div>
        <h2 className="font-display text-2xl font-semibold text-ink mb-4">
          Aucun terme dans le lexique
        </h2>
        <p className="text-ink-light mb-8">
          Extrayez les termes cles de votre cours pour commencer.
        </p>
        <button
          onClick={onExtraire}
          className="inline-flex items-center justify-center gap-2 px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          Extraire les termes avec l'IA
        </button>
      </div>
    )
  }

  if (extractionEnCours) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <ProcessingSection message="Extraction des termes en cours..." />
      </div>
    )
  }

  return (
    <div>
      {/* Barre de recherche et tri */}
      <div className="flex flex-col sm:flex-row gap-4 mb-8">
        <div className="flex-1 relative">
          <input
            type="text"
            placeholder="Rechercher un terme..."
            value={recherche}
            onChange={(e) => setRecherche(e.target.value)}
            className="w-full px-4 py-3 pl-10 bg-white border border-cream-dark rounded-lg text-ink placeholder:text-ink-muted focus:outline-none focus:ring-2 focus:ring-coral/30 focus:border-coral"
          />
          <svg className="absolute left-3 top-3.5 w-5 h-5 text-ink-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>

        <select
          value={tri}
          onChange={(e) => setTri(e.target.value as TriOption)}
          className="px-4 py-3 bg-white border border-cream-dark rounded-lg text-ink focus:outline-none focus:ring-2 focus:ring-coral/30 focus:border-coral"
        >
          <option value="alphabetique">Alphabetique</option>
          <option value="maitrise">Par maitrise</option>
          <option value="categorie">Par categorie</option>
        </select>

        <button
          onClick={onExtraire}
          className="px-6 py-3 bg-coral text-white rounded-lg font-medium hover:bg-coral-dark transition-colors whitespace-nowrap"
        >
          Re-extraire
        </button>
      </div>

      <div className="text-sm text-ink-muted mb-4">
        {termesFiltres.length} terme{termesFiltres.length !== 1 ? 's' : ''} {recherche ? 'trouves' : 'au total'}
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {termesFiltres.map((terme) => (
          <CarteTerme key={terme.id} terme={terme} onMaitriseChange={onMaitriseChange} />
        ))}
      </div>
    </div>
  )
}

// ===== Vue Quiz Vocabulaire =====
function VueQuizVocabulaire({
  coursId,
  termes,
}: {
  coursId: string
  termes: TermeLexique[]
}) {
  const [mode, setMode] = useState<ModeQuiz>('terme_vers_definition')
  const [quiz, setQuiz] = useState<QuizVocabulaire | null>(null)
  const [questionIndex, setQuestionIndex] = useState(0)
  const [reponse, setReponse] = useState('')
  const [afficherReponse, setAfficherReponse] = useState(false)
  const [indiceIndex, setIndiceIndex] = useState(0)
  const [score, setScore] = useState(0)
  const [termine, setTermine] = useState(false)
  const [chargement, setChargement] = useState(false)
  const [erreur, setErreur] = useState<string | null>(null)
  const [resultats, setResultats] = useState<{question: QuestionVocabulaire; reponseEleve: string; correct: boolean}[]>([])

  const lancerQuiz = useCallback(async () => {
    setChargement(true)
    setErreur(null)
    setQuiz(null)
    setQuestionIndex(0)
    setReponse('')
    setAfficherReponse(false)
    setIndiceIndex(0)
    setScore(0)
    setTermine(false)
    setResultats([])

    try {
      const res = await genererQuizVocabulaire(coursId, mode)
      if (res.succes && res.quiz) {
        setQuiz(res.quiz)
      } else {
        setErreur(res.erreur?.message || 'Erreur lors de la generation du quiz')
      }
    } catch (err) {
      setErreur(err instanceof Error ? err.message : 'Erreur lors de la generation du quiz')
    } finally {
      setChargement(false)
    }
  }, [coursId, mode])

  const verifierReponse = useCallback(() => {
    if (!quiz) return
    const question = quiz.questions[questionIndex]
    const reponseNormalisee = reponse.trim().toLowerCase()
    const attendueNormalisee = question.reponseAttendue.trim().toLowerCase()

    // Verification simple: la reponse contient les mots cles de la reponse attendue
    const correct = reponseNormalisee.length > 3 &&
      (attendueNormalisee.includes(reponseNormalisee) || reponseNormalisee.includes(attendueNormalisee.substring(0, Math.min(20, attendueNormalisee.length))))

    if (correct) {
      setScore((s) => s + 1)
    }

    setResultats((r) => [...r, { question, reponseEleve: reponse, correct }])
    setAfficherReponse(true)

    // Mettre a jour la maitrise du terme
    const terme = termes.find((t) => t.terme === question.terme)
    if (terme) {
      const nouvelleMaitrise = correct ? Math.min(5, terme.maitrise + 1) : Math.max(0, terme.maitrise - 1)
      mettreAJourMaitrise(terme.id, nouvelleMaitrise).catch(() => {})
    }
  }, [quiz, questionIndex, reponse, termes])

  const questionSuivante = useCallback(() => {
    if (!quiz) return
    if (questionIndex + 1 >= quiz.questions.length) {
      setTermine(true)
    } else {
      setQuestionIndex((i) => i + 1)
      setReponse('')
      setAfficherReponse(false)
      setIndiceIndex(0)
    }
  }, [quiz, questionIndex])

  if (termes.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-ink-light">Extrayez d'abord les termes du cours pour lancer un quiz.</p>
      </div>
    )
  }

  // Ecran de selection du mode
  if (!quiz && !chargement) {
    return (
      <div className="max-w-lg mx-auto text-center py-8">
        <h2 className="font-display text-2xl font-semibold text-ink mb-8">Quiz vocabulaire</h2>

        <div className="space-y-4 mb-8">
          <button
            onClick={() => setMode('terme_vers_definition')}
            className={`w-full p-4 rounded-lg border-2 text-left transition-all ${
              mode === 'terme_vers_definition'
                ? 'border-coral bg-coral/5'
                : 'border-cream-dark hover:border-coral/50'
            }`}
          >
            <div className="font-semibold text-ink">Terme {"\u2192"} Definition</div>
            <div className="text-sm text-ink-muted mt-1">On vous montre le terme, devinez la definition</div>
          </button>

          <button
            onClick={() => setMode('definition_vers_terme')}
            className={`w-full p-4 rounded-lg border-2 text-left transition-all ${
              mode === 'definition_vers_terme'
                ? 'border-coral bg-coral/5'
                : 'border-cream-dark hover:border-coral/50'
            }`}
          >
            <div className="font-semibold text-ink">Definition {"\u2192"} Terme</div>
            <div className="text-sm text-ink-muted mt-1">On vous montre la definition, devinez le terme</div>
          </button>
        </div>

        {erreur && <p className="text-red-600 mb-4">{erreur}</p>}

        <button
          onClick={lancerQuiz}
          className="px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
        >
          Lancer le quiz
        </button>
      </div>
    )
  }

  if (chargement) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <ProcessingSection message="Generation du quiz..." />
      </div>
    )
  }

  // Ecran de resultats
  if (termine && quiz) {
    return (
      <div className="max-w-2xl mx-auto py-8">
        <div className="text-center mb-8">
          <div className="text-5xl mb-4">
            {score >= quiz.questions.length * 0.8 ? String.fromCodePoint(0x1F389) : score >= quiz.questions.length * 0.5 ? String.fromCodePoint(0x1F44D) : String.fromCodePoint(0x1F4AA)}
          </div>
          <h2 className="font-display text-2xl font-semibold text-ink mb-2">Quiz termine !</h2>
          <p className="text-xl text-ink-light">
            Score : <span className="font-bold text-coral">{score}</span> / {quiz.questions.length}
          </p>
        </div>

        <div className="space-y-4 mb-8">
          {resultats.map((r, i) => (
            <div
              key={i}
              className={`p-4 rounded-lg border-2 ${
                r.correct ? 'border-green-300 bg-green-50' : 'border-red-300 bg-red-50'
              }`}
            >
              <div className="flex items-center gap-2 mb-2">
                <span>{r.correct ? String.fromCodePoint(0x2705) : String.fromCodePoint(0x274C)}</span>
                <span className="font-semibold text-ink">{r.question.terme}</span>
              </div>
              {!r.correct && (
                <div className="text-sm text-ink-muted">
                  <div>Votre reponse : {r.reponseEleve || '(vide)'}</div>
                  <div className="text-teal font-medium mt-1">Reponse attendue : {r.question.reponseAttendue}</div>
                </div>
              )}
            </div>
          ))}
        </div>

        <div className="text-center">
          <button
            onClick={() => {
              setQuiz(null)
              setTermine(false)
            }}
            className="px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
          >
            Recommencer
          </button>
        </div>
      </div>
    )
  }

  // Ecran de question
  if (!quiz) return null
  const question = quiz.questions[questionIndex]

  return (
    <div className="max-w-2xl mx-auto py-8">
      {/* Progression */}
      <div className="mb-8">
        <div className="flex justify-between text-sm mb-2">
          <span className="text-ink-muted">Question {questionIndex + 1} / {quiz.questions.length}</span>
          <span className="font-semibold text-ink">Score : {score}</span>
        </div>
        <div className="h-1.5 bg-cream-dark rounded-full overflow-hidden">
          <div
            className="h-full bg-coral rounded-full transition-all duration-300"
            style={{ width: `${((questionIndex + 1) / quiz.questions.length) * 100}%` }}
          />
        </div>
      </div>

      {/* Question */}
      <div className="bg-white rounded-lg p-8 shadow-sm mb-6">
        <div className="text-sm text-ink-muted mb-2 uppercase tracking-wide">
          {mode === 'terme_vers_definition' ? 'Definissez ce terme' : 'Quel est ce terme ?'}
        </div>
        <h3 className="font-display text-xl font-semibold text-ink mb-6">{question.question}</h3>

        {!afficherReponse && (
          <>
            <textarea
              value={reponse}
              onChange={(e) => setReponse(e.target.value)}
              placeholder={mode === 'terme_vers_definition' ? 'Ecrivez la definition...' : 'Ecrivez le terme...'}
              className="w-full p-4 border border-cream-dark rounded-lg text-ink placeholder:text-ink-muted focus:outline-none focus:ring-2 focus:ring-coral/30 focus:border-coral resize-none"
              rows={3}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault()
                  verifierReponse()
                }
              }}
            />

            {/* Indices */}
            {question.indices && question.indices.length > 0 && indiceIndex < question.indices.length && (
              <div className="mt-4">
                {indiceIndex > 0 && (
                  <div className="space-y-2 mb-3">
                    {question.indices.slice(0, indiceIndex).map((indice, i) => (
                      <div key={i} className="text-sm text-gold bg-gold/10 p-2 rounded">
                        {String.fromCodePoint(0x1F4A1)} {indice}
                      </div>
                    ))}
                  </div>
                )}
                <button
                  onClick={() => setIndiceIndex((i) => i + 1)}
                  className="text-sm text-gold hover:text-gold/80 transition-colors"
                >
                  {String.fromCodePoint(0x1F4A1)} Afficher un indice ({indiceIndex}/{question.indices.length})
                </button>
              </div>
            )}

            <div className="mt-6 flex gap-4">
              <button
                onClick={verifierReponse}
                className="px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
              >
                Verifier
              </button>
              <button
                onClick={() => {
                  setReponse('')
                  setAfficherReponse(true)
                  setResultats((r) => [...r, { question, reponseEleve: '', correct: false }])
                }}
                className="px-8 py-3 border-2 border-ink-muted text-ink-muted rounded-full font-medium hover:bg-ink-muted/5 transition-colors"
              >
                Je ne sais pas
              </button>
            </div>
          </>
        )}

        {afficherReponse && (
          <div className="mt-4">
            <div className={`p-4 rounded-lg ${
              resultats[resultats.length - 1]?.correct
                ? 'bg-green-50 border border-green-300'
                : 'bg-red-50 border border-red-300'
            }`}>
              <div className="font-semibold mb-2">
                {resultats[resultats.length - 1]?.correct ? String.fromCodePoint(0x2705) + ' Correct !' : String.fromCodePoint(0x274C) + ' Incorrect'}
              </div>
              <div className="text-sm text-ink-light">
                <span className="font-medium">Reponse attendue :</span> {question.reponseAttendue}
              </div>
            </div>

            <button
              onClick={questionSuivante}
              className="mt-6 px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
            >
              {questionIndex + 1 >= quiz.questions.length ? 'Voir les resultats' : 'Question suivante'}
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

// ===== Vue Flashcards =====
function VueFlashcards({
  termes,
  onMaitriseChange,
}: {
  termes: TermeLexique[]
  onMaitriseChange: (id: string, maitrise: number) => void
}) {
  const [index, setIndex] = useState(0)
  const [retourne, setRetourne] = useState(false)
  const [termesLocaux, setTermesLocaux] = useState(termes)

  useLayoutEffect(() => {
    // Melanger les termes au montage
    const shuffled = [...termes]
    for (let i = shuffled.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1))
      ;[shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]]
    }
    setTermesLocaux(shuffled)
    setIndex(0)
    setRetourne(false)
  }, [termes])

  if (termes.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-ink-light">Extrayez d'abord les termes du cours pour utiliser les flashcards.</p>
      </div>
    )
  }

  const terme = termesLocaux[index]
  if (!terme) return null

  const allerPrecedent = () => {
    if (index > 0) {
      setIndex(index - 1)
      setRetourne(false)
    }
  }

  const allerSuivant = () => {
    if (index < termesLocaux.length - 1) {
      setIndex(index + 1)
      setRetourne(false)
    }
  }

  const marquerConnu = () => {
    const nouvelleMaitrise = Math.min(5, terme.maitrise + 1)
    onMaitriseChange(terme.id, nouvelleMaitrise)
    allerSuivant()
  }

  const marquerInconnu = () => {
    const nouvelleMaitrise = Math.max(0, terme.maitrise - 1)
    onMaitriseChange(terme.id, nouvelleMaitrise)
    allerSuivant()
  }

  return (
    <div className="max-w-xl mx-auto py-8">
      {/* Barre de progression */}
      <div className="mb-8">
        <div className="flex justify-between text-sm mb-2">
          <span className="text-ink-muted">Progression</span>
          <span className="font-semibold text-ink">{index + 1} / {termesLocaux.length}</span>
        </div>
        <div className="h-1.5 bg-cream-dark rounded-full overflow-hidden">
          <div
            className="h-full bg-coral rounded-full transition-all duration-300"
            style={{ width: `${((index + 1) / termesLocaux.length) * 100}%` }}
          />
        </div>
      </div>

      {/* Flashcard */}
      <div
        onClick={() => setRetourne(!retourne)}
        className="bg-white rounded-xl p-10 shadow-md hover:shadow-lg transition-shadow cursor-pointer min-h-[250px] flex flex-col items-center justify-center text-center border border-cream-dark"
      >
        {!retourne ? (
          <>
            <div className="text-xs text-ink-muted uppercase tracking-wide mb-4">Terme</div>
            <h3 className="font-display text-2xl font-semibold text-ink mb-4">{terme.terme}</h3>
            {terme.categorie && (
              <span className="text-xs font-semibold uppercase tracking-wide px-2.5 py-1 rounded-full bg-teal/10 text-teal">
                {terme.categorie}
              </span>
            )}
            <div className="mt-6 text-sm text-ink-muted">Cliquez pour reveler la definition</div>
          </>
        ) : (
          <>
            <div className="text-xs text-ink-muted uppercase tracking-wide mb-4">Definition</div>
            <p className="text-ink-light leading-relaxed mb-4">{terme.definition}</p>
            {terme.exemple && (
              <p className="text-sm text-ink-muted italic">Exemple : {terme.exemple}</p>
            )}
          </>
        )}
      </div>

      {/* Boutons d'action */}
      <div className="flex items-center justify-center gap-4 mt-8">
        <button
          onClick={allerPrecedent}
          disabled={index === 0}
          className="p-3 rounded-full border border-cream-dark text-ink-muted hover:bg-cream transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
          aria-label="Precedent"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>

        <button
          onClick={marquerInconnu}
          className="px-6 py-3 bg-red-50 text-red-600 border border-red-200 rounded-full font-medium hover:bg-red-100 transition-colors"
        >
          {String.fromCodePoint(0x274C)} Je ne sais pas
        </button>

        <button
          onClick={marquerConnu}
          className="px-6 py-3 bg-green-50 text-green-600 border border-green-200 rounded-full font-medium hover:bg-green-100 transition-colors"
        >
          {String.fromCodePoint(0x2705)} Je sais !
        </button>

        <button
          onClick={allerSuivant}
          disabled={index === termesLocaux.length - 1}
          className="p-3 rounded-full border border-cream-dark text-ink-muted hover:bg-cream transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
          aria-label="Suivant"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>

      {/* Maitrise actuelle */}
      <div className="flex items-center justify-center gap-2 mt-6">
        <span className="text-sm text-ink-muted">Maitrise :</span>
        <IndicateurMaitrise niveau={terme.maitrise} />
      </div>
    </div>
  )
}

// ===== Hook Chargement Cours =====
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
          setListeCours(res.cours || [])
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

// ===== Hook Chargement Termes =====
function useChargementTermes(coursId: string | null) {
  const [termes, setTermes] = useState<TermeLexique[]>([])
  const [chargement, setChargement] = useState(() => !!coursId)
  const [erreur, setErreur] = useState<string | null>(null)
  const [prevCoursId, setPrevCoursId] = useState(coursId)

  if (coursId !== prevCoursId) {
    setPrevCoursId(coursId)
    if (coursId) {
      setChargement(true)
      setErreur(null)
      setTermes([])
    }
  }

  useLayoutEffect(() => {
    if (!coursId) return

    let cancelled = false

    getTermesLexique(coursId)
      .then((res) => {
        if (!cancelled) {
          if (res.succes) {
            setTermes(res.termes || [])
          } else {
            setErreur(res.erreur?.message || 'Erreur lors du chargement')
          }
          setChargement(false)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setErreur(err.message)
          setChargement(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [coursId])

  const ajouterTermes = useCallback((nouveauxTermes: TermeLexique[]) => {
    setTermes(nouveauxTermes)
  }, [])

  return { termes, chargement, erreur, ajouterTermes }
}

// ===== Page Principale Lexique =====
export default function Lexique() {
  const [searchParams] = useSearchParams()
  const coursId = searchParams.get('cours')

  const { listeCours, chargement: chargementCours, erreur: erreurCours } = useChargementCours(coursId)
  const { termes, chargement: chargementTermes, erreur: erreurTermes, ajouterTermes } = useChargementTermes(coursId)

  const [modeVue, setModeVue] = useState<ModeVue>('lexique')
  const [recherche, setRecherche] = useState('')
  const [tri, setTri] = useState<TriOption>('alphabetique')
  const [extractionEnCours, setExtractionEnCours] = useState(false)
  const [erreurExtraction, setErreurExtraction] = useState<string | null>(null)

  const handleExtraire = useCallback(async () => {
    if (!coursId || extractionEnCours) return

    setExtractionEnCours(true)
    setErreurExtraction(null)

    try {
      const resultat = await extraireTermesLexique(coursId)
      if (resultat.succes && resultat.termes) {
        ajouterTermes(resultat.termes)
      } else {
        setErreurExtraction(resultat.erreur?.message || 'Erreur lors de l extraction')
      }
    } catch (err) {
      setErreurExtraction(err instanceof Error ? err.message : 'Erreur lors de l extraction')
    } finally {
      setExtractionEnCours(false)
    }
  }, [coursId, extractionEnCours, ajouterTermes])

  const handleMaitriseChange = useCallback(async (termeId: string, maitrise: number) => {
    try {
      await mettreAJourMaitrise(termeId, maitrise)
      // Mettre a jour localement
      ajouterTermes(termes.map((t) => t.id === termeId ? { ...t, maitrise } : t))
    } catch {
      // Silencieusement ignorer
    }
  }, [termes, ajouterTermes])

  // Page de selection de cours
  if (!coursId) {
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
          <div className="text-5xl mb-6">{String.fromCodePoint(0x1F4DA)}</div>
          <h1 className="font-display text-2xl font-semibold text-ink mb-4">
            {erreurCours || 'Aucun cours disponible'}
          </h1>
          <p className="text-ink-light mb-8">
            Scannez d'abord un cours pour generer un lexique.
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
        <h1 className="font-display text-xl md:text-3xl font-semibold text-ink mb-8">
          Lexique
        </h1>
        <p className="text-ink-light mb-8">
          Selectionnez un cours pour consulter son lexique.
        </p>

        <div className="grid gap-6">
          {listeCours.map((c) => (
            <Link
              key={c.id}
              to={`/lexique?cours=${c.id}`}
              className="bg-white rounded-lg p-8 shadow-sm hover:shadow-md transition-shadow border border-cream-dark"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h2 className="font-display text-lg font-semibold text-ink mb-2">
                    {c.titre || 'Cours sans titre'}
                  </h2>
                  <p className="text-sm text-ink-muted">{c.matiere || 'Matiere non definie'}</p>
                </div>
                <span className="text-coral font-medium">Voir le lexique {"\u2192"}</span>
              </div>
            </Link>
          ))}
        </div>
      </div>
    )
  }

  // Page du lexique d'un cours
  if (chargementTermes) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <ProcessingSection message="Chargement du lexique..." />
      </div>
    )
  }

  if (erreurTermes) {
    return (
      <div className="text-center py-12">
        <div className="text-5xl mb-6">{String.fromCodePoint(0x1F615)}</div>
        <h1 className="font-display text-2xl font-semibold text-ink mb-4">Erreur</h1>
        <p className="text-ink-light mb-8">{erreurTermes}</p>
        <Link
          to="/lexique"
          className="inline-flex items-center gap-4 px-8 py-3 bg-coral text-white rounded-full font-medium hover:bg-coral-dark transition-colors"
        >
          {"\u2190"} Retour aux cours
        </Link>
      </div>
    )
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <Link to="/lexique" className="text-ink-light hover:text-ink transition-colors text-sm">
            {"\u2190"} Retour aux cours
          </Link>
          <h1 className="font-display text-xl md:text-3xl font-semibold text-ink mt-2">Lexique</h1>
        </div>
        {termes.length > 0 && (
          <div className="text-sm text-ink-muted">
            {termes.length} terme{termes.length !== 1 ? 's' : ''}
          </div>
        )}
      </div>

      {erreurExtraction && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
          {erreurExtraction}
        </div>
      )}

      {/* Onglets */}
      <div className="flex flex-wrap gap-2 bg-white p-1 rounded-full shadow-sm mb-8 w-fit">
        <button
          onClick={() => setModeVue('lexique')}
          className={`px-6 py-2.5 rounded-full text-sm font-medium transition-all ${
            modeVue === 'lexique' ? 'bg-ink text-white' : 'text-ink-light hover:bg-cream'
          }`}
        >
          {String.fromCodePoint(0x1F4D6)} Lexique
        </button>
        <button
          onClick={() => setModeVue('quiz')}
          className={`px-6 py-2.5 rounded-full text-sm font-medium transition-all ${
            modeVue === 'quiz' ? 'bg-ink text-white' : 'text-ink-light hover:bg-cream'
          }`}
        >
          {String.fromCodePoint(0x1F3AF)} Quiz vocabulaire
        </button>
        <button
          onClick={() => setModeVue('flashcards')}
          className={`px-6 py-2.5 rounded-full text-sm font-medium transition-all ${
            modeVue === 'flashcards' ? 'bg-ink text-white' : 'text-ink-light hover:bg-cream'
          }`}
        >
          {String.fromCodePoint(0x26A1)} Flashcards rapides
        </button>
      </div>

      {/* Contenu selon le mode */}
      {modeVue === 'lexique' && (
        <VueListeTermes
          termes={termes}
          recherche={recherche}
          setRecherche={setRecherche}
          tri={tri}
          setTri={setTri}
          onMaitriseChange={handleMaitriseChange}
          onExtraire={handleExtraire}
          extractionEnCours={extractionEnCours}
        />
      )}

      {modeVue === 'quiz' && (
        <VueQuizVocabulaire coursId={coursId} termes={termes} />
      )}

      {modeVue === 'flashcards' && (
        <VueFlashcards termes={termes} onMaitriseChange={handleMaitriseChange} />
      )}
    </div>
  )
}
