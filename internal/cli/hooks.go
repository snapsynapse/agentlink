package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/snapsynapse/agentlink/internal/fsutil"
	"github.com/spf13/cobra"
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Inspect and remove automatic sync triggers",
	Long: `Inspect or remove the automatic sync triggers that earlier agentlink
releases installed.

Agentlink is retired, and installing new triggers is no longer supported.
Triggers installed by earlier releases keep running until they are removed:

  --git       Global git hooks (post-checkout, post-merge) via core.hooksPath
  --zsh       Zsh chpwd hook in ~/.zshrc
  --launchd   macOS LaunchAgent (com.agentlink.sync)
  --all       All of the above

Run 'agentlink hooks remove --all' to uninstall them.

Removing git hooks deletes hook files that contained nothing but agentlink's
block. It also unsets the global core.hooksPath when that points at
agentlink's default directory (~/.config/git/hooks) and the directory is left
empty. While core.hooksPath is set, Git ignores every repository's own
.git/hooks, so removal reports it when it has to stay set.

Agentlink only edits an absolute global core.hooksPath. If yours is relative,
replace it with the absolute path before removing git hooks.`,
}

// errHooksInstallRetired is returned for every 'hooks install' invocation.
var errHooksInstallRetired = errors.New("agentlink is retired; installing automatic triggers is no longer supported. Existing triggers keep running until removed: run 'agentlink hooks remove --all'")

var hooksInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Retired: installing triggers is no longer supported",
	// Accept the old flags so 'hooks install --all' reaches the retirement
	// message instead of an unknown-flag error.
	DisableFlagParsing: true,
	SilenceUsage:       true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errHooksInstallRetired
	},
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
	hooksRemoveCmd.Flags().BoolVar(&hookGit, "git", false, "git global hooks (post-checkout, post-merge)")
	hooksRemoveCmd.Flags().BoolVar(&hookZsh, "zsh", false, "zsh chpwd hook")
	hooksRemoveCmd.Flags().BoolVar(&hookLaunchd, "launchd", false, "macOS LaunchAgent")
	hooksRemoveCmd.Flags().BoolVar(&hookAll, "all", false, "all triggers")

	hooksCmd.AddCommand(hooksInstallCmd)
	hooksCmd.AddCommand(hooksRemoveCmd)
	hooksCmd.AddCommand(hooksStatusCmd)
	rootCmd.AddCommand(hooksCmd)
}

// --- Remove ---

