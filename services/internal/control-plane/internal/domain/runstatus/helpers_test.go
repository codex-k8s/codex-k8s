package runstatus

import (
	"encoding/json"
	"fmt"
	"testing"

	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func TestResolveCommentTarget(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		payload    querytypes.RunPayload
		wantKind   commentTargetKind
		wantNumber int
		expectErr  bool
	}{
		{
			name: "issue label trigger",
			payload: querytypes.RunPayload{
				Issue: &querytypes.RunPayloadIssue{Number: 77},
				Trigger: &querytypes.RunPayloadTrigger{
					Source: triggerSourceIssueLabel,
				},
			},
			wantKind:   commentTargetKindIssue,
			wantNumber: 77,
		},
		{
			name: "pull request review trigger uses pull request number",
			payload: querytypes.RunPayload{
				Issue:       &querytypes.RunPayloadIssue{Number: 77},
				PullRequest: &querytypes.RunPayloadPullRequest{Number: 200},
				Trigger: &querytypes.RunPayloadTrigger{
					Source: triggerSourcePullRequestReview,
				},
			},
			wantKind:   commentTargetKindPullRequest,
			wantNumber: 200,
		},
		{
			name: "pull request review trigger falls back to issue number",
			payload: querytypes.RunPayload{
				Issue: &querytypes.RunPayloadIssue{Number: 20},
				Trigger: &querytypes.RunPayloadTrigger{
					Source: triggerSourcePullRequestReview,
				},
			},
			wantKind:   commentTargetKindPullRequest,
			wantNumber: 20,
		},
		{
			name: "missing target returns error",
			payload: querytypes.RunPayload{
				Trigger: &querytypes.RunPayloadTrigger{
					Source: triggerSourcePullRequestReview,
				},
			},
			expectErr: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotKind, gotNumber, err := resolveCommentTarget(testCase.payload)
			if testCase.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveCommentTarget returned error: %v", err)
			}
			if gotKind != testCase.wantKind {
				t.Fatalf("unexpected target kind: got %q want %q", gotKind, testCase.wantKind)
			}
			if gotNumber != testCase.wantNumber {
				t.Fatalf("unexpected target number: got %d want %d", gotNumber, testCase.wantNumber)
			}
		})
	}
}

func TestPhaseOrder_PreparingRuntimeBetweenCreatedAndStarted(t *testing.T) {
	t.Parallel()

	if gotCreated, gotPreparing := phaseOrder(PhaseCreated), phaseOrder(PhasePreparingRuntime); gotPreparing <= gotCreated {
		t.Fatalf("expected preparing phase order to be greater than created: created=%d preparing=%d", gotCreated, gotPreparing)
	}
	if gotPreparing, gotStarted := phaseOrder(PhasePreparingRuntime), phaseOrder(PhaseStarted); gotPreparing >= gotStarted {
		t.Fatalf("expected preparing phase order to be less than started: preparing=%d started=%d", gotPreparing, gotStarted)
	}
	if gotResolved, gotReady := phaseOrder(PhaseAuthResolved), phaseOrder(PhaseReady); gotReady <= gotResolved {
		t.Fatalf("expected ready phase order to be greater than auth_resolved: auth_resolved=%d ready=%d", gotResolved, gotReady)
	}
}

func TestResolveUpsertTriggerKind(t *testing.T) {
	t.Parallel()

	if got := resolveUpsertTriggerKind("design", "dev"); got != "design" {
		t.Fatalf("resolveUpsertTriggerKind(design, dev) = %q, want %q", got, "design")
	}
	if got := resolveUpsertTriggerKind("", "design"); got != "design" {
		t.Fatalf("resolveUpsertTriggerKind(empty, design) = %q, want %q", got, "design")
	}
	if got := resolveUpsertTriggerKind("  ", ""); got != "dev" {
		t.Fatalf("resolveUpsertTriggerKind(blank, empty) = %q, want %q", got, "dev")
	}
}

func TestNormalizeRuntimeMode_PreservesExplicitCodeOnly(t *testing.T) {
	t.Parallel()

	if got := normalizeRuntimeMode(runtimeModeCode, triggerKindDev); got != runtimeModeCode {
		t.Fatalf("normalizeRuntimeMode(code-only, dev) = %q, want %q", got, runtimeModeCode)
	}
}

func TestNormalizeRuntimeMode_DefaultsDiscussionToCodeOnly(t *testing.T) {
	t.Parallel()

	if got := normalizeRuntimeMode("", triggerKindDiscussion); got != runtimeModeCode {
		t.Fatalf("normalizeRuntimeMode(empty, discussion) = %q, want %q", got, runtimeModeCode)
	}
}

