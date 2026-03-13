package runtimedeploy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/manifesttpl"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

const runtimeFingerprintVersion = "v1"

const (
	runtimeFingerprintHashAnnotationKey    = "codex-k8s.dev/runtime-fingerprint-hash"
	runtimeFingerprintJSONAnnotationKey    = "codex-k8s.dev/runtime-fingerprint-json"
	runtimeFingerprintUpdatedAnnotationKey = "codex-k8s.dev/runtime-fingerprint-updated-at"
)

const (
	runtimeReuseReasonRuntimeModeNotFullEnv = "runtime_mode_not_full_env"
	runtimeReuseReasonDeployOnly            = "deploy_only_run"
	runtimeReuseReasonNamespaceNotManaged   = "namespace_not_managed"
	runtimeReuseReasonNamespaceTerminating  = "namespace_terminating"
	runtimeReuseReasonActiveTask            = "active_runtime_deploy_task"
)

const (
	runtimeReuseReasonBuildRefNotImmutable  = "build_ref_not_immutable"
	runtimeReuseReasonRepoSnapshotMissing   = "repo_snapshot_missing"
	runtimeReuseReasonRepoSnapshotStale     = "repo_snapshot_stale"
	runtimeReuseReasonServicesConfigMissing = "services_config_missing"
	runtimeReuseReasonNamespaceMismatch     = "namespace_mismatch"
)

const (
	runtimeReuseReasonFingerprintMissing   = "fingerprint_missing"
	runtimeReuseReasonFingerprintCorrupted = "fingerprint_corrupted"
	runtimeReuseReasonFingerprintMismatch  = "fingerprint_mismatch"
)

type runtimeFingerprintRecord struct {
	Version            string `json:"version"`
	RuntimeMode        string `json:"runtime_mode"`
	TargetEnv          string `json:"target_env"`
	Namespace          string `json:"namespace"`
	ProjectID          string `json:"project_id"`
	IssueNumber        int64  `json:"issue_number"`
	AgentKey           string `json:"agent_key"`
	RepositoryFullName string `json:"repository_full_name"`
	ServicesYAMLPath   string `json:"services_yaml_path"`
	BuildRef           string `json:"build_ref"`
	DeployOnly         bool   `json:"deploy_only"`
	ManifestHash       string `json:"manifest_hash"`
	Hash               string `json:"hash"`
}

type renderedManifestRecord struct {
	Unit string `json:"unit"`
	Path string `json:"path"`
	Hash string `json:"hash"`
}

type runtimeReuseInvalidation struct {
	reason            string
	effectiveBuildRef string
}

func (e runtimeReuseInvalidation) Error() string {
	return e.reason
}

// EvaluateRuntimeReuse determines whether a reusable namespace can skip runtime deploy/build.
func (s *Service) EvaluateRuntimeReuse(ctx context.Context, params EvaluateReuseParams) (EvaluateReuseResult, error) {
	params = normalizeEvaluateReuseParams(params)
	if params.RunID == "" {
		return EvaluateReuseResult{}, fmt.Errorf("run_id is required")
	}
	if params.Namespace == "" {
		return EvaluateReuseResult{}, fmt.Errorf("namespace is required")
	}

	result := EvaluateReuseResult{
		Namespace: params.Namespace,
		TargetEnv: params.TargetEnv,
	}

	if params.RuntimeMode != "full-env" {
		result.Reason = runtimeReuseReasonRuntimeModeNotFullEnv
		return result, nil
	}
	if params.DeployOnly {
		result.Reason = runtimeReuseReasonDeployOnly
		return result, nil
	}

	namespaceState, found, err := s.k8s.GetManagedRunNamespace(ctx, params.Namespace)
	if err != nil {
		return EvaluateReuseResult{}, fmt.Errorf("load managed namespace %s: %w", params.Namespace, err)
	}
	if !found {
		result.Reason = runtimeReuseReasonNamespaceNotManaged
		return result, nil
	}
	if namespaceState.Terminating {
		result.Reason = runtimeReuseReasonNamespaceTerminating
		return result, nil
	}

	activeTask, active, err := s.tasks.FindActiveByNamespace(ctx, params.Namespace)
	if err != nil {
		return EvaluateReuseResult{}, fmt.Errorf("load active runtime deploy task for namespace %s: %w", params.Namespace, err)
	}
	if active {
		result.Reason = runtimeReuseReasonActiveTask
		result.EffectiveBuildRef = strings.TrimSpace(activeTask.BuildRef)
		return result, nil
	}

	fingerprint, err := s.buildRuntimeFingerprint(ctx, params)
	if err != nil {
		var invalid runtimeReuseInvalidation
		if ok := asRuntimeReuseInvalidation(err, &invalid); ok {
			result.Reason = invalid.reason
			result.EffectiveBuildRef = invalid.effectiveBuildRef
			return result, nil
		}
		return EvaluateReuseResult{}, err
	}
	result.TargetEnv = fingerprint.TargetEnv
	result.EffectiveBuildRef = fingerprint.BuildRef
	result.FingerprintHash = fingerprint.Hash

	stored, ok := parseStoredRuntimeFingerprint(namespaceState.Annotations)
	if !ok {
		if strings.TrimSpace(namespaceState.Annotations[runtimeFingerprintJSONAnnotationKey]) != "" || strings.TrimSpace(namespaceState.Annotations[runtimeFingerprintHashAnnotationKey]) != "" {
			result.Reason = runtimeReuseReasonFingerprintCorrupted
		} else {
			result.Reason = runtimeReuseReasonFingerprintMissing
		}
		return result, nil
	}
	if strings.TrimSpace(stored.Hash) != strings.TrimSpace(fingerprint.Hash) {
		result.Reason = runtimeReuseReasonFingerprintMismatch
		return result, nil
	}

	result.Reusable = true
	return result, nil
}