func runHooksRemove(cmd *cobra.Command, args []string) error {
	if hookAll {
		hookGit = true
		hookZsh = true
		hookLaunchd = runtime.GOOS == "darwin"
	}

	if hookLaunchd && runtime.GOOS != "darwin" {
		return fmt.Errorf("launchd is supported only on macOS; use --git or --zsh on this platform")
	}
	if !hookGit && !hookZsh && !hookLaunchd {
		return fmt.Errorf("specify at least one trigger: --git, --zsh, --launchd, or --all")
	}

	var errs []string
	if hookGit {
		if err := removeGitHooks(); err != nil {
			errs = append(errs, fmt.Sprintf("git: %v", err))
		}
	}
	if hookZsh {
		if err := removeZshHook(); err != nil {
			errs = append(errs, fmt.Sprintf("zsh: %v", err))
		}
	}
	if hookLaunchd {
		if err := removeLaunchdAgent(); err != nil {
			errs = append(errs, fmt.Sprintf("launchd: %v", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("some hooks failed to remove:\n  %s", strings.Join(errs, "\n  "))
	}

	return nil
}

// --- Status ---

func runHooksStatus(cmd *cobra.Command, args []string) error {
	fmt.Printf("Trigger Status\n")
	fmt.Printf("==============\n\n")

	for _, line := range triggerStatusLines() {
		fmt.Printf("  %s\n", line)
	}
	if installedTriggerCount() > 0 {
		fmt.Printf("\nAgentlink is retired. Remove these triggers with 'agentlink hooks remove --all'.\n")
	}
	return nil
}

// triggerStatusLines reports each trigger in a stable order.
func triggerStatusLines() []string {
	var lines []string

	hooksDir := gitGlobalHooksDir()
	switch {
	case hooksDir == "":
		lines = append(lines, "git hooks:    not installed (no core.hooksPath set)")
	case !filepath.IsAbs(hooksDir):
		lines = append(lines, fmt.Sprintf("git hooks:    unsupported relative core.hooksPath (%s)", hooksDir))
	default:
		hasCheckout := fileContainsAgentlink(filepath.Join(hooksDir, "post-checkout"))
		hasMerge := fileContainsAgentlink(filepath.Join(hooksDir, "post-merge"))
		switch {
		case hasCheckout && hasMerge:
			lines = append(lines, fmt.Sprintf("git hooks:    installed (%s)", hooksDir))
		case hasCheckout || hasMerge:
			lines = append(lines, fmt.Sprintf("git hooks:    partial (%s)", hooksDir))
		default:
			lines = append(lines, "git hooks:    not installed")
		}
		lines = append(lines, fmt.Sprintf("core.hooksPath: %s (repository .git/hooks are ignored while set)", hooksDir))
	}

	if fileContainsAgentlink(zshrcPath()) {
		lines = append(lines, "zsh chpwd:    installed (~/.zshrc)")
	} else {
		lines = append(lines, "zsh chpwd:    not installed")
	}

	plistPath := launchdPlistPath()
	if _, err := os.Lstat(plistPath); err == nil {
		lines = append(lines, fmt.Sprintf("launchd:      installed (%s)", plistPath))
	} else {
		lines = append(lines, "launchd:      not installed")
	}
	return lines
}

// installedTriggerCount counts triggers that still run agentlink.
func installedTriggerCount() int {
	count := 0
	if hooksDir := gitGlobalHooksDir(); hooksDir != "" && filepath.IsAbs(hooksDir) {
		if fileContainsAgentlink(filepath.Join(hooksDir, "post-checkout")) || fileContainsAgentlink(filepath.Join(hooksDir, "post-merge")) {
			count++
		}
	}
	if fileContainsAgentlink(zshrcPath()) {
		count++
	}
	if _, err := os.Lstat(launchdPlistPath()); err == nil {
		count++
	}
	return count
}

// --- Git Hooks ---

const gitHookMarkerStart = "# >>> agentlink >>>"
const gitHookMarkerEnd = "# <<< agentlink <<<"

// legacyHookStub is what remains of a hook file that an earlier release
// created, once agentlink's block is removed.
const legacyHookStub = "#!/bin/sh\n"

var gitHookNames = []string{"post-checkout", "post-merge"}

func removeGitHooks() error {
	hooksDir := gitGlobalHooksDir()
	if hooksDir == "" {
		printInfo("No global git hooks directory configured")
		return nil
	}
	if !filepath.IsAbs(hooksDir) {
		return fmt.Errorf("global git core.hooksPath is relative (%s); refusing to remove hooks from an ambiguous path. Set core.hooksPath to the absolute hooks directory before removing agentlink hooks", hooksDir)
	}
	return removeGitHooksFromDir(hooksDir, defaultGitHooksDir(), unsetGlobalHooksPath)
}

// removeGitHooksFromDir strips agentlink's block from each hook, deletes hook
// files that held nothing else, and unsets core.hooksPath when it points at
// agentlink's default directory and nothing else is left there.
func removeGitHooksFromDir(hooksDir, defaultDir string, unsetHooksPath func() error) error {
	deleted := map[string]bool{}
	for _, hookName := range gitHookNames {
		hookPath := filepath.Join(hooksDir, hookName)
		removed, stubOnly, err := removeGitHookBlock(hookPath)
		if err != nil {
			return fmt.Errorf("failed to update %s: %w", hookName, err)
		}
		switch {
		case stubOnly && dryRun:
			printInfo("Would delete %s (it contains only agentlink's block)", hookPath)
			deleted[hookName] = true
		case stubOnly:
			printOK("Deleted %s (it contained only agentlink's block)", hookPath)
			deleted[hookName] = true
		case removed:
			printOK("Removed agentlink from %s", hookName)
		}
	}

	if filepath.Clean(hooksDir) != filepath.Clean(defaultDir) {
		printWarning("Global core.hooksPath is still %s, so Git ignores each repository's own .git/hooks. Agentlink did not create that directory and leaves it set; run 'git config --global --unset core.hooksPath' if nothing else relies on it.", hooksDir)
		return nil
	}

	remaining, err := remainingEntries(hooksDir, deleted)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		printWarning("Global core.hooksPath is still %s, which holds other hooks (%s), so Git ignores each repository's own .git/hooks. Unset it with 'git config --global --unset core.hooksPath' once those hooks are moved.", hooksDir, strings.Join(remaining, ", "))
		return nil
	}

	if dryRun {
		printInfo("Would remove empty %s and unset global core.hooksPath", hooksDir)
		return nil
	}
	if err := os.Remove(hooksDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove empty hooks directory %s: %w", hooksDir, err)
	}
	if err := unsetHooksPath(); err != nil {
		return fmt.Errorf("failed to unset core.hooksPath: %w", err)
	}
	printOK("Unset global core.hooksPath; repository .git/hooks run again")
	return nil
}

// removeGitHookBlock removes agentlink's block from a hook. When the hook then
// holds only the shebang an earlier release wrote, it deletes the file and
// reports stubOnly instead.
func removeGitHookBlock(path string) (removed, stubOnly bool, err error) {
	data, info, err := readManagedFile(path)
	if err != nil || data == nil {
		return false, false, err
	}
	stripped, found, err := stripMarkedSection(string(data))
	if err != nil || !found {
		return false, false, err
	}
	if stripped == legacyHookStub || strings.TrimSpace(stripped) == "" {
		if !dryRun {
			if err := os.Remove(path); err != nil {
				return false, false, err
			}
		}
		return true, true, nil
	}
	if dryRun {
		printInfo("Would update managed section in %s", path)
		return true, false, nil
	}
	return true, false, fsutil.AtomicWriteFile(path, []byte(stripped), info.Mode().Perm())
}

func remainingEntries(dir string, excluded map[string]bool) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !excluded[entry.Name()] {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

func defaultGitHooksDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "git", "hooks")
}

func unsetGlobalHooksPath() error {
	err := exec.Command("git", "config", "--global", "--unset", "core.hooksPath").Run()
	var exitErr *exec.ExitError
	// Exit status 5 means the key was already unset.
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 5 {
		return nil
	}
	return err
}

func gitGlobalHooksDir() string {
	out, err := exec.Command("git", "config", "--global", "core.hooksPath").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// --- Zsh Hook ---

func zshrcPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".zshrc")
}

