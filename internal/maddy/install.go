package maddy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type AdditionalDomain struct {
	MailDomain string
}

type InstallParameters struct {
	Version         string
	GitRef          string
	Repository      string
	DownloadBaseURL string
	ForceCompile    bool

	InstallRoot   string
	InstallPrefix string

	KeepArtifacts bool
}

type ConfigParameters struct {
	Hostname              string
	PrimaryMailDomain     string
	AdditionalDomains     []AdditionalDomain
	AcmeRegistrationEmail string

	InstallRoot   string
	InstallPrefix string
	MaddyUser     string
	MaddyGroup    string
}

func Install(ctx context.Context, params InstallParameters) error {
	if params.ForceCompile {
		return buildAndInstall(ctx, params)
	}

	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return downloadAndInstall(ctx, params, "x86_64-linux-musl")
		case "arm64":
			return buildAndInstall(ctx, params)
		default:
			return fmt.Errorf("unsupported architecture: %q", runtime.GOARCH)
		}
	default:
		return fmt.Errorf("unsupported operating system: %q", runtime.GOOS)
	}
}

func Config(ctx context.Context, params ConfigParameters) error {
	if err := configureDomains(ctx, params); err != nil {
		return fmt.Errorf("configuring domains: %w", err)
	}

	if err := configureCerts(ctx, params); err != nil {
		return fmt.Errorf("configuring certificates: %w", err)
	}

	if err := configurePermissions(ctx, params); err != nil {
		return fmt.Errorf("configuring certificates: %w", err)
	}

	return nil
}

func EnableAndStart() error {
	if _, err := exec.LookPath("systemctl"); err == nil {
		if out, err := exec.Command("systemctl", "daemon-reload").CombinedOutput(); err != nil {
			return fmt.Errorf("reloading systemctl: %w\nOutput: %s", err, string(out))
		}

		if out, err := exec.Command("systemctl", "enable", "maddy").CombinedOutput(); err != nil {
			return fmt.Errorf("enabling systemctl service: %w\nOutput: %s", err, string(out))
		}

		if out, err := exec.Command("systemctl", "start", "maddy").CombinedOutput(); err != nil {
			return fmt.Errorf("enabling systemctl service: %w\nOutput: %s", err, string(out))
		}
	} else {
		log.Printf("skipping systemctl steps; not on path")
	}

	return nil
}

