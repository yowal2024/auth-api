package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveConfigPathPrefersExecutableDirectory(t *testing.T) {
	root := t.TempDir()
	executableDir := filepath.Join(root, "opt", "auth-api")
	workingDir := filepath.Join(root, "root")
	if err := os.MkdirAll(filepath.Join(executableDir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workingDir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(executableDir, "configs", "config.yaml"),
		filepath.Join(workingDir, "config.yaml"),
	} {
		if err := os.WriteFile(path, []byte("installed: true\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	got := resolveConfigPathFrom("", filepath.Join(executableDir, "auth-api"), workingDir)
	want := filepath.Join(executableDir, "configs", "config.yaml")
	if got != want {
		t.Fatalf("default config path = %q, want executable config %q", got, want)
	}

	custom := resolveConfigPathFrom("custom/settings.yaml", filepath.Join(executableDir, "auth-api"), workingDir)
	if want := filepath.Join(workingDir, "custom", "settings.yaml"); custom != want {
		t.Fatalf("explicit relative config path = %q, want %q", custom, want)
	}
}

func TestResolveConfigPathUsesExecutableDefaultAndWorkingDirectoryFallback(t *testing.T) {
	root := t.TempDir()
	executableDir := filepath.Join(root, "opt", "auth-api")
	workingDir := filepath.Join(root, "root")
	if err := os.MkdirAll(executableDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workingDir, "configs"), 0755); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(executableDir, "auth-api")
	if got, want := resolveConfigPathFrom("", executable, workingDir), filepath.Join(executableDir, "configs", "config.yaml"); got != want {
		t.Fatalf("missing-config default path = %q, want %q", got, want)
	}
	workingConfig := filepath.Join(workingDir, "configs", "config.yaml")
	if err := os.WriteFile(workingConfig, []byte("installed: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := resolveConfigPathFrom("", executable, workingDir); got != workingConfig {
		t.Fatalf("working-directory fallback = %q, want %q", got, workingConfig)
	}
}

func TestRuntimeDirForConfigUsesParentOfConfigs(t *testing.T) {
	if got := runtimeDirForConfig(filepath.Join("configs", "config.yaml")); got != "." {
		t.Fatalf("runtime dir = %q", got)
	}
	if got := runtimeDirForConfig(filepath.Join("settings", "config.yaml")); got != "settings" {
		t.Fatalf("custom runtime dir = %q", got)
	}
}

func TestListenAddressPriority(t *testing.T) {
	tests := []struct {
		name       string
		configured string
		override   string
		port       int
		want       string
	}{
		{name: "default", want: ":8080"},
		{name: "yaml", configured: ":8081", want: ":8081"},
		{name: "explicit host", configured: "127.0.0.1:8080", want: "127.0.0.1:8080"},
		{name: "addr overrides yaml", configured: ":8080", override: "[::1]:9000", want: "[::1]:9000"},
		{name: "port overrides yaml", configured: ":8080", port: 8443, want: ":8443"},
		{name: "port binds all interfaces like original", configured: "127.0.0.1:8080", port: 8443, want: ":8443"},
		{name: "port has highest priority", configured: ":8080", override: "127.0.0.1:9000", port: 8443, want: ":8443"},
		{name: "maximum port", port: 65535, want: ":65535"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := resolveListenAddress(tt.configured, tt.override, tt.port); err != nil || got != tt.want {
				t.Fatalf("listen address = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
	for _, port := range []int{-1, 65536} {
		if _, err := resolveListenAddress(":8080", "", port); err == nil {
			t.Errorf("accepted invalid port %d", port)
		}
	}
}
