package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
)

func ensureBootstrapAllowedUsers(ctx context.Context, users userrepo.Repository, ownerEmail string, allowedEmails []string, logger *slog.Logger) error {
	return ensureBootstrapUsers(ctx, users, ownerEmail, allowedEmails, false, "allowed user", "CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS", logger)
}

func ensureBootstrapPlatformAdmins(ctx context.Context, users userrepo.Repository, ownerEmail string, adminEmails []string, logger *slog.Logger) error {
	return ensureBootstrapUsers(ctx, users, ownerEmail, adminEmails, true, "platform admin", "CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS", logger)
}

func ensureBootstrapUsers(ctx context.Context, users userrepo.Repository, ownerEmail string, emails []string, isPlatformAdmin bool, label string, envVarName string, logger *slog.Logger) error {
	emails, err := normalizeBootstrapEmails(emails, envVarName)
	if err != nil {
		return err
	}
	if len(emails) == 0 {
		return nil
	}

	owner := strings.ToLower(strings.TrimSpace(ownerEmail))
	for _, email := range emails {
		if email == owner {
			continue
		}
		if _, err := users.CreateAllowedUser(ctx, email, isPlatformAdmin); err != nil {
			return fmt.Errorf("create %s user %q: %w", label, email, err)
		}
	}

	if logger != nil {
		logger.Info("bootstrap users ensured", "label", label, "count", len(emails))
	}
	return nil
}

func normalizeBootstrapEmails(values []string, envVarName string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{})
	out := make([]string, 0, len(values))
	for _, value := range values {
		email := strings.ToLower(strings.TrimSpace(value))
		if email == "" {
			continue
		}
		if !strings.Contains(email, "@") {
			return nil, fmt.Errorf("invalid email in %s: %q", envVarName, email)
		}
		if _, ok := seen[email]; ok {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, email)
	}
	return out, nil
}
