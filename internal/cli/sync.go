package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/martinmose/agentlink/internal/config"
	"github.com/martinmose/agentlink/internal/symlink"
	"github.com/spf13/cobra"
)

var (
	syncBackup bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Create/fix symlinks based on configuration",
	Long: `Create or fix symlinks to keep instruction files in sync.

Reads .agentlink.yaml in current directory, or falls back to global config
at ~/.config/agentlink/config.yaml. Creates or fixes symlinks so they point
to the configured source file.

When a target path already contains a real file (not a symlink), sync will
refuse to overwrite it. You have three options:

  --backup    Back up existing files to <name>.bak before replacing
  --force     Replace existing files without backup
  --dry-run   Preview what would happen without making changes

Without any of these flags, sync reports the conflict and moves on.`,
	RunE: runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&syncBackup, "backup", false, "back up existing files before replacing")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	// Find config file
	configPath, isProject := config.FindConfigPath()

	// Load or create config
	cfg, err := loadOrCreateConfig(configPath, isProject)
	if err != nil {
		return err
	}

	if verbose {
		if isProject {
			printInfo("Using project config: %s", configPath)
		} else {
			printInfo("Using global config: %s", configPath)
		}
	}

	// If --backup is set, enable force (backup happens before replacement)
	effectiveForce := force || syncBackup
	manager := symlink.NewManager(dryRun, effectiveForce, verbose)

	// Validate source file
	if err := manager.ValidateSource(cfg.Source); err != nil {
		printError("Source validation failed: %v", err)
		return err
	}

	printOK("Source: %s", cfg.Source)

	// Process each link
	hasErrors := false
	for _, linkPath := range cfg.Links {
		if err := processLink(manager, linkPath, cfg.Source); err != nil {
			printError("Failed to process %s: %v", linkPath, err)
			hasErrors = true
		}
	}

	if hasErrors {
		return fmt.Errorf("sync completed with errors")
	}

	if dryRun {
		printInfo("Dry run completed - no changes made")
	}

	return nil
}

func loadOrCreateConfig(configPath string, isProject bool) (*config.Config, error) {
	// Try to load existing config
	if _, err := os.Stat(configPath); err == nil {
		return config.LoadConfig(configPath)
	}

	// If it's a project config and doesn't exist, error
	if isProject {
		printError("No .agentlink.yaml found in current directory")
		printInfo("Run 'agentlink init' to create one")
		return nil, fmt.Errorf("no project config found")
	}

	// Create default global config
	printInfo("Creating default global config at %s", configPath)
	if !dryRun {
		if err := config.CreateDefaultGlobalConfig(configPath); err != nil {
			printError("Failed to create default config: %v", err)
			return nil, err
		}
	}

	printWarning("Please edit %s to configure your source and links", configPath)
	return nil, fmt.Errorf("created default config - please edit it first")
}

func processLink(manager *symlink.Manager, linkPath, sourcePath string) error {
	if verbose {
		printInfo("Processing link: %s", linkPath)
	}

	// If backup mode, handle existing non-symlink files before FixLink
	if syncBackup && !dryRun {
		if info, err := os.Lstat(linkPath); err == nil {
			if info.Mode()&os.ModeSymlink == 0 && info.Mode().IsRegular() {
				if info.Size() == 0 {
					printWarning("%s exists but is empty, skipping backup", linkPath)
					// Still need to remove it so FixLink can create the symlink
					os.Remove(linkPath)
				} else {
					if err := backupFile(linkPath); err != nil {
						return fmt.Errorf("backup failed: %w", err)
					}
				}
			}
		}
	}

	action, err := manager.FixLink(linkPath, sourcePath)
	if err != nil {
		// Enhance error messages with actionable guidance
		if info, statErr := os.Lstat(linkPath); statErr == nil {
			if info.Mode()&os.ModeSymlink == 0 && info.Mode().IsRegular() {
				size := info.Size()
				if size == 0 {
					return fmt.Errorf("%s exists but is empty (safe to --force)", linkPath)
				}
				return fmt.Errorf(
					"%s exists (%d bytes, modified %s). Options:\n"+
						"         cat %s          # inspect the file\n"+
						"         agentlink sync --backup   # back up to %s.bak then replace\n"+
						"         agentlink sync --force    # replace without backup",
					linkPath, size, info.ModTime().Format("2006-01-02"),
					linkPath, filepath.Base(linkPath),
				)
			}
		}
		return err
	}

	switch action {
	case "skip":
		if verbose {
			printSkip("%s already links to %s", linkPath, sourcePath)
		}
	case "create":
		printCreate("%s -> %s", linkPath, sourcePath)
	case "fix":
		printOK("Fixed %s -> %s", linkPath, sourcePath)
	case "replace":
		printOK("Replaced %s -> %s", linkPath, sourcePath)
	case "fix broken":
		printOK("Fixed broken %s -> %s", linkPath, sourcePath)
	}

	return nil
}

func backupFile(path string) error {
	bakPath := path + ".bak"

	// If .bak already exists, use timestamped name
	if _, err := os.Stat(bakPath); err == nil {
		ts := time.Now().Format("20060102-150405")
		bakPath = fmt.Sprintf("%s.%s.bak", path, ts)
	}

	if err := os.Rename(path, bakPath); err != nil {
		return fmt.Errorf("failed to back up %s to %s: %w", path, bakPath, err)
	}

	printInfo("Backed up %s -> %s", path, bakPath)
	return nil
}
