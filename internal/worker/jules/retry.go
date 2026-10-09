package jules

import (
	"errors"
	"net"
	"time"
)

const (
	pollInterval        = 5 * time.Second
	sendRetryMax        = 3
	sendRetryDelay      = 10 * time.Second
	pollNotFoundRetries = 3
	pollNotFoundDelay   = 5 * time.Second
	nudgeInterval       = 30 * time.Minute
	nudgeText           = "please update"
	headChangeQuiet     = 2 * time.Minute
)

// WithRetries wraps a Doer with retry logic.
func WithRetries(inner Doer, sleep func(time.Duration)) Doer {
	if inner == nil {
		return nil
	}
	if _, ok := inner.(*retryDoer); ok {
		return inner
	}
	return &retryDoer{inner: inner, sleep: sleep}
}

type retryDoer struct {
	inner Doer
	sleep func(time.Duration)
}

func (r *retryDoer) Do(method, path string, body []byte) (int, []byte, error) {
	sleep := r.sleep
	if sleep == nil {
		sleep = time.Sleep
	}

	tries404 := 0
	triesOther := 0

	for {
		status, resp, err := r.inner.Do(method, path, body)
		if err == nil && status >= 200 && status < 300 {
			return status, resp, nil
		}

		isTimeoutOrTemp := false
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
				isTimeoutOrTemp = true
			}
		}

		if status == 404 {
			if tries404 < pollNotFoundRetries {
				tries404++
				sleep(pollNotFoundDelay)
				continue
			}
			return status, resp, err
		}

		if isTimeoutOrTemp || status == 429 || status == 500 || status == 502 || status == 503 || status == 504 {
			if triesOther < sendRetryMax {
				triesOther++
				sleep(sendRetryDelay)
				continue
			}
			return status, resp, err
		}

		// Non-retryable
		return status, resp, err
	}
}
