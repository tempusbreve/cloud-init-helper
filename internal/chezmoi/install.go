package chezmoi

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type Options struct {
	Display        string
	Email          string
	GithubID       string
	AdditionalKeys []string

	ChezmoiRepo string

	ChezmoiConfigDir string
	InstallDirectory string
	SourceDirectory  string

	DryRun  bool
	Debug   bool
	Verbose bool
}

func (o Options) ConfigFile() string {
	if len(o.ChezmoiConfigDir) > 0 {
		return filepath.Join(o.ChezmoiConfigDir, "chezmoi.toml")
	}

	return ""
}

func (o Options) DotfilesRepo() string {
	if len(o.ChezmoiRepo) > 0 {
		return o.ChezmoiRepo
	}

	return "https://github.com/jw4/min.files"
}

func Install(ctx context.Context, opts Options) error {
	for ix, step := range []func(context.Context, Options) error{
		runInstaller,
	} {
		if err := step(ctx, opts); err != nil {
			return fmt.Errorf("running step %d: %w", ix, err)
		}
	}

	return nil
}

func Apply(ctx context.Context, opts Options) error {
	for ix, step := range []func(context.Context, Options) error{
		createTarget,
		writeTemplateConfig,
		apply,
	} {
		if err := step(ctx, opts); err != nil {
			return fmt.Errorf("running step %d: %w", ix, err)
		}
	}

	return nil
}

func createTarget(ctx context.Context, opts Options) error {
	log.Printf("creating config directory: %s", opts.ChezmoiConfigDir)

	if opts.DryRun {
		log.Printf("skipping: dry run")
	} else {
		return os.MkdirAll(opts.ChezmoiConfigDir, 0700)
	}

	return nil
}

func writeTemplateConfig(ctx context.Context, opts Options) error {
	configFile := opts.ConfigFile()
	log.Printf("writing config file: %s", configFile)

	var dest io.Writer

	if opts.DryRun {
		log.Printf("only writing to stdout: dry run")
		dest = io.Discard
	} else {
		out, err := os.Create(configFile)
		if err != nil {
			return fmt.Errorf("creating config file %q: %w", configFile, err)
		}
		defer out.Close()
		dest = out
	}

	return tmpl.Execute(io.MultiWriter(os.Stdout, dest), opts)
}

func apply(ctx context.Context, opts Options) error {
	configFile := opts.ConfigFile()
	repo := opts.DotfilesRepo()
	log.Printf("applying chezmoi with config file: %s, and options: %+v", configFile, opts)

	args := []string{"init", repo, "--apply", "--no-tty"}
	if configFile != "" {
		args = append(args, "--config", configFile, "--config-path", configFile)
	}
	if opts.SourceDirectory != "" {
		args = append(args, "--source", opts.SourceDirectory)
	}
	if opts.InstallDirectory != "" {
		args = append(args, "--destination", opts.InstallDirectory)
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	if opts.Verbose {
		args = append(args, "--verbose")
	}
	if opts.Debug {
		args = append(args, "--debug")
	}

	if out, err := exec.Command("chezmoi", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("installing chezmoi: %w\nOutput: %s", err, string(out))
	} else {
		log.Printf("Installer output:\n%s\n", string(out))
	}

	return nil
}

func runInstaller(ctx context.Context, opts Options) error {
	log.Printf("installing chezmoi binary, with config %+v", opts)

	//
	// download installer script
	//

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://git.io/chezmoi", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	cl := &http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to download script")
	}

	// copy to tempfile

	f, err := os.CreateTemp("", "chezmoi-init")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err = io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("copying response body: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("unable to write script file: %w", err)
	}

	// make it executable

	if err = os.Chmod(f.Name(), 0755); err != nil {
		return fmt.Errorf("unable to set execute permissions on file: %w", err)
	}

	args := []string{}
	if opts.Debug {
		args = append(args, "-d")
	}
	if out, err := exec.Command(f.Name(), args...).CombinedOutput(); err != nil {
		return fmt.Errorf("installing chezmoi: %w\nOutput: %s", err, string(out))
	} else {
		log.Printf("Installer output:\n%s\n", string(out))
	}

	return nil
}

var tmpl = template.Must(template.New("").Parse(configTemplate))

const configTemplate = `[data]
  [data.user]
    name = "{{ .Display }}"
    email = "{{ .Email }}"
    github = "{{ .GithubID }}"
  [data.ssh]
    authorized_keys = [{{ range .AdditionalKeys }}
      "{{ . }}"{{ end }}
    ]
`
