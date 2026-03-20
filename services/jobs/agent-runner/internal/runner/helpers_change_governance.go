package runner

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
)

const (
	changeGovernanceDraftKindInternalWorkingDraft = "internal_working_draft"

	changeGovernanceSurfaceDomain        = "domain"
	changeGovernanceSurfaceTransport     = "transport"
	changeGovernanceSurfaceSchema        = "schema"
	changeGovernanceSurfaceReleasePolicy = "release_policy"
	changeGovernanceSurfaceUI            = "ui"
	changeGovernanceSurfaceDocs          = "docs"

	changeGovernanceRiskDriverBlastRadius    = "blast_radius"
	changeGovernanceRiskDriverContractData   = "contract_data"
	changeGovernanceRiskDriverSecurityPolicy = "security_policy"
	changeGovernanceRiskDriverRuntimeRelease = "runtime_release"

	changeGovernanceDominantIntentCodeBehavior = "code_behavior"
	changeGovernanceDominantIntentSchema       = "schema"
	changeGovernanceDominantIntentTransport    = "transport"
	changeGovernanceDominantIntentUI           = "ui"
	changeGovernanceDominantIntentOps          = "ops"
	changeGovernanceDominantIntentDocsOnly     = "docs_only"

	changeGovernanceBoundedScopeSingleContext          = "single_context"
	changeGovernanceBoundedScopeCrossContext           = "cross_context"
	changeGovernanceBoundedScopeMechanicalBoundedScope = "mechanical_bounded_scope"

	changeGovernanceEvidenceScopePackage = "package"
	changeGovernanceEvidenceScopeWave    = "wave"

	changeGovernanceEvidenceBlockIntentContract = "intent_contract"
	changeGovernanceEvidenceBlockVerification   = "verification"

	changeGovernanceVerificationNotStarted = "not_started"
	changeGovernanceVerificationInProgress = "in_progress"

	changeGovernanceArtifactKindIssue        = "issue"
	changeGovernanceArtifactKindPullRequest  = "pull_request"
	changeGovernanceArtifactKindRun          = "run"
	changeGovernanceArtifactKindAgentSession = "agent_session"

	changeGovernanceArtifactRelationPrimaryContext = "primary_context"
	changeGovernanceArtifactRelationEvidenceSource = "evidence_source"

	changeGovernanceVerificationTargetUnit             = "unit"
	changeGovernanceVerificationTargetIntegration      = "integration"
	changeGovernanceVerificationTargetContract         = "contract"
	changeGovernanceVerificationTargetRegression       = "regression"
	changeGovernanceVerificationTargetRollback         = "rollback"
	changeGovernanceVerificationTargetReleaseReadiness = "release_readiness"
)

type changeGovernanceWaveCategory struct {
	Key          string
	SurfaceKind  string
	Intent       string
	Summary      string
	TargetKinds  []string
	IsMechanical bool
}

type changeGovernanceWaveSeed struct {
	category changeGovernanceWaveCategory
	contexts map[string]struct{}
}

