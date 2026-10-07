package worker

// Worker is the relay-owned contract. Meeting code talks only through
// this interface. Adapters live in subpackages.
type Worker interface {
	Capabilities() Capabilities
	Start(job Job) Run
	Continue(job Job, answer string) Run
}