func normalizeEvaluateReuseParams(params EvaluateReuseParams) EvaluateReuseParams {
	params.RunID = strings.TrimSpace(params.RunID)
	params.ProjectID = strings.TrimSpace(params.ProjectID)
	params.AgentKey = strings.ToLower(strings.TrimSpace(params.AgentKey))
	params.RuntimeMode = strings.TrimSpace(params.RuntimeMode)
	if params.RuntimeMode == "" {
		params.RuntimeMode = "full-env"
	}
	params.Namespace = strings.TrimSpace(params.Namespace)
	params.TargetEnv = strings.TrimSpace(params.TargetEnv)
	if params.TargetEnv == "" {
		params.TargetEnv = "ai"
	}
	if params.SlotNo < 0 {
		params.SlotNo = 0
	}
	params.RepositoryFullName = strings.TrimSpace(params.RepositoryFullName)
	params.ServicesYAMLPath = strings.TrimSpace(params.ServicesYAMLPath)
	if params.ServicesYAMLPath == "" {
		params.ServicesYAMLPath = defaultServicesConfigPath
	}
	params.BuildRef = resolveRuntimeBuildRef(params.BuildRef)
	return params
}

func (s *Service) buildRuntimeFingerprint(ctx context.Context, params EvaluateReuseParams) (runtimeFingerprintRecord, error) {
	params = normalizeEvaluateReuseParams(params)
	params, err := s.enrichRuntimeFingerprintScope(ctx, params)
	if err != nil {
		return runtimeFingerprintRecord{}, err
	}
	if !isImmutableGitRef(params.BuildRef) {
		return runtimeFingerprintRecord{}, runtimeReuseInvalidation{reason: runtimeReuseReasonBuildRefNotImmutable}
	}

	repositoryRoot, err := s.resolveRunRepositoryRootForReuse(params)
	if err != nil {
		return runtimeFingerprintRecord{}, err
	}
	headCommit, err := resolveRepositoryHeadCommit(repositoryRoot)
	if err != nil {
		return runtimeFingerprintRecord{}, runtimeReuseInvalidation{
			reason:            runtimeReuseReasonRepoSnapshotMissing,
			effectiveBuildRef: params.BuildRef,
		}
	}
	if !strings.EqualFold(strings.TrimSpace(headCommit), strings.TrimSpace(params.BuildRef)) {
		return runtimeFingerprintRecord{}, runtimeReuseInvalidation{
			reason:            runtimeReuseReasonRepoSnapshotStale,
			effectiveBuildRef: strings.TrimSpace(headCommit),
		}
	}

	prepareParams := PrepareParams{
		RunID:              params.RunID,
		RuntimeMode:        params.RuntimeMode,
		Namespace:          params.Namespace,
		TargetEnv:          params.TargetEnv,
		SlotNo:             params.SlotNo,
		RepositoryFullName: params.RepositoryFullName,
		ServicesYAMLPath:   params.ServicesYAMLPath,
		BuildRef:           params.BuildRef,
		DeployOnly:         params.DeployOnly,
	}
	servicesConfigPath := s.resolveServicesConfigPath(repositoryRoot, params.ServicesYAMLPath)
	if !servicesConfigExists(servicesConfigPath) {
		return runtimeFingerprintRecord{}, runtimeReuseInvalidation{
			reason:            runtimeReuseReasonServicesConfigMissing,
			effectiveBuildRef: params.BuildRef,
		}
	}

	templateVars := s.buildTemplateVars(prepareParams, params.Namespace)
	loaded, err := servicescfg.Load(servicesConfigPath, servicescfg.LoadOptions{
		Env:       params.TargetEnv,
		Namespace: params.Namespace,
		Slot:      params.SlotNo,
		Vars:      templateVars,
	})
	if err != nil {
		return runtimeFingerprintRecord{}, fmt.Errorf("load services config for runtime fingerprint: %w", err)
	}

	targetNamespace := strings.TrimSpace(loaded.Context.Namespace)
	if targetNamespace == "" {
		targetNamespace = params.Namespace
	}
	if targetNamespace == "" || targetNamespace != params.Namespace {
		return runtimeFingerprintRecord{}, runtimeReuseInvalidation{
			reason:            runtimeReuseReasonNamespaceMismatch,
			effectiveBuildRef: params.BuildRef,
		}
	}
	targetEnv := strings.TrimSpace(loaded.Context.Env)
	if targetEnv == "" {
		targetEnv = params.TargetEnv
	}
	prepareParams.TargetEnv = targetEnv
	templateVars = s.buildTemplateVars(prepareParams, targetNamespace)
	applyStackImageVars(templateVars, loaded.Stack)
	templateVars["CODEXK8S_PRODUCTION_NAMESPACE"] = targetNamespace
	templateVars["CODEXK8S_WORKER_K8S_NAMESPACE"] = targetNamespace
	templateVars["CODEXK8S_REPOSITORY_ROOT"] = s.repositoryRootForRuntimeEnv(repositoryRoot)
	if params.RepositoryFullName != "" {
		templateVars["CODEXK8S_GITHUB_REPO"] = params.RepositoryFullName
	}
	applyEnvironmentDomainTemplate(templateVars, loaded.Stack, targetEnv)

	manifestHash, err := s.computeRenderedManifestHash(repositoryRoot, loaded.Stack, targetNamespace, templateVars)
	if err != nil {
		return runtimeFingerprintRecord{}, fmt.Errorf("compute runtime manifest hash: %w", err)
	}

	record := runtimeFingerprintRecord{
		Version:            runtimeFingerprintVersion,
		RuntimeMode:        params.RuntimeMode,
		TargetEnv:          targetEnv,
		Namespace:          targetNamespace,
		ProjectID:          params.ProjectID,
		IssueNumber:        params.IssueNumber,
		AgentKey:           params.AgentKey,
		RepositoryFullName: params.RepositoryFullName,
		ServicesYAMLPath:   params.ServicesYAMLPath,
		BuildRef:           params.BuildRef,
		DeployOnly:         params.DeployOnly,
		ManifestHash:       manifestHash,
	}
	record.Hash = hashRuntimeFingerprint(record)
	return record, nil
}

