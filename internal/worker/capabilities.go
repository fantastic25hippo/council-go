package worker

// Capabilities is declared before planning. The relay does not infer
// these from prompts or from worker output.
type Capabilities struct {
	CanAskAndWait          bool
	KeepsMemoryAcrossFixes bool
	ReachesGitHub          bool
	TouchesLocalCheckout   bool
	OpensPR                bool
}
