package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "skill-cli-bin-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not create temp dir:", err)
		os.Exit(1)
	}

	name := "skill-cli"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binaryPath = filepath.Join(tmp, name)

	out, err := exec.Command("go", "build", "-o", binaryPath, "../").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s\n", err, out)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// run executes the binary with the given XDG config home and args.
// Returns stdout, stderr, and the exit code.
func run(t *testing.T, xdgHome string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+xdgHome, "HOME="+xdgHome)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected run error: %v", err)
		}
	}
	return stdout.String(), stderr.String(), code
}

// newEnv returns a fresh temp directory to use as XDG_CONFIG_HOME.
func newEnv(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// fixture returns the absolute path to a testdata file.
func fixture(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

// serveSkill starts an httptest server serving content at /SKILL.md.
// Returns the URL to the raw file.
func serveSkill(t *testing.T, content string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, content)
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/SKILL.md"
}

// serve404 starts an httptest server that always returns 404.
func serve404(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/SKILL.md"
}

// localBareRepo creates a bare git repo and returns its file:/// URL.
func localBareRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	out, err := exec.Command("git", "init", "--bare", dir).CombinedOutput()
	if err != nil {
		t.Fatalf("git init --bare: %v\n%s", err, out)
	}
	return "file://" + dir
}

// readFixture reads a testdata file as a string.
func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(fixture(name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// ---- unknown command ----
func TestUnknownCommand(t *testing.T) {
	env := newEnv(t)
	_, stderr, code := run(t, env, "bogus")
	if code == 0 {
		t.Fatal("expected exit 1")
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Errorf("expected 'unknown command' in stderr, got: %q", stderr)
	}
}