func (s *Service) enrichRuntimeFingerprintScope(ctx context.Context, params EvaluateReuseParams) (EvaluateReuseParams, error) {
	if params.ProjectID != "" && params.IssueNumber > 0 && params.AgentKey != "" {
		return params, nil
	}
	if s.runs == nil {
		return EvaluateReuseParams{}, fmt.Errorf("runtime fingerprint scope requires run reader")
	}
	run, found, err := s.runs.GetByID(ctx, params.RunID)
	if err != nil {
		return EvaluateReuseParams{}, fmt.Errorf("load run %s for runtime fingerprint scope: %w", params.RunID, err)
	}
	if !found {
		return EvaluateReuseParams{}, fmt.Errorf("run %s for runtime fingerprint scope not found", params.RunID)
	}
	if params.ProjectID == "" {
		params.ProjectID = strings.TrimSpace(run.ProjectID)
	}
	if params.IssueNumber > 0 && params.AgentKey != "" {
		return normalizeEvaluateReuseParams(params), nil
	}
	if len(run.RunPayload) == 0 {
		return EvaluateReuseParams{}, fmt.Errorf("run %s payload is required for runtime fingerprint scope", params.RunID)
	}

	var payload querytypes.RunPayload
	if err := json.Unmarshal(run.RunPayload, &payload); err != nil {
		return EvaluateReuseParams{}, fmt.Errorf("decode run %s payload for runtime fingerprint scope: %w", params.RunID, err)
	}
	if params.ProjectID == "" {
		params.ProjectID = strings.TrimSpace(payload.Project.ID)
	}
	if params.IssueNumber <= 0 && payload.Issue != nil && payload.Issue.Number > 0 {
		params.IssueNumber = payload.Issue.Number
	}
	if params.AgentKey == "" && payload.Agent != nil {
		params.AgentKey = strings.ToLower(strings.TrimSpace(payload.Agent.Key))
	}
	if params.ProjectID == "" || params.IssueNumber <= 0 || params.AgentKey == "" {
		return EvaluateReuseParams{}, fmt.Errorf("runtime fingerprint scope for run %s is incomplete", params.RunID)
	}
	return normalizeEvaluateReuseParams(params), nil
}

