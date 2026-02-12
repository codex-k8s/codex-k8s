package runner

const (
	templateNamePromptEnvelope = "templates/prompt_envelope.tmpl"
	templateNamePromptWork     = "templates/prompt_work.tmpl"
	templateNamePromptReview   = "templates/prompt_review.tmpl"
	templateNameCodexConfig    = "templates/codex_config.toml.tmpl"

	promptTemplateKindWork   = "work"
	promptTemplateKindReview = "review"

	triggerKindDev       = "dev"
	triggerKindDevRevise = "dev_revise"

	runStatusSucceeded          = "succeeded"
	runStatusFailed             = "failed"
	runStatusFailedPrecondition = "failed_precondition"

	envContext7APIKey = "CODEXK8S_CONTEXT7_API_KEY"
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
    "reasoning_effort": { "type": "string" }
  },
  "required": ["summary", "branch", "pr_number", "pr_url"],
  "additionalProperties": true
}`
