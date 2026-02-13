package runner

const (
	templateNamePromptEnvelope = "templates/prompt_envelope.tmpl"
	templateNamePromptWork     = "templates/prompt_work.tmpl"
	templateNamePromptReview   = "templates/prompt_review.tmpl"
	templateNameCodexConfig    = "templates/codex_config.toml.tmpl"
	templateNameKubeconfig     = "templates/kubeconfig.tmpl"
	promptSeedsDirRelativePath = "docs/product/prompt-seeds"

	promptTemplateKindWork   = "work"
	promptTemplateKindReview = "review"

	promptLocaleRU = "ru"
	promptLocaleEN = "en"

	runtimeModeFullEnv  = "full-env"
	runtimeModeCodeOnly = "code-only"

	runStatusSucceeded          = "succeeded"
	runStatusFailed             = "failed"
	runStatusFailedPrecondition = "failed_precondition"

	envContext7APIKey = "CODEXK8S_CONTEXT7_API_KEY"

	sessionLogVersionV1      = "v1"
	maxCapturedCommandOutput = 256 * 1024

	envGitAskPass        = "GIT_ASKPASS"
	envGitTerminalPrompt = "GIT_TERMINAL_PROMPT"
	envGitAskPassRequire = "GIT_ASKPASS_REQUIRE"
	envGHToken           = "GH_TOKEN"
	envGitHubToken       = "GITHUB_TOKEN"
	envKubeconfig        = "KUBECONFIG"

	gitAskPassRequireForce = "force"
	redactedSecretValue    = "[REDACTED]"
)

const outputSchemaJSON = `{
  "type": "object",
  "properties": {
    "summary": { "type": "string" },
    "branch": { "type": "string" },
    "pr_number": { "type": "integer", "minimum": 1 },
    "pr_url": { "type": "string", "minLength": 1 },
    "session_id": { "type": "string" },
    "model": { "type": "string" },
    "reasoning_effort": { "type": "string" },
    "diagnosis": { "type": "string" },
    "action_items": {
      "type": "array",
      "items": { "type": "string" }
    },
    "evidence_refs": {
      "type": "array",
      "items": { "type": "string" }
    },
    "tool_gaps": {
      "type": "array",
      "items": { "type": "string" }
    }
  },
  "required": ["summary", "branch", "pr_number", "pr_url", "session_id", "model", "reasoning_effort"],
  "additionalProperties": false
}`
