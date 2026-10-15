package jules

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

type Live struct {
	mu       sync.Mutex
	Client   *Client
	Branch   string
	Timeout  time.Duration
	Interval time.Duration
	Sleep    func(time.Duration)
	Now      func() time.Time
	Runner   GHRunner
	Repo     string
	sess     *APISession
	approved bool
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewLive(c *Client, timeout time.Duration) *Worker {
	if timeout <= 0 {
		timeout = 1800 * time.Second
	}
	return &Worker{live: &Live{
		Client:   c,
		Timeout:  timeout,
		Interval: 5 * time.Second,
		Sleep:    time.Sleep,
		Runner:   execGH{},
	}}
}

func (l *Live) session() Session { return l.sess }

func (l *Live) newCtx() context.Context {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ctx, l.cancel = context.WithCancel(context.Background())
	return l.ctx
}

func (l *Live) start(job worker.Job) error {
	if l.Client == nil {
		return errors.New("jules client missing")
	}
	if job.SessionID != "" {
		ctx := l.newCtx()
		l.Repo = job.Repo
		l.sess = &APISession{C: l.Client, SID: job.SessionID}
		if job.PRURL != "" {
			l.sess.SetPRURL(job.PRURL)
		}
		l.approved = false
		err := l.poll(ctx)
		if l.sess != nil && l.sess.PRURL() == "" && job.PRURL != "" {
			l.sess.SetPRURL(job.PRURL)
		}
		return err
	}
	if job.Repo == "" {
		return errors.New("empty REPO")
	}

	ctx := l.newCtx()
	l.Repo = job.Repo
	branch := l.Branch
	if branch == "" {
		branch = "master"
	}
	id, err := l.Client.CreateSession(job.Repo, job.Text, branch, "Council task")
	if err != nil {
		return err
	}
	l.sess = &APISession{C: l.Client, SID: id}
	l.approved = false
	return l.poll(ctx)
}

func (l *Live) answer(text string) error {
	if l.sess == nil || l.sess.SID == "" {
		return errors.New("empty session id")
	}
	ctx := l.newCtx()
	if err := l.Client.Send(l.sess.SID, text); err != nil {
		return err
	}
	return l.poll(ctx)
}

