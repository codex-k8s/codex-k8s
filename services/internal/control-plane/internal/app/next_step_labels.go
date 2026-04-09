package app

import nextstepdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/nextstep"

func buildNextStepLabels(cfg Config) nextstepdomain.Labels {
	return nextstepdomain.NewLabels(nextstepdomain.Config{
		RunIntake:            cfg.RunIntakeLabel,
		RunIntakeRevise:      cfg.RunIntakeReviseLabel,
		RunVision:            cfg.RunVisionLabel,
		RunVisionRevise:      cfg.RunVisionReviseLabel,
		RunPRD:               cfg.RunPRDLabel,
		RunPRDRevise:         cfg.RunPRDReviseLabel,
		RunArch:              cfg.RunArchLabel,
		RunArchRevise:        cfg.RunArchReviseLabel,
		RunDesign:            cfg.RunDesignLabel,
		RunDesignRevise:      cfg.RunDesignReviseLabel,
		RunPlan:              cfg.RunPlanLabel,
		RunPlanRevise:        cfg.RunPlanReviseLabel,
		RunDev:               cfg.RunDevLabel,
		RunDevRevise:         cfg.RunDevReviseLabel,
		RunDocAudit:          cfg.RunDocAuditLabel,
		RunDocAuditRevise:    cfg.RunDocAuditReviseLabel,
		RunQA:                cfg.RunQALabel,
		RunQARevise:          cfg.RunQAReviseLabel,
		RunRelease:           cfg.RunReleaseLabel,
		RunReleaseRevise:     cfg.RunReleaseReviseLabel,
		RunPostDeploy:        cfg.RunPostDeployLabel,
		RunPostDeployRevise:  cfg.RunPostDeployReviseLabel,
		RunOps:               cfg.RunOpsLabel,
		RunOpsRevise:         cfg.RunOpsReviseLabel,
		RunSelfImprove:       cfg.RunSelfImproveLabel,
		RunSelfImproveRevise: cfg.RunSelfImproveReviseLabel,
		RunRethink:           cfg.RunRethinkLabel,
	})
}
