package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUploadSnapshotRestoresExistingFilesAndRemovesTestFiles(t *testing.T) {
	root := t.TempDir()
	images := filepath.Join(root, "uploads", "images")
	pdfs := filepath.Join(root, "uploads", "pdfs")
	snapshot := filepath.Join(root, "snapshots", "current")
	mustWriteTestFile(t, filepath.Join(images, "nested", "original.png"), "original image")
	mustWriteTestFile(t, filepath.Join(pdfs, "guide.pdf"), "original pdf")

	repo := &eventTestRunRepository{
		uploadSnapshotRoot: snapshot,
		uploadDirectories:  []string{images, pdfs},
	}
	if err := repo.createUploadSnapshot(); err != nil {
		t.Fatalf("createUploadSnapshot() error = %v", err)
	}

	mustWriteTestFile(t, filepath.Join(images, "nested", "original.png"), "changed image")
	mustWriteTestFile(t, filepath.Join(images, "test-only.png"), "test image")
	if err := os.Remove(filepath.Join(pdfs, "guide.pdf")); err != nil {
		t.Fatal(err)
	}

	if err := repo.restoreUploadSnapshot(); err != nil {
		t.Fatalf("restoreUploadSnapshot() error = %v", err)
	}
	assertTestFileContent(t, filepath.Join(images, "nested", "original.png"), "original image")
	assertTestFileContent(t, filepath.Join(pdfs, "guide.pdf"), "original pdf")
	if _, err := os.Stat(filepath.Join(images, "test-only.png")); !os.IsNotExist(err) {
		t.Fatalf("test-only image still exists or stat failed: %v", err)
	}
}

func TestUploadSnapshotRemovesDirectoryCreatedDuringTestRun(t *testing.T) {
	root := t.TempDir()
	images := filepath.Join(root, "uploads", "images")
	snapshot := filepath.Join(root, "snapshots", "current")
	repo := &eventTestRunRepository{
		uploadSnapshotRoot: snapshot,
		uploadDirectories:  []string{images},
	}

	if err := repo.createUploadSnapshot(); err != nil {
		t.Fatalf("createUploadSnapshot() error = %v", err)
	}
	mustWriteTestFile(t, filepath.Join(images, "test-only.png"), "test image")
	if err := repo.restoreUploadSnapshot(); err != nil {
		t.Fatalf("restoreUploadSnapshot() error = %v", err)
	}
	if _, err := os.Stat(images); !os.IsNotExist(err) {
		t.Fatalf("directory created during test run still exists or stat failed: %v", err)
	}
}

func TestUploadSnapshotRejectsUnsafePaths(t *testing.T) {
	repo := &eventTestRunRepository{
		uploadSnapshotRoot: ".",
		uploadDirectories:  []string{filepath.Join(t.TempDir(), "images")},
	}
	if err := repo.createUploadSnapshot(); err == nil {
		t.Fatal("expected current directory snapshot path to be rejected")
	}
}

func TestUploadSnapshotRejectsSymbolicLinks(t *testing.T) {
	root := t.TempDir()
	images := filepath.Join(root, "images")
	if err := os.MkdirAll(images, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "outside"), filepath.Join(images, "link")); err != nil {
		t.Fatal(err)
	}
	repo := &eventTestRunRepository{
		uploadSnapshotRoot: filepath.Join(root, "snapshot"),
		uploadDirectories:  []string{images},
	}
	if err := repo.createUploadSnapshot(); err == nil {
		t.Fatal("expected symbolic link to be rejected")
	}
}

func mustWriteTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertTestFileContent(t *testing.T, path, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != want {
		t.Fatalf("content = %q, want %q", string(content), want)
	}
}
