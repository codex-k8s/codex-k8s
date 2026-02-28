package staff

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	docsetdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/docset"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
	"github.com/google/uuid"
)

func (s *Service) UpsertRepositoryBotParams(ctx context.Context, principal Principal, projectID string, repositoryID string, botTokenRaw *string, botUsername *string, botEmail *string) error {
	if projectID == "" {
		return errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if repositoryID == "" {
		return errs.Validation{Field: "repository_id", Msg: "is required"}
	}

	role := "admin"
	if !principal.IsPlatformAdmin {
		r, ok, err := s.members.GetRole(ctx, projectID, principal.UserID)
		if err != nil {
			return err
		}
		if !ok {
			return errs.Forbidden{Msg: "project access required"}
		}
		role = r
	}
	if role != "admin" && role != "read_write" {
		return errs.Forbidden{Msg: "project write access required"}
	}

	var enc []byte
	if botTokenRaw != nil {
		raw := strings.TrimSpace(*botTokenRaw)
		if raw != "" {
			encrypted, err := s.tokencrypt.EncryptString(raw)
			if err != nil {
				return fmt.Errorf("encrypt repository bot token: %w", err)
			}
			enc = encrypted
		}
	} else {
		// Preserve existing bot token if caller does not provide one.
		current, ok, err := s.repos.GetBotTokenEncrypted(ctx, repositoryID)
		if err != nil {
			return err
		}
		if ok {
			enc = current
		}
	}

	username := ""
	if botUsername != nil {
		username = strings.TrimSpace(*botUsername)
	}
	email := ""
	if botEmail != nil {
		email = strings.TrimSpace(*botEmail)
	}

	return s.repos.UpsertBotParams(ctx, repocfgrepo.RepositoryBotParamsUpsertParams{
		RepositoryID:      repositoryID,
		BotTokenEncrypted: enc,
		BotUsername:       username,
		BotEmail:          email,
	})
}

func (s *Service) RunRepositoryPreflight(ctx context.Context, principal Principal, projectID string, repositoryID string) (valuetypes.GitHubPreflightReport, error) {
	if !principal.IsPlatformAdmin {
		return valuetypes.GitHubPreflightReport{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if projectID == "" {
		return valuetypes.GitHubPreflightReport{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if repositoryID == "" {
		return valuetypes.GitHubPreflightReport{}, errs.Validation{Field: "repository_id", Msg: "is required"}
	}

	repo, ok, err := s.repos.GetByID(ctx, repositoryID)
	if err != nil {
		return valuetypes.GitHubPreflightReport{}, err
	}
	if !ok {
		return valuetypes.GitHubPreflightReport{}, errs.Validation{Field: "repository_id", Msg: "not found"}
	}

	lockToken := uuid.NewString()
	acquiredToken, acquired, err := s.repos.AcquirePreflightLock(ctx, repocfgrepo.RepositoryPreflightLockAcquireParams{
		RepositoryID:   repositoryID,
		LockToken:      lockToken,
		LockedByUserID: principal.UserID,
		LockedUntilUTC: time.Now().UTC().Add(10 * time.Minute),
	})
	if err != nil {
		return valuetypes.GitHubPreflightReport{}, err
	}
	if !acquired {
		return valuetypes.GitHubPreflightReport{}, errs.Conflict{Msg: "repository preflight is already running"}
	}
	lockToken = acquiredToken
	defer func() {
		_ = s.repos.ReleasePreflightLock(ctx, repositoryID, lockToken)
	}()

	platformToken, botToken, platformScope, botScope, err := s.resolveEffectiveGitHubTokens(ctx, projectID, repositoryID)
	if err != nil {
		return valuetypes.GitHubPreflightReport{}, err
	}

	expectedHost, expectedIPs := resolveExpectedIngressIPs(s.cfg.WebhookSpec.URL)

	report := valuetypes.GitHubPreflightReport{
		Status: "running",
		TokenScopes: valuetypes.GitHubPreflightTokenScopes{
			Platform: platformScope,
			Bot:      botScope,
		},
		Checks:     make([]valuetypes.GitHubPreflightCheck, 0, 32),
		Artifacts:  make([]valuetypes.GitHubPreflightArtifact, 0),
		FinishedAt: time.Time{},
	}
	report.Checks = append(report.Checks,
		valuetypes.GitHubPreflightCheck{Name: "github:tokens:platform_scope", Status: "ok", Details: platformScope},
		valuetypes.GitHubPreflightCheck{Name: "github:tokens:bot_scope", Status: "ok", Details: botScope},
	)

	hasFailures := false

	dnsCandidates := make([]dnsCandidate, 0, 8)

	// Always validate that the platform webhook host resolves (best-effort expected ingress).
	if expectedHost != "" {
		dnsCandidates = append(dnsCandidates, dnsCandidate{CheckName: "dns:platform:webhook_host", Domain: expectedHost})
	}

	if s.githubMgmt == nil {
		report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:preflight", Status: "failed", Details: "github management client is not configured"})
		hasFailures = true
	} else {
		baseBranch, branchErr := s.githubMgmt.GetDefaultBranch(ctx, platformToken, repo.Owner, repo.Name)
		if branchErr != nil {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:default_branch", Status: "failed", Details: branchErr.Error()})
			hasFailures = true
		} else {
			servicesPath := strings.TrimSpace(repo.ServicesYAMLPath)
			if servicesPath == "" {
				servicesPath = "services.yaml"
			}

			servicesYAML, found, getErr := s.githubMgmt.GetFile(ctx, platformToken, repo.Owner, repo.Name, servicesPath, baseBranch)
			if getErr != nil {
				report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:get", Status: "failed", Details: getErr.Error()})
				hasFailures = true
			} else if !found {
				report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:get", Status: "failed", Details: fmt.Sprintf("%s not found on %s", servicesPath, baseBranch)})
				hasFailures = true
			} else {
				report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:get", Status: "ok"})

				envNames, parseErr := listServicesYAMLEnvironments(servicesYAML)
				if parseErr != nil {
					report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:parse", Status: "failed", Details: parseErr.Error()})
					hasFailures = true
				} else {
					report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:parse", Status: "ok"})

					vars := envVarsMap()

					for _, item := range servicesYAMLPreflightEnvSlots {
						if _, ok := envNames[item.Env]; !ok {
							report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:env:" + item.Env, Status: "skipped", Details: "environment not defined"})
							continue
						}

						domain, source, ns, err := resolveServicesYAMLDomain(servicesYAML, item.Env, item.Slot, vars)
						if err != nil {
							report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:domain:" + item.Env, Status: "failed", Details: err.Error()})
							hasFailures = true
							continue
						}
						if strings.TrimSpace(domain) == "" {
							report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "services_yaml:domain:" + item.Env, Status: "failed", Details: "resolved domain is empty"})
							hasFailures = true
							continue
						}

						report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{
							Name:    "services_yaml:domain:" + item.Env,
							Status:  "ok",
							Details: fmt.Sprintf("source=%s namespace=%s domain=%s", source, ns, domain),
						})
						dnsCandidates = append(dnsCandidates, dnsCandidate{
							CheckName: "dns:services_yaml:" + item.Env + ":" + domain,
							Domain:    domain,
						})
					}
				}
			}
		}
	}

	for _, candidate := range dnsCandidates {
		check := runDNSCheck(candidate.CheckName, candidate.Domain, expectedIPs)
		if check.Status != "ok" {
			hasFailures = true
		}
		report.Checks = append(report.Checks, check)
	}

	if s.githubMgmt != nil {
		ghReport, ghErr := s.githubMgmt.Preflight(ctx, valuetypes.GitHubPreflightParams{
			PlatformToken: platformToken,
			BotToken:      botToken,
			Owner:         repo.Owner,
			Repository:    repo.Name,
			WebhookURL:    s.cfg.WebhookSpec.URL,
			WebhookSecret: s.cfg.WebhookSpec.Secret,
		})
		if ghErr != nil {
			report.Checks = append(report.Checks, valuetypes.GitHubPreflightCheck{Name: "github:preflight", Status: "failed", Details: ghErr.Error()})
			hasFailures = true
		} else {
			report.Checks = append(report.Checks, ghReport.Checks...)
			report.Artifacts = append(report.Artifacts, ghReport.Artifacts...)
			if strings.TrimSpace(ghReport.Status) != "ok" {
				hasFailures = true
			}
		}
	}

	report.FinishedAt = time.Now().UTC()
	if hasFailures {
		report.Status = "failed"
	} else {
		report.Status = "ok"
	}

	encoded, _ := json.Marshal(report)
	_ = s.repos.UpsertPreflightReport(ctx, repocfgrepo.RepositoryPreflightReportUpsertParams{
		RepositoryID: repositoryID,
		ReportJSON:   encoded,
	})

	return report, nil
}

func (s *Service) ListDocsetGroups(ctx context.Context, principal Principal, docsetRef string, locale string) ([]querytypes.DocsetGroup, error) {
	if !principal.IsPlatformAdmin {
		return nil, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.githubMgmt == nil {
		return nil, fmt.Errorf("failed_precondition: github management client is not configured")
	}
	token, err := s.resolvePlatformManagementToken(ctx)
	if err != nil {
		return nil, err
	}

	docsetRef = strings.TrimSpace(docsetRef)
	if docsetRef == "" {
		docsetRef = "main"
	}
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "" {
		locale = "ru"
	}

	manifestBlob, ok, err := s.githubMgmt.GetFile(ctx, token, "codex-k8s", "agent-knowledge-base", "docset.manifest.json", docsetRef)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("docset.manifest.json not found at ref %q", docsetRef)
	}
	manifest, err := docsetdomain.ParseManifest(manifestBlob)
	if err != nil {
		return nil, err
	}

	out := make([]querytypes.DocsetGroup, 0, len(manifest.Groups))
	for _, g := range manifest.Groups {
		out = append(out, querytypes.DocsetGroup{
			ID:              g.ID,
			Title:           g.Title.ForLocale(locale),
			Description:     g.Description.ForLocale(locale),
			DefaultSelected: g.DefaultSelected,
		})
	}
	return out, nil
}

func (s *Service) ImportDocset(ctx context.Context, principal Principal, projectID string, repositoryID string, docsetRef string, locale string, groupIDs []string) (querytypes.DocsetImportResult, error) {
	if !principal.IsPlatformAdmin {
		return querytypes.DocsetImportResult{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.githubMgmt == nil {
		return querytypes.DocsetImportResult{}, fmt.Errorf("failed_precondition: github management client is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	repositoryID = strings.TrimSpace(repositoryID)
	if projectID == "" {
		return querytypes.DocsetImportResult{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if repositoryID == "" {
		return querytypes.DocsetImportResult{}, errs.Validation{Field: "repository_id", Msg: "is required"}
	}
	docsetRef = strings.TrimSpace(docsetRef)
	if docsetRef == "" {
		docsetRef = "main"
	}
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "" {
		locale = "ru"
	}

	targetRepo, ok, err := s.repos.GetByID(ctx, repositoryID)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}
	if !ok {
		return querytypes.DocsetImportResult{}, errs.Validation{Field: "repository_id", Msg: "not found"}
	}

	token, _, _, _, err := s.resolveEffectiveGitHubTokens(ctx, projectID, repositoryID)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}

	manifestBlob, ok, err := s.githubMgmt.GetFile(ctx, token, "codex-k8s", "agent-knowledge-base", "docset.manifest.json", docsetRef)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}
	if !ok {
		return querytypes.DocsetImportResult{}, fmt.Errorf("docset.manifest.json not found at ref %q", docsetRef)
	}
	manifest, err := docsetdomain.ParseManifest(manifestBlob)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}

	plan, selectedGroups, err := docsetdomain.BuildImportPlan(manifest, locale, groupIDs)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}

	files := make(map[string][]byte, len(plan.Files)+1)
	lockFiles := make([]valuetypes.DocsetLockFile, 0, len(plan.Files))
	for _, f := range plan.Files {
		blob, ok, err := s.githubMgmt.GetFile(ctx, token, "codex-k8s", "agent-knowledge-base", f.SrcPath, docsetRef)
		if err != nil {
			return querytypes.DocsetImportResult{}, err
		}
		if !ok {
			return querytypes.DocsetImportResult{}, fmt.Errorf("docset source file %q not found at ref %q", f.SrcPath, docsetRef)
		}
		if f.ExpectedSHA256 != "" {
			if got := docsetdomain.SHA256Hex(blob); got != f.ExpectedSHA256 {
				return querytypes.DocsetImportResult{}, fmt.Errorf("sha256 mismatch for %s: got %s want %s", f.SrcPath, got, f.ExpectedSHA256)
			}
		}
		files[f.DstPath] = blob
		lockFiles = append(lockFiles, valuetypes.DocsetLockFile{
			Path:       f.DstPath,
			SHA256:     docsetdomain.SHA256Hex(blob),
			SourcePath: f.SrcPath,
		})
	}

	lock := docsetdomain.NewLock(manifest.Docset.ID, docsetRef, locale, selectedGroups, lockFiles)
	lockBlob, err := docsetdomain.MarshalLock(lock)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}
	files["docs/.docset-lock.json"] = lockBlob

	baseBranch, err := s.githubMgmt.GetDefaultBranch(ctx, token, targetRepo.Owner, targetRepo.Name)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}
	branch := fmt.Sprintf("codex-k8s-docset-import/%s", time.Now().UTC().Format("20060102-150405"))
	title := fmt.Sprintf("chore(docs): import docset %s (%s)", manifest.Docset.ID, docsetRef)
	body := fmt.Sprintf("Docset import\n\n- docset: %s\n- ref: %s\n- locale: %s\n- groups: %s\n- files: %d\n", manifest.Docset.ID, docsetRef, locale, strings.Join(selectedGroups, ", "), len(plan.Files))
	prNumber, prURL, err := s.githubMgmt.CreatePullRequestWithFiles(ctx, token, targetRepo.Owner, targetRepo.Name, baseBranch, branch, title, body, files)
	if err != nil {
		return querytypes.DocsetImportResult{}, err
	}

	return querytypes.DocsetImportResult{
		RepositoryFullName: targetRepo.Owner + "/" + targetRepo.Name,
		PRNumber:           prNumber,
		PRURL:              prURL,
		Branch:             branch,
		FilesTotal:         len(plan.Files),
	}, nil
}

func (s *Service) SyncDocset(ctx context.Context, principal Principal, projectID string, repositoryID string, docsetRef string) (querytypes.DocsetSyncResult, error) {
	if !principal.IsPlatformAdmin {
		return querytypes.DocsetSyncResult{}, errs.Forbidden{Msg: "platform admin required"}
	}
	if s.githubMgmt == nil {
		return querytypes.DocsetSyncResult{}, fmt.Errorf("failed_precondition: github management client is not configured")
	}
	projectID = strings.TrimSpace(projectID)
	repositoryID = strings.TrimSpace(repositoryID)
	if projectID == "" {
		return querytypes.DocsetSyncResult{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if repositoryID == "" {
		return querytypes.DocsetSyncResult{}, errs.Validation{Field: "repository_id", Msg: "is required"}
	}
	docsetRef = strings.TrimSpace(docsetRef)
	if docsetRef == "" {
		return querytypes.DocsetSyncResult{}, errs.Validation{Field: "docset_ref", Msg: "is required"}
	}

	targetRepo, ok, err := s.repos.GetByID(ctx, repositoryID)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	if !ok {
		return querytypes.DocsetSyncResult{}, errs.Validation{Field: "repository_id", Msg: "not found"}
	}

	token, _, _, _, err := s.resolveEffectiveGitHubTokens(ctx, projectID, repositoryID)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	baseBranch, err := s.githubMgmt.GetDefaultBranch(ctx, token, targetRepo.Owner, targetRepo.Name)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}

	lockBlob, ok, err := s.githubMgmt.GetFile(ctx, token, targetRepo.Owner, targetRepo.Name, "docs/.docset-lock.json", baseBranch)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	if !ok {
		return querytypes.DocsetSyncResult{}, fmt.Errorf("docset lock not found: docs/.docset-lock.json (run import first)")
	}
	lock, err := docsetdomain.ParseLock(lockBlob)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	locale := strings.ToLower(strings.TrimSpace(lock.Docset.Locale))
	if locale == "" {
		locale = "ru"
	}

	manifestBlob, ok, err := s.githubMgmt.GetFile(ctx, token, "codex-k8s", "agent-knowledge-base", "docset.manifest.json", docsetRef)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	if !ok {
		return querytypes.DocsetSyncResult{}, fmt.Errorf("docset.manifest.json not found at ref %q", docsetRef)
	}
	manifest, err := docsetdomain.ParseManifest(manifestBlob)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}

	currentSHA := make(map[string]string, len(lock.Files))
	for _, f := range lock.Files {
		blob, ok, err := s.githubMgmt.GetFile(ctx, token, targetRepo.Owner, targetRepo.Name, f.Path, baseBranch)
		if err != nil {
			return querytypes.DocsetSyncResult{}, err
		}
		if !ok {
			currentSHA[f.Path] = ""
			continue
		}
		currentSHA[f.Path] = docsetdomain.SHA256Hex(blob)
	}

	plan, err := docsetdomain.BuildSafeSyncPlan(lock, manifest, locale, currentSHA)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}

	files := make(map[string][]byte, len(plan.Updates)+1)
	updatedLockFiles := make([]valuetypes.DocsetLockFile, 0, len(plan.Updates))
	for _, f := range plan.Updates {
		blob, ok, err := s.githubMgmt.GetFile(ctx, token, "codex-k8s", "agent-knowledge-base", f.SrcPath, docsetRef)
		if err != nil {
			return querytypes.DocsetSyncResult{}, err
		}
		if !ok {
			return querytypes.DocsetSyncResult{}, fmt.Errorf("docset source file %q not found at ref %q", f.SrcPath, docsetRef)
		}
		if f.ExpectedSHA256 != "" {
			if got := docsetdomain.SHA256Hex(blob); got != f.ExpectedSHA256 {
				return querytypes.DocsetSyncResult{}, fmt.Errorf("sha256 mismatch for %s: got %s want %s", f.SrcPath, got, f.ExpectedSHA256)
			}
		}
		files[f.DstPath] = blob
		updatedLockFiles = append(updatedLockFiles, valuetypes.DocsetLockFile{
			Path:       f.DstPath,
			SHA256:     docsetdomain.SHA256Hex(blob),
			SourcePath: f.SrcPath,
		})
	}

	nextLock, err := docsetdomain.UpdateLockForSync(lock, docsetRef, updatedLockFiles)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	lockOut, err := docsetdomain.MarshalLock(nextLock)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}
	files["docs/.docset-lock.json"] = lockOut

	branch := fmt.Sprintf("codex-k8s-docset-sync/%s", time.Now().UTC().Format("20060102-150405"))
	title := fmt.Sprintf("chore(docs): sync docset %s (%s)", manifest.Docset.ID, docsetRef)
	body := fmt.Sprintf("Docset sync\n\n- docset: %s\n- ref: %s\n- locale: %s\n- updated: %d\n- drift: %d\n", manifest.Docset.ID, docsetRef, locale, len(plan.Updates), len(plan.Drift))
	prNumber, prURL, err := s.githubMgmt.CreatePullRequestWithFiles(ctx, token, targetRepo.Owner, targetRepo.Name, baseBranch, branch, title, body, files)
	if err != nil {
		return querytypes.DocsetSyncResult{}, err
	}

	return querytypes.DocsetSyncResult{
		RepositoryFullName: targetRepo.Owner + "/" + targetRepo.Name,
		PRNumber:           prNumber,
		PRURL:              prURL,
		Branch:             branch,
		FilesUpdated:       len(plan.Updates),
		FilesDrift:         len(plan.Drift),
	}, nil
}

