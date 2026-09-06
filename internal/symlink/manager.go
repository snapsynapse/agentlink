package symlink

import (
	"fmt"
	"os"
	"path/filepath"
)

// LinkStatus represents the status of a symlink
type LinkStatus int

const (
	StatusOK LinkStatus = iota
	StatusMissing
	StatusWrongTarget
	StatusNotSymlink
	StatusBroken
)

// LinkInfo contains information about a symlink
type LinkInfo struct {
	Target string
	Status LinkStatus
	Error  error
}

// Manager handles symlink operations
type Manager struct {
	dryRun bool
	force  bool
}

// NewManager creates a new symlink manager
func NewManager(dryRun, force bool) *Manager {
	return &Manager{
		dryRun: dryRun,
		force:  force,
	}
}

// ValidateSource checks if the source file exists and is a regular file
func (m *Manager) ValidateSource(sourcePath string) error {
	info, err := os.Lstat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source file %s does not exist", sourcePath)
		}
		return fmt.Errorf("failed to stat source file %s: %w", sourcePath, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		if !m.force {
			return fmt.Errorf("source file %s is a symlink (use --force to override)", sourcePath)
		}
		info, err = os.Stat(sourcePath)
		if err != nil {
			return fmt.Errorf("resolve source %s: %w", sourcePath, err)
		}
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("source file %s is not a regular file", sourcePath)
	}

	return nil
}

// CheckLink checks the status of a symlink
func (m *Manager) CheckLink(linkPath, expectedTarget string) *LinkInfo {
	info := &LinkInfo{}

	// Check if link exists
	linkInfo, err := os.Lstat(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			info.Status = StatusMissing
			return info
		}
		info.Error = err
		info.Status = StatusBroken
		return info
	}

	// Check if it's a symlink
	if linkInfo.Mode()&os.ModeSymlink == 0 {
		info.Status = StatusNotSymlink
		return info
	}

	// Get the target
	target, err := os.Readlink(linkPath)
	if err != nil {
		info.Error = err
		info.Status = StatusBroken
		return info
	}

	info.Target = target

	// Resolve the link itself through the filesystem. Cleaning a lexical path
	// first changes the meaning of ".." beneath a symlinked directory.
	actual, err := filepath.EvalSymlinks(linkPath)
	if err != nil {
		info.Status = StatusBroken
		if !os.IsNotExist(err) {
			info.Error = err
		}
		return info
	}
	expected, err := filepath.EvalSymlinks(expectedTarget)
	if err != nil {
		info.Status = StatusBroken
		info.Error = err
		return info
	}
	actual, err = filepath.Abs(actual)
	if err != nil {
		info.Status = StatusBroken
		info.Error = err
		return info
	}
	expected, err = filepath.Abs(expected)
	if err != nil {
		info.Status = StatusBroken
		info.Error = err
		return info
	}
	if actual == expected {
		info.Status = StatusOK
	} else {
		info.Status = StatusWrongTarget
	}

	return info
}

// CreateLink creates or fixes a symlink
func (m *Manager) CreateLink(linkPath, targetPath string) error {
	if m.dryRun {
		return nil // Don't actually create in dry-run mode
	}

	// Ensure the parent directory exists
	if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", linkPath, err)
	}

	// Calculate relative path from link to target
	link, err := resolveParentSymlinks(linkPath)
	if err != nil {
		return fmt.Errorf("resolve link parent: %w", err)
	}
	target, err := resolveParentSymlinks(targetPath)
	if err != nil {
		return fmt.Errorf("resolve source parent: %w", err)
	}
	relTarget, err := filepath.Rel(filepath.Dir(link), target)
	if err != nil {
		return fmt.Errorf("failed to calculate relative path: %w", err)
	}

	// Create the symlink
	if err := os.Symlink(relTarget, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", linkPath, relTarget, err)
	}

	return nil
}

