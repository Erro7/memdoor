// `memdoor upgrade` checks the configured release source, compares it to the
// version baked into this binary, and runs its installer when newer.
//
// Why a thin wrapper around install.sh rather than re-implementing
// the swap logic here: scripts/install.sh already owns the gnarly
// part — detect existing install, SIGTERM the gateway, SIGKILL fallback, strip macOS
// quarantine, sudo/no-sudo branching. Forking that into Go would
// double the surface and split the source of truth. Better to keep
// install.sh canonical and have this command answer the "is there a
// newer build?" question, then delegate.
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// versionEndpointOverride lets tests replace the configured VERSION asset.
var versionEndpointOverride string

// installScriptURL is the curl|bash entrypoint install.sh publishes
// to. Same script handles fresh-install and upgrade-in-place (see
// scripts/install.sh, the "Upgrade-in-place" block).
func distributionBaseURL() string {
	if base := strings.TrimSpace(os.Getenv("MEMDOOR_DISTRIBUTION_URL")); base != "" {
		return strings.TrimRight(base, "/")
	}
	return "http://localhost:18789"
}

func versionEndpointURL() string {
	if versionEndpointOverride != "" {
		return versionEndpointOverride
	}
	if useGitHubReleases() {
		return githubReleaseBaseURL() + "/VERSION"
	}
	return distributionBaseURL() + "/dl/VERSION"
}

func installScriptURL() string {
	if useGitHubReleases() {
		return githubReleaseBaseURL() + "/install.sh"
	}
	return distributionBaseURL() + "/install.sh"
}

// installScriptURLWindows is the PowerShell installer the Windows binary
// hands itself to on upgrade.
func installScriptURLWindows() string {
	if useGitHubReleases() {
		return githubReleaseBaseURL() + "/install.ps1"
	}
	return distributionBaseURL() + "/install.ps1"
}

// installerLine is the one-liner the installer runs as on this machine.
func installerLine() string {
	if runtime.GOOS == "windows" {
		return "irm " + installScriptURLWindows() + " | iex"
	}
	return "curl -fsSL " + installScriptURL() + " | bash"
}

// installerCommand is that line in the shell it belongs to: PowerShell on
// Windows, bash elsewhere.
func installerCommand() *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("powershell", "-NoProfile", "-Command", installerLine())
	}
	return exec.Command("bash", "-c", installerLine())
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Check for a newer memdoor build and install it",
	Long: `Compares the version in this binary (memdoor version) with the latest
published build. When a newer one exists it runs the configured installer: the
running gateway is stopped, the binary swapped, the gateway started again.

  memdoor upgrade            ask, then upgrade
  memdoor upgrade --check    print the versions; exit 0 when current, 1 when an upgrade exists
  memdoor upgrade --yes      no question (a script)
  memdoor upgrade --force    run the installer even when the versions match (a build that says "dev")`,
	RunE: func(cmd *cobra.Command, args []string) error {
		checkOnly, _ := cmd.Flags().GetBool("check")
		force, _ := cmd.Flags().GetBool("force")
		yes, _ := cmd.Flags().GetBool("yes")

		remote, err := fetchRemoteVersion()
		if err != nil {
			return fmt.Errorf("fetch %s: %w", versionEndpointURL(), err)
		}
		local := Version

		fmt.Printf("Local:  %s\n", local)
		fmt.Printf("Remote: %s\n", remote)

		// "dev" or empty is always considered out-of-date for the
		// purposes of the version check — a locally-built binary
		// without ldflags has no provenance to compare against, so
		// the safe answer is "yes, you can upgrade to a real build".
		// --force keeps the door open for re-installing the same
		// version (e.g. after a botched swap that left a corrupted
		// binary on disk).
		isDev := local == "" || local == "dev"
		upToDate := !isDev && local == remote

		if upToDate && !force {
			fmt.Println("✓ Already on latest.")
			return nil
		}

		if checkOnly {
			if upToDate {
				return nil
			}
			fmt.Println("→ Upgrade available. Run 'memdoor upgrade' to install.")
			os.Exit(1)
		}

		if isDev {
			fmt.Println("Note: local version is \"dev\" — installer will replace this development binary with the published build.")
		}

		if !yes {
			fmt.Print("Proceed with install? [Y/n]: ")
			var ans string
			fmt.Scanln(&ans)
			ans = strings.TrimSpace(strings.ToLower(ans))
			if ans != "" && ans != "y" && ans != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		// Delegate to install.sh. Stream output through so the user
		// sees the same progress they'd get from the curl|bash flow.
		// install.sh owns every post-swap step, so every install path heals
		// the same way, including `curl|bash install.sh` users who never
		// invoke `memdoor upgrade`.
		fmt.Println("→ Running installer (" + installerLine() + ")…")
		c := installerCommand()
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		if err := c.Run(); err != nil {
			return err
		}
		return nil
	},
}

// fetchRemoteVersion reads the one-line plaintext VERSION file the
// deploy publishes alongside the binary. Trims surrounding whitespace
// since CDNs and editors occasionally add a trailing newline.
func fetchRemoteVersion() (string, error) { return fetchRemoteVersionFrom("", 10*time.Second) }

// fetchRemoteVersionFrom is fetchRemoteVersion with the door that asks named
// in the query (?from=cli for a headless command): memdoor-ops' metrics count
// only the window's daily check as a person that day, so the headless one is
// marked and left out (#90). Nothing about the person is sent.
func fetchRemoteVersionFrom(from string, timeout time.Duration) (string, error) {
	client := &http.Client{Timeout: timeout}
	u := versionEndpointURL()
	if from != "" {
		u += "?from=" + from
	}
	resp, err := client.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(string(body))
	if v == "" {
		return "", fmt.Errorf("VERSION endpoint returned empty body")
	}
	return v, nil
}

func init() {
	upgradeCmd.Flags().Bool("check", false, "Print the version delta and exit 0/1 — no download, no install")
	upgradeCmd.Flags().Bool("force", false, "Run the installer even when local matches remote")
	upgradeCmd.Flags().Bool("yes", false, "Skip the interactive confirm before running the installer")
	rootCmd.AddCommand(upgradeCmd)
}

func githubReleaseBaseURL() string {
	repository := strings.TrimSpace(os.Getenv("MEMDOOR_GITHUB_REPOSITORY"))
	if repository == "" {
		repository = "Erro7/memdoor"
	}
	release := strings.TrimSpace(os.Getenv("MEMDOOR_GITHUB_RELEASE"))
	if release == "" || release == "latest" {
		return "https://github.com/" + repository + "/releases/latest/download"
	}
	return "https://github.com/" + repository + "/releases/download/" + release
}

func useGitHubReleases() bool {
	source := strings.TrimSpace(os.Getenv("MEMDOOR_INSTALL_SOURCE"))
	return source == "" || strings.EqualFold(source, "github")
}
