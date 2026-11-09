package cmd

import "testing"

func TestDistributionURLsUseLocalHostForSourceBuild(t *testing.T) {
	t.Setenv("MEMDOOR_INSTALL_SOURCE", "")
	t.Setenv("MEMDOOR_DISTRIBUTION_URL", "")
	t.Setenv("MEMDOOR_GITHUB_REPOSITORY", "")
	t.Setenv("MEMDOOR_GITHUB_RELEASE", "")
	saved := ReleaseRepository
	ReleaseRepository = ""
	t.Cleanup(func() { ReleaseRepository = saved })

	if got, want := versionEndpointURL(), "http://localhost:18789/dl/VERSION"; got != want {
		t.Fatalf("versionEndpointURL() = %q, want %q", got, want)
	}
	if got, want := installScriptURL(), "http://localhost:18789/install.sh"; got != want {
		t.Fatalf("installScriptURL() = %q, want %q", got, want)
	}
}

func TestDistributionURLsUseBuildRepositoryForReleaseBinary(t *testing.T) {
	t.Setenv("MEMDOOR_INSTALL_SOURCE", "")
	t.Setenv("MEMDOOR_GITHUB_REPOSITORY", "")
	t.Setenv("MEMDOOR_GITHUB_RELEASE", "")
	saved := ReleaseRepository
	ReleaseRepository = "owner/repository"
	t.Cleanup(func() { ReleaseRepository = saved })

	if got, want := versionEndpointURL(), "https://github.com/owner/repository/releases/latest/download/VERSION"; got != want {
		t.Fatalf("versionEndpointURL() = %q, want %q", got, want)
	}
}

func TestDistributionURLsUseEnvironmentHost(t *testing.T) {
	t.Setenv("MEMDOOR_INSTALL_SOURCE", "local")
	t.Setenv("MEMDOOR_DISTRIBUTION_URL", " https://downloads.example.test/memdoor/ ")

	if got, want := versionEndpointURL(), "https://downloads.example.test/memdoor/dl/VERSION"; got != want {
		t.Fatalf("versionEndpointURL() = %q, want %q", got, want)
	}
	if got, want := installScriptURLWindows(), "https://downloads.example.test/memdoor/install.ps1"; got != want {
		t.Fatalf("installScriptURLWindows() = %q, want %q", got, want)
	}
}

func TestDistributionURLsUseLatestGitHubRelease(t *testing.T) {
	t.Setenv("MEMDOOR_INSTALL_SOURCE", "github")
	t.Setenv("MEMDOOR_GITHUB_REPOSITORY", "acme/memdoor")
	t.Setenv("MEMDOOR_GITHUB_RELEASE", "")
	saved := ReleaseRepository
	ReleaseRepository = ""
	t.Cleanup(func() { ReleaseRepository = saved })

	if got, want := versionEndpointURL(), "https://github.com/acme/memdoor/releases/latest/download/VERSION"; got != want {
		t.Fatalf("versionEndpointURL() = %q, want %q", got, want)
	}
	if got, want := installScriptURL(), "https://github.com/acme/memdoor/releases/latest/download/install.sh"; got != want {
		t.Fatalf("installScriptURL() = %q, want %q", got, want)
	}
}

func TestDistributionURLsUseTaggedGitHubRelease(t *testing.T) {
	t.Setenv("MEMDOOR_INSTALL_SOURCE", "github")
	t.Setenv("MEMDOOR_GITHUB_REPOSITORY", "acme/memdoor")
	t.Setenv("MEMDOOR_GITHUB_RELEASE", "v1.2.3")
	saved := ReleaseRepository
	ReleaseRepository = ""
	t.Cleanup(func() { ReleaseRepository = saved })

	if got, want := installScriptURLWindows(), "https://github.com/acme/memdoor/releases/download/v1.2.3/install.ps1"; got != want {
		t.Fatalf("installScriptURLWindows() = %q, want %q", got, want)
	}
}
