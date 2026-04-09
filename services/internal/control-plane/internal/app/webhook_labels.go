package app

import "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/webhook"

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
	labels.RunDocAuditRevise = cfg.RunDocAuditReviseLabel
	labels.RunAIRepair = cfg.RunAIRepairLabel
	labels.RunQA = cfg.RunQALabel
	labels.RunQARevise = cfg.RunQAReviseLabel
	labels.RunRelease = cfg.RunReleaseLabel
	labels.RunReleaseRevise = cfg.RunReleaseReviseLabel
	labels.RunPostDeploy = cfg.RunPostDeployLabel
	labels.RunPostDeployRevise = cfg.RunPostDeployReviseLabel
	labels.RunOps = cfg.RunOpsLabel
	labels.RunOpsRevise = cfg.RunOpsReviseLabel
	labels.RunSelfImprove = cfg.RunSelfImproveLabel
	labels.RunSelfImproveRevise = cfg.RunSelfImproveReviseLabel
	labels.RunRethink = cfg.RunRethinkLabel
	labels.ModeDiscussion = cfg.ModeDiscussionLabel
	labels.NeedReviewer = cfg.NeedReviewerLabel
	return labels
}
