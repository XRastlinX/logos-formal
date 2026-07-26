package cyonicvalidate

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func FindRepositoryRoot(start string) (string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve starting directory: %w", err)
	}
	if info, statErr := os.Stat(current); statErr == nil && !info.IsDir() {
		current = filepath.Dir(current)
	}

	for {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return filepath.EvalSymlinks(current)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("no Git repository found from %q", start)
		}
		current = parent
	}
}

func ResolveBoundPath(repoRoot, candidate string) (string, string, error) {
	if candidate == "" {
		candidate = "."
	}
	root, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve repository root: %w", err)
	}

	path := candidate
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("resolve target path: %w", err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", "", fmt.Errorf("resolve target symlinks: %w", err)
	}
	if err := ensureWithin(root, path); err != nil {
		return "", "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", "", fmt.Errorf("stat target: %w", err)
	}
	if !info.IsDir() {
		return "", "", fmt.Errorf("validation target %q is not a directory", candidate)
	}

	relative, err := filepath.Rel(root, path)
	if err != nil {
		return "", "", fmt.Errorf("make target repository-relative: %w", err)
	}
	if relative == "." {
		return path, ".", nil
	}
	return path, filepath.ToSlash(relative), nil
}

func ResolveBoundFile(repoRoot, candidate string) (string, error) {
	root, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	path := candidate
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve file path: %w", err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve file symlinks: %w", err)
	}
	if err := ensureWithin(root, path); err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("path %q is not a regular file", candidate)
	}
	return path, nil
}

func ensureWithin(root, candidate string) error {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return fmt.Errorf("compare repository and candidate paths: %w", err)
	}
	if relative == ".." || startsWithParent(relative) || filepath.IsAbs(relative) {
		return fmt.Errorf("path %q escapes repository boundary %q", candidate, root)
	}
	return nil
}

func startsWithParent(path string) bool {
	return strings.HasPrefix(path, ".."+string(filepath.Separator))
}

func ValidateRequiredPaths(repoRoot string, required []string) error {
	for _, requiredPath := range required {
		candidate := filepath.Join(repoRoot, filepath.FromSlash(requiredPath))
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			return fmt.Errorf("required path %q: %w", requiredPath, err)
		}
		if err := ensureWithin(repoRoot, resolved); err != nil {
			return fmt.Errorf("required path %q: %w", requiredPath, err)
		}
		if _, err := os.Stat(resolved); err != nil {
			return fmt.Errorf("required path %q: %w", requiredPath, err)
		}
	}
	return nil
}

func HashTree(repoRoot, target string) (string, error) {
	root, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		return "", fmt.Errorf("resolve repository root for hashing: %w", err)
	}
	target, err = filepath.EvalSymlinks(target)
	if err != nil {
		return "", fmt.Errorf("resolve target for hashing: %w", err)
	}
	if err := ensureWithin(root, target); err != nil {
		return "", err
	}

	digest := sha256.New()
	writeHashField(digest, "logos-formal.validation-tree.v1")

	err = filepath.WalkDir(target, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == filepath.Join(root, ".git") {
			return filepath.SkipDir
		}

		relative, err := filepath.Rel(target, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if relative == "." {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		writeHashField(digest, relative)

		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				return fmt.Errorf("resolve symlink %q: %w", relative, err)
			}
			if err := ensureWithin(root, resolved); err != nil {
				return fmt.Errorf("symlink %q: %w", relative, err)
			}
			writeHashField(digest, "symlink")
			writeHashField(digest, filepath.ToSlash(linkTarget))
			return nil
		}
		if info.IsDir() {
			writeHashField(digest, "directory")
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported filesystem object %q with mode %s", relative, info.Mode())
		}

		writeHashField(digest, "file")
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(info.Size()))
		if _, err := digest.Write(size[:]); err != nil {
			_ = file.Close()
			return err
		}
		if _, err := io.Copy(digest, file); err != nil {
			_ = file.Close()
			return err
		}
		return file.Close()
	})
	if err != nil {
		return "", fmt.Errorf("hash validation target: %w", err)
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}

func writeHashField(writer io.Writer, value string) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = writer.Write(size[:])
	_, _ = io.WriteString(writer, value)
}
