package placetime13d

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path"
	"strings"
)

type GitObserver struct{}

func (GitObserver) ObserveIndex(ctx context.Context, repoRoot string, excludes []string) (GitBinding, error) {
	objectFormat, err := gitText(ctx, repoRoot, "rev-parse", "--show-object-format")
	if err != nil {
		return GitBinding{}, err
	}
	treeOID, err := gitText(ctx, repoRoot, "write-tree")
	if err != nil {
		return GitBinding{}, err
	}
	parentOID, err := gitText(ctx, repoRoot, "rev-parse", "HEAD^{commit}")
	parentOIDs := []string{}
	if err == nil {
		parentOIDs = append(parentOIDs, parentOID)
	}
	ref, err := gitText(ctx, repoRoot, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		ref = "DETACHED"
	}
	repository, err := gitText(ctx, repoRoot, "remote", "get-url", "origin")
	if err != nil {
		repository = repoRoot
	}
	artifactRoot, err := hashGitTree(ctx, repoRoot, treeOID, excludes)
	if err != nil {
		return GitBinding{}, err
	}
	return GitBinding{
		Repository:       repository,
		ObjectFormat:     objectFormat,
		CommitOID:        nil,
		TreeOID:          nil,
		ParentOIDs:       parentOIDs,
		RefAtObservation: ref,
		ArtifactRoot:     HashRef{Alg: "sha256", Value: artifactRoot},
	}, nil
}

func (GitObserver) Observe(ctx context.Context, repoRoot, revision string, excludes []string) (GitBinding, error) {
	objectFormat, err := gitText(ctx, repoRoot, "rev-parse", "--show-object-format")
	if err != nil {
		return GitBinding{}, err
	}
	commitOID, err := gitText(ctx, repoRoot, "rev-parse", revision+"^{commit}")
	if err != nil {
		return GitBinding{}, err
	}
	treeOID, err := gitText(ctx, repoRoot, "show", "-s", "--format=%T", commitOID)
	if err != nil {
		return GitBinding{}, err
	}
	parentText, err := gitText(ctx, repoRoot, "show", "-s", "--format=%P", commitOID)
	if err != nil {
		return GitBinding{}, err
	}
	parentOIDs := []string{}
	if parentText != "" {
		parentOIDs = strings.Fields(parentText)
	}
	ref, err := gitText(ctx, repoRoot, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		ref = commitOID
	}
	repository, err := gitText(ctx, repoRoot, "remote", "get-url", "origin")
	if err != nil {
		repository = repoRoot
	}
	artifactRoot, err := hashGitTree(ctx, repoRoot, treeOID, excludes)
	if err != nil {
		return GitBinding{}, err
	}
	return GitBinding{
		Repository:       repository,
		ObjectFormat:     objectFormat,
		CommitOID:        stringPointer(commitOID),
		TreeOID:          stringPointer(treeOID),
		ParentOIDs:       parentOIDs,
		RefAtObservation: ref,
		ArtifactRoot:     HashRef{Alg: "sha256", Value: artifactRoot},
	}, nil
}

func hashGitTree(ctx context.Context, repoRoot, treeOID string, excludes []string) (string, error) {
	output, err := gitBytes(ctx, repoRoot, "ls-tree", "-r", "-z", "--full-tree", treeOID)
	if err != nil {
		return "", err
	}
	hasher := sha256.New()
	writeField(hasher, []byte("logos-formal.placetime13d.artifact-tree.v1"))
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Split(splitNull)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		entry := scanner.Text()
		if entry == "" {
			continue
		}
		tab := strings.IndexByte(entry, '\t')
		if tab < 0 {
			return "", fmt.Errorf("invalid git ls-tree entry")
		}
		header, filePath := entry[:tab], entry[tab+1:]
		if excludedPath(filePath, excludes) {
			continue
		}
		fields := strings.Fields(header)
		if len(fields) != 3 {
			return "", fmt.Errorf("invalid git ls-tree header %q", header)
		}
		mode, objectType, oid := fields[0], fields[1], fields[2]
		writeField(hasher, []byte(mode))
		writeField(hasher, []byte(objectType))
		writeField(hasher, []byte(filePath))
		if objectType == "blob" {
			content, err := gitBytes(ctx, repoRoot, "cat-file", "blob", oid)
			if err != nil {
				return "", err
			}
			writeField(hasher, content)
		} else {
			writeField(hasher, []byte(oid))
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func excludedPath(filePath string, excludes []string) bool {
	clean := path.Clean(strings.ReplaceAll(filePath, "\\", "/"))
	for _, exclude := range excludes {
		prefix := strings.TrimSuffix(path.Clean(strings.ReplaceAll(exclude, "\\", "/")), "/")
		if clean == prefix || strings.HasPrefix(clean, prefix+"/") {
			return true
		}
	}
	return false
}

func gitText(ctx context.Context, repoRoot string, arguments ...string) (string, error) {
	output, err := gitBytes(ctx, repoRoot, arguments...)
	return strings.TrimSpace(string(output)), err
}

func gitBytes(ctx context.Context, repoRoot string, arguments ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, "git", arguments...)
	command.Dir = repoRoot
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func splitNull(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if index := bytes.IndexByte(data, 0); index >= 0 {
		return index + 1, data[:index], nil
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func writeField(hasher interface{ Write([]byte) (int, error) }, value []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = hasher.Write(length[:])
	_, _ = hasher.Write(value)
}

func stringPointer(value string) *string {
	return &value
}
