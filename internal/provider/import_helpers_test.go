package provider

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDashboardDirs(t *testing.T) string {
	t.Helper()

	// Create a temp directory structure simulating multiple dashboard exports
	root := t.TempDir()

	// Dashboard 1
	dash1Dir := filepath.Join(root, "dashboard_one")
	require.NoError(t, os.MkdirAll(filepath.Join(dash1Dir, "datasets", "db_alpha"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dash1Dir, "databases"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dash1Dir, "charts"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(dash1Dir, "metadata.yaml"),
		[]byte("version: 1.0.0\ntype: Dashboard\ntimestamp: '2024-01-01T00:00:00+00:00'\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dash1Dir, "datasets", "db_alpha", "dataset_a.yaml"),
		[]byte("table_name: dataset_a\nuuid: aaaa-1111\ndatabase_uuid: db-uuid-1\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dash1Dir, "datasets", "db_alpha", "shared_dataset.yaml"),
		[]byte("table_name: shared_dataset\nuuid: shared-uuid\ndatabase_uuid: db-uuid-1\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dash1Dir, "databases", "db_alpha.yaml"),
		[]byte("database_name: db_alpha\nuuid: db-uuid-1\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dash1Dir, "charts", "chart_a.yaml"),
		[]byte("slice_name: Chart A\nuuid: chart-a-uuid\n"), 0644))

	// Dashboard 2 (shares a dataset with dashboard 1)
	dash2Dir := filepath.Join(root, "dashboard_two")
	require.NoError(t, os.MkdirAll(filepath.Join(dash2Dir, "datasets", "db_alpha"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dash2Dir, "databases"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(dash2Dir, "charts"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(dash2Dir, "metadata.yaml"),
		[]byte("version: 1.0.0\ntype: Dashboard\ntimestamp: '2024-01-01T00:00:00+00:00'\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dash2Dir, "datasets", "db_alpha", "dataset_b.yaml"),
		[]byte("table_name: dataset_b\nuuid: bbbb-2222\ndatabase_uuid: db-uuid-1\n"), 0644))
	// Same file path as dashboard 1 — should be deduplicated
	require.NoError(t, os.WriteFile(filepath.Join(dash2Dir, "datasets", "db_alpha", "shared_dataset.yaml"),
		[]byte("table_name: shared_dataset\nuuid: shared-uuid\ndatabase_uuid: db-uuid-1\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dash2Dir, "databases", "db_alpha.yaml"),
		[]byte("database_name: db_alpha\nuuid: db-uuid-1\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dash2Dir, "charts", "chart_b.yaml"),
		[]byte("slice_name: Chart B\nuuid: chart-b-uuid\n"), 0644))

	return root
}

func TestCollectDedupedFiles_Datasets(t *testing.T) {
	root := setupTestDashboardDirs(t)

	collected, err := collectDedupedFiles(root, []string{"datasets", "databases"}, nil)
	require.NoError(t, err)

	// Should have: 3 unique datasets + 1 database = 4 files
	// shared_dataset.yaml appears in both dashboards but should be deduped
	paths := make(map[string]bool)
	for _, f := range collected {
		paths[f.RelPath] = true
	}

	assert.True(t, paths["datasets/db_alpha/dataset_a.yaml"], "dataset_a should be collected")
	assert.True(t, paths["datasets/db_alpha/shared_dataset.yaml"], "shared_dataset should be collected")
	assert.True(t, paths["datasets/db_alpha/dataset_b.yaml"], "dataset_b should be collected")
	assert.True(t, paths["databases/db_alpha.yaml"], "database should be collected")
	assert.Equal(t, 4, len(collected), "should have exactly 4 deduplicated files")
}

func TestCollectDedupedFiles_Charts(t *testing.T) {
	root := setupTestDashboardDirs(t)

	collected, err := collectDedupedFiles(root, []string{"charts", "datasets", "databases"}, nil)
	require.NoError(t, err)

	paths := make(map[string]bool)
	for _, f := range collected {
		paths[f.RelPath] = true
	}

	// 2 charts + 3 datasets + 1 database = 6 files
	assert.True(t, paths["charts/chart_a.yaml"])
	assert.True(t, paths["charts/chart_b.yaml"])
	assert.True(t, paths["datasets/db_alpha/dataset_a.yaml"])
	assert.True(t, paths["datasets/db_alpha/shared_dataset.yaml"])
	assert.True(t, paths["datasets/db_alpha/dataset_b.yaml"])
	assert.True(t, paths["databases/db_alpha.yaml"])
	assert.Equal(t, 6, len(collected))
}

func TestCollectDedupedFiles_DatabaseOverrides(t *testing.T) {
	root := setupTestDashboardDirs(t)

	overrides := map[string]map[string]interface{}{
		"db-uuid-1": {
			"sqlalchemy_uri": "starrocks://new-host:9030",
		},
	}

	collected, err := collectDedupedFiles(root, []string{"databases"}, overrides)
	require.NoError(t, err)

	require.Equal(t, 1, len(collected))
	assert.Contains(t, string(collected[0].Data), "starrocks://new-host:9030")
}

func TestHashCollectedFiles(t *testing.T) {
	files := []collectedFile{
		{RelPath: "a.yaml", Data: []byte("hello")},
		{RelPath: "b.yaml", Data: []byte("world")},
	}

	hashes := hashCollectedFiles(files)

	assert.Equal(t, 2, len(hashes))
	assert.NotEmpty(t, hashes["a.yaml"])
	assert.NotEmpty(t, hashes["b.yaml"])
	assert.NotEqual(t, hashes["a.yaml"], hashes["b.yaml"])

	// Same content should produce same hash
	files2 := []collectedFile{
		{RelPath: "a.yaml", Data: []byte("hello")},
	}
	hashes2 := hashCollectedFiles(files2)
	assert.Equal(t, hashes["a.yaml"], hashes2["a.yaml"])
}

func TestBuildPasswordMapFromCollected(t *testing.T) {
	files := []collectedFile{
		{RelPath: "databases/db_alpha.yaml", Data: []byte("database_name: db_alpha\nuuid: db-uuid-1\n")},
		{RelPath: "databases/db_beta.yaml", Data: []byte("database_name: db_beta\nuuid: db-uuid-2\n")},
		{RelPath: "datasets/db_alpha/test.yaml", Data: []byte("table_name: test\n")},
	}

	secrets := map[string]string{
		"db-uuid-1": "secret123",
	}

	passwordMap := buildPasswordMapFromCollected(files, secrets)

	assert.Equal(t, "secret123", passwordMap["databases/db_alpha.yaml"])
	assert.Equal(t, "", passwordMap["databases/db_beta.yaml"])
	assert.NotContains(t, passwordMap, "datasets/db_alpha/test.yaml")
}

func TestBuildPasswordMapFromCollected_NoSecrets(t *testing.T) {
	files := []collectedFile{
		{RelPath: "databases/db_alpha.yaml", Data: []byte("database_name: db_alpha\nuuid: db-uuid-1\n")},
	}

	passwordMap := buildPasswordMapFromCollected(files, map[string]string{})
	assert.Nil(t, passwordMap)
}

func TestBuildImportZip(t *testing.T) {
	root := setupTestDashboardDirs(t)

	files := []collectedFile{
		{RelPath: "datasets/db_alpha/dataset_a.yaml", Data: []byte("table_name: dataset_a\n")},
		{RelPath: "databases/db_alpha.yaml", Data: []byte("database_name: db_alpha\n")},
	}

	zipData, err := buildImportZip(files, "SqlaTable", root)
	require.NoError(t, err)
	require.NotEmpty(t, zipData)

	// Read the ZIP and verify contents
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	require.NoError(t, err)

	zipPaths := make(map[string]bool)
	var metadataContent []byte
	for _, f := range reader.File {
		zipPaths[f.Name] = true
		if f.Name == "import_export/metadata.yaml" {
			rc, err := f.Open()
			require.NoError(t, err)
			metadataContent = make([]byte, f.UncompressedSize64)
			_, err = rc.Read(metadataContent)
			rc.Close()
		}
	}

	assert.True(t, zipPaths["import_export/"])
	assert.True(t, zipPaths["import_export/datasets/"])
	assert.True(t, zipPaths["import_export/datasets/db_alpha/"])
	assert.True(t, zipPaths["import_export/datasets/db_alpha/dataset_a.yaml"])
	assert.True(t, zipPaths["import_export/databases/"])
	assert.True(t, zipPaths["import_export/databases/db_alpha.yaml"])
	assert.True(t, zipPaths["import_export/metadata.yaml"])

	// Verify metadata has correct type
	assert.Contains(t, string(metadataContent), "SqlaTable")
}

func TestBuildImportZip_ChartType(t *testing.T) {
	root := setupTestDashboardDirs(t)

	files := []collectedFile{
		{RelPath: "charts/chart_a.yaml", Data: []byte("slice_name: Chart A\n")},
	}

	zipData, err := buildImportZip(files, "Slice", root)
	require.NoError(t, err)

	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	require.NoError(t, err)

	for _, f := range reader.File {
		if f.Name == "import_export/metadata.yaml" {
			rc, err := f.Open()
			require.NoError(t, err)
			buf := make([]byte, 1024)
			n, _ := rc.Read(buf)
			rc.Close()
			assert.Contains(t, string(buf[:n]), "Slice")
			return
		}
	}
	t.Fatal("metadata.yaml not found in ZIP")
}
