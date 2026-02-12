package mcp

import (
	"context"
	"fmt"
	"strings"
)

func (s *Service) resolveRunContext(ctx context.Context, session SessionContext, requireGitHubToken bool) (resolvedRunContext, error) {
	runID := strings.TrimSpace(session.RunID)
	if runID == "" {
		return resolvedRunContext{}, fmt.Errorf("run_id is required")
	}

	run, ok, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		return resolvedRunContext{}, fmt.Errorf("get run: %w", err)
	}
	if !ok {
		return resolvedRunContext{}, fmt.Errorf("run not found")
	}
	if !isRunActive(run.Status) {
		return resolvedRunContext{}, fmt.Errorf("run status %q is not active", run.Status)
	}

	payload, err := parseRunPayload(run.RunPayload)
	if err != nil {
		return resolvedRunContext{}, err
	}

	repositoryID := strings.TrimSpace(payload.Project.RepositoryID)
	if repositoryID == "" {
		return resolvedRunContext{}, fmt.Errorf("run payload missing repository_id")
	}

	repository, ok, err := s.repos.GetByID(ctx, repositoryID)
	if err != nil {
		return resolvedRunContext{}, fmt.Errorf("get repository binding: %w", err)
	}
	if !ok {
		return resolvedRunContext{}, fmt.Errorf("repository binding not found")
	}

	owner := strings.TrimSpace(repository.Owner)
	repoName := strings.TrimSpace(repository.Name)
	if owner == "" || repoName == "" {
		fallbackOwner, fallbackName := splitRepoFullName(payload.Repository.FullName)
		if owner == "" {
			owner = fallbackOwner
		}
		if repoName == "" {
			repoName = fallbackName
		}
	}
	if owner == "" || repoName == "" {
		return resolvedRunContext{}, fmt.Errorf("repository owner/name are required")
	}
	repository.Owner = owner
	repository.Name = repoName
	if strings.TrimSpace(repository.ServicesYAMLPath) == "" {
		repository.ServicesYAMLPath = "services.yaml"
	}

	token := ""
	if requireGitHubToken {
		token, err = s.loadBotToken(ctx)
		if err != nil {
			return resolvedRunContext{}, err
		}
	}

	sessionContext := session
	if sessionContext.CorrelationID == "" {
		sessionContext.CorrelationID = run.CorrelationID
	}
	if sessionContext.ProjectID == "" {
		switch {
		case run.ProjectID != "":
			sessionContext.ProjectID = run.ProjectID
		case repository.ProjectID != "":
			sessionContext.ProjectID = repository.ProjectID
		case payload.Project.ID != "":
			sessionContext.ProjectID = payload.Project.ID
		}
	}
	if sessionContext.RuntimeMode == "" {
		triggerKind := ""
		if payload.Trigger != nil {
			triggerKind = payload.Trigger.Kind
		}
		sessionContext.RuntimeMode = parseRuntimeMode(triggerKind)
	}
	sessionContext.RuntimeMode = normalizeRuntimeMode(sessionContext.RuntimeMode)

	return resolvedRunContext{
		Session:    sessionContext,
		Run:        run,
		Repository: repository,
		Token:      token,
		Payload:    payload,
	}, nil
}

func (s *Service) toolCapability(name ToolName) (ToolCapability, error) {
	tool, ok := toolCapabilityByName(s.toolCatalog, name)
	if !ok {
		return ToolCapability{}, fmt.Errorf("tool %q is not registered", name)
	}
	return tool, nil
}
