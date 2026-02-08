import { useState, useCallback, useLayoutEffect } from 'react'
import type { Fiche } from '../services/api'

interface CarteFicheProps {
  fiche: Fiche
  onFlip?: (estRetournee: boolean) => void
}

export default function CarteFiche({ fiche, onFlip }: CarteFicheProps) {
  const [estRetournee, setEstRetournee] = useState(false)
  const [ficheId, setFicheId] = useState(fiche.id)

  // Reset quand la fiche change (pattern derived state)
  if (fiche.id !== ficheId) {
    setFicheId(fiche.id)
    setEstRetournee(false)
  }

  const retourner = useCallback(() => {
    setEstRetournee((prev) => {
      const nouvelEtat = !prev
      onFlip?.(nouvelEtat)
      return nouvelEtat
    })
  }, [onFlip])

  // Gestion clavier (Space ou Enter)
  useLayoutEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === ' ' || e.key === 'Enter') {
        e.preventDefault()
        retourner()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [retourner])

  return (
    <div
      className="w-full max-w-[600px] cursor-pointer"
      style={{ perspective: '1000px' }}
      onClick={retourner}
      role="button"
      tabIndex={0}
      aria-label={estRetournee ? 'Afficher la question' : 'Afficher la réponse'}
    >
      <div
        className="relative w-full transition-transform duration-500"
        style={{
          transformStyle: 'preserve-3d',
          transform: estRetournee ? 'rotateY(180deg)' : 'rotateY(0deg)',
          aspectRatio: '4/3',
        }}
      >
        {/* Face avant - Question */}
        <div
          className="absolute inset-0 bg-white rounded-lg p-12 flex flex-col shadow-[0_20px_60px_rgba(0,0,0,0.12)]"
          style={{ backfaceVisibility: 'hidden' }}
        >
          <div className="text-xs font-semibold uppercase tracking-widest text-coral mb-4 opacity-60">
            Question
          </div>
          <div className="flex-1 flex items-center justify-center text-center">
            <p className="font-display text-2xl font-semibold leading-relaxed text-ink">
              {fiche.question}
            </p>
          </div>
          <p className="text-center text-sm text-ink-muted opacity-50">
            Clique ou appuie sur Espace pour retourner
          </p>
        </div>

        {/* Face arrière - Réponse */}
        <div
          className="absolute inset-0 bg-teal rounded-lg p-12 flex flex-col shadow-[0_20px_60px_rgba(0,0,0,0.12)]"
          style={{
            backfaceVisibility: 'hidden',
            transform: 'rotateY(180deg)',
          }}
        >
          <div className="text-xs font-semibold uppercase tracking-widest text-white/60 mb-4">
            Réponse
          </div>
          <div className="flex-1 flex items-center justify-center text-center">
            <p className="text-lg leading-relaxed text-white">
              {fiche.reponse}
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
