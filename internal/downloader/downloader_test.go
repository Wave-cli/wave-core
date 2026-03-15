package downloader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// --- Asset name selection ---

func TestSelectAsset(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "flow-linux-amd64.tar.gz", URL: "https://example.com/flow-linux-amd64.tar.gz"},
		{Name: "flow-darwin-arm64.tar.gz", URL: "https://example.com/flow-darwin-arm64.tar.gz"},
		{Name: "flow-windows-amd64.zip", URL: "https://example.com/flow-windows-amd64.zip"},
		{Name: "Waveplugin", URL: "https://example.com/Waveplugin"},
	}

	// Should find the binary for current OS/arch
	selected, err := SelectAsset(assets, "flow", runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Fatalf("SelectAsset failed: %v", err)
	}
	if selected.Name == "" {
		t.Error("Selected asset should have a name")
	}
}

func TestSelectAssetNotFound(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "flow-linux-amd64.tar.gz", URL: "https://example.com/flow-linux-amd64.tar.gz"},
	}
	_, err := SelectAsset(assets, "flow", "plan9", "mips")
	if err == nil {
		t.Error("Should fail when no matching asset")
	}
}

func TestSelectAssetEmptyList(t *testing.T) {
	_, err := SelectAsset([]ReleaseAsset{}, "flow", "linux", "amd64")
	if err == nil {
		t.Error("Should fail for empty asset list")
	}
}

func TestSelectWavepluginAsset(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "flow-linux-amd64.tar.gz", URL: "https://example.com/a"},
		{Name: "Waveplugin", URL: "https://example.com/wp"},
	}

	wp := FindWavepluginAsset(assets)
	if wp == nil {
		t.Fatal("Should find Waveplugin asset")
	}
	if wp.Name != "Waveplugin" {
		t.Errorf("Name = %q", wp.Name)
	}
}

func TestSelectWavepluginAssetMissing(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "flow-linux-amd64.tar.gz", URL: "https://example.com/a"},
	}
	wp := FindWavepluginAsset(assets)
	if wp != nil {
		t.Error("Should return nil when no Waveplugin asset")
	}
}

// --- Release fetching (mocked HTTP) ---

func TestFetchRelease(t *testing.T) {
	release := Release{
		TagName: "v1.2.0",
		Assets: []ReleaseAsset{
			{Name: "flow-linux-amd64.tar.gz", URL: "https://example.com/a"},
			{Name: "Waveplugin", URL: "https://example.com/wp"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	rel, err := client.FetchRelease("wave-cli", "flow", "")
	if err != nil {
		t.Fatalf("FetchRelease failed: %v", err)
	}
	if rel.TagName != "v1.2.0" {
		t.Errorf("TagName = %q", rel.TagName)
	}
	if len(rel.Assets) != 2 {
		t.Errorf("Assets len = %d, want 2", len(rel.Assets))
	}
}

func TestFetchReleaseSpecificVersion(t *testing.T) {
	release := Release{
		TagName: "v1.0.0",
		Assets:  []ReleaseAsset{{Name: "flow-linux-amd64.tar.gz", URL: "https://example.com/a"}},
	}

	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	_, err := client.FetchRelease("wave-cli", "flow", "1.0.0")
	if err != nil {
		t.Fatalf("FetchRelease failed: %v", err)
	}
	if requestedPath != "/repos/wave-cli/flow/releases/tags/v1.0.0" {
		t.Errorf("Requested path = %q", requestedPath)
	}
}

func TestFetchReleaseNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	_, err := client.FetchRelease("wave-cli", "nonexistent", "")
	if err == nil {
		t.Error("Should fail for 404")
	}
}

// --- tar.gz extraction ---

func TestExtractTarGz(t *testing.T) {
	// Create a tar.gz in memory with a single file
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	content := []byte("#!/bin/sh\necho hello")
	hdr := &tar.Header{
		Name: "flow",
		Mode: 0755,
		Size: int64(len(content)),
	}
	tw.WriteHeader(hdr)
	tw.Write(content)
	tw.Close()
	gw.Close()

	destDir := t.TempDir()
	err := ExtractTarGz(bytes.NewReader(buf.Bytes()), destDir)
	if err != nil {
		t.Fatalf("ExtractTarGz failed: %v", err)
	}

	extracted := filepath.Join(destDir, "flow")
	if _, err := os.Stat(extracted); os.IsNotExist(err) {
		t.Error("Extracted file should exist")
	}

	data, _ := os.ReadFile(extracted)
	if string(data) != string(content) {
		t.Errorf("Content mismatch: got %q", string(data))
	}
}

func TestExtractTarGzWithSubdir(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add a directory entry
	tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     "templates/",
		Mode:     0755,
	})
	// Add a file in that directory
	content := []byte("template content")
	tw.WriteHeader(&tar.Header{
		Name: "templates/default.toml",
		Mode: 0644,
		Size: int64(len(content)),
	})
	tw.Write(content)
	tw.Close()
	gw.Close()

	destDir := t.TempDir()
	err := ExtractTarGz(bytes.NewReader(buf.Bytes()), destDir)
	if err != nil {
		t.Fatalf("ExtractTarGz failed: %v", err)
	}

	extracted := filepath.Join(destDir, "templates", "default.toml")
	if _, err := os.Stat(extracted); os.IsNotExist(err) {
		t.Error("Nested file should exist after extraction")
	}
}

