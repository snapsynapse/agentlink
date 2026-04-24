package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/martinmose/agentlink/internal/registry"
	"github.com/martinmose/agentlink/internal/symlink"
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

The scan directory defaults to ~/Git. Override with the --dir flag or by
passing a directory argument. Set a permanent default at build time with
-ldflags "-X github.com/martinmose/agentlink/internal/cli.DefaultScanDir=/your/path".

Only creates symlinks in repos that already have an AGENTS.md file.
Does not inject files into repos that lack one.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

var scanDir string

func init() {
	scanCmd.Flags().StringVar(&scanDir, "dir", "", "directory to scan (default: ~/Git)")
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

	// Find git repos
	repos := findGitRepos(dir)
	if len(repos) == 0 {
		printInfo("No git repositories found in %s", dir)
		return nil
	}

	fmt.Printf("Found %d git repositories\n\n", len(repos))

	manager := symlink.NewManager(dryRun, force, verbose)
	created := 0
	skipped := 0
	errors := 0

	for _, repo := range repos {
		agentsPath := filepath.Join(repo, "AGENTS.md")

		// Only act on repos that already have AGENTS.md
		if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
			if verbose {
				printSkip("%s (no AGENTS.md)", relativeTo(repo, dir))
			}
			skipped++
			continue
		}

		if verbose {
			printInfo("Processing %s", relativeTo(repo, dir))
		}

		for _, target := range linkTargets {
			linkPath := filepath.Join(repo, target)
			action, err := manager.FixLink(linkPath, agentsPath)
			if err != nil {
				printError("%s/%s: %v", relativeTo(repo, dir), target, err)
				errors++
				continue
			}

			switch action {
			case "skip":
				if verbose {
					printSkip("%s/%s already linked", relativeTo(repo, dir), target)
				}
			case "create":
				printCreate("%s/%s -> AGENTS.md", relativeTo(repo, dir), target)
				created++
			case "fix", "replace", "fix broken":
				printOK("Fixed %s/%s -> AGENTS.md", relativeTo(repo, dir), target)
				created++
			}
		}
	}

	fmt.Printf("\nScan complete: %d links created/fixed, %d repos skipped, %d errors\n", created, skipped, errors)

	if dryRun {
		printInfo("Dry run - no changes made")
	}

	return nil
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
