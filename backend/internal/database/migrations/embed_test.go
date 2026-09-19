package migrations

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var prefixPattern = regexp.MustCompile(`^[0-9]+[a-z]?$`)

func TestEmbeddedMigrations(t *testing.T) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		t.Fatalf("failed to read embedded migrations dir: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("no migration files embedded")
	}

	t.Logf("Total embedded migrations discovered: %d", len(entries))

	fileNames := make([]string, 0, len(entries))
	contentHashes := make(map[string]string)

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			t.Errorf("unexpected non-sql file in embedded migrations: %s", name)
			continue
		}

		fileNames = append(fileNames, name)

		data, err := FS.ReadFile(name)
		if err != nil {
			t.Errorf("failed to read embedded file %s: %v", name, err)
			continue
		}

		if len(data) == 0 {
			t.Errorf("embedded migration %s is empty", name)
		}

		hash := sha256.Sum256(data)
		hashStr := hex.EncodeToString(hash[:])
		if prevFile, exists := contentHashes[hashStr]; exists {
			t.Logf("Notice: identical migration content between %s and %s", prevFile, name)
		} else {
			contentHashes[hashStr] = name
		}

		// Verify prefix format (e.g. "001_", "013b_", "123_")
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			t.Errorf("migration filename missing underscore separator: %s", name)
			continue
		}
		if !prefixPattern.MatchString(parts[0]) {
			t.Errorf("migration prefix %q does not match expected pattern: %s", parts[0], name)
		}
	}

	// Verify deterministic sorting
	isSorted := sort.SliceIsSorted(fileNames, func(i, j int) bool {
		return fileNames[i] < fileNames[j]
	})
	if !isSorted {
		t.Errorf("migration files are not lexicographically sorted")
	}
}
