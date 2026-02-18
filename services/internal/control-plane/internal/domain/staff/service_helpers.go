package staff

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
	"gopkg.in/yaml.v3"
)

type dnsCandidate struct {
	CheckName string
	Domain    string
}

func parseNamespaceNameSpec(spec string) (namespace string, name string, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", "", fmt.Errorf("k8s target is empty")
	}

	// Accept both "<namespace>/<name>" and "<namespace>:<name>" forms.
	if strings.Contains(spec, "/") {
		parts := strings.SplitN(spec, "/", 2)
		namespace = strings.TrimSpace(parts[0])
		name = strings.TrimSpace(parts[1])
	} else if strings.Contains(spec, ":") {
		parts := strings.SplitN(spec, ":", 2)
		namespace = strings.TrimSpace(parts[0])
		name = strings.TrimSpace(parts[1])
	} else {
		return "", "", fmt.Errorf("k8s target must be <namespace>/<name>")
	}
	if namespace == "" || name == "" {
		return "", "", fmt.Errorf("k8s target must be <namespace>/<name>")
	}
	return namespace, name, nil
}

func parseGitHubFullName(fullName string) (owner string, repo string, err error) {
	fullName = strings.TrimSpace(fullName)
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub repo %q (expected owner/name)", fullName)
	}
	owner = strings.TrimSpace(parts[0])
	repo = strings.TrimSpace(parts[1])
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid GitHub repo %q (expected owner/name)", fullName)
	}
	return owner, repo, nil
}

func resolveExpectedIngressIPs(webhookURL string) (host string, ips []net.IP) {
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL == "" {
		return "", nil
	}
	// Best-effort: derive expected ingress IPs from the platform public host (webhook url host).
	parsed, err := urlParse(webhookURL)
	if err != nil || parsed == "" {
		return "", nil
	}
	host = parsed
	items, err := net.LookupIP(host)
	if err != nil {
		return host, nil
	}
	return host, items
}

func urlParse(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	h := strings.TrimSpace(u.Hostname())
	if h == "" {
		return "", fmt.Errorf("empty hostname")
	}
	return h, nil
}

func getOptionalEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func ipIntersects(a []net.IP, b []net.IP) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	lookup := make(map[string]struct{}, len(b))
	for _, ip := range b {
		if ip == nil {
			continue
		}
		lookup[ip.String()] = struct{}{}
	}
	for _, ip := range a {
		if ip == nil {
			continue
		}
		if _, ok := lookup[ip.String()]; ok {
			return true
		}
	}
	return false
}

func envVarsMap() map[string]string {
	out := make(map[string]string, 64)
	for _, item := range os.Environ() {
		key, value, ok := strings.Cut(item, "=")
		if !ok || strings.TrimSpace(key) == "" {
			continue
		}
		out[key] = value
	}
	return out
}

func listServicesYAMLEnvironments(raw []byte) (map[string]struct{}, error) {
	var stack servicescfg.Stack
	if err := yaml.Unmarshal(raw, &stack); err != nil {
		return nil, fmt.Errorf("parse services.yaml: %w", err)
	}
	out := make(map[string]struct{}, len(stack.Spec.Environments))
	for k := range stack.Spec.Environments {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		out[key] = struct{}{}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("spec.environments is empty")
	}
	return out, nil
}

func resolveServicesYAMLDomain(raw []byte, envName string, slot int, vars map[string]string) (domain string, source string, namespace string, err error) {
	envName = strings.TrimSpace(envName)
	if envName == "" {
		return "", "", "", fmt.Errorf("env is required")
	}
	if vars == nil {
		vars = envVarsMap()
	}

	result, err := servicescfg.LoadFromYAML(raw, servicescfg.LoadOptions{
		Env:  envName,
		Slot: slot,
		Vars: vars,
	})
	if err != nil {
		return "", "", "", err
	}
	namespace = strings.TrimSpace(result.Context.Namespace)

	envCfg, err := servicescfg.ResolveEnvironment(result.Stack, envName)
	if err != nil {
		return "", "", namespace, err
	}
	host := strings.TrimSpace(envCfg.DomainTemplate)
	if host != "" {
		source = "domainTemplate"
	} else if strings.EqualFold(envName, "ai") {
		base := strings.TrimSpace(vars["CODEXK8S_AI_DOMAIN"])
		if base == "" {
			base = getOptionalEnv("CODEXK8S_AI_DOMAIN")
		}
		if base != "" && namespace != "" {
			host = namespace + "." + base
			source = "default:namespace.CODEXK8S_AI_DOMAIN"
		}
	} else {
		base := strings.TrimSpace(vars["CODEXK8S_PRODUCTION_DOMAIN"])
		if base == "" {
			base = getOptionalEnv("CODEXK8S_PRODUCTION_DOMAIN")
		}
		host = base
		source = "default:CODEXK8S_PRODUCTION_DOMAIN"
	}

	host = strings.TrimSpace(host)
	if host == "" {
		return "", source, namespace, nil
	}
	// Domain template must yield a hostname (no scheme/path/port).
	switch {
	case strings.Contains(host, "://"):
		return "", source, namespace, fmt.Errorf("domain must be a hostname, got url %q", host)
	case strings.Contains(host, "/"):
		return "", source, namespace, fmt.Errorf("domain must be a hostname, got path %q", host)
	case strings.Contains(host, ":"):
		return "", source, namespace, fmt.Errorf("domain must be a hostname without port, got %q", host)
	}
	return host, source, namespace, nil
}

func runDNSCheck(name string, domain string, expectedIPs []net.IP) valuetypes.GitHubPreflightCheck {
	domain = strings.TrimSpace(domain)
	check := valuetypes.GitHubPreflightCheck{Name: strings.TrimSpace(name), Status: "ok"}
	if domain == "" {
		check.Status = "failed"
		check.Details = "domain is empty"
		return check
	}

	ips, lookupErr := net.LookupIP(domain)
	if lookupErr != nil || len(ips) == 0 {
		check.Status = "failed"
		if lookupErr != nil {
			check.Details = "dns lookup failed: " + lookupErr.Error()
		} else {
			check.Details = "dns lookup returned empty result"
		}
		return check
	}

	resolved := formatIPs(ips)
	if len(expectedIPs) > 0 && !ipIntersects(ips, expectedIPs) {
		check.Status = "failed"
		check.Details = fmt.Sprintf("domain does not resolve to ingress IPs (resolved_ips=%s expected_ingress_ips=%s)", resolved, formatIPs(expectedIPs))
		return check
	}

	check.Details = fmt.Sprintf("resolved_ips=%s", resolved)
	return check
}

func formatIPs(ips []net.IP) string {
	if len(ips) == 0 {
		return ""
	}
	seen := make(map[string]struct{}, len(ips))
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		if ip == nil {
			continue
		}
		s := strings.TrimSpace(ip.String())
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return strings.Join(out, ",")
}