func (s *Service) resolveRunRepositoryRootForReuse(params EvaluateReuseParams) (string, error) {
	configuredRoot := strings.TrimSpace(s.cfg.RepositoryRoot)
	if configuredRoot == "" {
		return "", runtimeReuseInvalidation{
			reason:            runtimeReuseReasonRepoSnapshotMissing,
			effectiveBuildRef: params.BuildRef,
		}
	}
	if !filepath.IsAbs(configuredRoot) {
		if looksLikeRepositoryRoot(configuredRoot) {
			return configuredRoot, nil
		}
		return "", runtimeReuseInvalidation{
			reason:            runtimeReuseReasonRepoSnapshotMissing,
			effectiveBuildRef: params.BuildRef,
		}
	}

	repositoryFullName := strings.TrimSpace(params.RepositoryFullName)
	if shouldUseDirectRepositoryRoot(configuredRoot, repositoryFullName) {
		return configuredRoot, nil
	}
	if repositoryFullName == "" {
		return "", runtimeReuseInvalidation{
			reason:            runtimeReuseReasonRepoSnapshotMissing,
			effectiveBuildRef: params.BuildRef,
		}
	}
	owner, name, ok := strings.Cut(repositoryFullName, "/")
	if !ok || strings.TrimSpace(owner) == "" || strings.TrimSpace(name) == "" {
		return "", runtimeReuseInvalidation{
			reason:            runtimeReuseReasonRepoSnapshotMissing,
			effectiveBuildRef: params.BuildRef,
		}
	}
	repositoryRoot := s.repoSnapshotPath(params.TargetEnv, strings.TrimSpace(owner), strings.TrimSpace(name), params.BuildRef)
	if repositoryRoot == "" {
		return "", runtimeReuseInvalidation{
			reason:            runtimeReuseReasonRepoSnapshotMissing,
			effectiveBuildRef: params.BuildRef,
		}
	}
	if _, err := os.Stat(filepath.Join(repositoryRoot, ".git")); err != nil {
		return "", runtimeReuseInvalidation{
			reason:            runtimeReuseReasonRepoSnapshotMissing,
			effectiveBuildRef: params.BuildRef,
		}
	}
	return repositoryRoot, nil
}

