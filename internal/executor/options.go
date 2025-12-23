package executor

// Options configure filesystem application of a planned layout.
type Options struct {
	InputRoot  string
	OutputRoot string
	DryRun     bool
	Overwrite  bool
	VerifyOnly bool
}
