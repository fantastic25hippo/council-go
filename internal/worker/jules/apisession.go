package jules

import "errors"

// APISession adapts Client to the existing Session interface
type APISession struct {
	C    *Client
	SID  string
	info SessionInfo
}

var _ Session = (*APISession)(nil)

// Refresh fetches the latest session info from the API and stores it.
func (s *APISession) Refresh() error {
	if s.SID == "" {
		return errors.New("empty session id")
	}

	info, err := s.C.GetSession(s.SID)
	if err != nil {
		return err
	}
	s.info = info
	return nil
}

// State returns the current session state.
func (s *APISession) State() string {
	return s.info.State
}

// LastAgentMessage returns the last message from the agent.
func (s *APISession) LastAgentMessage() string {
	return s.info.Message
}

// ID returns the session ID.
func (s *APISession) ID() string {
	return s.SID
}

// PRURL returns the PR URL if one was created.
func (s *APISession) PRURL() string {
	return s.info.PRURL
}
