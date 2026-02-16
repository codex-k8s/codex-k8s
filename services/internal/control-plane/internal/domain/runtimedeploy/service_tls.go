package runtimedeploy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/manifesttpl"
)

const (
	tlsSecretKeyCert = "tls.crt"
	tlsSecretKeyKey  = "tls.key"

	tlsValidityBuffer = 24 * time.Hour
	echoProbeTimeout  = 3 * time.Minute
	echoProbeInterval = 2 * time.Second
)

type tlsValidationResult struct {
	Valid  bool
	Reason string
}

func (s *Service) prepareTLS(ctx context.Context, repositoryRoot string, targetEnv string, namespace string, vars map[string]string, runID string) error {
	host := strings.TrimSpace(valueOr(vars, "CODEXK8S_PUBLIC_DOMAIN", ""))
	if host == "" {
		host = strings.TrimSpace(valueOr(vars, "CODEXK8S_PRODUCTION_DOMAIN", ""))
	}
	if host == "" {
		return nil
	}

	if err := s.k8s.EnsureNamespace(ctx, namespace); err != nil {
		return fmt.Errorf("ensure namespace %s: %w", namespace, err)
	}

	tlsSecretName := strings.TrimSpace(valueOr(vars, "CODEXK8S_TLS_SECRET_NAME", ""))
	if tlsSecretName == "" {
		tlsSecretName = defaultTLSSecretName(targetEnv)
		vars["CODEXK8S_TLS_SECRET_NAME"] = tlsSecretName
	}

	systemNamespace := strings.TrimSpace(valueOr(vars, "CODEXK8S_TLS_SYSTEM_NAMESPACE", ""))
	if systemNamespace == "" {
		return fmt.Errorf("CODEXK8S_TLS_SYSTEM_NAMESPACE is required")
	}
	if err := s.k8s.EnsureNamespace(ctx, systemNamespace); err != nil {
		return fmt.Errorf("ensure tls system namespace %s: %w", systemNamespace, err)
	}

	systemSecretName := tlsSystemSecretNameForHost(host)
	if systemSecretName == "" {
		return fmt.Errorf("resolve tls system secret name for host %q: empty", host)
	}

	now := time.Now().UTC()

	targetSecretData, targetFound, err := s.k8s.GetSecretData(ctx, namespace, tlsSecretName)
	if err != nil {
		return fmt.Errorf("load tls secret %s/%s: %w", namespace, tlsSecretName, err)
	}
	if targetFound {
		validation := validateTLSSecretForHost(targetSecretData, host, now)
		if validation.Valid {
			s.appendTaskLogBestEffort(ctx, runID, "tls", "info", "TLS secret already present in namespace "+namespace+", reusing")
			_ = s.bestEffortUpsertTLSSecretToSystem(ctx, systemNamespace, systemSecretName, host, targetSecretData, runID)
			vars["CODEXK8S_CERT_ISSUER_ENABLED"] = "false"
			vars["CODEXK8S_CERT_MANAGER_ANNOTATE"] = "false"
			return nil
		}
		s.appendTaskLogBestEffort(ctx, runID, "tls", "info", "Existing TLS secret in namespace "+namespace+" is not reusable: "+validation.Reason)
	}

	systemSecretData, systemFound, err := s.k8s.GetSecretData(ctx, systemNamespace, systemSecretName)
	if err != nil {
		return fmt.Errorf("load tls system secret %s/%s: %w", systemNamespace, systemSecretName, err)
	}
	if systemFound {
		validation := validateTLSSecretForHost(systemSecretData, host, now)
		if validation.Valid {
			if err := s.k8s.UpsertTLSSecret(ctx, namespace, tlsSecretName, tlsSecretDataOnly(systemSecretData)); err != nil {
				return fmt.Errorf("upsert reused tls secret into %s/%s: %w", namespace, tlsSecretName, err)
			}
			s.appendTaskLogBestEffort(ctx, runID, "tls", "info", "Reused TLS secret from system namespace "+systemNamespace+" for host "+host)
			vars["CODEXK8S_CERT_ISSUER_ENABLED"] = "false"
			vars["CODEXK8S_CERT_MANAGER_ANNOTATE"] = "false"
			return nil
		}
		s.appendTaskLogBestEffort(ctx, runID, "tls", "info", "TLS system secret is not reusable: "+validation.Reason)
	}

	// No reusable secret found. Only enable cert-manager after confirming that the domain
	// resolves to this cluster via echo-probe (HTTP-01 prerequisite).
	if strings.TrimSpace(valueOr(vars, "CODEXK8S_LETSENCRYPT_EMAIL", "")) == "" || strings.TrimSpace(valueOr(vars, "CODEXK8S_LETSENCRYPT_SERVER", "")) == "" {
		s.appendTaskLogBestEffort(ctx, runID, "tls", "warning", "LetsEncrypt config is incomplete; skipping issuer deployment")
		vars["CODEXK8S_CERT_ISSUER_ENABLED"] = "false"
		vars["CODEXK8S_CERT_MANAGER_ANNOTATE"] = "false"
		return nil
	}

	if err := s.runEchoProbe(ctx, repositoryRoot, namespace, host, vars, runID); err != nil {
		return err
	}

	vars["CODEXK8S_CERT_ISSUER_ENABLED"] = "true"
	vars["CODEXK8S_CERT_MANAGER_ANNOTATE"] = "true"
	s.appendTaskLogBestEffort(ctx, runID, "tls", "info", "Echo probe succeeded, enabling cert-manager issuer for host "+host)
	return nil
}

