package attach

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantOk  bool
		wantReq Request
	}{
		{
			name:   "1. Jules URL + PR URL",
			text:   "https://jules.google.com/session/123456 https://github.com/o/r/pull/8",
			wantOk: true,
			wantReq: Request{
				SessionID: "123456",
				PRNumber:  "8",
				Repo:      "o/r",
			},
		},
		{
			name:   "2. Bare 20-digit id + #8",
			text:   "12345678901234567890 #8",
			wantOk: true,
			wantReq: Request{
				SessionID: "12345678901234567890",
				PRNumber:  "8",
			},
		},
		{
			name:   "3. Only a Jules URL",
			text:   "https://jules.google.com/session/123456",
			wantOk: true,
			wantReq: Request{
				SessionID:  "123456",
				Incomplete: "Attach needs a PR (GitHub pull URL or #N).",
			},
		},
		{
			name:   "4. Only #8",
			text:   "fix #8",
			wantOk: false,
		},
		{
			name:   "5. text with no session",
			text:   "see PR #31 for the launcher",
			wantOk: false,
		},
		{
			name:   "6. text with no ids",
			text:   "fix extra_path",
			wantOk: false,
		},
		{
			name:   "7. Two Jules URLs with different ids",
			text:   "https://jules.google.com/session/123456 https://jules.google.com/session/654321",
			wantOk: true,
			wantReq: Request{
				Error: "Two different Jules sessions given.",
			},
		},
		{
			name:   "8. Bare 31 is not a PR. Plus session still needs PR",
			text:   "https://jules.google.com/session/123456 31",
			wantOk: true,
			wantReq: Request{
				SessionID:  "123456",
				Incomplete: "Attach needs a PR (GitHub pull URL or #N).",
			},
		},
		{
			name:   "9. /attach alone does not attach",
			text:   "/attach",
			wantOk: false,
		},
		{
			name:   "Two different PR numbers",
			text:   "https://jules.google.com/session/123456 #8 #9",
			wantOk: true,
			wantReq: Request{
				Error: "Two different pull requests given.",
			},
		},
		{
			name:   "Jules u/n/session URL",
			text:   "https://jules.google.com/u/1/session/123456 #8",
			wantOk: true,
			wantReq: Request{
				SessionID: "123456",
				PRNumber:  "8",
			},
		},
		{
			name:   "Bare ID not matched inside word",
			text:   "abc12345678901234567890def",
			wantOk: false,
		},
		{
			name:   "Bare ID correctly matched with newlines",
			text:   "hello\n12345678901234567890\nworld #8",
			wantOk: true,
			wantReq: Request{
				SessionID: "12345678901234567890",
				PRNumber:  "8",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, ok := Parse(tc.text)
			if ok != tc.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOk)
			}
			if ok {
				if !reflect.DeepEqual(req, tc.wantReq) {
					t.Fatalf("req = %+v\nwant = %+v", req, tc.wantReq)
				}
			}
		})
	}
}
