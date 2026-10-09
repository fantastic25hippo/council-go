package prompt

import (
	"fmt"
	"strings"

	"github.com/fantastic25hippo/council-go/internal/worker"
)

func System(name, other string, debateTurns, fixMax int, workerIDs []string) string {
	ids := strings.Join(workerIDs, ", ")
	return fmt.Sprintf(`You are one of two members of a council: %s and %s. Sahil is the user and watches.
You will see only the other member's proposals and messages, never its reasoning.
Independent first proposal, then %d debate turns each. Disputed leftovers go to Sahil.
First to agree writes the full plan plus AGREED. Second AGREED launches the worker.
Independent verification, maximum %d fix rounds, then adjourn.
Workers: %s.
Don't do the worker's typing unless the agreed plan says so. Never use pay-per-token API keys.

Required plan block:
WORKER: <id from: %s>
REPO: <owner/repo> (omit if not needed)
OWNED_PATHS:
- <path>
CHANGE: <self-contained change>
VERIFY: <ONE bare shell command only>
AGREED`, name, other, debateTurns, fixMax, ids, ids)
}

func Greet(name, other string) string {
	return fmt.Sprintf("Please briefly greet Sahil and %s in one or two sentences. Introduce yourself as %s. Do not start proposing or debating yet.", other, name)
}

func Proposal(task string, workerIDs []string, resolved string, caps worker.Capabilities) string {
	ids := strings.Join(workerIDs, ", ")
	capsText := fmt.Sprintf("CanAskAndWait: %v. ReachesGitHub: %v. OpensPR: %v. TouchesLocalCheckout: %v. KeepsMemoryAcrossFixes: %v.",
		caps.CanAskAndWait, caps.ReachesGitHub, caps.OpensPR, caps.TouchesLocalCheckout, caps.KeepsMemoryAcrossFixes)
	return fmt.Sprintf("Task from Sahil:\n%s\n\nPlease provide your independent initial proposal. Do not debate yet.\nAvailable workers: %s.\nThis task will be run by %s. %s\nIf proposing a full plan, use the required plan block ending with AGREED.", task, ids, resolved, capsText)
}

func Debate(other, message string) string {
	return fmt.Sprintf("%s's proposal/message:\n%s\n\nReview their points. If you agree on the complete final plan, output the full plan block ending with exactly the line AGREED.", other, message)
}

func Confirm(other, planText, worker string) string {
	return fmt.Sprintf("%s proposed this plan and agreed:\n\n%s\n\nIf you agree with this exact plan, reply ending with:\nLAUNCH: %s\nAGREED\nIf you reject it, state objections. Do not propose a modified plan while agreeing.", other, planText, worker)
}

func NackPlan(workerIDs []string) string {
	ids := strings.Join(workerIDs, ", ")
	return fmt.Sprintf("AGREED was detected, but the plan block was missing or invalid. Workers: %s. Resend the full plan ending with AGREED.", ids)
}

func LaunchNack(worker string) string {
	return fmt.Sprintf("You agreed, but did not include a valid launch command. Reply with:\nLAUNCH: %s\nAGREED", worker)
}

func Disputed(turns int) string {
	return fmt.Sprintf("The debate cap (%d turns each) was reached without dual agreement. List ONLY remaining disputed points for Sahil. Do not repeat agreed points.", turns)
}

func Ruling(text string) string {
	return fmt.Sprintf("Sahil has ruled:\n%s\n\nFollow up. If you agree on the plan, output the full plan block ending with AGREED.", text)
}

func SplitReview(otherName, otherText string) string {
	return fmt.Sprintf("%s's verdict:\n%s\n\nYou may keep or change your verdict after seeing theirs. Reply with exactly one line VERDICT: PASS or VERDICT: FAIL, then at most five short bullets.", otherName, otherText)
}

func ReviewAsk(diff, verifyOut string, verifyExit int, claim string) string {
	return fmt.Sprintf(`Worker finished. Claim (worker's claim, unverified):
%s

Diff:
%s

VERIFY output (exit %d):
%s

Reply with exactly one line VERDICT: PASS or VERDICT: FAIL, then at most five short bullets. Do not police python or timeout wrappers.`, claim, diff, verifyExit, verifyOut)
}

func WorkerQuestion(question string) string {
	return fmt.Sprintf("The worker asked this question:\n%s\n\nReply with one line ANSWER: <text> if you have a decision, otherwise say you disagree.", question)
}

func WorkerFailed(report string) string {
	return fmt.Sprintf("Worker failed/timed out. Report:\n%s\n\nReply ending with exactly one line: FIX to start a fix round, or ESCALATE to send to Sahil.", report)
}

func Interject(text string) string {
	return fmt.Sprintf("Message from Sahil:\n%s", text)
}