func (s *Service) reportChangeGovernanceSignals(ctx context.Context, repoDir string, baselineHead string, result *runResult) error {
	if s == nil || s.cp == nil || result == nil || !s.cfg.QualityGovernanceEnabled {
		return nil
	}
	if s.cfg.IssueNumber <= 0 || strings.TrimSpace(s.cfg.ProjectID) == "" {
		return nil
	}

	currentHead, err := gitCurrentHead(ctx, repoDir)
	if err != nil {
		return fmt.Errorf("resolve change governance current head: %w", err)
	}
	changedPaths, err := collectChangedPathsSince(ctx, repoDir, baselineHead, currentHead)
	if err != nil {
		return fmt.Errorf("collect change governance changed paths: %w", err)
	}
	if len(changedPaths) == 0 {
		return nil
	}

	scopeHints := deriveChangeGovernanceScopeHints(changedPaths)
	waves := deriveChangeGovernanceWaves(changedPaths, strings.TrimSpace(result.report.Summary))
	if len(waves) == 0 {
		return nil
	}

	draftRef := strings.TrimSpace(result.sessionID)
	if draftRef == "" {
		draftRef = "run:" + strings.TrimSpace(s.cfg.RunID)
	}
	draftResult, err := s.cp.ReportChangeGovernanceDraftSignal(ctx, cpclient.ReportChangeGovernanceDraftSignalParams{
		RunID:                s.cfg.RunID,
		SignalID:             changeGovernanceSignalID(s.cfg.RunID, "draft"),
		CorrelationID:        s.cfg.CorrelationID,
		ProjectID:            s.cfg.ProjectID,
		RepositoryFullName:   s.cfg.RepositoryFullName,
		IssueNumber:          int(s.cfg.IssueNumber),
		PRNumber:             optionalInt(result.prNumber),
		BranchName:           result.targetBranch,
		DraftRef:             draftRef,
		DraftKind:            changeGovernanceDraftKindInternalWorkingDraft,
		ChangeScopeHints:     scopeHints,
		CandidateRiskDrivers: deriveChangeGovernanceRiskDrivers(changedPaths),
		DraftChecksum:        strings.TrimSpace(currentHead),
		OccurredAt:           time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("report change governance draft signal: %w", err)
	}

	waveResult, err := s.cp.PublishChangeGovernanceWaveMap(ctx, cpclient.PublishChangeGovernanceWaveMapParams{
		RunID:         s.cfg.RunID,
		PackageID:     draftResult.PackageID,
		WaveMapID:     changeGovernanceSignalID(s.cfg.RunID, "wave-map"),
		CorrelationID: s.cfg.CorrelationID,
		Waves:         waves,
		PublishedAt:   time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("publish change governance wave map: %w", err)
	}
	if err := s.upsertChangeGovernanceEvidenceSignals(ctx, waveResult.PackageID, waves, result); err != nil {
		return err
	}
	return nil
}

func (s *Service) upsertChangeGovernanceEvidenceSignals(ctx context.Context, packageID string, waves []cpclient.ChangeGovernanceWaveDraft, result *runResult) error {
	packageArtifacts := buildChangeGovernanceArtifactLinks(s.cfg.RepositoryFullName, int(s.cfg.IssueNumber), optionalIntValue(result.prNumber), s.cfg.RunID, result.sessionID)
	_, err := s.cp.UpsertChangeGovernanceEvidenceSignal(ctx, cpclient.UpsertChangeGovernanceEvidenceSignalParams{
		RunID:                 s.cfg.RunID,
		PackageID:             packageID,
		SignalID:              changeGovernanceSignalID(s.cfg.RunID, "evidence-package-intent"),
		CorrelationID:         s.cfg.CorrelationID,
		ScopeKind:             changeGovernanceEvidenceScopePackage,
		ScopeRef:              packageID,
		BlockKind:             changeGovernanceEvidenceBlockIntentContract,
		ArtifactLinks:         packageArtifacts,
		VerificationStateHint: changeGovernanceVerificationNotStarted,
		RequiredByTier:        true,
		OccurredAt:            time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("upsert change governance package evidence signal: %w", err)
	}

	waveArtifacts := buildChangeGovernanceArtifactLinks(s.cfg.RepositoryFullName, int(s.cfg.IssueNumber), optionalIntValue(result.prNumber), s.cfg.RunID, "")
	for _, wave := range waves {
		_, err := s.cp.UpsertChangeGovernanceEvidenceSignal(ctx, cpclient.UpsertChangeGovernanceEvidenceSignalParams{
			RunID:                 s.cfg.RunID,
			PackageID:             packageID,
			SignalID:              changeGovernanceSignalID(s.cfg.RunID, "evidence-wave-"+wave.WaveKey),
			CorrelationID:         s.cfg.CorrelationID,
			ScopeKind:             changeGovernanceEvidenceScopeWave,
			ScopeRef:              wave.WaveKey,
			BlockKind:             changeGovernanceEvidenceBlockVerification,
			ArtifactLinks:         waveArtifacts,
			VerificationStateHint: changeGovernanceVerificationInProgress,
			RequiredByTier:        true,
			OccurredAt:            time.Now().UTC(),
		})
		if err != nil {
			return fmt.Errorf("upsert change governance wave evidence signal for %s: %w", wave.WaveKey, err)
		}
	}
	return nil
}

func deriveChangeGovernanceScopeHints(changedPaths []string) []cpclient.ChangeGovernanceScopeHint {
	seen := make(map[string]struct{}, len(changedPaths))
	result := make([]cpclient.ChangeGovernanceScopeHint, 0, len(changedPaths))
	for _, changedPath := range changedPaths {
		contextKey, surfaceKind := changeGovernanceScopeForPath(changedPath)
		if contextKey == "" || surfaceKind == "" {
			continue
		}
		key := contextKey + "|" + surfaceKind
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, cpclient.ChangeGovernanceScopeHint{
			ContextKey:  contextKey,
			SurfaceKind: surfaceKind,
		})
	}
	slices.SortFunc(result, func(left, right cpclient.ChangeGovernanceScopeHint) int {
		return strings.Compare(
			left.ContextKey+"\x00"+left.SurfaceKind,
			right.ContextKey+"\x00"+right.SurfaceKind,
		)
	})
	return result
}

func deriveChangeGovernanceRiskDrivers(changedPaths []string) []string {
	contextRoots := make(map[string]struct{}, len(changedPaths))
	drivers := make(map[string]struct{}, 4)
	for _, changedPath := range changedPaths {
		category := changeGovernanceCategoryForPath(changedPath)
		contextRoots[changeGovernanceContextKey(changedPath)] = struct{}{}
		switch category.Key {
		case "schema", "transport":
			drivers[changeGovernanceRiskDriverContractData] = struct{}{}
		case "ops":
			drivers[changeGovernanceRiskDriverRuntimeRelease] = struct{}{}
		}
		lower := strings.ToLower(changedPath)
		if strings.Contains(lower, "auth") || strings.Contains(lower, "rbac") || strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "security") {
			drivers[changeGovernanceRiskDriverSecurityPolicy] = struct{}{}
		}
	}
	if len(contextRoots) > 1 {
		drivers[changeGovernanceRiskDriverBlastRadius] = struct{}{}
	}
	return sortedChangeGovernanceKeys(drivers)
}

func deriveChangeGovernanceWaves(changedPaths []string, reportSummary string) []cpclient.ChangeGovernanceWaveDraft {
	seedsByCategory := make(map[string]*changeGovernanceWaveSeed, 6)
	categoryOrder := make([]string, 0, 6)
	for _, changedPath := range changedPaths {
		category := changeGovernanceCategoryForPath(changedPath)
		if category.Key == "" {
			continue
		}
		seed, exists := seedsByCategory[category.Key]
		if !exists {
			seed = &changeGovernanceWaveSeed{
				category: category,
				contexts: make(map[string]struct{}, 4),
			}
			seedsByCategory[category.Key] = seed
			categoryOrder = append(categoryOrder, category.Key)
		}
		seed.contexts[changeGovernanceContextKey(changedPath)] = struct{}{}
	}
	if len(categoryOrder) == 0 {
		return nil
	}

	result := make([]cpclient.ChangeGovernanceWaveDraft, 0, len(categoryOrder))
	for index, key := range categoryOrder {
		seed := seedsByCategory[key]
		summary := seed.category.Summary
		if len(categoryOrder) == 1 && strings.TrimSpace(reportSummary) != "" {
			summary = strings.TrimSpace(reportSummary)
		}
		scopeKind := changeGovernanceBoundedScopeCrossContext
		if seed.category.IsMechanical {
			scopeKind = changeGovernanceBoundedScopeMechanicalBoundedScope
		} else if len(seed.contexts) <= 1 {
			scopeKind = changeGovernanceBoundedScopeSingleContext
		}
		result = append(result, cpclient.ChangeGovernanceWaveDraft{
			WaveKey:             seed.category.Key,
			PublishOrder:        index + 1,
			DominantIntent:      seed.category.Intent,
			BoundedScopeKind:    scopeKind,
			Summary:             summary,
			VerificationTargets: changeGovernanceVerificationTargets(seed.category.TargetKinds),
		})
	}
	return result
}

func changeGovernanceVerificationTargets(targetKinds []string) []cpclient.ChangeGovernanceVerificationTarget {
	result := make([]cpclient.ChangeGovernanceVerificationTarget, 0, len(targetKinds))
	for _, targetKind := range targetKinds {
		result = append(result, cpclient.ChangeGovernanceVerificationTarget{
			TargetKind: targetKind,
			TargetRef:  targetKind,
		})
	}
	return result
}

func changeGovernanceScopeForPath(path string) (string, string) {
	category := changeGovernanceCategoryForPath(path)
	return changeGovernanceContextKey(path), category.SurfaceKind
}

func changeGovernanceContextKey(path string) string {
	normalized := normalizeRepoRelativePath(path)
	parts := strings.Split(normalized, "/")
	switch {
	case len(parts) >= 3 && parts[0] == "services":
		return filepath.ToSlash(strings.Join(parts[:3], "/"))
	case len(parts) >= 3 && parts[0] == "deploy":
		return filepath.ToSlash(strings.Join(parts[:3], "/"))
	case len(parts) >= 4 && parts[0] == "proto":
		return filepath.ToSlash(strings.Join(parts[:4], "/"))
	case len(parts) >= 2 && parts[0] == "libs":
		return filepath.ToSlash(strings.Join(parts[:2], "/"))
	case len(parts) >= 2 && parts[0] == "docs":
		return filepath.ToSlash(strings.Join(parts[:2], "/"))
	case len(parts) >= 2:
		return filepath.ToSlash(strings.Join(parts[:2], "/"))
	default:
		return normalized
	}
}

func changeGovernanceCategoryForPath(path string) changeGovernanceWaveCategory {
	normalized := normalizeRepoRelativePath(path)
	lower := strings.ToLower(normalized)
	switch {
	case strings.HasPrefix(lower, "services/staff/web-console/"), strings.HasPrefix(lower, "libs/vue/"), strings.HasPrefix(lower, "libs/ts/"):
		return changeGovernanceWaveCategory{
			Key:         "ui",
			SurfaceKind: changeGovernanceSurfaceUI,
			Intent:      changeGovernanceDominantIntentUI,
			Summary:     "UI and operator-surface changes",
			TargetKinds: []string{changeGovernanceVerificationTargetRegression},
		}
	case strings.Contains(lower, "/migrations/"):
		return changeGovernanceWaveCategory{
			Key:         "schema",
			SurfaceKind: changeGovernanceSurfaceSchema,
			Intent:      changeGovernanceDominantIntentSchema,
			Summary:     "Schema and migration changes",
			TargetKinds: []string{changeGovernanceVerificationTargetContract, changeGovernanceVerificationTargetRegression, changeGovernanceVerificationTargetReleaseReadiness},
		}
	case strings.HasPrefix(lower, "proto/"), strings.HasPrefix(lower, "api/server/"), strings.HasPrefix(lower, "services/external/"):
		return changeGovernanceWaveCategory{
			Key:         "transport",
			SurfaceKind: changeGovernanceSurfaceTransport,
			Intent:      changeGovernanceDominantIntentTransport,
			Summary:     "Transport and contract changes",
			TargetKinds: []string{changeGovernanceVerificationTargetContract, changeGovernanceVerificationTargetRegression},
		}
	case strings.HasPrefix(lower, "deploy/"), strings.HasPrefix(lower, "bootstrap/"), lower == "services.yaml" || strings.HasSuffix(lower, "/dockerfile") || strings.HasSuffix(lower, ".yaml.tpl"):
		return changeGovernanceWaveCategory{
			Key:         "ops",
			SurfaceKind: changeGovernanceSurfaceReleasePolicy,
			Intent:      changeGovernanceDominantIntentOps,
			Summary:     "Runtime and release-orchestration changes",
			TargetKinds: []string{changeGovernanceVerificationTargetReleaseReadiness, changeGovernanceVerificationTargetRollback},
		}
	case strings.HasSuffix(lower, ".md"), strings.HasPrefix(lower, "docs/"), strings.HasSuffix(lower, "readme.md"), lower == "agents.md":
		return changeGovernanceWaveCategory{
			Key:          "docs",
			SurfaceKind:  changeGovernanceSurfaceDocs,
			Intent:       changeGovernanceDominantIntentDocsOnly,
			Summary:      "Documentation-only updates",
			TargetKinds:  nil,
			IsMechanical: true,
		}
	default:
		return changeGovernanceWaveCategory{
			Key:         "code",
			SurfaceKind: changeGovernanceSurfaceDomain,
			Intent:      changeGovernanceDominantIntentCodeBehavior,
			Summary:     "Service and domain behavior changes",
			TargetKinds: []string{changeGovernanceVerificationTargetUnit, changeGovernanceVerificationTargetIntegration},
		}
	}
}

func buildChangeGovernanceArtifactLinks(repositoryFullName string, issueNumber int, prNumber *int, runID string, sessionID string) []cpclient.ChangeGovernanceArtifactLinkSeed {
	result := []cpclient.ChangeGovernanceArtifactLinkSeed{
		{
			ArtifactKind: changeGovernanceArtifactKindIssue,
			ArtifactRef:  fmt.Sprintf("%s#%d", strings.TrimSpace(repositoryFullName), issueNumber),
			RelationKind: changeGovernanceArtifactRelationPrimaryContext,
			DisplayLabel: fmt.Sprintf("Issue #%d", issueNumber),
		},
		{
			ArtifactKind: changeGovernanceArtifactKindRun,
			ArtifactRef:  strings.TrimSpace(runID),
			RelationKind: changeGovernanceArtifactRelationEvidenceSource,
			DisplayLabel: fmt.Sprintf("Run %s", strings.TrimSpace(runID)),
		},
	}
	if prNumber != nil && *prNumber > 0 {
		result = append(result, cpclient.ChangeGovernanceArtifactLinkSeed{
			ArtifactKind: changeGovernanceArtifactKindPullRequest,
			ArtifactRef:  fmt.Sprintf("%s#%d", strings.TrimSpace(repositoryFullName), *prNumber),
			RelationKind: changeGovernanceArtifactRelationEvidenceSource,
			DisplayLabel: fmt.Sprintf("PR #%d", *prNumber),
		})
	}
	if strings.TrimSpace(sessionID) != "" {
		result = append(result, cpclient.ChangeGovernanceArtifactLinkSeed{
			ArtifactKind: changeGovernanceArtifactKindAgentSession,
			ArtifactRef:  strings.TrimSpace(sessionID),
			RelationKind: changeGovernanceArtifactRelationEvidenceSource,
			DisplayLabel: strings.TrimSpace(sessionID),
		})
	}
	return result
}

func changeGovernanceSignalID(runID string, suffix string) string {
	return strings.TrimSpace(runID) + ":" + strings.TrimSpace(suffix)
}

func optionalIntValue(value int) *int {
	if value <= 0 {
		return nil
	}
	result := value
	return &result
}

func sortedChangeGovernanceKeys(items map[string]struct{}) []string {
	result := make([]string, 0, len(items))
	for item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		result = append(result, item)
	}
	slices.Sort(result)
	return result
}
