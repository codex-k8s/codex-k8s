package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
)

func ensureBootstrapAllowedUsers(ctx context.Context, users userrepo.Repository, ownerEmail string, allowedEmailsCSV string, logger *slog.Logger) error {
	return ensureBootstrapUsers(ctx, users, ownerEmail, allowedEmailsCSV, false, "allowed user", logger)
}

func ensureBootstrapPlatformAdmins(ctx context.Context, users userrepo.Repository, ownerEmail string, adminEmailsCSV string, logger *slog.Logger) error {
	return ensureBootstrapUsers(ctx, users, ownerEmail, adminEmailsCSV, true, "platform admin", logger)
}

func ensureBootstrapUsers(ctx context.Context, users userrepo.Repository, ownerEmail string, emailsCSV string, isPlatformAdmin bool, label string, logger *slog.Logger) error {
	emails, err := parseAllowedEmailsCSV(emailsCSV)
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

func parseAllowedEmailsCSV(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	seen := make(map[string]struct{})
	var out []string
	for _, part := range strings.Split(s, ",") {
		email := strings.ToLower(strings.TrimSpace(part))
		if email == "" {
			continue
		}
		if !strings.Contains(email, "@") {
			return nil, fmt.Errorf("invalid email in CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS: %q", email)
		}
		if _, ok := seen[email]; ok {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, email)
	}
	return out, nil
}
