package webhook

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	"gopkg.in/yaml.v3"
)

const autoBumpCommitMessagePrefix = "chore(services): auto-bump image versions"

var lastNumericTokenPattern = regexp.MustCompile(`\d+`)

type versionBumpChange struct {
	Name string
	From string
	To   string
}

func (s *Service) maybeAutoBumpMainVersions(ctx context.Context, envelope githubWebhookEnvelope, servicesYAMLPath string, buildRef string) (bool, error) {
	if !s.autoVersionBump {
		return false, nil
	}
	if s.githubMgmt == nil {
		return false, fmt.Errorf("push-main auto bump is enabled, but github management client is not configured")
	}
	if strings.TrimSpace(s.githubToken) == "" {
		return false, fmt.Errorf("push-main auto bump is enabled, but github token is empty")
	}

	owner, repo, err := splitRepositoryFullName(envelope.Repository.FullName)
	if err != nil {
		return false, err
	}

	changedPaths, err := s.collectPushMainChangedPaths(ctx, owner, repo, envelope, buildRef)
	if err != nil {
		return false, err
	}
	if len(changedPaths) == 0 {
		return false, nil
	}

	servicesYAMLPath = strings.TrimSpace(servicesYAMLPath)
	if servicesYAMLPath == "" {
		servicesYAMLPath = "services.yaml"
	}

	rawServicesYAML, found, err := s.githubMgmt.GetFile(ctx, s.githubToken, owner, repo, servicesYAMLPath, buildRef)
	if err != nil {
		return false, fmt.Errorf("load %s from %s/%s@%s: %w", servicesYAMLPath, owner, repo, buildRef, err)
	}
	if !found {
		return false, fmt.Errorf("%s not found in %s/%s@%s", servicesYAMLPath, owner, repo, buildRef)
	}

	updatedServicesYAML, changes, err := bumpServicesYAMLVersions(rawServicesYAML, changedPaths)
	if err != nil {
		return false, fmt.Errorf("bump versions in %s: %w", servicesYAMLPath, err)
	}
	if len(changes) == 0 {
		return false, nil
	}

	branch := resolvePushBranch(envelope.Ref)
	if branch == "" {
		branch = "main"
	}
	commitMessage := buildAutoBumpCommitMessage(changes)
	_, err = s.githubMgmt.CommitFilesOnBranch(
		ctx,
		s.githubToken,
		owner,
		repo,
		branch,
		strings.TrimSpace(buildRef),
		commitMessage,
		map[string][]byte{
			servicesYAMLPath: updatedServicesYAML,
		},
	)
	if err != nil {
		return false, fmt.Errorf("commit auto-bumped services versions to %s/%s@%s: %w", owner, repo, branch, err)
	}

	return true, nil
}

func (s *Service) collectPushMainChangedPaths(ctx context.Context, owner string, repo string, envelope githubWebhookEnvelope, buildRef string) ([]string, error) {
	afterSHA := strings.TrimSpace(buildRef)
	beforeSHA := strings.TrimSpace(envelope.Before)
	fallback := collectPathsFromPushCommits(envelope.Commits)

	if beforeSHA == "" || afterSHA == "" || isDeletedGitCommitSHA(beforeSHA) {
		return fallback, nil
	}

	paths, err := s.githubMgmt.ListChangedFilesBetweenCommits(ctx, s.githubToken, owner, repo, beforeSHA, afterSHA)
	if err != nil {
		if len(fallback) > 0 {
			return fallback, nil
		}
		return nil, fmt.Errorf("list changed files between commits %s...%s: %w", beforeSHA, afterSHA, err)
	}
	if len(paths) > 0 {
		return normalizeChangedPaths(paths), nil
	}
	return fallback, nil
}

func collectPathsFromPushCommits(commits []githubPushCommitRecord) []string {
	paths := make([]string, 0, len(commits)*3)
	for _, commit := range commits {
		paths = append(paths, commit.Added...)
		paths = append(paths, commit.Modified...)
		paths = append(paths, commit.Removed...)
	}
	return normalizeChangedPaths(paths)
}

func normalizeChangedPaths(paths []string) []string {
	return servicescfg.NormalizeRepositoryRelativePaths(paths)
}

func splitRepositoryFullName(fullName string) (string, string, error) {
	fullName = strings.TrimSpace(fullName)
	owner, repo, ok := strings.Cut(fullName, "/")
	if !ok {
		return "", "", fmt.Errorf("repository full_name must be in owner/name form, got %q", fullName)
	}
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("repository full_name must be in owner/name form, got %q", fullName)
	}
	return owner, repo, nil
}

func resolvePushBranch(ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	const branchPrefix = "refs/heads/"
	if strings.HasPrefix(ref, branchPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(ref, branchPrefix))
	}
	return ""
}