func (s *Service) finalizeTLS(ctx context.Context, targetEnv string, namespace string, vars map[string]string, runID string) error {
	host := strings.TrimSpace(valueOr(vars, "CODEXK8S_PUBLIC_DOMAIN", ""))
	if host == "" {
		host = strings.TrimSpace(valueOr(vars, "CODEXK8S_PRODUCTION_DOMAIN", ""))
	}
	if host == "" {
		return nil
	}

	systemNamespace := strings.TrimSpace(valueOr(vars, "CODEXK8S_TLS_SYSTEM_NAMESPACE", ""))
	if systemNamespace == "" {
		return nil
	}
	if err := s.k8s.EnsureNamespace(ctx, systemNamespace); err != nil {
		return fmt.Errorf("ensure tls system namespace %s: %w", systemNamespace, err)
	}
	tlsSecretName := strings.TrimSpace(valueOr(vars, "CODEXK8S_TLS_SECRET_NAME", ""))
	if tlsSecretName == "" {
		tlsSecretName = defaultTLSSecretName(targetEnv)
	}
	systemSecretName := tlsSystemSecretNameForHost(host)
	if systemSecretName == "" {
		return nil
	}

	now := time.Now().UTC()
	if existingSystem, found, err := s.k8s.GetSecretData(ctx, systemNamespace, systemSecretName); err == nil && found {
		if validation := validateTLSSecretForHost(existingSystem, host, now); validation.Valid {
			return nil
		}
	}

	shouldWait := strings.EqualFold(strings.TrimSpace(valueOr(vars, "CODEXK8S_CERT_MANAGER_ANNOTATE", "false")), "true")
	timeout := s.cfg.RolloutTimeout
	if timeout <= 0 {
		timeout = 20 * time.Minute
	}
	if !shouldWait {
		timeout = 2 * time.Second
	}

	secretData, found, err := s.waitForValidTLSSecret(ctx, namespace, tlsSecretName, host, timeout)
	if err != nil {
		if shouldWait {
			return err
		}
		s.appendTaskLogBestEffort(ctx, runID, "tls", "warning", "TLS secret reuse check failed: "+err.Error())
		return nil
	}
	if !found {
		if shouldWait {
			return fmt.Errorf("tls secret %s/%s was not created within %s", namespace, tlsSecretName, timeout.String())
		}
		return nil
	}

	if err := s.k8s.UpsertTLSSecret(ctx, systemNamespace, systemSecretName, secretData); err != nil {
		if shouldWait {
			return fmt.Errorf("persist tls secret to system namespace %s/%s: %w", systemNamespace, systemSecretName, err)
		}
		s.appendTaskLogBestEffort(ctx, runID, "tls", "warning", "Persist TLS secret to system namespace failed: "+err.Error())
		return nil
	}
	s.appendTaskLogBestEffort(ctx, runID, "tls", "info", "Persisted TLS secret for host "+host+" into system namespace "+systemNamespace)
	return nil
}

func (s *Service) bestEffortUpsertTLSSecretToSystem(ctx context.Context, systemNamespace string, systemSecretName string, host string, data map[string][]byte, runID string) error {
	validation := validateTLSSecretForHost(data, host, time.Now().UTC())
	if !validation.Valid {
		return nil
	}
	if err := s.k8s.UpsertTLSSecret(ctx, systemNamespace, systemSecretName, tlsSecretDataOnly(data)); err != nil {
		s.appendTaskLogBestEffort(ctx, runID, "tls", "warning", "Persist existing TLS secret into system namespace failed: "+err.Error())
		return err
	}
	return nil
}

func (s *Service) waitForValidTLSSecret(ctx context.Context, namespace string, secretName string, host string, timeout time.Duration) (map[string][]byte, bool, error) {
	targetNamespace := strings.TrimSpace(namespace)
	targetName := strings.TrimSpace(secretName)
	targetHost := strings.TrimSpace(host)
	if targetNamespace == "" || targetName == "" || targetHost == "" {
		return nil, false, fmt.Errorf("namespace, secretName and host are required for tls wait")
	}
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		data, found, err := s.k8s.GetSecretData(waitCtx, targetNamespace, targetName)
		if err != nil {
			return nil, false, err
		}
		if found {
			validation := validateTLSSecretForHost(data, targetHost, time.Now().UTC())
			if validation.Valid {
				return tlsSecretDataOnly(data), true, nil
			}
		}

		select {
		case <-waitCtx.Done():
			return nil, false, nil
		case <-ticker.C:
		}
	}
}

