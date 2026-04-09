package runner

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	cpclient "github.com/codex-k8s/kodex/services/jobs/agent-runner/internal/controlplane"
)

type outputRepairPromptTemplateData struct {
	PromptLocale     string
	MissingFields    []string
	TargetBranch     string
	ExistingPRNumber int
}

func (s *Service) resolveCodexReport(ctx context.Context, state codexState, result *runResult, runStartedAt time.Time, outputSchemaFile string, codexOutput []byte, requiresPRFlow bool) (codexReport, []byte, error) {
	report, _, parseErr := parseCodexReportOutput(codexOutput)
	if parseErr == nil {
		report = enrichCodexReportWithLocalState(report, *result)
	}

	recoveredReport, recovered, recoverErr := s.recoverCodexReportFromControlPlane(ctx, *result, report, requiresPRFlow)
	if recoverErr != nil {
		s.logger.Warn("recover codex report from control-plane failed", "run_id", s.cfg.RunID, "err", recoverErr)
	}
	if recovered {
		report = recoveredReport
	}
	report = enrichCodexReportWithLocalState(report, *result)
	if missing := missingCriticalCodexReportFields(report, requiresPRFlow); len(missing) == 0 {
		return report, nil, nil
	}

	repairOutput, repairErr := s.repairCodexStructuredOutput(ctx, state, result, runStartedAt, outputSchemaFile, report, requiresPRFlow)
	if repairErr != nil {
		if parseErr != nil {
			return codexReport{}, nil, fmt.Errorf("parse codex structured output: %w", parseErr)
		}
		return codexReport{}, nil, repairErr
	}

	repairedReport, _, err := parseCodexReportOutput(repairOutput)
	if err != nil {
		return codexReport{}, nil, fmt.Errorf("parse repaired codex structured output: %w", err)
	}
	repairedReport = enrichCodexReportWithLocalState(repairedReport, *result)

	recoveredReport, recovered, recoverErr = s.recoverCodexReportFromControlPlane(ctx, *result, repairedReport, requiresPRFlow)
	if recoverErr != nil {
		s.logger.Warn("recover repaired codex report from control-plane failed", "run_id", s.cfg.RunID, "err", recoverErr)
	}
	if recovered {
		repairedReport = recoveredReport
	}
	repairedReport = enrichCodexReportWithLocalState(repairedReport, *result)
	if missing := missingCriticalCodexReportFields(repairedReport, requiresPRFlow); len(missing) > 0 {
		return codexReport{}, nil, fmt.Errorf("invalid codex result: missing required fields after repair: %s", strings.Join(missing, ", "))
	}
	return repairedReport, repairOutput, nil
}

func enrichCodexReportWithLocalState(report codexReport, result runResult) codexReport {
	if strings.TrimSpace(report.Branch) == "" {
		report.Branch = strings.TrimSpace(result.targetBranch)
	}
	if strings.TrimSpace(report.SessionID) == "" {
		report.SessionID = strings.TrimSpace(result.sessionID)
	}
	if report.PRNumber <= 0 && result.existingPRNumber > 0 {
		report.PRNumber = result.existingPRNumber
	}
	if strings.TrimSpace(report.PRURL) == "" {
		report.PRURL = strings.TrimSpace(result.prURL)
	}
	return report
}

func missingCriticalCodexReportFields(report codexReport, requiresPRFlow bool) []string {
	fields := make([]string, 0, 2)
	if !requiresPRFlow {
		return fields
	}
	if report.PRNumber <= 0 {
		fields = append(fields, outputFieldPRNumber)
	}
	if strings.TrimSpace(report.PRURL) == "" {
		fields = append(fields, outputFieldPRURL)
	}
	sort.Strings(fields)
	return fields
}

