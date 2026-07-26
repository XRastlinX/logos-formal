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
	"sort"
	"strings"
	"unicode/utf8"
)

type treeEntry struct {
	absolutePath  string
	canonicalPath string
	kind          string
}

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
	if err := rejectSymlinkComponents(root, path); err != nil {
		return "", "", err
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
	if err := rejectSymlinkComponents(root, path); err != nil {
		return "", err
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
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("canonicalize repository boundary: %w", err)
	}
	canonicalCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return fmt.Errorf("canonicalize candidate path: %w", err)
	}
	relative, err := filepath.Rel(canonicalRoot, canonicalCandidate)
	if err != nil {
		return fmt.Errorf("compare repository and candidate paths: %w", err)
	}
	if relative == ".." || startsWithParent(relative) || filepath.IsAbs(relative) {
		return fmt.Errorf("path %q escapes repository boundary %q", canonicalCandidate, canonicalRoot)
	}
	return nil
}

func startsWithParent(path string) bool {
	return strings.HasPrefix(path, ".."+string(filepath.Separator))
}

func rejectSymlinkComponents(root, candidate string) error {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return fmt.Errorf("derive path components: %w", err)
	}
	if relative == "." {
		return nil
	}
	if relative == ".." || startsWithParent(relative) || filepath.IsAbs(relative) {
		return fmt.Errorf("path %q escapes repository boundary %q", candidate, root)
	}
	current := root
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("inspect path component %q: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink or reparse-point path component %q is forbidden", current)
		}
	}
	return nil
}

func ValidateRequiredPaths(repoRoot string, required []string) error {
	for _, requiredPath := range required {
		candidate := filepath.Join(repoRoot, filepath.FromSlash(requiredPath))
		if err := rejectSymlinkComponents(repoRoot, candidate); err != nil {
			return fmt.Errorf("required path %q: %w", requiredPath, err)
		}
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

	entries := make([]treeEntry, 0, 128)
	caseFoldedPaths := map[string]string{}

	err = filepath.WalkDir(target, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != target && entry.Name() == ".git" && entry.IsDir() {
			return filepath.SkipDir
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink or reparse point %q is forbidden", path)
		}

		relative, err := filepath.Rel(target, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if relative == "." {
			return nil
		}
		if !utf8.ValidString(relative) {
			return fmt.Errorf("path %q is not valid UTF-8", relative)
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		kind := ""
		if info.IsDir() {
			kind = "directory"
		} else if info.Mode().IsRegular() {
			kind = "file"
		} else {
			return fmt.Errorf("unsupported filesystem object %q with mode %s", relative, info.Mode())
		}

		folded := strings.ToLower(relative)
		if prior, exists := caseFoldedPaths[folded]; exists && prior != relative {
			return fmt.Errorf("case-colliding paths %q and %q are forbidden", prior, relative)
		}
		caseFoldedPaths[folded] = relative
		entries = append(entries, treeEntry{
			absolutePath:  path,
			canonicalPath: relative,
			kind:          kind,
		})
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("enumerate validation target: %w", err)
	}

	sort.Slice(entries, func(left, right int) bool {
		return entries[left].canonicalPath < entries[right].canonicalPath
	})

	digest := sha256.New()
	writeHashField(digest, "logos-formal.validation-tree.v1")
	for _, entry := range entries {
		writeHashField(digest, entry.canonicalPath)
		writeHashField(digest, entry.kind)
		if entry.kind != "file" {
			continue
		}

		file, err := os.Open(entry.absolutePath)
		if err != nil {
			return "", fmt.Errorf("open %q for hashing: %w", entry.canonicalPath, err)
		}
		info, err := file.Stat()
		if err != nil {
			_ = file.Close()
			return "", fmt.Errorf("stat %q for hashing: %w", entry.canonicalPath, err)
		}
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(info.Size()))
		if _, err := digest.Write(size[:]); err != nil {
			_ = file.Close()
			return "", err
		}
		if _, err := io.Copy(digest, file); err != nil {
			_ = file.Close()
			return "", fmt.Errorf("hash %q: %w", entry.canonicalPath, err)
		}
		if err := file.Close(); err != nil {
			return "", fmt.Errorf("close %q after hashing: %w", entry.canonicalPath, err)
		}
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}

func writeHashField(writer io.Writer, value string) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = writer.Write(size[:])
	_, _ = io.WriteString(writer, value)
}
