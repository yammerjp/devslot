package main

import "testing"

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name      string
		repoURL   string
		wantName  string
		wantLocal bool
	}{
		{
			name:      "https URL with .git",
			repoURL:   "https://github.com/yammerjp/devslot.git",
			wantName:  "devslot.git",
			wantLocal: false,
		},
		{
			name:      "https URL without .git",
			repoURL:   "https://github.com/yammerjp/devslot",
			wantName:  "devslot.git",
			wantLocal: false,
		},
		{
			name:      "SSH URL",
			repoURL:   "git@github.com:yammerjp/devslot.git",
			wantName:  "devslot.git",
			wantLocal: false,
		},
		{
			name:      "file URL",
			repoURL:   "file:///home/user/repos/myrepo",
			wantName:  "myrepo.git",
			wantLocal: true,
		},
		{
			name:      "absolute local path",
			repoURL:   "/home/user/repos/myrepo",
			wantName:  "myrepo.git",
			wantLocal: true,
		},
		{
			name:      "relative local path",
			repoURL:   "./repos/myrepo",
			wantName:  "myrepo.git",
			wantLocal: true,
		},
		{
			name:      "git protocol URL",
			repoURL:   "git://github.com/user/repo.git",
			wantName:  "repo.git",
			wantLocal: false,
		},
		{
			name:      "complex path",
			repoURL:   "https://gitlab.com/group/subgroup/project.git",
			wantName:  "project.git",
			wantLocal: false,
		},
		{
			name:      "local path with .git",
			repoURL:   "/repos/myrepo.git",
			wantName:  "myrepo.git",
			wantLocal: true,
		},
		{
			name:      "invalid URL returns empty",
			repoURL:   "",
			wantName:  "",
			wantLocal: true,
		},
		{
			name:      "root path returns empty",
			repoURL:   "/",
			wantName:  "",
			wantLocal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotLocal := parseRepoURL(tt.repoURL)
			if gotName != tt.wantName {
				t.Errorf("parseRepoURL() name = %v, want %v", gotName, tt.wantName)
			}
			if gotLocal != tt.wantLocal {
				t.Errorf("parseRepoURL() isLocal = %v, want %v", gotLocal, tt.wantLocal)
			}
		})
	}
}