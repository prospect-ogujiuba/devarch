//go:build integration

package integration_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/priz/devarch-api/internal/compose"
	"github.com/stretchr/testify/require"
)

func testdataDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "services-library")
}

func TestImportAll(t *testing.T) {
	truncateAll(t, testDB)

	importer := compose.NewImporter(testDB, testdataDir())
	err := importer.ImportAll()
	require.NoError(t, err)

	var catCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM categories WHERE name = 'testcat'").Scan(&catCount)
	require.NoError(t, err)
	require.Equal(t, 1, catCount)

	var svcCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM services").Scan(&svcCount)
	require.NoError(t, err)
	require.Equal(t, 2, svcCount)

	var names []string
	rows, err := testDB.Query("SELECT name FROM services ORDER BY name")
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var n string
		require.NoError(t, rows.Scan(&n))
		names = append(names, n)
	}
	require.Equal(t, []string{"svc-plain", "svc-xconfig"}, names)
}

func TestImportConfigFiles(t *testing.T) {
	truncateAll(t, testDB)

	importer := compose.NewImporter(testDB, testdataDir())
	err := importer.ImportAll()
	require.NoError(t, err)

	var fileCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM service_config_files").Scan(&fileCount)
	require.NoError(t, err)
	require.Equal(t, 2, fileCount)

	var paths []string
	rows, err := testDB.Query(`
		SELECT scf.file_path FROM service_config_files scf
		JOIN services s ON s.id = scf.service_id
		WHERE s.name = 'svc-xconfig'
		ORDER BY scf.file_path
	`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var p string
		require.NoError(t, rows.Scan(&p))
		paths = append(paths, p)
	}
	require.Equal(t, []string{"main.conf", "sub/nested.yml"}, paths)
}

func TestImportConfigMountFKs(t *testing.T) {
	truncateAll(t, testDB)

	importer := compose.NewImporter(testDB, testdataDir())
	err := importer.ImportAll()
	require.NoError(t, err)

	var nullCount int
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM service_config_mounts
		WHERE config_file_id IS NULL
	`).Scan(&nullCount)
	require.NoError(t, err)
	require.Equal(t, 0, nullCount)

	var resolvedCount int
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM service_config_mounts
		WHERE config_file_id IS NOT NULL
	`).Scan(&resolvedCount)
	require.NoError(t, err)
	require.Equal(t, 2, resolvedCount)
}

func TestFullImportPipeline(t *testing.T) {
	truncateAll(t, testDB)

	importer := compose.NewImporter(testDB, testdataDir())
	err := importer.ImportAll()
	require.NoError(t, err)

	var svcCount int
	testDB.QueryRow("SELECT COUNT(*) FROM services").Scan(&svcCount)
	require.Equal(t, 2, svcCount)

	var configFileCount int
	testDB.QueryRow("SELECT COUNT(*) FROM service_config_files").Scan(&configFileCount)
	require.Equal(t, 2, configFileCount)

	var mountCount int
	testDB.QueryRow("SELECT COUNT(*) FROM service_config_mounts WHERE config_file_id IS NOT NULL").Scan(&mountCount)
	require.Equal(t, 2, mountCount)

	var target string
	err = testDB.QueryRow(`
		SELECT scm.target_path FROM service_config_mounts scm
		JOIN service_config_files scf ON scf.id = scm.config_file_id
		WHERE scf.file_path = 'main.conf'
	`).Scan(&target)
	require.NoError(t, err)
	require.Equal(t, "/etc/app/main.conf", target)
}
