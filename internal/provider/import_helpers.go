package provider

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// collectedFile represents a single file collected for import with its content already resolved.
type collectedFile struct {
	// RelPath is the deduplication key, e.g. "datasets/<database_name>/dataset_name.yaml"
	RelPath string
	// Data is the file content (with database overrides applied if applicable)
	Data []byte
}

// collectDedupedFiles walks all dashboard export subdirectories under sourceDir,
// collects files from the specified subdirectory prefixes (e.g. "datasets", "databases"),
// and deduplicates them by relative file path. Database YAML overrides are applied.
//
// sourceDir is expected to contain one or more dashboard export directories, each
// containing subdirs like datasets/, databases/, charts/.
func collectDedupedFiles(sourceDir string, prefixes []string, overrides map[string]map[string]interface{}) ([]collectedFile, error) {
	seen := make(map[string]bool)
	var result []collectedFile

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("reading source_dir %s: %w", sourceDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dashDir := filepath.Join(sourceDir, entry.Name())

		for _, prefix := range prefixes {
			subDir := filepath.Join(dashDir, prefix)
			if _, err := os.Stat(subDir); os.IsNotExist(err) {
				continue
			}

			err := filepath.WalkDir(subDir, func(p string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}

				// Relative path from the dashboard subdir, e.g. "datasets/StarRocks/ngr_daily.yaml"
				rel, err := filepath.Rel(dashDir, p)
				if err != nil {
					return err
				}
				relSlash := filepath.ToSlash(rel)

				if seen[relSlash] {
					return nil // deduplicated
				}
				seen[relSlash] = true

				data, err := os.ReadFile(p)
				if err != nil {
					return err
				}

				// Apply database overrides if this is a database YAML
				if strings.HasPrefix(relSlash, "databases/") && strings.HasSuffix(relSlash, ".yaml") {
					data, _ = applyDatabaseOverrides(data, overrides)
				}

				result = append(result, collectedFile{RelPath: relSlash, Data: data})
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("walking %s: %w", subDir, err)
			}
		}
	}

	return result, nil
}

// hashCollectedFiles computes SHA256 hashes for each collected file.
func hashCollectedFiles(files []collectedFile) map[string]string {
	hashes := make(map[string]string, len(files))
	for _, f := range files {
		h := sha256.Sum256(f.Data)
		hashes[f.RelPath] = fmt.Sprintf("%x", h)
	}
	return hashes
}

// buildPasswordMapFromCollected builds the password map from collected database files.
// Keys are "databases/<filename>.yaml", values are passwords from secrets map (matched by UUID).
func buildPasswordMapFromCollected(files []collectedFile, secrets map[string]string) map[string]string {
	if len(secrets) == 0 {
		return nil
	}

	result := make(map[string]string)
	for _, f := range files {
		if !strings.HasPrefix(f.RelPath, "databases/") || !strings.HasSuffix(f.RelPath, ".yaml") {
			continue
		}
		result[f.RelPath] = ""
		var db struct {
			UUID string `yaml:"uuid"`
		}
		if err := yaml.Unmarshal(f.Data, &db); err != nil {
			continue
		}
		if pw, ok := secrets[db.UUID]; ok {
			result[f.RelPath] = pw
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// buildImportZip creates a ZIP from collected files with a metadata.yaml containing the specified type.
// The ZIP structure uses "import_export/" as the root directory.
// It reads metadata.yaml from the first dashboard subdir found in sourceDir and overrides the type field.
func buildImportZip(files []collectedFile, metadataType string, sourceDir string) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	const root = "import_export"

	// Create root directory entry
	if _, err := w.Create(root + "/"); err != nil {
		return nil, err
	}

	// Track which subdirectories we've created
	createdDirs := make(map[string]bool)

	for _, f := range files {
		// Ensure parent directories exist in ZIP
		parts := strings.Split(f.RelPath, "/")
		for i := 1; i < len(parts); i++ {
			dir := strings.Join(parts[:i], "/")
			dirPath := root + "/" + dir + "/"
			if !createdDirs[dirPath] {
				if _, err := w.Create(dirPath); err != nil {
					return nil, err
				}
				createdDirs[dirPath] = true
			}
		}

		zipPath := root + "/" + f.RelPath
		fw, err := w.Create(zipPath)
		if err != nil {
			return nil, err
		}
		if _, err := fw.Write(f.Data); err != nil {
			return nil, err
		}
	}

	// Find and read metadata.yaml from the first dashboard subdir, override the type field
	metaContent, err := buildMetadataYAML(sourceDir, metadataType)
	if err != nil {
		return nil, fmt.Errorf("building metadata.yaml: %w", err)
	}

	metaPath := root + "/metadata.yaml"
	fw, err := w.Create(metaPath)
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(metaContent); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// buildMetadataYAML reads metadata.yaml from the first dashboard subdir and overrides the type field.
// It also sets the timestamp to the current time for traceability.
func buildMetadataYAML(sourceDir string, metadataType string) ([]byte, error) {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metaPath := filepath.Join(sourceDir, entry.Name(), "metadata.yaml")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}

		var doc map[string]interface{}
		if err := yaml.Unmarshal(data, &doc); err != nil {
			continue
		}
		doc["type"] = metadataType
		doc["timestamp"] = time.Now().UTC().Format(time.RFC3339)

		out, err := yaml.Marshal(doc)
		if err != nil {
			return nil, err
		}
		return out, nil
	}

	// Fallback if no metadata.yaml found in any subdir
	fallback := fmt.Sprintf("version: 1.0.0\ntype: %s\ntimestamp: '%s'\n", metadataType, time.Now().UTC().Format(time.RFC3339))
	return []byte(fallback), nil
}