func bumpServicesYAMLVersions(raw []byte, changedPaths []string) ([]byte, []versionBumpChange, error) {
	if len(raw) == 0 {
		return nil, nil, fmt.Errorf("services yaml is empty")
	}

	var document yaml.Node
	if err := yaml.Unmarshal(raw, &document); err != nil {
		return nil, nil, fmt.Errorf("parse services yaml: %w", err)
	}
	if document.Kind != yaml.DocumentNode || len(document.Content) == 0 {
		return nil, nil, fmt.Errorf("invalid services yaml: expected document node")
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("invalid services yaml: expected mapping root")
	}

	specNode, ok := yamlMapLookup(root, "spec")
	if !ok || specNode.Kind != yaml.MappingNode {
		return raw, nil, nil
	}
	versionsNode, ok := yamlMapLookup(specNode, "versions")
	if !ok || versionsNode.Kind != yaml.MappingNode {
		return raw, nil, nil
	}

	normalizedChanged := normalizeChangedPaths(changedPaths)
	if len(normalizedChanged) == 0 {
		return raw, nil, nil
	}

	changes := make([]versionBumpChange, 0)
	for i := 0; i+1 < len(versionsNode.Content); i += 2 {
		keyNode := versionsNode.Content[i]
		valueNode := versionsNode.Content[i+1]
		versionName := strings.TrimSpace(keyNode.Value)
		if versionName == "" {
			continue
		}

		var versionSpec servicescfg.VersionSpec
		if err := valueNode.Decode(&versionSpec); err != nil {
			return nil, nil, fmt.Errorf("decode spec.versions[%s]: %w", versionName, err)
		}
		if len(versionSpec.BumpOn) == 0 {
			continue
		}
		if !shouldBumpVersion(versionSpec.BumpOn, normalizedChanged) {
			continue
		}

		nextVersion, err := bumpLastNumericToken(versionSpec.Value)
		if err != nil {
			return nil, nil, fmt.Errorf("bump spec.versions[%s]=%q: %w", versionName, versionSpec.Value, err)
		}
		if nextVersion == versionSpec.Value {
			continue
		}
		if err := setVersionValue(valueNode, nextVersion); err != nil {
			return nil, nil, fmt.Errorf("set spec.versions[%s] value: %w", versionName, err)
		}
		changes = append(changes, versionBumpChange{
			Name: versionName,
			From: versionSpec.Value,
			To:   nextVersion,
		})
	}
	if len(changes) == 0 {
		return raw, nil, nil
	}

	sort.Slice(changes, func(i, j int) bool { return changes[i].Name < changes[j].Name })
	rendered, err := encodeYAMLDocument(&document)
	if err != nil {
		return nil, nil, err
	}
	return rendered, changes, nil
}

func shouldBumpVersion(bumpOn []string, changedPaths []string) bool {
	if len(bumpOn) == 0 || len(changedPaths) == 0 {
		return false
	}
	for _, ruleRaw := range bumpOn {
		rule := servicescfg.NormalizeRepositoryRelativePath(ruleRaw)
		if rule == "" {
			continue
		}
		for _, changed := range changedPaths {
			if changed == rule || strings.HasPrefix(changed, rule+"/") {
				return true
			}
		}
	}
	return false
}

func bumpLastNumericToken(version string) (string, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return "", fmt.Errorf("version is empty")
	}

	matches := lastNumericTokenPattern.FindAllStringIndex(version, -1)
	if len(matches) == 0 {
		return "", fmt.Errorf("version has no numeric token")
	}
	last := matches[len(matches)-1]
	from := version[last[0]:last[1]]
	current, err := strconv.ParseUint(from, 10, 64)
	if err != nil {
		return "", fmt.Errorf("parse numeric token %q: %w", from, err)
	}
	next := current + 1
	nextToken := strconv.FormatUint(next, 10)
	if width := len(from); len(nextToken) < width {
		nextToken = strings.Repeat("0", width-len(nextToken)) + nextToken
	}
	return version[:last[0]] + nextToken + version[last[1]:], nil
}

func setVersionValue(node *yaml.Node, value string) error {
	if node == nil {
		return fmt.Errorf("version node is nil")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("version value is empty")
	}

	switch node.Kind {
	case yaml.ScalarNode:
		node.Tag = "!!str"
		node.Value = value
		return nil
	case yaml.MappingNode:
		if valueNode, ok := yamlMapLookup(node, "value"); ok {
			valueNode.Kind = yaml.ScalarNode
			valueNode.Tag = "!!str"
			valueNode.Value = value
			return nil
		}
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "value"},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value},
		)
		return nil
	default:
		return fmt.Errorf("unsupported version node kind %d", node.Kind)
	}
}

func yamlMapLookup(mapping *yaml.Node, key string) (*yaml.Node, bool) {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil, false
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		keyNode := mapping.Content[i]
		if keyNode == nil || keyNode.Kind != yaml.ScalarNode {
			continue
		}
		if strings.TrimSpace(keyNode.Value) == key {
			return mapping.Content[i+1], true
		}
	}
	return nil, false
}

func encodeYAMLDocument(document *yaml.Node) ([]byte, error) {
	if document == nil {
		return nil, fmt.Errorf("yaml document is nil")
	}
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)
	if err := encoder.Encode(document); err != nil {
		_ = encoder.Close()
		return nil, fmt.Errorf("encode services yaml: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("close yaml encoder: %w", err)
	}
	return out.Bytes(), nil
}

func buildAutoBumpCommitMessage(changes []versionBumpChange) string {
	if len(changes) == 0 {
		return autoBumpCommitMessagePrefix
	}

	parts := make([]string, 0, len(changes))
	for _, item := range changes {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		parts = append(parts, name)
	}
	sort.Strings(parts)
	if len(parts) == 0 {
		return autoBumpCommitMessagePrefix
	}
	if len(parts) > 3 {
		return autoBumpCommitMessagePrefix + ": " + strings.Join(parts[:3], ", ") + ", +more"
	}
	return autoBumpCommitMessagePrefix + ": " + strings.Join(parts, ", ")
}
