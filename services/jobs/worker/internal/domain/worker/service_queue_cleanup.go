package worker

import (
	"context"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

func (s *Service) cleanupExpiredNamespaces(ctx context.Context) error {
	cleaned, err := s.launcher.CleanupExpiredNamespaces(ctx, NamespaceCleanupParams{
		Now:   s.now().UTC(),
		Limit: s.cfg.NamespaceLeaseSweepLimit,
	})
	if err != nil {
		return err
	}
	for _, item := range cleaned {
		s.logger.Info(
			"cleaned expired run namespace",
			"namespace", item.Namespace,
			"run_id", item.RunID,
			"expires_at", item.ExpiresAt.Format(time.RFC3339),
		)
		runID := strings.TrimSpace(item.RunID)
		if runID == "" {
			continue
		}
		if _, upsertErr := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
			RunID:       runID,
			Phase:       RunStatusPhaseNamespaceDeleted,
			RuntimeMode: string(agentdomain.RuntimeModeFullEnv),
			Namespace:   item.Namespace,
			Deleted:     true,
		}); upsertErr != nil {
			s.logger.Warn("upsert run status comment (namespace ttl cleanup) failed", "run_id", runID, "namespace", item.Namespace, "err", upsertErr)
		}
	}
	return nil
}
