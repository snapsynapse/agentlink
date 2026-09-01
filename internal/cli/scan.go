package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snapsynapse/agentlink/internal/config"
	"github.com/snapsynapse/agentlink/internal/registry"
	"github.com/snapsynapse/agentlink/internal/symlink"
	"github.com/spf13/cobra"
)

// DefaultScanDir is the default directory to scan for git repositories.
// Override at build time with -ldflags or via the --dir flag.
var DefaultScanDir = "~/Git"

var scanCmd = &cobra.Command{
	Use:   "scan [directory]",
	Short: "Scan git repos and manage CLAUDE.md symlinks",
	Long: `Walk a directory tree, find git repositories, and ensure each repo that
contains an AGENTS.md also has the appropriate symlinks for tools that use
different filenames (e.g., CLAUDE.md, GEMINI.md).

If a repository contains .agentlink.yaml, that explicit source and link list
is authoritative. Otherwise scan preserves existing real alias files as
unmanaged wrappers. Use --nested to also manage aliases beside nested AGENTS.md
files in unconfigured repositories.

The scan directory defaults to ~/Git. Override with the --dir flag or by
passing a directory argument. Set a permanent default at build time with
-ldflags "-X github.com/snapsynapse/agentlink/internal/cli.DefaultScanDir=/your/path".

Unconfigured repositories require an existing AGENTS.md. Configured
repositories use the source declared in .agentlink.yaml. Scan never injects a
source instruction file.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

var scanDir string
var scanNested bool

func init() {
	scanCmd.Flags().StringVar(&scanDir, "dir", "", "directory to scan (default: ~/Git)")
	scanCmd.Flags().BoolVar(&scanNested, "nested", false, "also manage documented nested aliases beside AGENTS.md files in unconfigured repos")
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	// Determine scan directory: arg > flag > default
	dir := DefaultScanDir
	if scanDir != "" {
		dir = scanDir
	}
	if len(args) > 0 {
		dir = args[0]
	}

	// Expand ~ in path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	if strings.HasPrefix(dir, "~/") {
		dir = filepath.Join(homeDir, dir[2:])
	}

	// Verify directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("scan directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	fmt.Printf("Scanning %s for git repositories...\n\n", dir)

	// Build the set of repo-level filenames that tools expect,
	// excluding AGENTS.md itself (that's the source).
	linkTargets := repoLinkTargets()
	nestedLinkTargets := nestedRepoLinkTargets()

	// Find git repos
	repos := findGitRepos(dir)
	if len(repos) == 0 {
		printInfo("No git repositories found in %s", dir)
		return nil
	}

	fmt.Printf("Found %d git repositories\n\n", len(repos))

	manager := symlink.NewManager(dryRun, force)
	stats := scanStats{}

	for _, repo := range repos {
		configPath := filepath.Join(repo, ".agentlink.yaml")
		if _, err := os.Stat(configPath); err == nil {
			cfg, loadErr := config.LoadConfig(configPath)
			if loadErr != nil {
				printError("%s/.agentlink.yaml: %v", relativeTo(repo, dir), loadErr)
				stats.errors++
				continue
			}
			if verbose {
				printInfo("Processing %s using explicit .agentlink.yaml", relativeTo(repo, dir))
			}
			processScanSource(manager, dir, cfg.Source, cfg.Links, false, &stats)
			continue
		} else if !os.IsNotExist(err) {
			printError("%s/.agentlink.yaml: %v", relativeTo(repo, dir), err)
			stats.errors++
			continue
		}

		agentsPath := filepath.Join(repo, "AGENTS.md")
		if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
			if verbose {
				printSkip("%s (no AGENTS.md)", relativeTo(repo, dir))
			}
			stats.skippedRepos++
			continue
		}

		if verbose {
			printInfo("Processing %s", relativeTo(repo, dir))
		}
		links := joinLinkTargets(filepath.Dir(agentsPath), linkTargets)
		processScanSource(manager, dir, agentsPath, links, true, &stats)

		if scanNested {
			for _, nestedSource := range findNestedAgentsFiles(repo) {
				nestedLinks := joinLinkTargets(filepath.Dir(nestedSource), nestedLinkTargets)
				processScanSource(manager, dir, nestedSource, nestedLinks, true, &stats)
			}
		}
	}

	fmt.Printf("\nScan complete: %d links created/fixed, %d existing files preserved, %d repos skipped, %d errors\n", stats.created, stats.preserved, stats.skippedRepos, stats.errors)

	if dryRun {
		printInfo("Dry run - no changes made")
	}

	if stats.errors > 0 {
		return fmt.Errorf("scan completed with %d link error(s)", stats.errors)
	}

	return nil
}

type scanStats struct {
	created      int
	preserved    int
	skippedRepos int
	errors       int
}

func processScanSource(manager *symlink.Manager, scanRoot, source string, links []string, preserveRegular bool, stats *scanStats) {
	if err := manager.ValidateSource(source); err != nil {
		printError("%s: %v", relativeTo(source, scanRoot), err)
		stats.errors++
		return
	}

	for _, linkPath := range links {
		if preserveRegular && !force {
			if info, err := os.Lstat(linkPath); err == nil && info.Mode().IsRegular() {
				printSkip("%s (existing regular file preserved; unmanaged)", relativeTo(linkPath, scanRoot))
				stats.preserved++
				continue
			}
		}

		action, err := manager.FixLink(linkPath, source)
		if err != nil {
			printError("%s: %v", relativeTo(linkPath, scanRoot), err)
			stats.errors++
			continue
		}
		switch action {
		case "skip":
			if verbose {
				printSkip("%s already linked", relativeTo(linkPath, scanRoot))
			}
		case "create":
			printCreate("%s -> %s", relativeTo(linkPath, scanRoot), filepath.Base(source))
			stats.created++
		case "fix", "replace", "fix broken":
			printOK("Fixed %s -> %s", relativeTo(linkPath, scanRoot), filepath.Base(source))
			stats.created++
		}
	}
}

func joinLinkTargets(dir string, targets []string) []string {
	links := make([]string, 0, len(targets))
	for _, target := range targets {
		links = append(links, filepath.Join(dir, target))
	}
	return links
}

func findNestedAgentsFiles(repo string) []string {
	var sources []string
	_ = filepath.Walk(repo, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if path == repo {
				return nil
			}
			if shouldSkipNestedDir(info.Name()) {
				return filepath.SkipDir
			}
			if _, err := os.Lstat(filepath.Join(path, ".git")); err == nil {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Name() == "AGENTS.md" && filepath.Dir(path) != repo {
			sources = append(sources, path)
		}
		return nil
	})
	return sources
}

func shouldSkipNestedDir(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "node_modules", "vendor", "dist", "build", "handoffs", "working":
		return true
	}
	return false
}

// repoLinkTargets returns the set of filenames (other than AGENTS.md) that
// known tools expect at the repo root. These become symlink targets.
func repoLinkTargets() []string {
	seen := map[string]bool{"AGENTS.md": true}
	var targets []string

	for _, tool := range registry.All() {
		name := tool.RepoFileName
		if name == "" || name == "AGENTS.md" {
			continue
		}
		// Skip paths with directories (e.g., .github/copilot-instructions.md,
		// .junie/AGENTS.md) -- those require different handling and the user
		// should configure them explicitly via .agentlink.yaml.
		if strings.Contains(name, "/") {
			continue
		}
		if !seen[name] {
			seen[name] = true
			targets = append(targets, name)
		}
	}

	return targets
}

// nestedRepoLinkTargets returns only aliases whose tools are documented to
// discover instruction files below the repository root. Unknown behavior is
// excluded rather than treated as equivalent across harnesses.
func nestedRepoLinkTargets() []string {
	seen := map[string]bool{"AGENTS.md": true}
	var targets []string
	for _, tool := range registry.All() {
		name := tool.RepoFileName
		if !tool.SupportsNestedRepoFile || name == "" || seen[name] {
			continue
		}
		seen[name] = true
		targets = append(targets, name)
	}
	return targets
}

// findGitRepos walks a directory tree and returns paths to directories
// containing a .git folder. Does not recurse into .git directories or
// into nested git repos (stops at the first .git found in each subtree).
func findGitRepos(root string) []string {
	var repos []string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible directories
		}

		// Skip hidden directories (except the root itself)
		if info.IsDir() && path != root {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
		}

		// Check for .git directory
		if info.IsDir() {
			gitDir := filepath.Join(path, ".git")
			if fi, err := os.Lstat(gitDir); err == nil && (fi.IsDir() || fi.Mode().IsRegular() || fi.Mode()&os.ModeSymlink != 0) {
				repos = append(repos, path)
				return filepath.SkipDir // don't recurse into this repo
			}
		}

		return nil
	})

	return repos
}

// relativeTo returns path relative to base, or the original path on error.
func relativeTo(path, base string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return rel
}
