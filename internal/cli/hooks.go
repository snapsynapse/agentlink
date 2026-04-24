package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage automatic sync triggers",
	Long: `Install or remove automatic triggers that run agentlink sync.

Supported triggers:
  --git       Global git hook (post-checkout, post-merge) via core.hooksPath
  --zsh       Zsh chpwd hook that syncs on directory change into git repos
  --launchd   macOS LaunchAgent for periodic sync (every 60 minutes)
  --all       Install all triggers

Use 'agentlink hooks install' to set up triggers and 'agentlink hooks remove'
to uninstall them.`,
}

var hooksInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install sync triggers",
	RunE:  runHooksInstall,
}

var hooksRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove sync triggers",
	RunE:  runHooksRemove,
}

var hooksStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed trigger status",
	RunE:  runHooksStatus,
}

var (
	hookGit     bool
	hookZsh     bool
	hookLaunchd bool
	hookAll     bool
)

func init() {
	// Shared flags for install and remove
	for _, cmd := range []*cobra.Command{hooksInstallCmd, hooksRemoveCmd} {
		cmd.Flags().BoolVar(&hookGit, "git", false, "git global hooks (post-checkout, post-merge)")
		cmd.Flags().BoolVar(&hookZsh, "zsh", false, "zsh chpwd hook")
		cmd.Flags().BoolVar(&hookLaunchd, "launchd", false, "macOS LaunchAgent (60-minute heartbeat)")
		cmd.Flags().BoolVar(&hookAll, "all", false, "all triggers")
	}

	hooksCmd.AddCommand(hooksInstallCmd)
	hooksCmd.AddCommand(hooksRemoveCmd)
	hooksCmd.AddCommand(hooksStatusCmd)
	rootCmd.AddCommand(hooksCmd)
}

// --- Install ---

