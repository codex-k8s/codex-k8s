package runner

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestGitCleanArgs(t *testing.T) {
	t.Run("full-env keeps hard cleanup in isolated checkout", func(t *testing.T) {
		if got, want := gitCleanArgs(runtimeModeFullEnv), []string{"clean", "-fdx"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("gitCleanArgs(full-env) = %v, want %v", got, want)
		}
	})

	t.Run("code-only keeps hard cleanup", func(t *testing.T) {
		if got, want := gitCleanArgs(runtimeModeCodeOnly), []string{"clean", "-fdx"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("gitCleanArgs(code-only) = %v, want %v", got, want)
		}
	})

	t.Run("unknown mode defaults to code-only cleanup", func(t *testing.T) {
		if got, want := gitCleanArgs("unexpected"), []string{"clean", "-fdx"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("gitCleanArgs(unexpected) = %v, want %v", got, want)
		}
	})
}

func TestRunnerRepoDir(t *testing.T) {
	t.Run("code-only keeps repo checkout under workspace", func(t *testing.T) {
		if got, want := runnerRepoDir(runtimeModeCodeOnly), filepath.Join("/workspace", "repo"); got != want {
			t.Fatalf("runnerRepoDir(code-only) = %q, want %q", got, want)
		}
	})

	t.Run("full-env uses live workspace repo prepared by runtime deploy", func(t *testing.T) {
		want := "/workspace"
		if got := runnerRepoDir(runtimeModeFullEnv); got != want {
			t.Fatalf("runnerRepoDir(full-env) = %q, want %q", got, want)
		}
	})
}

func TestShouldRestoreLatestSession(t *testing.T) {
	t.Run("returns true for revise trigger", func(t *testing.T) {
		if !shouldRestoreLatestSession("dev_revise", false, "") {
			t.Fatal("expected revise trigger to require latest session restore")
		}
	})

	t.Run("returns true for discussion mode", func(t *testing.T) {
		if !shouldRestoreLatestSession("dev", true, "") {
			t.Fatal("expected discussion mode to require latest session restore")
		}
	})

	t.Run("returns true for interaction resume payload", func(t *testing.T) {
		if !shouldRestoreLatestSession("dev", false, `{"interaction_id":"interaction-1"}`) {
			t.Fatal("expected interaction resume payload to require latest session restore")
		}
	})

	t.Run("returns false for plain work run", func(t *testing.T) {
		if shouldRestoreLatestSession("dev", false, "") {
			t.Fatal("expected plain work run to skip latest session restore")
		}
	})
}

func TestIsInteractionResumeRun(t *testing.T) {
	t.Run("returns true for interaction resume correlation prefix", func(t *testing.T) {
		if !isInteractionResumeRun("interaction-resume:interaction-1") {
			t.Fatal("expected interaction resume correlation id to be detected")
		}
	})

	t.Run("returns false for regular correlation id", func(t *testing.T) {
		if isInteractionResumeRun("corr-1") {
			t.Fatal("expected regular correlation id to skip interaction resume detection")
		}
	})
}
