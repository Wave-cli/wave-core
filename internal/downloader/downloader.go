// Package downloader handles fetching plugin releases from GitHub and
// installing them into the local plugin store.
package downloader

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Release represents a GitHub release.
type Release struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

// ReleaseAsset represents a single asset in a GitHub release.
type ReleaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// Client is a GitHub Releases API client.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient creates a new downloader client.
// baseURL is the GitHub API base (e.g. "https://api.github.com").
// token is an optional GitHub token for authenticated requests.
func NewClient(baseURL, token string) *Client {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{},
	}
}

// FetchRelease fetches a release from GitHub. If version is empty, fetches latest.
func (c *Client) FetchRelease(org, name, version string) (*Release, error) {
	var url string
	if version == "" {
		url = fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.baseURL, org, name)
	} else {
		// Ensure version has v prefix for GitHub tag
		v := version
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
		url = fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", c.baseURL, org, name, v)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d for %s/%s", resp.StatusCode, org, name)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding release: %w", err)
	}

	return &release, nil
}

// SelectAsset finds the binary asset matching the given OS and architecture.
// It looks for patterns like: <name>-<os>-<arch>.tar.gz
func SelectAsset(assets []ReleaseAsset, pluginName, goos, goarch string) (*ReleaseAsset, error) {
	if len(assets) == 0 {
		return nil, fmt.Errorf("no assets in release")
	}

	pattern := fmt.Sprintf("%s-%s-%s", pluginName, goos, goarch)

	for i, a := range assets {
		if strings.HasPrefix(a.Name, pattern) {
			return &assets[i], nil
		}
	}

	return nil, fmt.Errorf("no asset matching %s found in release", pattern)
}

// FindWavepluginAsset finds the Waveplugin metadata asset in the release.
func FindWavepluginAsset(assets []ReleaseAsset) *ReleaseAsset {
	for i, a := range assets {
		if a.Name == "Waveplugin" {
			return &assets[i]
		}
	}
	return nil
}

// DownloadFile downloads a URL to a local file path.
func (c *Client) DownloadFile(url, destPath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %d for %s", resp.StatusCode, url)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// ExtractTarGz extracts a tar.gz archive from r into destDir.
func ExtractTarGz(r io.Reader, destDir string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}

		target := filepath.Join(destDir, hdr.Name)

		// Security: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("tar entry %q escapes destination", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("creating parent directory: %w", err)
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return fmt.Errorf("creating file: %w", err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("writing file: %w", err)
			}
			f.Close()
		}
	}

	return nil
}

// InstallPlugin performs the full installation flow:
// 1. Fetch release from GitHub
// 2. Download binary archive + Waveplugin
// 3. Extract to plugins dir
func (c *Client) InstallPlugin(org, name, version, pluginsDir string) error {
	// 1. Fetch release
	release, err := c.FetchRelease(org, name, version)
	if err != nil {
		return fmt.Errorf("fetching release: %w", err)
	}

	// 2. Select binary asset
	binaryAsset, err := SelectAsset(release.Assets, name, currentOS(), currentArch())
	if err != nil {
		return fmt.Errorf("selecting binary: %w", err)
	}

	// 3. Set up directories - now directly under plugins/<name>
	pluginDir := filepath.Join(pluginsDir, name)
	binDir := filepath.Join(pluginDir, "bin")
	assetsDir := filepath.Join(pluginDir, "assets")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("creating bin dir: %w", err)
	}
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return fmt.Errorf("creating assets dir: %w", err)
	}

	// 4. Download and extract binary
	tmpFile := filepath.Join(os.TempDir(), "wave-download-"+name+".tar.gz")
	defer os.Remove(tmpFile)

	if err := c.DownloadFile(binaryAsset.URL, tmpFile); err != nil {
		return fmt.Errorf("downloading binary: %w", err)
	}

	f, err := os.Open(tmpFile)
	if err != nil {
		return fmt.Errorf("opening archive: %w", err)
	}
	defer f.Close()

	if err := ExtractTarGz(f, binDir); err != nil {
		return fmt.Errorf("extracting binary: %w", err)
	}

	// Make binary executable
	binPath := filepath.Join(binDir, name)
	os.Chmod(binPath, 0755)

	// 5. Download Waveplugin
	wpAsset := FindWavepluginAsset(release.Assets)
	if wpAsset != nil {
		wpDest := filepath.Join(pluginDir, "Waveplugin")
		if err := c.DownloadFile(wpAsset.URL, wpDest); err != nil {
			return fmt.Errorf("downloading Waveplugin: %w", err)
		}
	}

	return nil
}

// currentOS returns the runtime OS.
func currentOS() string {
	return os_() // indirection for testing
}

// currentArch returns the runtime architecture.
func currentArch() string {
	return arch_() // indirection for testing
}
