package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	staffdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

const promptTemplateSeedSyncModeApply = "apply"

var bootstrapPromptTemplateLocales = []string{"ru", "en"}

type promptTemplateSeedSyncer interface {
	SyncPromptTemplateSeeds(ctx context.Context, principal staffdomain.Principal, params querytypes.PromptTemplateSeedSyncParams) (entitytypes.PromptTemplateSeedSyncResult, error)
}

func syncBootstrapPromptTemplateSeeds(ctx context.Context, syncer promptTemplateSeedSyncer, ownerUserID string, logger *slog.Logger) error {
	if syncer == nil {
		return fmt.Errorf("prompt template seed syncer is required")
	}

	ownerUserID = strings.TrimSpace(ownerUserID)
	if ownerUserID == "" {
		return fmt.Errorf("bootstrap owner user id is required")
	}

	result, err := syncer.SyncPromptTemplateSeeds(ctx, staffdomain.Principal{
		UserID:          ownerUserID,
		IsPlatformAdmin: true,
	}, querytypes.PromptTemplateSeedSyncParams{
		Mode:           promptTemplateSeedSyncModeApply,
		Scope:          "global",
		IncludeLocales: bootstrapPromptTemplateLocales,
	})
	if err != nil {
		return fmt.Errorf("sync prompt template seeds: %w", err)
	}

	if logger != nil {
		logger.Info("bootstrap prompt template seeds synced", "created", result.CreatedCount, "updated", result.UpdatedCount, "skipped", result.SkippedCount, "locales", strings.Join(bootstrapPromptTemplateLocales, ","))
	}

	return nil
}
