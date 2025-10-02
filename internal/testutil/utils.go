package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
)

func CreateTestDb(t *testing.T) *sql.DB {
	dsnWithoutDb := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)
	defaultDsn := dsnWithoutDb + " dbname=postgres"
	defaultDb, err := sql.Open("postgres", defaultDsn)
	if err != nil {
		defaultDb.Close()
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	testDbName := fmt.Sprintf("test_db_part_dump_%d", os.Getpid())
	dropDB := fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDbName)
	_, _ = defaultDb.Exec(dropDB)
	createDB := fmt.Sprintf("CREATE DATABASE %s", testDbName)
	_, err = defaultDb.Exec(createDB)
	if err != nil {
		defaultDb.Close()
		t.Fatalf("failed to create test database: %v", err)
	}
	dsn := fmt.Sprintf("%s dbname=%s", dsnWithoutDb, testDbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		defaultDb.Close()
		t.Fatalf("failed to connect to testDb: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		defaultDb.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDbName))
		defaultDb.Close()
	})
	return db

}

func getTestFilename(t *testing.T) string {
	_, filename, _, ok := runtime.Caller(0)
	splitedDir := strings.Split(filename, "/")
	dir := strings.Join(splitedDir[:len(splitedDir)-1], "/")
	if !ok {
		t.Fatalf("no caller info")
	}
	return dir + "/data.sql"
}

func InsertTestData(t *testing.T, db *sql.DB) {
	filename := getTestFilename(t)
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	statements := strings.SplitSeq(string(fileContent), ";")
	for statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("failed to exec statement: %v", statement)
		}
	}
}