func updateFile(ctx context.Context, name string, re regexp.Regexp, update string) error {
	of, err := os.CreateTemp("", "update-file-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	w := bufio.NewWriter(of)

	inf, err := os.Open(name)
	if err != nil {
		_ = of.Close()
		_ = os.Remove(of.Name())
		return fmt.Errorf("opening file %q for reading: %w", name, err)
	}
	defer inf.Close()

	nl := "\n"
	if runtime.GOOS == "windows" {
		nl = "\r\n"
	}

	fs := bufio.NewScanner(inf)
	for fs.Scan() {
		select {
		case <-ctx.Done():
			return fmt.Errorf("processing file interrupted: %w", ctx.Err())
		default:
			if _, err = w.WriteString(re.ReplaceAllString(fs.Text(), update) + nl); err != nil {
				_ = of.Close()
				_ = os.Remove(of.Name())
				return fmt.Errorf("writing to tempfile: %w", err)
			}
		}
	}

	if err = w.Flush(); err != nil {
		_ = of.Close()
		_ = os.Remove(of.Name())
		return fmt.Errorf("flushing tempfile: %w", err)
	}

	newFile := of.Name()
	_ = of.Close()

	backup := name + ".bak-" + time.Now().Format(time.RFC3339Nano)
	if err = os.Rename(name, backup); err != nil {
		return fmt.Errorf("renaming %q -> %q: %w", name, backup, err)
	}

	if err = os.Rename(newFile, name); err != nil {
		return fmt.Errorf("renaming %q -> %q: %w", newFile, name, err)
	}

	return nil
}

func downloadAndInstall(ctx context.Context, params InstallParameters, variant string) error {
	log.Printf("downloadAndInstall: %v", params)

	// set up build directory

	workingDir, cleanup, err := switchToTempDir(params.KeepArtifacts, "maddy-download-")
	defer cleanup()
	if err != nil {
		return fmt.Errorf("unable to switch to temp dir: %w", err)
	}

	// download and extract

	extractCmd := exec.Command("tar", "--zstd", "-xf", "-")
	stdin, err := extractCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("unable to get pipe to stdin from the extractCommand (%+v): %w", extractCmd, err)
	}

	baseURL := params.DownloadBaseURL
	if baseURL == "" {
		baseURL = "https://github.com/foxcpp/maddy/releases/download"
	}

	version := params.Version
	if version == "" {
		version = "0.8.1"
	}

	dir := fmt.Sprintf("maddy-%s-%s", version, variant)
	url := fmt.Sprintf("%s/v%s/%s.tar.zst", baseURL, version, dir)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}
	req.Header.Set("User-Agent", "maddy installer")

	log.Printf("begin download: %v", url)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to download %q: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected error downloading archive %q: %s", url, resp.Status)
	}

	log.Printf("begin extract")
	if err = extractCmd.Start(); err != nil {
		return fmt.Errorf("unable to start extract command %v: %w", extractCmd, err)
	}

	if _, err = io.Copy(stdin, resp.Body); err != nil {
		return fmt.Errorf("unable to pipe downloaded file to extract command %v: %w", extractCmd, err)
	}

	if err = stdin.Close(); err != nil {
		return fmt.Errorf("unable to close stdin pipe to extract command %v: %w", extractCmd, err)
	}

	log.Printf("wait for extract")
	if err = extractCmd.Wait(); err != nil {
		return fmt.Errorf("extract command %v failed: %w", extractCmd, err)
	}

	// install

	buildDir := filepath.Join(workingDir, dir)
	root := params.InstallRoot
	if root == "" {
		root = "/"
	}

	binDir := filepath.Join(root, params.InstallPrefix, "bin")
	configDir := filepath.Join(root, "etc/maddy")
	confTarget := filepath.Join(configDir, "maddy.conf")

	if out, err := exec.Command("install", "-m", "0755", "-d", binDir).CombinedOutput(); err != nil {
		return fmt.Errorf("installing bin directory %q: %w\nOutput: %s", binDir, err, string(out))
	}

	if out, err := exec.Command("install", "-m", "0755", filepath.Join(buildDir, "maddy"), binDir).CombinedOutput(); err != nil {
		return fmt.Errorf("installing maddy binary: %w\nOutput: %s", err, string(out))
	}

	if out, err := exec.Command("install", "-m", "0755", "-d", configDir).CombinedOutput(); err != nil {
		return fmt.Errorf("installing bin directory %q: %w\nOutput: %s", binDir, err, string(out))
	}

	if _, err = os.Stat(confTarget); err == nil {
		confTarget = filepath.Join(configDir, "maddy.conf.new-"+time.Now().Format(time.RFC3339Nano))
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("attempt to stat %q: %w", confTarget, err)
	}

	if out, err := exec.Command("install", "-m", "0644", filepath.Join(buildDir, "maddy.conf"), confTarget).CombinedOutput(); err != nil {
		return fmt.Errorf("installing maddy config %q: %w\nOutput: %s", confTarget, err, string(out))
	}

	if runtime.GOOS == "linux" {
		systemdDir := filepath.Join(root, params.InstallPrefix, "lib/systemd/system")
		if out, err := exec.Command("install", "-m", "0755", "-d", systemdDir).CombinedOutput(); err != nil {
			return fmt.Errorf("installing bin directory %q: %w\nOutput: %s", systemdDir, err, string(out))
		}

		systemdSrcFiles, err := filepath.Glob(filepath.Join(buildDir, "systemd/*.service"))
		if err != nil {
			return fmt.Errorf("could not find systemd service files: %w", err)
		}
		args := []string{"-m", "0644"}
		args = append(args, systemdSrcFiles...)
		args = append(args, systemdDir)
		if out, err := exec.Command("install", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("installing maddy systemd service: %w\nOutput: %s", err, string(out))
		}
	}

	manSrcDir := filepath.Join(buildDir, "man")
	if s, err := os.Stat(manSrcDir); err == nil && s.IsDir() {
		manSrcFiles, err := filepath.Glob(filepath.Join(manSrcDir, "*.1"))
		if err != nil {
			return fmt.Errorf("could not find man files: %w", err)
		}
		manTarget := filepath.Join(root, params.InstallPrefix, "share/man/man1")
		if out, err := exec.Command("install", "-m", "0755", "-d", manTarget).CombinedOutput(); err != nil {
			return fmt.Errorf("installing man directory %q: %w\nOutput: %s", manTarget, err, string(out))
		}
		for _, manFile := range manSrcFiles {
			if out, err := exec.Command("install", "-m", "0644", manFile, manTarget).CombinedOutput(); err != nil {
				return fmt.Errorf("installing maddy man file %q: %w\nOutput: %s", manTarget, err, string(out))
			}
		}
	}

	return nil
}