// --- Download file (mocked) ---

func TestDownloadFile(t *testing.T) {
	content := "binary content here"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "downloaded")
	client := NewClient("", "")
	err := client.DownloadFile(server.URL, dest)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}

	data, _ := os.ReadFile(dest)
	if string(data) != content {
		t.Errorf("Downloaded content = %q", string(data))
	}
}

func TestDownloadFileServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "downloaded")
	client := NewClient("", "")
	err := client.DownloadFile(server.URL, dest)
	if err == nil {
		t.Error("Should fail for 500")
	}
}

// --- InstallPlugin (full integration with fake server) ---

func TestInstallPlugin(t *testing.T) {
	home := t.TempDir()
	pluginsDir := filepath.Join(home, ".wave", "plugins")
	os.MkdirAll(pluginsDir, 0755)

	// Create a tar.gz with a fake binary
	var tarBuf bytes.Buffer
	gw := gzip.NewWriter(&tarBuf)
	tw := tar.NewWriter(gw)
	binContent := []byte("#!/bin/sh\necho hello")
	tw.WriteHeader(&tar.Header{Name: "flow", Mode: 0755, Size: int64(len(binContent))})
	tw.Write(binContent)
	tw.Close()
	gw.Close()

	wavepluginContent := `[plugin]
name = "flow"
version = "1.2.0"
description = "Workflow automation"
creator = "wave-cli"
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/wave-cli/flow/releases/latest":
			release := Release{
				TagName: "v1.2.0",
				Assets: []ReleaseAsset{
					{Name: "flow-" + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz", URL: r.Host + "/download/binary"},
					{Name: "Waveplugin", URL: r.Host + "/download/waveplugin"},
				},
			}
			// Rewrite URLs to point to our test server
			for i := range release.Assets {
				release.Assets[i].URL = "http://" + r.Host + "/download/" + release.Assets[i].Name
			}
			json.NewEncoder(w).Encode(release)
		case r.URL.Path == "/download/flow-"+runtime.GOOS+"-"+runtime.GOARCH+".tar.gz":
			w.Write(tarBuf.Bytes())
		case r.URL.Path == "/download/Waveplugin":
			w.Write([]byte(wavepluginContent))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "")
	err := client.InstallPlugin("wave-cli", "flow", "", pluginsDir)
	if err != nil {
		t.Fatalf("InstallPlugin failed: %v", err)
	}

	// Verify binary exists
	binPath := filepath.Join(pluginsDir, "wave-cli", "flow", "v1.2.0", "bin", "flow")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Error("Binary should be installed")
	}

	// Verify Waveplugin exists
	wpPath := filepath.Join(pluginsDir, "wave-cli", "flow", "v1.2.0", "Waveplugin")
	if _, err := os.Stat(wpPath); os.IsNotExist(err) {
		t.Error("Waveplugin should be installed")
	}

	// Verify current symlink
	currentLink := filepath.Join(pluginsDir, "wave-cli", "flow", "current")
	target, err := os.Readlink(currentLink)
	if err != nil {
		t.Fatalf("current symlink: %v", err)
	}
	expected := filepath.Join(pluginsDir, "wave-cli", "flow", "v1.2.0")
	if target != expected {
		t.Errorf("current -> %q, want %q", target, expected)
	}
}
