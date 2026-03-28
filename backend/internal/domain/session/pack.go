package session

// PackConstraints defines session composition rules tied to a chapter's pack (Z4-AC05).
type PackConstraints struct {
	MaxWritingPerSession  int  // max writing exercises per session (default: 1)
	SessionMustIncludeDoc bool // must include at least 1 document exercise
}

// DefaultPackConstraints returns the MVP default pack constraints.
func DefaultPackConstraints() PackConstraints {
	return PackConstraints{
		MaxWritingPerSession:  1,
		SessionMustIncludeDoc: true,
	}
}