func runHooksInstall(cmd *cobra.Command, args []string) error {
	if hookAll {
		hookGit = true
		hookZsh = true
		hookLaunchd = true
	}

	if !hookGit && !hookZsh && !hookLaunchd {
		return fmt.Errorf("specify at least one trigger: --git, --zsh, --launchd, or --all")
	}

	binaryPath, err := resolveAgentlinkBinary()
	if err != nil {
		return err
	}

	var errs []string

	if hookGit {
		if err := installGitHooks(binaryPath); err != nil {
			errs = append(errs, fmt.Sprintf("git: %v", err))
		}
	}

	if hookZsh {
		if err := installZshHook(binaryPath); err != nil {
			errs = append(errs, fmt.Sprintf("zsh: %v", err))
		}
	}

	if hookLaunchd {
		if err := installLaunchdAgent(binaryPath); err != nil {
			errs = append(errs, fmt.Sprintf("launchd: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("some hooks failed to install:\n  %s", strings.Join(errs, "\n  "))
	}

	return nil
}

// --- Remove ---

func runHooksRemove(cmd *cobra.Command, args []string) error {
	if hookAll {
		hookGit = true
		hookZsh = true
		hookLaunchd = true
	}

	if !hookGit && !hookZsh && !hookLaunchd {
		return fmt.Errorf("specify at least one trigger: --git, --zsh, --launchd, or --all")
	}

	if hookGit {
		removeGitHooks()
	}
	if hookZsh {
		removeZshHook()
	}
	if hookLaunchd {
		removeLaunchdAgent()
	}

	return nil
}

// --- Status ---

func runHooksStatus(cmd *cobra.Command, args []string) error {
	fmt.Printf("Trigger Status\n")
	fmt.Printf("==============\n\n")

	// Git hooks
	hooksDir := gitGlobalHooksDir()
	if hooksDir != "" {
		postCheckout := filepath.Join(hooksDir, "post-checkout")
		postMerge := filepath.Join(hooksDir, "post-merge")
		hasCheckout := fileContainsAgentlink(postCheckout)
		hasMerge := fileContainsAgentlink(postMerge)
		if hasCheckout && hasMerge {
			fmt.Printf("  git hooks:    installed (%s)\n", hooksDir)
		} else if hasCheckout || hasMerge {
			fmt.Printf("  git hooks:    partial (%s)\n", hooksDir)
		} else {
			fmt.Printf("  git hooks:    not installed\n")
		}
	} else {
		fmt.Printf("  git hooks:    not installed (no core.hooksPath set)\n")
	}

	// Zsh hook
	homeDir, _ := os.UserHomeDir()
	zshrc := filepath.Join(homeDir, ".zshrc")
	if fileContainsAgentlink(zshrc) {
		fmt.Printf("  zsh chpwd:    installed (~/.zshrc)\n")
	} else {
		fmt.Printf("  zsh chpwd:    not installed\n")
	}

	// Launchd
	plistPath := launchdPlistPath()
	if _, err := os.Stat(plistPath); err == nil {
		fmt.Printf("  launchd:      installed (%s)\n", plistPath)
	} else {
		fmt.Printf("  launchd:      not installed\n")
	}

	return nil
}

// --- Git Hooks ---

const gitHookMarkerStart = "# >>> agentlink >>>"
const gitHookMarkerEnd = "# <<< agentlink <<<"

func gitHookContent(binaryPath string) string {
	quotedBinaryPath := shellQuote(binaryPath)
	return fmt.Sprintf(`%s
# Sync agent instruction symlinks after git operations.
# Installed by: agentlink hooks install --git
%s sync --quiet 2>/dev/null || true
%s
`, gitHookMarkerStart, quotedBinaryPath, gitHookMarkerEnd)
}

func installGitHooks(binaryPath string) error {
	// Determine or create global hooks directory
	hooksDir := gitGlobalHooksDir()
	if hooksDir == "" {
		homeDir, _ := os.UserHomeDir()
		hooksDir = filepath.Join(homeDir, ".config", "git", "hooks")
	}

	if dryRun {
		printInfo("Would install git hooks in %s", hooksDir)
		return nil
	}

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("cannot create hooks directory: %w", err)
	}

	// Set core.hooksPath if not already set
	currentDir := gitGlobalHooksDir()
	if currentDir == "" {
		cmd := exec.Command("git", "config", "--global", "core.hooksPath", hooksDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set core.hooksPath: %w", err)
		}
		printOK("Set git core.hooksPath to %s", hooksDir)
	}

	hookContent := gitHookContent(binaryPath)

	for _, hookName := range []string{"post-checkout", "post-merge"} {
		hookPath := filepath.Join(hooksDir, hookName)
		if err := appendOrCreateHook(hookPath, hookContent); err != nil {
			return fmt.Errorf("failed to install %s: %w", hookName, err)
		}
		printOK("Installed %s hook", hookName)
	}

	return nil
}

func removeGitHooks() {
	hooksDir := gitGlobalHooksDir()
	if hooksDir == "" {
		printInfo("No global git hooks directory configured")
		return
	}

	for _, hookName := range []string{"post-checkout", "post-merge"} {
		hookPath := filepath.Join(hooksDir, hookName)
		if removeMarkedSection(hookPath) {
			printOK("Removed agentlink from %s", hookName)
		}
	}
}

func gitGlobalHooksDir() string {
	out, err := exec.Command("git", "config", "--global", "core.hooksPath").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// --- Zsh Hook ---

func zshHookContent(binaryPath string) string {
	quotedBinaryPath := shellQuote(binaryPath)
	return fmt.Sprintf(`
%s
# Sync agent instruction symlinks when entering a git repo.
# Installed by: agentlink hooks install --zsh
agentlink_chpwd() {
  if [ -d .git ] && [ -f .agentlink.yaml ]; then
    %s sync --quiet 2>/dev/null &!
  fi
}
autoload -U add-zsh-hook
add-zsh-hook chpwd agentlink_chpwd
%s
`, gitHookMarkerStart, quotedBinaryPath, gitHookMarkerEnd)
}

func installZshHook(binaryPath string) error {
	homeDir, _ := os.UserHomeDir()
	zshrc := filepath.Join(homeDir, ".zshrc")

	if fileContainsAgentlink(zshrc) {
		printSkip("zsh hook already installed in %s", zshrc)
		return nil
	}

	if dryRun {
		printInfo("Would append chpwd hook to %s", zshrc)
		return nil
	}

	f, err := os.OpenFile(zshrc, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", zshrc, err)
	}
	defer f.Close()

	if _, err := f.WriteString(zshHookContent(binaryPath)); err != nil {
		return fmt.Errorf("cannot write to %s: %w", zshrc, err)
	}

	printOK("Installed chpwd hook in %s", zshrc)
	return nil
}

func removeZshHook() {
	homeDir, _ := os.UserHomeDir()
	zshrc := filepath.Join(homeDir, ".zshrc")
	if removeMarkedSection(zshrc) {
		printOK("Removed agentlink hook from ~/.zshrc")
	} else {
		printInfo("No agentlink hook found in ~/.zshrc")
	}
}

// --- Launchd ---

const launchdLabel = "com.agentlink.sync"

func launchdPlistPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Library", "LaunchAgents", launchdLabel+".plist")
}

func launchdPlistContent(binaryPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>sync</string>
        <string>--quiet</string>
    </array>
    <key>StartInterval</key>
    <integer>3600</integer>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/agentlink-sync.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/agentlink-sync.err</string>
</dict>
</plist>
`, launchdLabel, binaryPath)
}

func installLaunchdAgent(binaryPath string) error {
	plistPath := launchdPlistPath()

	if dryRun {
		printInfo("Would install LaunchAgent at %s", plistPath)
		return nil
	}

	// Ensure LaunchAgents directory exists
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return fmt.Errorf("cannot create LaunchAgents directory: %w", err)
	}

	// Unload existing agent if present
	if _, err := os.Stat(plistPath); err == nil {
		exec.Command("launchctl", "unload", plistPath).Run()
	}

	if err := os.WriteFile(plistPath, []byte(launchdPlistContent(binaryPath)), 0644); err != nil {
		return fmt.Errorf("cannot write plist: %w", err)
	}

	// Load the agent
	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		printWarning("Plist written but launchctl load failed: %v", err)
		printInfo("Try: launchctl load %s", plistPath)
		return nil
	}

	printOK("Installed LaunchAgent (60-minute heartbeat)")
	return nil
}

func removeLaunchdAgent() {
	plistPath := launchdPlistPath()

	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		printInfo("No LaunchAgent installed")
		return
	}

	exec.Command("launchctl", "unload", plistPath).Run()
	os.Remove(plistPath)
	printOK("Removed LaunchAgent")
}

// --- Helpers ---

func resolveAgentlinkBinary() (string, error) {
	// Try to find ourselves in PATH first
	if path, err := exec.LookPath("agentlink"); err == nil {
		return path, nil
	}

	// Fall back to current executable path
	if path, err := os.Executable(); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("cannot determine agentlink binary path; ensure it is in PATH")
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func fileContainsAgentlink(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), gitHookMarkerStart)
}

func appendOrCreateHook(path, content string) error {
	// If file exists and already has our marker, skip
	if fileContainsAgentlink(path) {
		return nil
	}

	// If file doesn't exist, create with shebang
	if _, err := os.Stat(path); os.IsNotExist(err) {
		full := "#!/bin/sh\n" + content
		if err := os.WriteFile(path, []byte(full), 0755); err != nil {
			return err
		}
		return nil
	}

	// Append to existing hook
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("\n" + content)
	return err
}

func removeMarkedSection(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	content := string(data)
	startIdx := strings.Index(content, gitHookMarkerStart)
	endIdx := strings.Index(content, gitHookMarkerEnd)

	if startIdx == -1 || endIdx == -1 {
		return false
	}

	// Remove from start marker to end marker (inclusive of trailing newline)
	endIdx += len(gitHookMarkerEnd)
	if endIdx < len(content) && content[endIdx] == '\n' {
		endIdx++
	}

	newContent := content[:startIdx] + content[endIdx:]

	// Clean up double blank lines left behind
	for strings.Contains(newContent, "\n\n\n") {
		newContent = strings.ReplaceAll(newContent, "\n\n\n", "\n\n")
	}

	os.WriteFile(path, []byte(newContent), 0644)
	return true
}
