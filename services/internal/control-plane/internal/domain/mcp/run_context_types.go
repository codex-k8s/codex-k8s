package mcp

import (
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type resolvedRunContext struct {
	Session    SessionContext
	Run        agentrunrepo.Run
	Repository repocfgrepo.RepositoryBinding
	Token      string
	Payload    querytypes.RunPayload
}
