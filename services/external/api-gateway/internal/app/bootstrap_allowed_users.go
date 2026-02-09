package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	userrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/user"
)

func ensureBootstrapAllowedUsers(ctx context.Context, users userrepo.Repository, ownerEmail string, allowedEmailsCSV string, logger *slog.Logger) error {
	emails, err := parseAllowedEmailsCSV(allowedEmailsCSV)
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
		if _, err := users.CreateAllowedUser(ctx, email, false); err != nil {
			return fmt.Errorf("create allowed user %q: %w", email, err)
		}
	}

	if logger != nil {
		logger.Info("bootstrap allowed users ensured", "count", len(emails))
	}
	return nil
}

func ensureBootstrapPlatformAdmins(ctx context.Context, users userrepo.Repository, ownerEmail string, adminEmailsCSV string, logger *slog.Logger) error {
	emails, err := parseAllowedEmailsCSV(adminEmailsCSV)
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
		if _, err := users.CreateAllowedUser(ctx, email, true); err != nil {
			return fmt.Errorf("create platform admin user %q: %w", email, err)
		}
	}

	if logger != nil {
		logger.Info("bootstrap platform admins ensured", "count", len(emails))
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
