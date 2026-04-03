import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter } from 'react-router-dom'
import Validation from '../../pages/Validation'
import type { TacheValidation } from '../../services/api'

// Mock the API module
vi.mock('../../services/api', () => ({
  listerValidations: vi.fn(),
  resoudreValidation: vi.fn(),
}))

import { listerValidations, resoudreValidation } from '../../services/api'

const mockLister = vi.mocked(listerValidations)
const mockResoudre = vi.mocked(resoudreValidation)

const tachesMock: TacheValidation[] = [
  {
    id: 'task-1',
    item_id: 'item-1',
    chapter_id: 'ch-1',
    suggestion: 'Le theoreme de Pythagore',
    priority: 2,
    status: 'pending',
    source: 'ocr',
    item_term: 'Pythagore',
    created_at: '2026-04-01T10:00:00Z',
  },
  {
    id: 'task-2',
    item_id: 'item-2',
    suggestion: 'La loi de Newton',
    priority: 4,
    status: 'pending',
    source: 'llm',
    item_term: 'Newton',
    created_at: '2026-04-02T14:00:00Z',
  },
]

function renderValidation() {
  return render(
    <BrowserRouter>
      <Validation />
    </BrowserRouter>
  )
}

describe('Validation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('affiche la liste des tasks pending', async () => {
    mockLister.mockResolvedValue({
      succes: true,
      validations: tachesMock,
    })

    renderValidation()

    await waitFor(() => {
      expect(screen.getByTestId('validation-task-list')).toBeInTheDocument()
    })

    expect(screen.getByText('Pythagore')).toBeInTheDocument()
    expect(screen.getByText('Newton')).toBeInTheDocument()
    expect(screen.getByText('2 en attente')).toBeInTheDocument()
  })

  it('affiche empty state quand 0 tasks', async () => {
    mockLister.mockResolvedValue({
      succes: true,
      validations: [],
    })

    renderValidation()

    await waitFor(() => {
      expect(screen.getByTestId('validation-empty')).toBeInTheDocument()
    })

    expect(screen.getByText('Aucune validation en attente')).toBeInTheDocument()
  })

  it('affiche le detail quand on clique sur une task', async () => {
    mockLister.mockResolvedValue({
      succes: true,
      validations: tachesMock,
    })

    const user = userEvent.setup()
    renderValidation()

    await waitFor(() => {
      expect(screen.getByTestId('validation-task-list')).toBeInTheDocument()
    })

    // Click on the first task
    const items = screen.getAllByTestId('validation-task-item')
    await user.click(items[0])

    // Should show detail view with action buttons
    expect(screen.getByTestId('validation-confirm-btn')).toBeInTheDocument()
    expect(screen.getByTestId('validation-correct-btn')).toBeInTheDocument()
    expect(screen.getByTestId('validation-ignore-btn')).toBeInTheDocument()
    expect(screen.getByTestId('validation-skip-btn')).toBeInTheDocument()
  })

  it('confirmer une task appelle l\'API', async () => {
    mockLister.mockResolvedValue({
      succes: true,
      validations: tachesMock,
    })
    mockResoudre.mockResolvedValue({
      succes: true,
      validation: { ...tachesMock[0], status: 'resolved', resolved_at: '2026-04-03T10:00:00Z' },
    })

    const user = userEvent.setup()
    renderValidation()

    await waitFor(() => {
      expect(screen.getByTestId('validation-task-list')).toBeInTheDocument()
    })

    // Navigate to detail
    const items = screen.getAllByTestId('validation-task-item')
    await user.click(items[0])

    // Click confirm
    await user.click(screen.getByTestId('validation-confirm-btn'))

    await waitFor(() => {
      expect(mockResoudre).toHaveBeenCalledWith('task-1', 'approve', undefined)
    })
  })

  it('affiche une erreur quand le chargement echoue', async () => {
    mockLister.mockRejectedValue(new Error('Erreur reseau'))

    renderValidation()

    await waitFor(() => {
      expect(screen.getByTestId('validation-error')).toBeInTheDocument()
    })

    expect(screen.getByText('Erreur reseau')).toBeInTheDocument()
  })

  it('filtre les tasks non-pending', async () => {
    mockLister.mockResolvedValue({
      succes: true,
      validations: [
        ...tachesMock,
        {
          id: 'task-3',
          item_id: 'item-3',
          suggestion: 'Deja resolue',
          priority: 1,
          status: 'resolved',
          source: 'manual',
          item_term: 'Resolue',
          created_at: '2026-04-01T08:00:00Z',
          resolved_at: '2026-04-01T09:00:00Z',
        },
      ],
    })

    renderValidation()

    await waitFor(() => {
      expect(screen.getByTestId('validation-task-list')).toBeInTheDocument()
    })

    // Only 2 pending tasks should be visible
    expect(screen.getByText('2 en attente')).toBeInTheDocument()
    expect(screen.queryByText('Resolue')).not.toBeInTheDocument()
  })
})