func (s *Service) computeRenderedManifestHash(repositoryRoot string, stack *servicescfg.Stack, namespace string, vars map[string]string) (string, error) {
	if stack == nil {
		return "", fmt.Errorf("stack is nil")
	}

	records := make([]renderedManifestRecord, 0, len(stack.Spec.Infrastructure)+len(stack.Spec.Services))
	appendRecord := func(unit string, path string, rendered []byte) {
		sum := sha256.Sum256(rendered)
		records = append(records, renderedManifestRecord{
			Unit: strings.TrimSpace(unit),
			Path: filepath.Clean(strings.TrimSpace(path)),
			Hash: hex.EncodeToString(sum[:]),
		})
	}

	enabledInfra := make(map[string]servicescfg.InfrastructureItem, len(stack.Spec.Infrastructure))
	for _, item := range stack.Spec.Infrastructure {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return "", fmt.Errorf("infrastructure item name is required")
		}
		include, err := evaluateWhen(item.When)
		if err != nil {
			return "", fmt.Errorf("infrastructure %q when expression: %w", name, err)
		}
		if !include {
			continue
		}
		enabledInfra[name] = item
	}
	order, err := topoSortInfrastructure(enabledInfra)
	if err != nil {
		return "", err
	}
	for _, name := range order {
		item := enabledInfra[name]
		if err := s.collectRenderedManifests(repositoryRoot, name, item.Manifests, vars, appendRecord); err != nil {
			return "", err
		}
	}

	enabledServices := make(map[string]servicescfg.Service, len(stack.Spec.Services))
	groupToNames := make(map[string][]string)
	for _, service := range stack.Spec.Services {
		name := strings.TrimSpace(service.Name)
		if name == "" {
			return "", fmt.Errorf("service name is required")
		}
		includeScope, _ := shouldApplyServiceScope(service.Scope, namespace, vars)
		if !includeScope {
			continue
		}
		include, err := evaluateWhen(service.When)
		if err != nil {
			return "", fmt.Errorf("service %q when expression: %w", name, err)
		}
		if !include {
			continue
		}
		enabledServices[name] = service
		group := strings.TrimSpace(service.DeployGroup)
		groupToNames[group] = append(groupToNames[group], name)
	}

	applied := make(map[string]struct{}, len(enabledInfra))
	for name := range enabledInfra {
		applied[name] = struct{}{}
	}
	groupOrder := buildServiceGroupOrder(stack.Spec.Orchestration.DeployOrder, groupToNames)
	for _, group := range groupOrder {
		names := append([]string(nil), groupToNames[group]...)
		sort.Strings(names)
		for len(names) > 0 {
			progress := false
			for idx := 0; idx < len(names); idx++ {
				name := names[idx]
				service := enabledServices[name]
				if !dependenciesSatisfied(service.DependsOn, applied, enabledServices) {
					continue
				}
				if err := s.collectRenderedManifests(repositoryRoot, name, service.Manifests, vars, appendRecord); err != nil {
					return "", err
				}
				applied[name] = struct{}{}
				names = append(names[:idx], names[idx+1:]...)
				progress = true
				break
			}
			if !progress {
				return "", fmt.Errorf("service dependency deadlock in group %q: unresolved %s", group, strings.Join(names, ", "))
			}
		}
	}

	payload := struct {
		ServicesYAML string                   `json:"services_yaml"`
		Manifests    []renderedManifestRecord `json:"manifests"`
	}{
		ServicesYAML: strings.TrimSpace(string(stackMustMarshalYAML(stack))),
		Manifests:    records,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal rendered manifest payload: %w", err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func (s *Service) collectRenderedManifests(repositoryRoot string, unitName string, manifests []servicescfg.ManifestRef, vars map[string]string, appendRecord func(unit string, path string, rendered []byte)) error {
	repoRoot := strings.TrimSpace(repositoryRoot)
	if repoRoot == "" {
		repoRoot = s.cfg.RepositoryRoot
	}
	for _, manifest := range manifests {
		path := strings.TrimSpace(manifest.Path)
		if path == "" {
			continue
		}
		fullPath := path
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(repoRoot, path)
		}
		raw, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read manifest %s for %s: %w", fullPath, unitName, err)
		}
		rendered, err := manifesttpl.Render(fullPath, raw, vars)
		if err != nil {
			return fmt.Errorf("render manifest template %s for %s: %w", fullPath, unitName, err)
		}
		appendRecord(unitName, path, rendered)
	}
	return nil
}

func (s *Service) persistRuntimeFingerprint(ctx context.Context, namespace string, fingerprint runtimeFingerprintRecord) error {
	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" {
		return fmt.Errorf("namespace is required")
	}
	raw, err := json.Marshal(fingerprint)
	if err != nil {
		return fmt.Errorf("marshal runtime fingerprint: %w", err)
	}
	return s.k8s.UpsertNamespaceAnnotations(ctx, targetNamespace, map[string]string{
		runtimeFingerprintHashAnnotationKey:    fingerprint.Hash,
		runtimeFingerprintJSONAnnotationKey:    string(raw),
		runtimeFingerprintUpdatedAnnotationKey: time.Now().UTC().Format(time.RFC3339),
	})
}

func parseStoredRuntimeFingerprint(annotations map[string]string) (runtimeFingerprintRecord, bool) {
	if len(annotations) == 0 {
		return runtimeFingerprintRecord{}, false
	}
	raw := strings.TrimSpace(annotations[runtimeFingerprintJSONAnnotationKey])
	if raw == "" {
		return runtimeFingerprintRecord{}, false
	}
	var record runtimeFingerprintRecord
	if err := json.Unmarshal([]byte(raw), &record); err != nil {
		return runtimeFingerprintRecord{}, false
	}
	hash := strings.TrimSpace(annotations[runtimeFingerprintHashAnnotationKey])
	if hash == "" {
		hash = strings.TrimSpace(record.Hash)
	}
	if hash == "" {
		return runtimeFingerprintRecord{}, false
	}
	record.Hash = hash
	return record, true
}

