package app

import "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/webhook"

func buildWebhookTriggerLabels(cfg Config) webhook.TriggerLabels {
	labels := webhook.TriggerLabels{}
	labels.RunIntake = cfg.RunIntakeLabel
	labels.RunIntakeRevise = cfg.RunIntakeReviseLabel
	labels.RunVision = cfg.RunVisionLabel
	labels.RunVisionRevise = cfg.RunVisionReviseLabel
	labels.RunPRD = cfg.RunPRDLabel
	labels.RunPRDRevise = cfg.RunPRDReviseLabel
	labels.RunArch = cfg.RunArchLabel
	labels.RunArchRevise = cfg.RunArchReviseLabel
	labels.RunDesign = cfg.RunDesignLabel
	labels.RunDesignRevise = cfg.RunDesignReviseLabel
	labels.RunPlan = cfg.RunPlanLabel
	labels.RunPlanRevise = cfg.RunPlanReviseLabel
	labels.RunDev = cfg.RunDevLabel
	labels.RunDevRevise = cfg.RunDevReviseLabel
	labels.RunDocAudit = cfg.RunDocAuditLabel
	labels.RunAIRepair = cfg.RunAIRepairLabel
	labels.RunQA = cfg.RunQALabel
	labels.RunRelease = cfg.RunReleaseLabel
	labels.RunPostDeploy = cfg.RunPostDeployLabel
	labels.RunOps = cfg.RunOpsLabel
	labels.RunSelfImprove = cfg.RunSelfImproveLabel
	labels.RunRethink = cfg.RunRethinkLabel
	return labels
}
