package golembic

// ApplyConfig provides configurable fields for "up" commands that will apply
// migrations.
type ApplyConfig struct {
	VerifyHistory bool
	Revision      string
}

// NewApplyConfig creates a new `ApplyConfig` and applies options.
func NewApplyConfig(opts ...ApplyOption) (*ApplyConfig, error) {
	ac := &ApplyConfig{}
	for _, opt := range opts {
		err := opt(ac)
		if err != nil {
			return nil, err
		}
	}

	return ac, nil
}

// OptApplyVerifyHistory sets `VerifyHistory` on an `ApplyConfig`.
func OptApplyVerifyHistory(verify bool) ApplyOption {
	return func(ac *ApplyConfig) error {
		ac.VerifyHistory = verify
		return nil
	}
}

// OptApplyRevision sets `Revision` on an `ApplyConfig`.
func OptApplyRevision(revision string) ApplyOption {
	return func(ac *ApplyConfig) error {
		ac.Revision = revision
		return nil
	}
}