func hashRuntimeFingerprint(record runtimeFingerprintRecord) string {
	normalized := runtimeFingerprintRecord{
		Version:            strings.TrimSpace(record.Version),
		RuntimeMode:        strings.TrimSpace(record.RuntimeMode),
		TargetEnv:          strings.TrimSpace(record.TargetEnv),
		Namespace:          strings.TrimSpace(record.Namespace),
		ProjectID:          strings.TrimSpace(record.ProjectID),
		IssueNumber:        record.IssueNumber,
		AgentKey:           strings.TrimSpace(record.AgentKey),
		RepositoryFullName: strings.TrimSpace(record.RepositoryFullName),
		ServicesYAMLPath:   strings.TrimSpace(record.ServicesYAMLPath),
		BuildRef:           strings.TrimSpace(record.BuildRef),
		DeployOnly:         record.DeployOnly,
		ManifestHash:       strings.TrimSpace(record.ManifestHash),
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func stackMustMarshalYAML(stack *servicescfg.Stack) []byte {
	if stack == nil {
		return nil
	}
	raw, err := json.Marshal(stack)
	if err != nil {
		return nil
	}
	return raw
}

func asRuntimeReuseInvalidation(err error, target *runtimeReuseInvalidation) bool {
	if err == nil || target == nil {
		return false
	}
	invalid, ok := err.(runtimeReuseInvalidation)
	if !ok {
		return false
	}
	*target = invalid
	return true
}

func resolveRepositoryHeadCommit(repositoryRoot string) (string, error) {
	gitDir, err := resolveGitDir(repositoryRoot)
	if err != nil {
		return "", err
	}
	headRaw, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return "", fmt.Errorf("read git HEAD: %w", err)
	}
	head := strings.TrimSpace(string(headRaw))
	if head == "" {
		return "", fmt.Errorf("git HEAD is empty")
	}
	if isImmutableGitRef(head) {
		return strings.ToLower(head), nil
	}
	if !strings.HasPrefix(head, "ref:") {
		return "", fmt.Errorf("git HEAD is not a ref or commit")
	}
	refName := strings.TrimSpace(strings.TrimPrefix(head, "ref:"))
	if refName == "" {
		return "", fmt.Errorf("git HEAD ref is empty")
	}
	refPath := filepath.Join(gitDir, filepath.FromSlash(refName))
	if refRaw, err := os.ReadFile(refPath); err == nil {
		refValue := strings.TrimSpace(string(refRaw))
		if isImmutableGitRef(refValue) {
			return strings.ToLower(refValue), nil
		}
	}
	return resolvePackedRef(gitDir, refName)
}

func resolveGitDir(repositoryRoot string) (string, error) {
	root := strings.TrimSpace(repositoryRoot)
	if root == "" {
		return "", fmt.Errorf("repository root is required")
	}
	gitPath := filepath.Join(root, ".git")
	info, err := os.Stat(gitPath)
	if err == nil && info.IsDir() {
		return gitPath, nil
	}
	if err == nil {
		raw, readErr := os.ReadFile(gitPath)
		if readErr != nil {
			return "", fmt.Errorf("read gitdir file: %w", readErr)
		}
		content := strings.TrimSpace(string(raw))
		if !strings.HasPrefix(content, "gitdir:") {
			return "", fmt.Errorf("unsupported gitdir file format")
		}
		target := strings.TrimSpace(strings.TrimPrefix(content, "gitdir:"))
		if target == "" {
			return "", fmt.Errorf("gitdir target is empty")
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(root, target)
		}
		return filepath.Clean(target), nil
	}
	if os.IsNotExist(err) {
		return "", fmt.Errorf("git metadata is missing")
	}
	return "", fmt.Errorf("stat git metadata: %w", err)
}

func resolvePackedRef(gitDir string, refName string) (string, error) {
	packedRefs, err := os.ReadFile(filepath.Join(gitDir, "packed-refs"))
	if err != nil {
		return "", fmt.Errorf("read packed-refs: %w", err)
	}
	for _, line := range strings.Split(string(packedRefs), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "^") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) != 2 {
			continue
		}
		if strings.TrimSpace(fields[1]) != strings.TrimSpace(refName) {
			continue
		}
		if isImmutableGitRef(fields[0]) {
			return strings.ToLower(strings.TrimSpace(fields[0])), nil
		}
	}
	return "", fmt.Errorf("packed ref %s not found", refName)
}
