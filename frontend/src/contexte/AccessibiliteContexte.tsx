import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import type { ReactNode } from 'react'

// Types pour les options d'accessibilité
export type TailleTexte = 'normal' | 'large' | 'xlarge'
export type ModeDaltonien = 'off' | 'protanopia' | 'deuteranopia' | 'tritanopia'

export interface EtatAccessibilite {
  tailleTexte: TailleTexte
  modeDaltonien: ModeDaltonien
  contrasteEleve: boolean
}

interface ContexteAccessibiliteType extends EtatAccessibilite {
  setTailleTexte: (taille: TailleTexte) => void
  setModeDaltonien: (mode: ModeDaltonien) => void
  setContrasteEleve: (actif: boolean) => void
  reinitialiser: () => void
}

// Valeurs par défaut
const ETAT_INITIAL: EtatAccessibilite = {
  tailleTexte: 'normal',
  modeDaltonien: 'off',
  contrasteEleve: false,
}

// Clés localStorage
const CLES_STORAGE = {
  tailleTexte: 'accessibilite-taille',
  modeDaltonien: 'accessibilite-daltonien',
  contrasteEleve: 'accessibilite-contraste',
} as const

// Contexte
const AccessibiliteContexte = createContext<ContexteAccessibiliteType | null>(null)

// Charger l'état depuis localStorage (synchrone, pour initialisation)
function chargerDepuisStorage<T>(cle: string, defaut: T): T {
  if (typeof window === 'undefined') return defaut
  try {
    const valeur = localStorage.getItem(cle)
    if (valeur === null) return defaut
    // Pour les booléens
    if (typeof defaut === 'boolean') {
      return (valeur === 'true') as T
    }
    return valeur as T
  } catch {
    return defaut
  }
}

// Appliquer les classes CSS au document
function appliquerClassesCSS(etat: EtatAccessibilite) {
  if (typeof document === 'undefined') return

  const html = document.documentElement

  // Supprimer les anciennes classes
  html.classList.remove(
    'text-scale-normal',
    'text-scale-large',
    'text-scale-xlarge',
    'daltonien-protanopia',
    'daltonien-deuteranopia',
    'daltonien-tritanopia',
    'contraste-eleve'
  )

  // Appliquer la taille de texte
  html.classList.add(`text-scale-${etat.tailleTexte}`)

  // Appliquer le mode daltonien
  if (etat.modeDaltonien !== 'off') {
    html.classList.add(`daltonien-${etat.modeDaltonien}`)
  }

  // Appliquer le contraste élevé
  if (etat.contrasteEleve) {
    html.classList.add('contraste-eleve')
  }
}

// Provider
interface ProviderProps {
  children: ReactNode
}

export function AccessibiliteProvider({ children }: ProviderProps) {
  // Utiliser lazy initialization pour charger depuis localStorage au premier rendu
  const [tailleTexte, setTailleTexteState] = useState<TailleTexte>(() =>
    chargerDepuisStorage(CLES_STORAGE.tailleTexte, ETAT_INITIAL.tailleTexte)
  )
  const [modeDaltonien, setModeDaltonienState] = useState<ModeDaltonien>(() =>
    chargerDepuisStorage(CLES_STORAGE.modeDaltonien, ETAT_INITIAL.modeDaltonien)
  )
  const [contrasteEleve, setContrasteEleveState] = useState(() =>
    chargerDepuisStorage(CLES_STORAGE.contrasteEleve, ETAT_INITIAL.contrasteEleve)
  )

  // Appliquer les classes CSS quand l'état change
  useEffect(() => {
    appliquerClassesCSS({ tailleTexte, modeDaltonien, contrasteEleve })
  }, [tailleTexte, modeDaltonien, contrasteEleve])

  // Setters avec persistance
  const setTailleTexte = useCallback((taille: TailleTexte) => {
    setTailleTexteState(taille)
    localStorage.setItem(CLES_STORAGE.tailleTexte, taille)
  }, [])

  const setModeDaltonien = useCallback((mode: ModeDaltonien) => {
    setModeDaltonienState(mode)
    localStorage.setItem(CLES_STORAGE.modeDaltonien, mode)
  }, [])

  const setContrasteEleve = useCallback((actif: boolean) => {
    setContrasteEleveState(actif)
    localStorage.setItem(CLES_STORAGE.contrasteEleve, String(actif))
  }, [])

  const reinitialiser = useCallback(() => {
    setTailleTexte(ETAT_INITIAL.tailleTexte)
    setModeDaltonien(ETAT_INITIAL.modeDaltonien)
    setContrasteEleve(ETAT_INITIAL.contrasteEleve)
  }, [setTailleTexte, setModeDaltonien, setContrasteEleve])

  const valeur: ContexteAccessibiliteType = {
    tailleTexte,
    modeDaltonien,
    contrasteEleve,
    setTailleTexte,
    setModeDaltonien,
    setContrasteEleve,
    reinitialiser,
  }

  return (
    <AccessibiliteContexte.Provider value={valeur}>
      {children}
    </AccessibiliteContexte.Provider>
  )
}

// Hook personnalisé pour utiliser le contexte
// eslint-disable-next-line react-refresh/only-export-components
export function useAccessibilite(): ContexteAccessibiliteType {
  const contexte = useContext(AccessibiliteContexte)
  if (!contexte) {
    throw new Error('useAccessibilite doit être utilisé dans un AccessibiliteProvider')
  }
  return contexte
}