func removeZshHook() error {
	removed, err := removeMarkedSection(zshrcPath())
	if err != nil {
		return err
	}
	if removed {
		printOK("Removed agentlink hook from ~/.zshrc")
	} else {
		printInfo("No agentlink hook found in ~/.zshrc")
	}
	return nil
}

// --- Launchd ---

const launchdLabel = "com.agentlink.sync"

// Earlier releases logged here. The files sit in the shared /tmp, so removal
// deletes them only when they are regular files the current user can remove.
var legacyLaunchdLogs = []string{"/tmp/agentlink-sync.log", "/tmp/agentlink-sync.err"}

func launchdPlistPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Library", "LaunchAgents", launchdLabel+".plist")
}

func removeLaunchdAgent() error {
	plistPath := launchdPlistPath()

	if _, err := os.Lstat(plistPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("cannot inspect LaunchAgent plist: %w", err)
		}
		printInfo("No LaunchAgent installed")
	} else if dryRun {
		printInfo("Would unload LaunchAgent and remove %s", plistPath)
	} else {
		if err := exec.Command("launchctl", "unload", plistPath).Run(); err != nil {
			printWarning("launchctl unload failed: %v", err)
		}
		if err := os.Remove(plistPath); err != nil {
			return fmt.Errorf("cannot remove LaunchAgent plist: %w", err)
		}
		printOK("Removed LaunchAgent")
	}

	removeLegacyLaunchdLogs(legacyLaunchdLogs)
	return nil
}

func removeLegacyLaunchdLogs(paths []string) {
	for _, path := range paths {
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		if dryRun {
			printInfo("Would remove legacy LaunchAgent log %s", path)
			continue
		}
		if err := os.Remove(path); err != nil {
			printWarning("Left legacy LaunchAgent log %s: %v", path, err)
			continue
		}
		printOK("Removed legacy LaunchAgent log %s", path)
	}
}

// --- Helpers ---

func fileContainsAgentlink(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), gitHookMarkerStart)
}

// readManagedFile returns nil data and no error for a missing file.
func readManagedFile(path string) ([]byte, os.FileInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	return data, info, nil
}

// removeMarkedSection deletes agentlink's marked block from path, replacing the
// file atomically and keeping its permissions.
func removeMarkedSection(path string) (bool, error) {
	data, info, err := readManagedFile(path)
	if err != nil || data == nil {
		return false, err
	}
	stripped, found, err := stripMarkedSection(string(data))
	if err != nil || !found {
		return false, err
	}
	if dryRun {
		printInfo("Would update managed section in %s", path)
		return true, nil
	}
	if err := fsutil.AtomicWriteFile(path, []byte(stripped), info.Mode().Perm()); err != nil {
		return false, err
	}
	return true, nil
}

// stripMarkedSection removes the first marked block, including the end
// marker's trailing newline.
func stripMarkedSection(content string) (string, bool, error) {
	startIdx := strings.Index(content, gitHookMarkerStart)
	if startIdx == -1 {
		return content, false, nil
	}
	endOffset := strings.Index(content[startIdx:], gitHookMarkerEnd)
	if endOffset == -1 {
		return content, false, fmt.Errorf("found start marker without end marker")
	}
	endIdx := startIdx + endOffset + len(gitHookMarkerEnd)
	if endIdx < len(content) && content[endIdx] == '\n' {
		endIdx++
	}
	return content[:startIdx] + content[endIdx:], true, nil
}
