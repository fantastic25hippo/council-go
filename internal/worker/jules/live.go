package jules

import (
	"errors"
	"strings"
	"time"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

type Live struct {
	Client   *Client
	Branch   string
	Timeout  time.Duration
	Interval time.Duration
	Sleep    func(time.Duration)
	sess     *APISession
	approved bool
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
	}}
}

func (l *Live) session() Session { return l.sess }

func (l *Live) start(job worker.Job) error {
	if l.Client == nil {
		return errors.New("jules client missing")
	}
	if job.Repo == "" {
		return errors.New("empty REPO")
	}
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
	return l.poll()
}

func (l *Live) answer(text string) error {
	if l.sess == nil || l.sess.SID == "" {
		return errors.New("empty session id")
	}
	if err := l.Client.Send(l.sess.SID, text); err != nil {
		return err
	}
	return l.poll()
}

func (l *Live) poll() error {
	if l.sess == nil {
		return errors.New("empty session id")
	}
	sleep := l.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	interval := l.Interval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(l.Timeout)
	for time.Now().Before(deadline) {
		if err := l.sess.Refresh(); err != nil {
			return err
		}
		sum, _ := l.Client.Activities(l.sess.SID)
		if sum.LastAgent != "" {
			l.sess.SetMessage(sum.LastAgent)
		}
		st := strings.ToUpper(l.sess.State())
		if !l.approved && (sum.PlanGenerated || strings.Contains(st, "PLAN")) {
			if err := l.Client.ApprovePlan(l.sess.SID); err == nil {
				l.approved = true
			}
		}
		switch st {
		case "COMPLETED", "FAILED", "AWAITING_USER_FEEDBACK":
			return nil
		}
		sleep(interval)
	}
	return errors.New("jules poll timed out")
}
