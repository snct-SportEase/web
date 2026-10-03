package repository

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const missingUploadDirectorySuffix = ".missing"

func (r *eventTestRunRepository) hasUploadSnapshot() bool {
	return r.uploadSnapshotRoot != "" && len(r.uploadDirectories) > 0
}

func (r *eventTestRunRepository) createUploadSnapshot() error {
	if err := validateUploadSnapshotPaths(r.uploadSnapshotRoot, r.uploadDirectories); err != nil {
		return err
	}

	temporaryRoot := r.uploadSnapshotRoot + ".tmp"
	if err := removeManagedPath(temporaryRoot); err != nil {
		return err
	}
	if err := os.MkdirAll(temporaryRoot, 0o700); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = removeManagedPath(temporaryRoot)
		}
	}()

	for _, directory := range r.uploadDirectories {
		name := filepath.Base(filepath.Clean(directory))
		info, err := os.Lstat(directory)
		if os.IsNotExist(err) {
			marker := filepath.Join(temporaryRoot, name+missingUploadDirectorySuffix)
			if err := os.WriteFile(marker, nil, 0o600); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("upload path %q is not a directory", directory)
		}
		if err := copyDirectory(directory, filepath.Join(temporaryRoot, name)); err != nil {
			return fmt.Errorf("copy %q: %w", directory, err)
		}
	}

	if err := removeManagedPath(r.uploadSnapshotRoot); err != nil {
		return err
	}
	if err := os.Rename(temporaryRoot, r.uploadSnapshotRoot); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *eventTestRunRepository) restoreUploadSnapshot() error {
	if err := validateUploadSnapshotPaths(r.uploadSnapshotRoot, r.uploadDirectories); err != nil {
		return err
	}
	info, err := os.Lstat(r.uploadSnapshotRoot)
	if os.IsNotExist(err) {
		return errorsJoinSnapshotNotFound(r.uploadSnapshotRoot)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("upload snapshot path %q is not a directory", r.uploadSnapshotRoot)
	}

	for _, directory := range r.uploadDirectories {
		name := filepath.Base(filepath.Clean(directory))
		marker := filepath.Join(r.uploadSnapshotRoot, name+missingUploadDirectorySuffix)
		_, markerErr := os.Lstat(marker)
		if markerErr == nil {
			if err := removeManagedPath(directory); err != nil {
				return err
			}
			continue
		}
		if !os.IsNotExist(markerErr) {
			return markerErr
		}

		source := filepath.Join(r.uploadSnapshotRoot, name)
		sourceInfo, err := os.Lstat(source)
		if os.IsNotExist(err) {
			return errorsJoinSnapshotNotFound(source)
		}
		if err != nil {
			return err
		}
		if !sourceInfo.IsDir() {
			return fmt.Errorf("upload snapshot %q is not a directory", source)
		}
		if err := removeManagedPath(directory); err != nil {
			return err
		}
		if err := copyDirectory(source, directory); err != nil {
			return fmt.Errorf("restore %q: %w", directory, err)
		}
	}
	return nil
}

func (r *eventTestRunRepository) removeUploadSnapshot() error {
	if !r.hasUploadSnapshot() {
		return nil
	}
	if err := validateManagedPath(r.uploadSnapshotRoot); err != nil {
		return err
	}
	return removeManagedPath(r.uploadSnapshotRoot)
}

func validateUploadSnapshotPaths(snapshotRoot string, uploadDirectories []string) error {
	if err := validateManagedPath(snapshotRoot); err != nil {
		return fmt.Errorf("invalid upload snapshot path: %w", err)
	}
	seen := make(map[string]struct{}, len(uploadDirectories))
	for _, directory := range uploadDirectories {
		if err := validateManagedPath(directory); err != nil {
			return fmt.Errorf("invalid upload path: %w", err)
		}
		name := filepath.Base(filepath.Clean(directory))
		if _, exists := seen[name]; exists {
			return fmt.Errorf("upload directories must have unique base names: %q", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func validateManagedPath(path string) error {
	if path == "" {
		return fmt.Errorf("path is empty")
	}
	clean := filepath.Clean(path)
	if clean == "." {
		return fmt.Errorf("refusing to manage current directory")
	}
	abs, err := filepath.Abs(clean)
	if err != nil {
		return err
	}
	volumeRoot := filepath.Clean(filepath.VolumeName(abs) + string(os.PathSeparator))
	if abs == volumeRoot {
		return fmt.Errorf("refusing to manage filesystem root")
	}
	return nil
}

func removeManagedPath(path string) error {
	if err := validateManagedPath(path); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func copyDirectory(source, destination string) error {
	sourceInfo, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !sourceInfo.IsDir() {
		return fmt.Errorf("source %q is not a directory", source)
	}
	if err := os.MkdirAll(destination, sourceInfo.Mode().Perm()); err != nil {
		return err
	}
	sourceRoot, err := os.OpenRoot(source)
	if err != nil {
		return err
	}
	defer sourceRoot.Close()
	destinationRoot, err := os.OpenRoot(destination)
	if err != nil {
		return err
	}
	defer destinationRoot.Close()

	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == source {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not supported: %q", path)
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return destinationRoot.MkdirAll(relative, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported upload file type: %q", path)
		}
		return copyRegularFile(sourceRoot, destinationRoot, relative, info.Mode().Perm())
	})
}

func copyRegularFile(sourceRoot, destinationRoot *os.Root, name string, mode fs.FileMode) error {
	input, err := sourceRoot.Open(name)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := destinationRoot.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func errorsJoinSnapshotNotFound(path string) error {
	return fmt.Errorf("%w: %s", ErrTestRunNotFound, path)
}
