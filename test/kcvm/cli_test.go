package test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	cmds "github.com/raqqun/kcvm/pkg/kcvm"
	log "github.com/sirupsen/logrus"
)

// Fake kubectl releases served by the download server. A nil checksum means no
// .sha256 file is published, as for old kubectl releases.
var releases = map[string]struct {
	binary   string
	checksum *string
}{
	"v1.19.4": {binary: "kubectl v1.19.4", checksum: sha256Of("kubectl v1.19.4")},
	"v1.20.0": {binary: "kubectl v1.20.0", checksum: sha256Of("kubectl v1.20.0")},
	"v1.9.0":  {binary: "kubectl v1.9.0", checksum: nil},
	"v1.21.0": {binary: "kubectl v1.21.0", checksum: sha256Of("tampered")},
}

func sha256Of(s string) *string {
	sum := sha256.Sum256([]byte(s))
	hexSum := hex.EncodeToString(sum[:])
	return &hexSum
}

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)

	downloads := http.NewServeMux()
	for v, release := range releases {
		release := release
		binURL := fmt.Sprintf("/%s/bin/%s/%s/kubectl", v, runtime.GOOS, runtime.GOARCH)
		downloads.HandleFunc(binURL, func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, release.binary)
		})
		if release.checksum != nil {
			downloads.HandleFunc(binURL+".sha256", func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, *release.checksum)
			})
		}
	}
	downloadServer := httptest.NewServer(downloads)
	defer downloadServer.Close()

	gitServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/git-upload-pack" || r.Header.Get("Git-Protocol") != "version=2" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		for _, tag := range []string{"v0.2", "v1.19.4", "v1.2.0", "v1.20.0", "v1.21.0-rc.0", "v1.10.1", "v1.1.0-beta"} {
			line := fmt.Sprintf("0123456789abcdef0123456789abcdef01234567 refs/tags/%s\n", tag)
			fmt.Fprintf(w, "%04x%s", len(line)+4, line)
		}
		io.WriteString(w, "0000")
	}))
	defer gitServer.Close()

	cmds.KubectlDownloadURL = downloadServer.URL
	cmds.KubernetesRepoURL = gitServer.URL

	code := m.Run()
	downloadServer.Close()
	gitServer.Close()
	os.Exit(code)
}

// run runs kcvm with the given arguments against dir and returns its output
func run(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()

	var out bytes.Buffer
	kcvm := cmds.InitCLI()
	kcvm.Writer = &out
	kcvm.ErrWriter = io.Discard

	err := kcvm.Run(append([]string{"kcvm", "--dir", dir}, args...))
	return out.String(), err
}

func mustRun(t *testing.T, dir string, args ...string) string {
	t.Helper()

	out, err := run(t, dir, args...)
	if err != nil {
		t.Fatalf("kcvm %s: %v", strings.Join(args, " "), err)
	}
	return out
}

func assertActive(t *testing.T, dir string, kubectlVersion string) {
	t.Helper()

	target, err := os.Readlink(filepath.Join(dir, "kubectl"))
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(dir, "bin", "kubectl-"+kubectlVersion)
	if target != expected {
		t.Fatalf("kubectl points to %s, expected %s", target, expected)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != releases[kubectlVersion].binary {
		t.Fatalf("kubectl content is %q, expected %q", content, releases[kubectlVersion].binary)
	}
}

func assertBinDirContains(t *testing.T, dir string, expected ...string) {
	t.Helper()

	files, err := os.ReadDir(filepath.Join(dir, "bin"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	var names []string
	for _, file := range files {
		names = append(names, file.Name())
	}
	if strings.Join(names, ",") != strings.Join(expected, ",") {
		t.Fatalf("bin contains %v, expected %v", names, expected)
	}
}

func TestInitCmd(t *testing.T) {
	dir := t.TempDir()

	mustRun(t, dir, "init")
	mustRun(t, dir, "init")

	info, err := os.Stat(filepath.Join(dir, "bin"))
	if err != nil || !info.IsDir() {
		t.Fatalf("bin directory not created: %v", err)
	}
}

func TestInstallCmd(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "init")

	mustRun(t, dir, "install", "v1.20.0")

	assertActive(t, dir, "v1.20.0")
	info, err := os.Stat(filepath.Join(dir, "bin", "kubectl-v1.20.0"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Fatalf("kubectl is not executable: %v", info.Mode())
	}
}

func TestInstallWithoutInit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "kcvm")

	mustRun(t, dir, "install", "v1.20.0")

	assertActive(t, dir, "v1.20.0")
}

func TestInstallWithoutVPrefix(t *testing.T) {
	dir := t.TempDir()

	mustRun(t, dir, "install", "1.20.0")

	assertActive(t, dir, "v1.20.0")
}

func TestInstallWithRelativeDir(t *testing.T) {
	dir, _ := filepath.EvalSymlinks(t.TempDir())
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(cwd)

	mustRun(t, "kcvm", "install", "v1.20.0")

	assertActive(t, filepath.Join(dir, "kcvm"), "v1.20.0")
}

func TestInstallAlreadyInstalled(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "install", "v1.20.0")
	mustRun(t, dir, "install", "v1.19.4")

	mustRun(t, dir, "install", "v1.20.0")

	assertActive(t, dir, "v1.20.0")
}

