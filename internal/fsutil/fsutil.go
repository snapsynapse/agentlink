// Package fsutil holds the file-writing primitives agentlink uses when it
// edits files it does not own, such as ~/.zshrc, Git hooks and config files.
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWriteFile replaces the contents of path without ever leaving it
// truncated: it writes a temporary file in the same directory, syncs it, and
// renames it over the original. When path is a symlink, as dotfile managers
// often make ~/.zshrc, the link is resolved first so the link survives and
// its target is the file replaced.
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	target := path
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		target = resolved
	} else if !os.IsNotExist(err) {
		return err
	}
	return replace(target, data, perm)
}

// WriteNewFile creates path with data and never writes through a symlink. A
// symlink at path, dangling or not, is an error. An existing regular file is
// an error unless overwrite is set, in which case it is replaced atomically.
func WriteNewFile(path string, data []byte, perm os.FileMode, overwrite bool) error {
	info, err := os.Lstat(path)
	switch {
	case err == nil && info.Mode()&os.ModeSymlink != 0:
		return fmt.Errorf("%s is a symlink; refusing to write through it", path)
	case err == nil && !info.Mode().IsRegular():
		return fmt.Errorf("%s is not a regular file", path)
	case err == nil && !overwrite:
		return fmt.Errorf("%s: %w", path, os.ErrExist)
	case err == nil:
		return replace(path, data, perm)
	case !os.IsNotExist(err):
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return err
	}
	return nil
}

func replace(target string, data []byte, perm os.FileMode) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(target), "."+filepath.Base(target)+".agentlink-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()
	if _, err = tmp.Write(data); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	if err = tmp.Chmod(perm); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, target)
}
