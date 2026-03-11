package mastery

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/event"
)

var (
	ErrTransitionBlocked = errors.New("transition blocked: spacing requirement not met")
	ErrInvalidState      = errors.New("invalid mastery state")
)

// SpacingHours defines the minimum hours between OK and SOLID transition.
const SpacingHours = 24

// Mastery is the aggregate root for tracking a user's mastery of a single item.
type Mastery struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	ItemID               uuid.UUID
	State                State
	NextDueAt            *time.Time
	LastReviewAt         *time.Time
	LastSuccessAt        *time.Time
	ConsecutiveSuccesses int
	ConsecutiveFailures  int
	CurrentDifficulty    *int
	CappedAtOK           bool // Items with validation_required cap at OK
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewMastery creates a new Mastery in UNKNOWN state.
func NewMastery(idGen event.IDGenerator, userID, itemID uuid.UUID, now time.Time) *Mastery {
	return &Mastery{
		ID:        idGen.New(),
		UserID:    userID,
		ItemID:    itemID,
		State:     Unknown,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// RecordAttempt processes a score (0.0 = fail, 1.0 = success) and transitions
// the state machine accordingly. The now parameter enables testable time control.
//
// State transitions:
//   UNKNOWN + success → FRAGILE
//   UNKNOWN + fail    → UNKNOWN
//   FRAGILE + success → OK
//   FRAGILE + fail    → FRAGILE (cs reset)
//   OK      + success → SOLID (if spacing met and cs >= 2) or stays OK
//   OK      + fail    → FRAGILE
//   SOLID   + success → SOLID
//   SOLID   + fail    → OK
func (m *Mastery) RecordAttempt(score float64, now time.Time) error {
	success := score >= 0.7
	m.LastReviewAt = &now
	m.UpdatedAt = now

	// Save previous values for Z1-AC04 (spacing check may need to revert)
	prevCS := m.ConsecutiveSuccesses
	prevLastSuccessAt := m.LastSuccessAt

	if success {
		m.ConsecutiveSuccesses++
		m.ConsecutiveFailures = 0
		m.LastSuccessAt = &now
	} else {
		m.ConsecutiveFailures++
		m.ConsecutiveSuccesses = 0
	}

	switch m.State {
	case Unknown:
		if success {
			m.State = Fragile
		}
		due := now.Add(24 * time.Hour)
		m.NextDueAt = &due

	case Fragile:
		if success && m.ConsecutiveSuccesses >= 2 {
			// Z1-AC02: cs was ≥1 before this attempt → transition to OK
			m.State = OK
			due := now.Add(3 * 24 * time.Hour)
			m.NextDueAt = &due
		} else {
			// Z1-AC07b: cs was 0 (post-regression) → stay FRAGILE, accumulate cs
			// Z1-AC07: fail → stay FRAGILE
			due := now.Add(24 * time.Hour)
			m.NextDueAt = &due
		}

	case OK:
		if success {
			if m.CappedAtOK {
				// Z1-AC13: validation_required items cannot go beyond OK
				due := now.Add(3 * 24 * time.Hour)
				m.NextDueAt = &due
				return nil
			}
			if m.ConsecutiveSuccesses >= 2 && m.spacingMet(prevLastSuccessAt, now) {
				// Z1-AC03: OK → SOLID (cs ≥ 2 AND spacing met)
				m.State = Solid
				due := now.Add(7 * 24 * time.Hour)
				m.NextDueAt = &due
			} else if m.ConsecutiveSuccesses >= 2 && !m.spacingMet(prevLastSuccessAt, now) {
				// Z1-AC04: spacing not met — revert cs, don't touch anything
				m.ConsecutiveSuccesses = prevCS
				m.LastSuccessAt = prevLastSuccessAt
				// State stays OK, next_due_at unchanged
				return nil
			} else {
				// Z1-AC07c: cs < 2, stay OK, accumulate cs normally
				due := now.Add(3 * 24 * time.Hour)
				m.NextDueAt = &due
			}
		} else {
			// Z1-AC06: OK → FRAGILE
			m.State = Fragile
			due := now.Add(24 * time.Hour)
			m.NextDueAt = &due
		}

	case Solid:
		if !success {
			// Z1-AC05: SOLID → OK, next_due_at = now + 2 jours
			m.State = OK
			due := now.Add(2 * 24 * time.Hour)
			m.NextDueAt = &due
		} else {
			// Z1-AC12: stay SOLID, next_due_at = now + 7 jours
			due := now.Add(7 * 24 * time.Hour)
			m.NextDueAt = &due
		}

	default:
		return ErrInvalidState
	}

	return nil
}

// spacingMet checks if ≥ 24h have elapsed since the previous success.
// prevSuccessAt is the LastSuccessAt value BEFORE the current attempt updated it.
func (m *Mastery) spacingMet(prevSuccessAt *time.Time, now time.Time) bool {
	if prevSuccessAt == nil {
		return false
	}
	return now.Sub(*prevSuccessAt) >= time.Duration(SpacingHours)*time.Hour
}

// IsDue returns true if the mastery is due for review at the given time.
func (m *Mastery) IsDue(at time.Time) bool {
	if m.NextDueAt == nil {
		return true
	}
	return !at.Before(*m.NextDueAt)
}