// RemoveLink removes a symlink if it's managed by agentlink
func (m *Manager) RemoveLink(linkPath, expectedTarget string) error {
	if m.dryRun {
		return nil
	}

	info := m.CheckLink(linkPath, expectedTarget)
	if info.Status == StatusOK {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("failed to remove symlink %s: %w", linkPath, err)
		}
	}

	return nil
}

// PlanLink validates a destination without mutating it. Sync uses the same
// preflight before any backup, so backup cannot bypass source protection.
func (m *Manager) PlanLink(linkPath, targetPath string) (string, error) {
	samePath, err := sameEffectivePath(linkPath, targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to validate link path %s: %w", linkPath, err)
	}
	if samePath {
		return "", fmt.Errorf("link path %s resolves to source %s; refusing to replace source", linkPath, targetPath)
	}
	info := m.CheckLink(linkPath, targetPath)
	if info.Error != nil {
		return "", fmt.Errorf("cannot inspect %s: %w", linkPath, info.Error)
	}
	switch info.Status {
	case StatusOK:
		return "skip", nil
	case StatusMissing:
		return "create", nil
	case StatusWrongTarget:
		if !m.force {
			return "", fmt.Errorf("symlink %s points to wrong target %s (expected %s), use --force to fix", linkPath, info.Target, targetPath)
		}
		return "fix", nil
	case StatusNotSymlink:
		entry, err := os.Lstat(linkPath)
		if err != nil {
			return "", fmt.Errorf("cannot inspect %s: %w", linkPath, err)
		}
		if !entry.Mode().IsRegular() {
			return "", fmt.Errorf("%s is not a regular file; refusing to replace directories or special files", linkPath)
		}
		if !m.force {
			return "", fmt.Errorf("file %s exists and is not a symlink; inspect it, then use --backup or --force to replace", linkPath)
		}
		return "replace", nil
	case StatusBroken:
		if !m.force {
			return "", fmt.Errorf("broken symlink %s points to %s; ownership cannot be verified, use --force to replace", linkPath, info.Target)
		}
		return "fix broken", nil
	default:
		return "", fmt.Errorf("unknown link status for %s", linkPath)
	}
}

// FixLink applies the validated operation. Revalidate immediately before each
// mutation, including after a caller has backed up an existing regular file.
func (m *Manager) FixLink(linkPath, targetPath string) (string, error) {
	action, err := m.PlanLink(linkPath, targetPath)
	if err != nil {
		return "", err
	}
	if m.dryRun || action == "skip" {
		return action, nil
	}
	if action != "create" {
		if err := os.Remove(linkPath); err != nil {
			return "", fmt.Errorf("failed to remove %s: %w", linkPath, err)
		}
	}
	if err := m.CreateLink(linkPath, targetPath); err != nil {
		return "", err
	}
	return action, nil
}

// sameEffectivePath compares paths after resolving symlinks in their parent
// directories. The leaf is deliberately left unresolved so an existing,
// correct managed symlink is not mistaken for an alias of its source.
func sameEffectivePath(linkPath, targetPath string) (bool, error) {
	link, err := resolveParentSymlinks(linkPath)
	if err != nil {
		return false, err
	}
	target, err := resolveParentSymlinks(targetPath)
	if err != nil {
		return false, err
	}
	if link == target {
		return true, nil
	}
	// A forced source symlink must not allow replacing its real backing file.
	resolved, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		return false, err
	}
	resolved, err = filepath.Abs(resolved)
	return link == resolved, err
}

// resolveParentSymlinks also supports nonexistent parent directories by
// resolving the nearest existing ancestor, then restoring the missing suffix.
func resolveParentSymlinks(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	parent := filepath.Dir(filepath.Clean(abs))
	suffix := []string{filepath.Base(abs)}
	for {
		resolved, err := filepath.EvalSymlinks(parent)
		if err == nil {
			parts := append([]string{resolved}, suffix...)
			return filepath.Clean(filepath.Join(parts...)), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}

		next := filepath.Dir(parent)
		if next == parent {
			return "", err
		}
		suffix = append([]string{filepath.Base(parent)}, suffix...)
		parent = next
	}
}