func TestInstallForce(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "install", "v1.20.0")
	os.WriteFile(filepath.Join(dir, "bin", "kubectl-v1.20.0"), []byte("corrupt"), 0755)

	mustRun(t, dir, "install", "--force", "v1.20.0")

	assertActive(t, dir, "v1.20.0")
}

func TestInstallWithoutPublishedChecksum(t *testing.T) {
	dir := t.TempDir()

	mustRun(t, dir, "install", "v1.9.0")

	assertActive(t, dir, "v1.9.0")
}

func TestInstallChecksumMismatch(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "install", "v1.20.0")

	_, err := run(t, dir, "install", "v1.21.0")

	if err == nil || !strings.Contains(err.Error(), "Checksum mismatch") {
		t.Fatalf("expected a checksum error, got %v", err)
	}
	assertBinDirContains(t, dir, "kubectl-v1.20.0")
	assertActive(t, dir, "v1.20.0")
}

func TestInstallUnknownVersion(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "init")

	_, err := run(t, dir, "install", "v1.99.0")

	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected a missing version error, got %v", err)
	}
	assertBinDirContains(t, dir)
}

func TestInstallInvalidVersion(t *testing.T) {
	for _, arg := range []string{"", "latest", "v1.20", "../../v1.20.0", "v1.20.0/../x"} {
		dir := t.TempDir()

		args := []string{"install"}
		if arg != "" {
			args = append(args, arg)
		}
		_, err := run(t, dir, args...)

		if err == nil {
			t.Fatalf("expected an error for version %q", arg)
		}
		assertBinDirContains(t, dir)
	}
}

func TestUseCmd(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "install", "v1.19.4")
	mustRun(t, dir, "install", "v1.20.0")

	mustRun(t, dir, "use", "v1.19.4")
	assertActive(t, dir, "v1.19.4")

	mustRun(t, dir, "use", "1.20.0")
	assertActive(t, dir, "v1.20.0")
}

func TestUseNotInstalled(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "install", "v1.20.0")

	_, err := run(t, dir, "use", "v1.19.4")

	if err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Fatalf("expected a not installed error, got %v", err)
	}
	assertActive(t, dir, "v1.20.0")
}

func TestUseReplacesBrokenSymlink(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "install", "v1.19.4")
	mustRun(t, dir, "install", "v1.20.0")
	os.Remove(filepath.Join(dir, "bin", "kubectl-v1.20.0"))

	mustRun(t, dir, "use", "v1.19.4")

	assertActive(t, dir, "v1.19.4")
}

func TestListLocalCmd(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "install", "v1.9.0")
	mustRun(t, dir, "install", "v1.20.0")
	mustRun(t, dir, "install", "v1.19.4")
	os.WriteFile(filepath.Join(dir, "bin", "notes.txt"), nil, 0644)
	os.WriteFile(filepath.Join(dir, "bin", ".kubectl-v1.22.0.123.tmp"), nil, 0644)

	out := mustRun(t, dir, "list", "local")

	expected := "  v1.20.0\n* v1.19.4\n  v1.9.0\n"
	if out != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s", out, expected)
	}
}

func TestListLocalNotInitialized(t *testing.T) {
	_, err := run(t, filepath.Join(t.TempDir(), "kcvm"), "list", "local")

	if err == nil || !strings.Contains(err.Error(), "kcvm init") {
		t.Fatalf("expected a not initialized error, got %v", err)
	}
}

func TestListRemoteCmd(t *testing.T) {
	out := mustRun(t, t.TempDir(), "list", "remote")

	expected := "v1.20.0\nv1.19.4\nv1.10.1\nv1.2.0\n"
	if out != expected {
		t.Fatalf("got:\n%s\nexpected:\n%s", out, expected)
	}
}

func TestListRemoteServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	repoURL := cmds.KubernetesRepoURL
	cmds.KubernetesRepoURL = server.URL
	defer func() { cmds.KubernetesRepoURL = repoURL }()

	_, err := run(t, t.TempDir(), "list", "remote")

	if err == nil {
		t.Fatal("expected an error")
	}
}
