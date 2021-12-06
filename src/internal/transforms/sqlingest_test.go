package transforms

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const outputName = "0000"

func TestSQLIngest(t *testing.T) {
	ctx := context.Background()
	inputDir, outputDir := t.TempDir(), t.TempDir()
	u := dockertestenv.NewMySQLURL(t)
	const N = 100
	loadDB(t, u, N)

	// write queries
	const Shards = 2
	for i := 0; i < Shards; i++ {
		name := fmt.Sprintf("%04d", i)
		// query would normally be different per shard
		query := "select * from test_data"
		err := ioutil.WriteFile(filepath.Join(inputDir, name), []byte(query), 0755)
		require.NoError(t, err)
	}

	err := SQLIngest(ctx, SQLIngestParams{
		Logger: logrus.StandardLogger(),

		InputDir:  inputDir,
		OutputDir: outputDir,

		URL:      u,
		Password: dockertestenv.MySQLPassword,
		Format:   "json",
	})
	require.NoError(t, err)

	// check the file exists
	dirEnts, err := os.ReadDir(outputDir)
	require.NoError(t, err)
	require.Len(t, dirEnts, Shards)
	require.Equal(t, outputName, dirEnts[0].Name())
	lineCount := countLinesInFile(t, filepath.Join(outputDir, outputName))
	require.Equal(t, N, lineCount)
}

func countLinesInFile(t testing.TB, p string) int {
	f, err := os.Open(p)
	require.NoError(t, err)
	defer f.Close()
	br := bufio.NewReader(f)
	var count int
	for {
		ru, _, err := br.ReadRune()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if ru == '\n' {
			count++
		}
	}
	return count
}

func TestSQLQueryGeneration(t *testing.T) {
	ctx := context.Background()
	log := logrus.StandardLogger()
	inputDir, outputDir := t.TempDir(), t.TempDir()
	writeCronFile(t, inputDir)

	err := SQLQueryGeneration(ctx, SQLQueryGenerationParams{
		Logger:    log,
		InputDir:  inputDir,
		OutputDir: outputDir,
		Query:     "select * from test_data",
	})
	require.NoError(t, err)

	dirEnts, err := os.ReadDir(outputDir)
	require.NoError(t, err)
	require.Len(t, dirEnts, 1)
	require.Equal(t, outputName, dirEnts[0].Name())
	data, err := ioutil.ReadFile(filepath.Join(outputDir, outputName))
	require.NoError(t, err)
	t.Log(string(data))
}

func writeCronFile(t testing.TB, inputDir string) {
	now := time.Now().UTC()
	timestampStr := now.Format(time.RFC3339)
	err := ioutil.WriteFile(filepath.Join(inputDir, timestampStr), nil, 0755)
	require.NoError(t, err)
}

func loadDB(t testing.TB, u pachsql.URL, numRows int) {
	// load  DB
	db := testutil.OpenDBURL(t, u, dockertestenv.MySQLPassword)
	_, err := db.Exec(`CREATE TABLE test_data (
			id SERIAL PRIMARY KEY,
			col_a VARCHAR(100)
	)`)
	require.NoError(t, err)
	for i := 0; i < numRows; i++ {
		_, err := db.Exec(`INSERT INTO test_data (col_a) VALUES (?)`, randutil.UniqueString(""))
		require.NoError(t, err)
	}
}

func TestSQLChain(t *testing.T) {
	ctx := context.Background()
	log := logrus.StandardLogger()
	u := dockertestenv.NewMySQLURL(t)
	const N = 100
	loadDB(t, u, N)
	inputDir, outputDir := t.TempDir(), t.TempDir()
	t.Logf("input: %v, output: %v", inputDir, outputDir)
	writeCronFile(t, inputDir)

	runChain(t, inputDir, outputDir, []func(string, string) error{
		func(inDir, outDir string) error {
			return SQLQueryGeneration(ctx, SQLQueryGenerationParams{
				Logger:    log,
				InputDir:  inDir,
				OutputDir: outDir,
				Query:     "select * FROM test_data",
			})
		},
		func(inDir, outDir string) error {
			return SQLIngest(ctx, SQLIngestParams{
				Logger:    log,
				InputDir:  inDir,
				OutputDir: outDir,
				URL:       u,
				Password:  dockertestenv.MySQLPassword,
				Format:    "csv",
			})
		},
	})
	lineCount := countLinesInFile(t, filepath.Join(outputDir, outputName))
	require.Equal(t, N, lineCount)
}

func runChain(t testing.TB, inDir, outDir string, stages []func(inDir, outDir string) error) {
	for i, stage := range stages {
		var outDir2 string
		if i == len(stages)-1 {
			outDir2 = outDir
		} else {
			outDir2 = t.TempDir()
		}
		err := stage(inDir, outDir2)
		require.NoError(t, err, "error in stage %d/%d", i+1, len(stages))
		inDir = outDir2
	}
}