func TestFindRunStatusComment_PrefersHigherPhaseAndFallsBackToLatestCommentByID(t *testing.T) {
	t.Parallel()

	bodyOld := testRunStatusCommentBody(t, commentState{RunID: "run-1", Phase: PhaseCreated})
	bodyNew := testRunStatusCommentBody(t, commentState{RunID: "run-1", Phase: PhaseStarted})

	comments := []mcpdomain.GitHubIssueComment{
		{ID: 101, Body: bodyOld},
		{ID: 109, Body: bodyNew},
	}

	gotComment, gotState, found := findRunStatusComment(comments, "run-1")
	if !found {
		t.Fatal("expected to find run status comment")
	}
	if gotComment.ID != 109 {
		t.Fatalf("expected latest comment id 109, got %d", gotComment.ID)
	}
	if gotState.Phase != PhaseStarted {
		t.Fatalf("expected phase %q from latest comment, got %q", PhaseStarted, gotState.Phase)
	}
}

func TestFindRunStatusComment_PrefersSucceededTerminalStatusOverLaterFailedDuplicate(t *testing.T) {
	t.Parallel()

	bodySucceeded := testRunStatusCommentBody(t, commentState{
		RunID:     "run-terminal",
		Phase:     PhaseFinished,
		RunStatus: runStatusSucceeded,
	})
	bodyFailed := testRunStatusCommentBody(t, commentState{
		RunID:     "run-terminal",
		Phase:     PhaseFinished,
		RunStatus: runStatusFailed,
	})

	comments := []mcpdomain.GitHubIssueComment{
		{ID: 401, Body: bodySucceeded},
		{ID: 402, Body: bodyFailed},
	}

	gotComment, gotState, found := findRunStatusComment(comments, "run-terminal")
	if !found {
		t.Fatal("expected to find run status comment")
	}
	if gotComment.ID != 401 {
		t.Fatalf("expected succeeded comment id 401 to win, got %d", gotComment.ID)
	}
	if gotState.RunStatus != runStatusSucceeded {
		t.Fatalf("expected terminal status %q, got %q", runStatusSucceeded, gotState.RunStatus)
	}
}

func TestFindRunStatusComment_IgnoresMalformedMarkerAndFindsValid(t *testing.T) {
	t.Parallel()

	malformed := "<!-- codex-k8s:run-status {not-valid-json} -->"
	valid := testRunStatusCommentBody(t, commentState{RunID: "run-2", Phase: PhaseFinished})

	comments := []mcpdomain.GitHubIssueComment{
		{ID: 201, Body: malformed},
		{ID: 202, Body: valid},
	}

	gotComment, gotState, found := findRunStatusComment(comments, "run-2")
	if !found {
		t.Fatal("expected to find valid run status comment")
	}
	if gotComment.ID != 202 {
		t.Fatalf("expected comment id 202, got %d", gotComment.ID)
	}
	if gotState.Phase != PhaseFinished {
		t.Fatalf("expected phase %q, got %q", PhaseFinished, gotState.Phase)
	}
}

func TestMergeState_PreservesSucceededTerminalStatusAgainstFailedDuplicate(t *testing.T) {
	t.Parallel()

	merged := mergeState(
		commentState{RunID: "run-1", Phase: PhaseFinished, RunStatus: runStatusSucceeded},
		commentState{RunID: "run-1", Phase: PhaseFinished, RunStatus: runStatusFailed},
	)

	if merged.RunStatus != runStatusSucceeded {
		t.Fatalf("expected merged terminal status %q, got %q", runStatusSucceeded, merged.RunStatus)
	}
}

func TestMergeState_PreservesTerminalStatusAgainstEarlierRunningUpdate(t *testing.T) {
	t.Parallel()

	merged := mergeState(
		commentState{RunID: "run-2", Phase: PhaseFinished, RunStatus: runStatusSucceeded},
		commentState{RunID: "run-2", Phase: PhaseStarted, RunStatus: "running"},
	)

	if merged.RunStatus != runStatusSucceeded {
		t.Fatalf("expected merged terminal status %q, got %q", runStatusSucceeded, merged.RunStatus)
	}
	if merged.Phase != PhaseFinished {
		t.Fatalf("expected phase %q, got %q", PhaseFinished, merged.Phase)
	}
}

func TestMergeState_PreservesAuthRequestedFlag(t *testing.T) {
	t.Parallel()

	merged := mergeState(
		commentState{RunID: "run-3", Phase: PhaseAuthRequired, AuthRequested: true},
		commentState{RunID: "run-3", Phase: PhaseReady},
	)

	if !merged.AuthRequested {
		t.Fatal("expected auth_requested flag to be preserved")
	}
}

func testRunStatusCommentBody(t *testing.T, state commentState) string {
	t.Helper()

	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal state marker: %v", err)
	}
	return fmt.Sprintf("status\n%s%s%s", commentMarkerPrefix, string(raw), commentMarkerSuffix)
}
