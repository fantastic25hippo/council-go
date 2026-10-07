package worker

import "testing"

func TestParseQuestionMarkerOnLastLine(t *testing.T) {
	report := "did some work\nCOUNCIL-QUESTION: should we merge 4255ed2?"
	q, asked := ParseQuestion(report)
	if !asked {
		t.Fatalf("asked=false, want true")
	}
	if q != "should we merge 4255ed2?" {
		t.Fatalf("question=%q", q)
	}
}

func TestParseQuestionIgnoresQuestionMarkWithoutMarker(t *testing.T) {
	report := "accept 4255ed2?\nall tests passed"
	q, asked := ParseQuestion(report)
	if asked || q != "" {
		t.Fatalf("asked=%v question=%q, want not asked", asked, q)
	}
}

func TestParseQuestionMarkerMustBeLastNonEmptyLine(t *testing.T) {
	report := "COUNCIL-QUESTION: wait?\nthen I continued anyway"
	if _, asked := ParseQuestion(report); asked {
		t.Fatalf("marker not last, asked=true")
	}
}

func TestParseQuestionAllowsTrailingBlankLines(t *testing.T) {
	report := "COUNCIL-QUESTION: which branch?\n\n"
	q, asked := ParseQuestion(report)
	if !asked || q != "which branch?" {
		t.Fatalf("asked=%v question=%q", asked, q)
	}
}

func TestParseQuestionEmptyQuestionIsNotAsked(t *testing.T) {
	if _, asked := ParseQuestion("COUNCIL-QUESTION:   \n"); asked {
		t.Fatalf("empty question counted as ASKED")
	}
}
