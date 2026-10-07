package member

// Member is a live council member session. Send returns visible text only.
type Member interface {
	ID() string
	DisplayName() string
	Start(cwd string) error
	Send(prompt string) (string, error)
	Close() error
}
