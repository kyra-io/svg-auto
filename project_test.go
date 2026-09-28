package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSSHHelperProcess(t *testing.T) {
	if os.Getenv("SVG_AUTO_SSH_HELPER") == "" {
		return
	}
	data, _ := io.ReadAll(os.Stdin)
	if os.Getenv("SVG_AUTO_SSH_HELPER") == "fail" {
		fmt.Fprint(os.Stderr, "simulated SSH failure")
		os.Exit(2)
	}
	if len(data) > 0 {
		fmt.Fprintf(os.Stdout, "stdin:%s", data)
	} else {
		fmt.Fprint(os.Stdout, "remote-data")
	}
	os.Exit(0)
}

func installSSHHelper(t *testing.T, mode string, calls *[][2]string) {
	t.Helper()
	original := sshCommand
	sshCommand = func(host, command string) *exec.Cmd {
		*calls = append(*calls, [2]string{host, command})
		cmd := exec.Command(os.Args[0], "-test.run=^TestSSHHelperProcess$")
		cmd.Env = append(os.Environ(), "SVG_AUTO_SSH_HELPER="+mode)
		return cmd
	}
	t.Cleanup(func() { sshCommand = original })
}

func TestIsRemote(t *testing.T) {
	remote := []string{
		"user@host:/srv/app",
		"host:/srv/app",
		"host:~/app",
		"user@host:~/app",
	}
	local := []string{
		"/srv/app",
		"/home/user/app",
		".",
		"C:\\Users\\x\\app",
		"relative/path",
		"",
		"host",
	}
	for _, p := range remote {
		if !isRemote(p) {
			t.Errorf("expected %q to be remote", p)
		}
	}
	for _, p := range local {
		if isRemote(p) {
			t.Errorf("expected %q to be local", p)
		}
	}
}

func TestProjectTarget(t *testing.T) {
	if got, want := newProject("/srv/app").target("icons/sprite.svg"), filepath.Join("/srv/app", "icons/sprite.svg"); got != want {
		t.Errorf("unexpected target: %s", got)
	}
	p := newProject("user@host:/srv/app")
	if !p.remote() {
		t.Fatal("expected remote project")
	}
	if got := p.target("icons/sprite.svg"); got != "/srv/app/icons/sprite.svg" {
		t.Errorf("unexpected remote target: %s", got)
	}
	if got := newProject("user@host:~/app").target("icons/sprite.svg"); got != "~/app/icons/sprite.svg" {
		t.Errorf("unexpected remote home target: %s", got)
	}
}

func TestProjectLocalReadWriteBackup(t *testing.T) {
	base := t.TempDir()
	p := newProject(base)

	if err := p.writeFile("sprite.svg", []byte("<svg></svg>")); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	data, err := p.readFile("sprite.svg")
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "<svg></svg>" {
		t.Errorf("unexpected content: %s", data)
	}

	if err := p.backup("sprite.svg"); err != nil {
		t.Fatalf("backup failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "sprite.svg.orig")); err != nil {
		t.Errorf("backup file not created: %v", err)
	}
}

func TestProjectLocalReadMissing(t *testing.T) {
	p := newProject(t.TempDir())
	if _, err := p.readFile("missing.svg"); err == nil {
		t.Fatal("expected error reading missing file, got nil")
	} else if !strings.Contains(err.Error(), "missing.svg") {
		t.Fatalf("error should mention the file: %v", err)
	}
}

func TestRunSSH(t *testing.T) {
	var calls [][2]string
	installSSHHelper(t, "success", &calls)

	out, err := runSSH("user@host", "cat '/srv/app/file'", []byte("payload"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "stdin:payload" {
		t.Errorf("unexpected output: %q", out)
	}
	if len(calls) != 1 || calls[0] != [2]string{"user@host", "cat '/srv/app/file'"} {
		t.Errorf("unexpected SSH call: %v", calls)
	}
}

func TestRunSSHError(t *testing.T) {
	var calls [][2]string
	installSSHHelper(t, "fail", &calls)

	_, err := runSSH("host", "false", nil)
	if err == nil || !strings.Contains(err.Error(), "simulated SSH failure") {
		t.Fatalf("expected SSH stderr in the error, got %v", err)
	}
}

func TestProjectRemoteOperations(t *testing.T) {
	var calls [][2]string
	installSSHHelper(t, "success", &calls)
	p := newProject("user@host:/srv/app")

	data, err := p.readFile("icons/sprite.svg")
	if err != nil || string(data) != "remote-data" {
		t.Fatalf("unexpected remote read: data=%q err=%v", data, err)
	}
	if err := p.writeFile("icons/sprite.svg", []byte("new-content")); err != nil {
		t.Fatalf("remote write failed: %v", err)
	}
	if err := p.backup("icons/sprite.svg"); err != nil {
		t.Fatalf("remote backup failed: %v", err)
	}
	if err := p.removeFile("icons/sprite.svg.orig"); err != nil {
		t.Fatalf("remote remove failed: %v", err)
	}

	want := [][2]string{
		{"user@host", "cat '/srv/app/icons/sprite.svg'"},
		{"user@host", "cat > '/srv/app/icons/sprite.svg'"},
		{"user@host", "cp '/srv/app/icons/sprite.svg' '/srv/app/icons/sprite.svg.orig'"},
		{"user@host", "rm -f '/srv/app/icons/sprite.svg.orig'"},
	}
	if fmt.Sprint(calls) != fmt.Sprint(want) {
		t.Errorf("unexpected SSH calls:\n got: %v\nwant: %v", calls, want)
	}
}
