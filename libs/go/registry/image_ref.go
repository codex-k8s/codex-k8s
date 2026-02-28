package registry

import "strings"

// ExtractRepositoryPath strips protocol/internal registry host prefix from image repository reference.
func ExtractRepositoryPath(imageRepository string, internalHost string) string {
	repository := strings.TrimSpace(imageRepository)
	host := strings.TrimSpace(internalHost)
	if repository == "" {
		return ""
	}
	repository = strings.TrimPrefix(repository, "http://")
	repository = strings.TrimPrefix(repository, "https://")
	if host == "" {
		return repository
	}
	prefix := host + "/"
	if strings.HasPrefix(repository, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(repository, prefix))
	}
	return ""
}

// SplitImageRef separates image repository and tag.
func SplitImageRef(ref string) (string, string) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return "", ""
	}
	if at := strings.Index(trimmed, "@"); at >= 0 {
		trimmed = trimmed[:at]
	}
	lastSlash := strings.LastIndex(trimmed, "/")
	lastColon := strings.LastIndex(trimmed, ":")
	if lastColon == -1 || lastColon < lastSlash {
		return trimmed, ""
	}
	return trimmed[:lastColon], trimmed[lastColon+1:]
}
