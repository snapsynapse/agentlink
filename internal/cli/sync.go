package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/snapsynapse/agentlink/internal/config"
	"github.com/snapsynapse/agentlink/internal/symlink"
	"github.com/spf13/cobra"
)

var (
	syncBackup bool
	backupNow  = time.Now
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Create/fix symlinks based on configuration",
	Long: `Create or fix symlinks to keep instruction files in sync.

Reads .agentlink.yaml in current directory, or falls back to global config
at ~/.config/agentlink/config.yaml. Creates or fixes symlinks so they point
to the configured source file. Use --global to select global scope explicitly.
An existing but unreadable project config is an error, never a fallback.

When a target path already contains a real file (not a symlink), sync will
refuse to overwrite it. You have three options:

  --backup    Back up existing regular files to <name>.bak before replacing
  --force     Replace conflicting regular files or symlinks without backup
  --dry-run   Preview what would happen without creating, removing, or backing up files

Broken or misdirected symlinks require --force; --backup does not authorize them.
Directories and special files are never replaced recursively. Without any of
these flags, sync reports the conflict and moves on.`,
	RunE: runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&syncBackup, "backup", false, "back up existing regular files before replacing")
	syncCmd.Flags().BoolVar(&useGlobalConfig, "global", false, "use global configuration even inside a configured project")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	// Find config file
	configPath, isProject, err := selectConfigPath()
	if err != nil {
		return err
	}

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

	// Backup permits regular-file replacement only, never unknown symlink ownership.
	manager := symlink.NewManager(dryRun, force)

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
	if _, err := os.Lstat(configPath); err == nil {
		return config.LoadConfig(configPath)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect config %s: %w", configPath, err)
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
	if info, err := os.Lstat(linkPath); syncBackup && err == nil && info.Mode().IsRegular() {
		manager = symlink.NewManager(dryRun, true)
	}
	if verbose {
		printInfo("Processing link: %s", linkPath)
	}

	// Validate before backing up, removing empty files, or suggesting a force
	// override. Preserve the safety error instead of masking its cause.
	if _, err := manager.PlanLink(linkPath, sourcePath); err != nil {
		return err
	}
	if syncBackup {
		info, err := os.Lstat(linkPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil && info.Mode().IsRegular() {
			if dryRun {
				if info.Size() == 0 {
					printInfo("Would remove empty file %s (no backup)", linkPath)
				} else {
					path, err := nextBackupPath(linkPath)
					if err != nil {
						return err
					}
					printInfo("Would back up %s -> %s before linking", linkPath, path)
				}
			} else if info.Size() == 0 {
				printWarning("%s exists but is empty, skipping backup", linkPath)
				if err := os.Remove(linkPath); err != nil {
					return err
				}
			} else if err := backupFile(linkPath); err != nil {
				return fmt.Errorf("backup failed: %w", err)
			}
		}
	}
	action, err := manager.FixLink(linkPath, sourcePath)
	if err != nil {
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

func nextBackupPath(path string) (string, error) {
	timestamped := fmt.Sprintf("%s.%s.bak", path, backupNow().Format("20060102-150405"))
	candidate := path + ".bak"
	for suffix := 0; ; suffix++ {
		if _, err := os.Lstat(candidate); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil {
			return "", err
		}
		candidate = timestamped
		if suffix > 0 {
			candidate = fmt.Sprintf("%s.%d", timestamped, suffix)
		}
	}
}

func backupFile(path string) error {
	for {
		bakPath, err := nextBackupPath(path)
		if err != nil {
			return err
		}
		// Exclusive hard-link creation cannot clobber a concurrent backup.
		if err := os.Link(path, bakPath); err != nil {
			if os.IsExist(err) {
				continue
			}
			return fmt.Errorf("failed to create no-clobber backup %s (original left unchanged): %w", bakPath, err)
		}
		if err := os.Remove(path); err != nil {
			_ = os.Remove(bakPath)
			return fmt.Errorf("failed to remove %s after backing up to %s: %w", path, bakPath, err)
		}
		printInfo("Backed up %s -> %s", path, bakPath)
		return nil
	}
}