func (s *Service) recoverCodexReportFromControlPlane(ctx context.Context, result runResult, report codexReport, requiresPRFlow bool) (codexReport, bool, error) {
	if !requiresPRFlow || s.cp == nil {
		return report, false, nil
	}

	recovered := false
	prNumber := report.PRNumber
	prURL := strings.TrimSpace(report.PRURL)

	if prNumber <= 0 && result.existingPRNumber > 0 {
		prNumber = result.existingPRNumber
		recovered = true
	}

	if prNumber > 0 && prURL == "" {
		snapshot, found, err := s.lookupRunPullRequest(ctx, prNumber, "")
		if err != nil {
			return report, recovered, err
		}
		if found && strings.TrimSpace(snapshot.URL) != "" {
			prURL = strings.TrimSpace(snapshot.URL)
			recovered = true
		}
	}

	if prNumber <= 0 || prURL == "" {
		snapshot, found, err := s.lookupRunPullRequest(ctx, 0, result.targetBranch)
		if err != nil {
			return report, recovered, err
		}
		if found {
			if prNumber <= 0 && snapshot.Number > 0 {
				prNumber = snapshot.Number
				recovered = true
			}
			if prURL == "" && strings.TrimSpace(snapshot.URL) != "" {
				prURL = strings.TrimSpace(snapshot.URL)
				recovered = true
			}
		}
	}

	if !recovered {
		return report, false, nil
	}
	report.PRNumber = prNumber
	report.PRURL = prURL
	return report, true, nil
}

func (s *Service) lookupRunPullRequest(ctx context.Context, prNumber int, headBranch string) (pullRequestSnapshot, bool, error) {
	if s.cp == nil {
		return pullRequestSnapshot{}, false, nil
	}

	item, found, err := s.cp.LookupRunPullRequest(ctx, runPullRequestLookupParams(s.cfg, prNumber, headBranch))
	if err != nil {
		return pullRequestSnapshot{}, false, err
	}
	if !found || item.PRNumber <= 0 {
		return pullRequestSnapshot{}, false, nil
	}

	return pullRequestSnapshot{
		Number: item.PRNumber,
		URL:    strings.TrimSpace(item.PRURL),
		State:  strings.TrimSpace(item.PRState),
		Head:   strings.TrimSpace(item.HeadBranch),
		Base:   strings.TrimSpace(item.BaseBranch),
	}, true, nil
}

func (s *Service) repairCodexStructuredOutput(ctx context.Context, state codexState, result *runResult, runStartedAt time.Time, outputSchemaFile string, report codexReport, requiresPRFlow bool) ([]byte, error) {
	repairPrompt, err := renderTemplate(templateNamePromptOutputRepair, outputRepairPromptTemplateData{
		PromptLocale:     normalizePromptLocale(s.cfg.PromptTemplateLocale),
		MissingFields:    missingCriticalCodexReportFields(report, requiresPRFlow),
		TargetBranch:     result.targetBranch,
		ExistingPRNumber: maxInt(result.existingPRNumber, report.PRNumber),
	})
	if err != nil {
		return nil, fmt.Errorf("render output repair prompt: %w", err)
	}

	repairOutput, err := s.runCodexExecWithAuthRecovery(ctx, state, result, runStartedAt, codexExecParams{
		RepoDir:          state.repoDir,
		Resume:           result.restoredSessionPath != "" || result.sessionID != "" || result.sessionFilePath != "",
		ResumeSessionID:  result.sessionID,
		OutputSchemaFile: outputSchemaFile,
		Prompt:           strings.TrimSpace(repairPrompt),
	})
	if err != nil {
		return nil, fmt.Errorf("codex structured output repair failed: %w", err)
	}
	return repairOutput, nil
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

type pullRequestSnapshot struct {
	Number int
	URL    string
	State  string
	Head   string
	Base   string
}

func runPullRequestLookupParams(cfg Config, prNumber int, headBranch string) cpclient.RunPullRequestLookupParams {
	return cpclient.RunPullRequestLookupParams{
		ProjectID:          strings.TrimSpace(cfg.ProjectID),
		RepositoryFullName: strings.TrimSpace(cfg.RepositoryFullName),
		PRNumber:           prNumber,
		HeadBranch:         strings.TrimSpace(headBranch),
	}
}