func (s *Service) runEchoProbe(ctx context.Context, repositoryRoot string, namespace string, host string, vars map[string]string, runID string) error {
	token, err := randomHex(12)
	if err != nil {
		return fmt.Errorf("generate echo probe token: %w", err)
	}

	repoRoot := strings.TrimSpace(repositoryRoot)
	if repoRoot == "" {
		repoRoot = s.cfg.RepositoryRoot
	}

	templatePath := filepath.Join(repoRoot, "deploy/base/echo-probe/echo-probe.yaml.tpl")
	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read echo probe manifest template %s: %w", templatePath, err)
	}

	renderVars := cloneStringMap(vars)
	renderVars["CODEXK8S_PRODUCTION_NAMESPACE"] = namespace
	renderVars["CODEXK8S_ECHO_PROBE_HOST"] = host
	renderVars["CODEXK8S_ECHO_PROBE_TOKEN"] = token

	rendered, err := manifesttpl.Render(templatePath, raw, renderVars)
	if err != nil {
		return fmt.Errorf("render echo probe manifest: %w", err)
	}

	if _, err := s.k8s.ApplyManifest(ctx, rendered, namespace, s.cfg.KanikoFieldManager); err != nil {
		return fmt.Errorf("apply echo probe manifest: %w", err)
	}
	if err := s.k8s.WaitForDeploymentReady(ctx, namespace, "codex-k8s-echo-probe", s.cfg.RolloutTimeout); err != nil {
		return fmt.Errorf("wait echo probe deployment: %w", err)
	}

	url := fmt.Sprintf("http://%s/.well-known/codex-k8s-probe/%s", host, token)
	s.appendTaskLogBestEffort(ctx, runID, "tls", "info", "Echo probe request: "+url)

	probeCtx, cancel := context.WithTimeout(ctx, echoProbeTimeout)
	defer cancel()

	ticker := time.NewTicker(echoProbeInterval)
	defer ticker.Stop()

	client := &http.Client{Timeout: 5 * time.Second}
	for {
		req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("build echo probe request: %w", err)
		}
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			body, readErr := io.ReadAll(io.LimitReader(resp.Body, 256))
			_ = resp.Body.Close()
			if readErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 && strings.TrimSpace(string(body)) == token {
				return nil
			}
		}

		select {
		case <-probeCtx.Done():
			return fmt.Errorf("echo probe failed for host %s: %w", host, probeCtx.Err())
		case <-ticker.C:
		}
	}
}

func tlsSystemSecretNameForHost(host string) string {
	normalized := strings.ToLower(strings.TrimSpace(host))
	if normalized == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(normalized))
	return "codex-k8s-tls-" + hex.EncodeToString(sum[:8])
}

func tlsSecretDataOnly(data map[string][]byte) map[string][]byte {
	out := make(map[string][]byte, 2)
	if data == nil {
		return out
	}
	if raw := bytes.TrimSpace(data[tlsSecretKeyCert]); len(raw) > 0 {
		out[tlsSecretKeyCert] = append([]byte(nil), raw...)
	}
	if raw := bytes.TrimSpace(data[tlsSecretKeyKey]); len(raw) > 0 {
		out[tlsSecretKeyKey] = append([]byte(nil), raw...)
	}
	return out
}

func validateTLSSecretForHost(data map[string][]byte, host string, now time.Time) tlsValidationResult {
	targetHost := strings.TrimSpace(host)
	if targetHost == "" {
		return tlsValidationResult{Valid: false, Reason: "host is empty"}
	}
	if data == nil {
		return tlsValidationResult{Valid: false, Reason: "secret data is nil"}
	}
	certPEM := bytes.TrimSpace(data[tlsSecretKeyCert])
	keyPEM := bytes.TrimSpace(data[tlsSecretKeyKey])
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		return tlsValidationResult{Valid: false, Reason: "tls.crt or tls.key is empty"}
	}

	certs, err := parsePEMCertificates(certPEM)
	if err != nil {
		return tlsValidationResult{Valid: false, Reason: "parse certificate: " + err.Error()}
	}

	leaf := certs[0]
	if err := leaf.VerifyHostname(targetHost); err != nil {
		return tlsValidationResult{Valid: false, Reason: "hostname mismatch"}
	}
	if leaf.NotAfter.Before(now.Add(tlsValidityBuffer)) {
		return tlsValidationResult{Valid: false, Reason: "certificate expires soon"}
	}
	return tlsValidationResult{Valid: true}
}

func parsePEMCertificates(data []byte) ([]*x509.Certificate, error) {
	rest := data
	out := make([]*x509.Certificate, 0)
	for len(rest) > 0 {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Bytes) == 0 {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		out = append(out, cert)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}
	return out, nil
}
