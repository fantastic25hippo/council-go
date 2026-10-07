package member

import "sync"

// Fake is a scripted Member for tests. Replies are consumed in order.
type Fake struct {
	Name    string
	Shown   string
	Replies []string
	Sent    []string
	Started int
	Closed  int
	i       int
	SendErr error
	mu      sync.Mutex
}

func (f *Fake) ID() string          { return f.Name }
func (f *Fake) DisplayName() string { return f.Shown }
func (f *Fake) Start(string) error  { f.mu.Lock(); f.Started++; f.mu.Unlock(); return nil }
func (f *Fake) Close() error        { f.mu.Lock(); f.Closed++; f.mu.Unlock(); return nil }

func (f *Fake) Send(prompt string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Sent = append(f.Sent, prompt)
	if f.SendErr != nil {
		return "", f.SendErr
	}
	if f.i >= len(f.Replies) {
		return "", nil
	}
	r := f.Replies[f.i]
	f.i++
	return r, nil
}
