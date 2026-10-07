package worker

// Fake is an in-memory Worker for tests. It records Start/Continue
// calls and returns Next.
type Fake struct {
	Caps       Capabilities
	Next       Run
	Starts     int
	Continues  int
	LastJob    Job
	LastAnswer string
}

func (f *Fake) Capabilities() Capabilities {
	return f.Caps
}

func (f *Fake) Start(job Job) Run {
	f.Starts++
	f.LastJob = job
	return f.Next
}

func (f *Fake) Continue(job Job, answer string) Run {
	f.Continues++
	f.LastJob = job
	f.LastAnswer = answer
	return f.Next
}
