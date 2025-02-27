//go:build integration

package maddy

import (
	"context"
	"os/user"
	"testing"
	"time"
)

var preserve = false

func TestBuild(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	wd, cleanup, err := switchToTempDir(preserve, "maddy-TestBuild-")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	t.Logf("Test Artifacts: %q", wd)

	testParams := InstallParameters{
		InstallRoot:   wd + "/testroot/",
		InstallPrefix: "usr/local",
	}

	if err := buildAndInstall(ctx, testParams); err != nil {
		t.Fatal(err)
	}
}

func TestDownloadAndInstall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	wd, cleanup, err := switchToTempDir(preserve, "maddy-TestDownloadAndInstall-")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	t.Logf("Test Artifacts: %q", wd)

	testParams := InstallParameters{
		InstallRoot:   wd + "/testroot/",
		InstallPrefix: "usr/local",
	}

	if err := downloadAndInstall(ctx, testParams, "x86_64-linux-musl"); err != nil {
		t.Fatal(err)
	}

	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}

	gids, err := u.GroupIds()
	if err != nil {
		t.Fatal(err)
	}

	group := "nobody"
	if len(gids) > 0 {
		g, err := user.LookupGroupId(gids[0])
		if err != nil {
			t.Fatal(err)
		}
		group = g.Name
	}

	cp := ConfigParameters{
		InstallRoot:       testParams.InstallRoot,
		InstallPrefix:     testParams.InstallPrefix,
		MaddyUser:         u.Username,
		MaddyGroup:        group,
		Hostname:          "mx.foo.com",
		PrimaryMailDomain: "foo.com",
		AdditionalDomains: []AdditionalDomain{
			{MailDomain: "foo.org"},
			{MailDomain: "foo.net"},
		},
		AcmeRegistrationEmail: "foo@foo.net",
	}
	if err := Config(ctx, cp); err != nil {
		t.Fatal(err)
	}

	if err := EnableAndStart(); err != nil {
		t.Fatal(err)
	}
}
