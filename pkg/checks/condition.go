package checks

// Condition is the common interface
// implemented by all condition types
type Condition interface {
	Verify(ctx VerificationContext) error
}
