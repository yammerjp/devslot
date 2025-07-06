package git

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// ParseRepoURL parses various repository URL formats and returns the repository name
// Supports:
// - HTTPS URLs: https://github.com/user/repo.git
// - SSH URLs: git@github.com:user/repo.git
// - File URLs: file:///path/to/repo
// - Local paths: /path/to/repo or ./relative/path
func ParseRepoURL(repoURL string) (name string, isLocal bool) {
	// Remove trailing .git if present
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Try to parse as URL
	u, err := url.Parse(repoURL)
	if err == nil && u.Scheme != "" {
		switch u.Scheme {
		case "file":
			// file:///path/to/repo -> repo
			name = filepath.Base(u.Path)
			isLocal = true
		case "http", "https":
			// https://github.com/user/repo -> repo
			name = path.Base(u.Path)
			isLocal = false
		default:
			// Other schemes (git://, ssh://, etc.)
			name = path.Base(u.Path)
			isLocal = false
		}
	} else if strings.Contains(repoURL, ":") && !strings.Contains(repoURL, "://") {
		// SSH format: git@github.com:user/repo
		parts := strings.Split(repoURL, ":")
		if len(parts) == 2 {
			name = path.Base(parts[1])
			isLocal = false
		}
	} else {
		// Local path: /path/to/repo or ./relative/path
		name = filepath.Base(repoURL)
		isLocal = true
	}

	// Ensure name is not empty and add .git suffix for bare repository
	if name != "" && name != "." && name != "/" {
		return name + ".git", isLocal
	}

	return "", isLocal
}
