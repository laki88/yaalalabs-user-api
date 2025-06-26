package userservice_test

import (
	"context"
	"database/sql"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/db"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/internal/repository"
	"github.com/laki88/yaalalabs-user-api/user-rest-api/pkg/userservice"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var testService userservice.UserService

const (
	adminConnStr = "postgres://user:pass@localhost:5432/postgres?sslmode=disable"
	testConnStr  = "postgres://user:pass@localhost:5432/userdb_test?sslmode=disable"
	schemaPath   = "../../db/schema.sql" // relative to project root
)

func initTestDB() *sql.DB {
	// Connect to admin DB
	adminDB, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to admin DB: %v", err)
	}
	defer func(adminDB *sql.DB) {
		err := adminDB.Close()
		if err != nil {

		}
	}(adminDB)

	// Disconnect any active sessions first
	_, _ = adminDB.Exec(`
	SELECT pg_terminate_backend(pid)
	FROM pg_stat_activity
	WHERE datname = 'userdb_test' AND pid <> pg_backend_pid();
`)

	// Drop and recreate test DB
	_, _ = adminDB.Exec("DROP DATABASE IF EXISTS userdb_test")
	_, err = adminDB.Exec("CREATE DATABASE userdb_test")
	if err != nil {
		log.Fatalf("Failed to create test DB: %v", err)
	}

	// Connect to the new test DB
	testDB, err := sql.Open("postgres", testConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to test DB: %v", err)
	}

	// Load and apply schema
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Fatalf("Failed to read schema.sql: %v", err)
	}
	if _, err := testDB.Exec(string(schema)); err != nil {
		log.Fatalf("Failed to execute schema.sql: %v", err)
	}

	return testDB
}

func TestMain(m *testing.M) {
	dbConn := initTestDB()

	queries := db.New(dbConn)
	repo := repository.NewPostgresUserRepository(queries)
	testService = userservice.NewService(repo)

	code := m.Run()

	_ = dbConn.Close()
	os.Exit(code)
}

func TestIntegration_CreateGetDeleteUser(t *testing.T) {
	ctx := context.Background()

	phone := "5551234"
	status := "Active"
	age := int32(28)

	user, err := testService.CreateUser(ctx, userservice.CreateUserParams{
		FirstName: "Integration",
		LastName:  "Test",
		Email:     "integration@test.com",
		Phone:     &phone,
		Age:       &age,
		Status:    &status,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Integration", user.FirstName)

	fetched, err := testService.GetUser(ctx, user.UserID)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, fetched.Email)

	err = testService.DeleteUser(ctx, user.UserID)
	assert.NoError(t, err)

	_, err = testService.GetUser(ctx, user.UserID)
	assert.Error(t, err)
}
