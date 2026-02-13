package postgres

import (
	"fmt"
	"regexp"
	"strings"
)

var databaseNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]{0,62}$`)

// NormalizeDatabaseName validates and normalizes PostgreSQL database name input.
func NormalizeDatabaseName(databaseName string) (string, error) {
	name := strings.TrimSpace(databaseName)
	if name == "" {
		return "", fmt.Errorf("database_name is required")
	}
	if !databaseNamePattern.MatchString(name) {
		return "", fmt.Errorf("database_name %q is invalid", name)
	}
	return name, nil
}