func buildAndInstall(ctx context.Context, params InstallParameters) error {
	log.Printf("build: %v", params)

	// set up build directory

	_, cleanup, err := switchToTempDir(params.KeepArtifacts, "maddy-build-")
	defer cleanup()
	if err != nil {
		return fmt.Errorf("unable to switch to temp dir: %w", err)
	}

	// clone repo

	repo := params.Repository
	if repo == "" {
		repo = "https://github.com/foxcpp/maddy"
	}

	version := params.Version
	if version == "" {
		version = "0.8.1"
	}

	ref := params.GitRef
	if ref == "" {
		ref = "refs/tags/v" + version
	}

	log.Printf("begin clone: %v", repo)
	if _, err = git.PlainCloneContext(ctx, "maddy", false, &git.CloneOptions{
		URL:           repo,
		ReferenceName: plumbing.ReferenceName(ref),
		SingleBranch:  true,
		Depth:         1,
	}); err != nil {
		return fmt.Errorf("cloning repository %q: %w", params.Repository, err)
	}

	// cd maddy

	if err = os.Chdir("./maddy"); err != nil {
		return fmt.Errorf("switching to maddy directory: %w", err)
	}

	// go mod download

	if out, err := exec.Command("go", "mod", "download").CombinedOutput(); err != nil {
		return fmt.Errorf("running go mod download: %w\n\nOutput: %s", err, string(out))
	}

	// go mod verify

	if out, err := exec.Command("go", "mod", "verify").CombinedOutput(); err != nil {
		return fmt.Errorf("running go mod verify: %w\n\nOutput: %s", err, string(out))
	}

	// build.sh

	if out, err := exec.Command("./build.sh").CombinedOutput(); err != nil {
		return fmt.Errorf("running build.sh: %w\n\nOutput: %s", err, string(out))
	}

	// build.sh install

	root := params.InstallRoot
	if root == "" {
		root = "/"
	}

	if out, err := exec.Command(
		"./build.sh",
		"--destdir", root,
		"--prefix", params.InstallPrefix,
		"install").CombinedOutput(); err != nil {
		return fmt.Errorf("running build.sh install: %w\n\nOutput: %s", err, string(out))
	}

	return nil
}

func configureDomains(ctx context.Context, params ConfigParameters) error {
	root := params.InstallRoot
	if root == "" {
		root = "/"
	}

	confFile := filepath.Join(root, "etc/maddy/maddy.conf")

	additionalDomains := ""
	for _, a := range params.AdditionalDomains {
		additionalDomains += " " + a.MailDomain
	}

	for name, edit := range map[string]struct {
		File        string
		RE          regexp.Regexp
		Replacement string
	}{
		"hostname": {File: confFile, RE: *regexp.MustCompile(`^\$\(hostname\) =.*$`), Replacement: "$(hostname) = " + params.Hostname},
		"primary":  {File: confFile, RE: *regexp.MustCompile(`^\$\(primary_domain\) =.*$`), Replacement: "$(primary_domain) = " + params.PrimaryMailDomain},
		"local":    {File: confFile, RE: *regexp.MustCompile(`^\$\(local_domains\) =.*$`), Replacement: "$(local_domains) = $(primary_domain)" + additionalDomains},
	} {
		if err := updateFile(ctx, edit.File, edit.RE, edit.Replacement); err != nil {
			return fmt.Errorf("unable to make edit %q (%+v): %w", name, edit, err)
		}
	}

	return nil
}

func configureCerts(_ context.Context, params ConfigParameters) error {
	root := params.InstallRoot
	if root == "" {
		root = "/"
	}

	maddyCertsDir := filepath.Join(root, "etc/maddy/certs")
	certsDir := filepath.Join(root, "etc/letsencrypt")
	for _, d := range []string{"live", "archive"} {
		dir := filepath.Join(certsDir, d)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("ensuring certs directories: %w", err)
		}
	}

	if err := os.Symlink(filepath.Join(certsDir, "live"), maddyCertsDir); err != nil {
		return fmt.Errorf("linking certs directories: %w", err)
	}

	return nil
}

func configurePermissions(_ context.Context, params ConfigParameters) error {
	root := params.InstallRoot
	if root == "" {
		root = "/"
	}

	user := params.MaddyUser
	if user == "" {
		user = "maddy"
	}

	group := params.MaddyGroup
	if group == "" {
		group = "maddy"
	}

	confDir := filepath.Join(root, "etc/maddy")
	certsDir := filepath.Join(root, "etc/letsencrypt")
	var leDirs []string
	for _, d := range []string{"live", "archive"} {
		leDirs = append(leDirs, filepath.Join(certsDir, d))
	}

	if _, err := exec.LookPath("setfacl"); err == nil {
		args := []string{
			"-R",
			"-m",
			"u:" + user + ":rX",
		}
		args = append(args, leDirs...)
		if out, err := exec.Command("setfacl", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("setting posix ACL permissions: %w\nOutput: %s", err, string(out))
		}
	} else {
		log.Printf("skipping ACL config: setfacl not found")
	}

	if out, err := exec.Command("chown", "-R", user+":"+group, confDir).CombinedOutput(); err != nil {
		return fmt.Errorf("setting permissions: %w\nOutput: %s", err, string(out))
	}

	return nil
}

func switchToTempDir(preserve bool, prefix string) (string, func(), error) {
	var cleanup func()

	prevDir, err := os.Getwd()
	if err != nil {
		return "", cleanup, fmt.Errorf("unable to get current directory: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", prefix+"*")
	if err != nil {
		return "", cleanup, fmt.Errorf("creating temporary directory: %w", err)
	}
	log.Printf("Temp Dir: %s", tmpDir)

	cleanup = func() {
		if preserve {
			return
		}

		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "error removing temporary directory %q: %v", tmpDir, err)
		}
	}

	if err = os.Chdir(tmpDir); err != nil {
		return "", cleanup, fmt.Errorf("switching to temporary build directory: %w", err)
	}

	return tmpDir, func() {
		if err := os.Chdir(prevDir); err != nil {
			fmt.Fprintf(os.Stderr, "error switching back to %q: %v", prevDir, err)
		}

		cleanup()
	}, nil
}
