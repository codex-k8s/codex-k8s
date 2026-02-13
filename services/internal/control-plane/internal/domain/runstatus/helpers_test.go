package runstatus

import (
	"testing"

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