func (s *Service) resolvePlatformManagementToken(ctx context.Context) (string, error) {
	if s.platformTokens == nil {
		return "", fmt.Errorf("failed_precondition: platform tokens repository is not configured")
	}
	item, ok, err := s.platformTokens.Get(ctx)
	if err != nil {
		return "", err
	}
	if !ok || len(item.PlatformTokenEncrypted) == 0 {
		return "", fmt.Errorf("failed_precondition: platform token is not configured")
	}
	raw, err := s.tokencrypt.DecryptString(item.PlatformTokenEncrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt platform token: %w", err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("failed_precondition: platform token is empty after decrypt")
	}
	return raw, nil
}

func (s *Service) resolveEffectiveGitHubTokens(ctx context.Context, projectID string, repositoryID string) (platformToken string, botToken string, platformScope string, botScope string, err error) {
	repoPlatformEnc, _, encErr := s.repos.GetTokenEncrypted(ctx, repositoryID)
	if encErr != nil {
		return "", "", "", "", encErr
	}
	if len(repoPlatformEnc) > 0 {
		raw, decErr := s.tokencrypt.DecryptString(repoPlatformEnc)
		if decErr == nil && strings.TrimSpace(raw) != "" {
			platformToken = strings.TrimSpace(raw)
			platformScope = "repository"
		}
	}

	repoBotEnc, _, botErr := s.repos.GetBotTokenEncrypted(ctx, repositoryID)
	if botErr != nil {
		return "", "", "", "", botErr
	}
	if len(repoBotEnc) > 0 {
		raw, decErr := s.tokencrypt.DecryptString(repoBotEnc)
		if decErr == nil && strings.TrimSpace(raw) != "" {
			botToken = strings.TrimSpace(raw)
			botScope = "repository"
		}
	}

	if (platformToken == "" || botToken == "") && s.projectTokens != nil && projectID != "" {
		projPlatformEnc, projBotEnc, _, _, ok, projErr := s.projectTokens.GetEncryptedByProjectID(ctx, projectID)
		if projErr != nil {
			return "", "", "", "", projErr
		}
		if ok {
			if platformToken == "" && len(projPlatformEnc) > 0 {
				raw, decErr := s.tokencrypt.DecryptString(projPlatformEnc)
				if decErr == nil && strings.TrimSpace(raw) != "" {
					platformToken = strings.TrimSpace(raw)
					platformScope = "project"
				}
			}
			if botToken == "" && len(projBotEnc) > 0 {
				raw, decErr := s.tokencrypt.DecryptString(projBotEnc)
				if decErr == nil && strings.TrimSpace(raw) != "" {
					botToken = strings.TrimSpace(raw)
					botScope = "project"
				}
			}
		}
	}

	if (platformToken == "" || botToken == "") && s.platformTokens != nil {
		item, ok, tokErr := s.platformTokens.Get(ctx)
		if tokErr != nil {
			return "", "", "", "", tokErr
		}
		if ok {
			if platformToken == "" && len(item.PlatformTokenEncrypted) > 0 {
				raw, decErr := s.tokencrypt.DecryptString(item.PlatformTokenEncrypted)
				if decErr == nil && strings.TrimSpace(raw) != "" {
					platformToken = strings.TrimSpace(raw)
					platformScope = "platform"
				}
			}
			if botToken == "" && len(item.BotTokenEncrypted) > 0 {
				raw, decErr := s.tokencrypt.DecryptString(item.BotTokenEncrypted)
				if decErr == nil && strings.TrimSpace(raw) != "" {
					botToken = strings.TrimSpace(raw)
					botScope = "platform"
				}
			}
		}
	}

	if platformToken == "" {
		return "", "", "", "", fmt.Errorf("failed_precondition: effective platform token is not configured (repo/project/platform fallback empty)")
	}
	if botToken == "" {
		return "", "", "", "", fmt.Errorf("failed_precondition: effective bot token is not configured (repo/project/platform fallback empty)")
	}
	return platformToken, botToken, strings.TrimSpace(platformScope), strings.TrimSpace(botScope), nil
}
