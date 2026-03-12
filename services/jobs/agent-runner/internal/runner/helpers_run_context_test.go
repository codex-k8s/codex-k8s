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
