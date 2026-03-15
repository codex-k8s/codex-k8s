package worker

import (
	"fmt"
	"net"
	"strings"
)

const (
	controlPlaneMCPHTTPPort = "8081"
	controlPlaneMCPPath     = "/mcp"
)

func resolveControlPlaneMCPBaseURL(explicitURL string, grpcTarget string) string {
	if resolved := strings.TrimSpace(explicitURL); resolved != "" {
		return resolved
	}

	host := strings.TrimSpace(grpcTarget)
	if host == "" {
		return ""
	}

	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	host = strings.Trim(strings.TrimSpace(host), "[]")
	if host == "" {
		return ""
	}

	return fmt.Sprintf("http://%s:%s%s", host, controlPlaneMCPHTTPPort, controlPlaneMCPPath)
}
