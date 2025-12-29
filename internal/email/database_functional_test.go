package email

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func getTestDB(t *testing.T) *sql.DB {
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	database := os.Getenv("MYSQL_DATABASE")

	if host == "" || user == "" || database == "" {
		t.Skip("MySQL environment variables not set, skipping functional test")
	}

	if port == "" {
		port = "3306"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	return db
}

func cleanupEmail(t *testing.T, db *sql.DB, id string) {
	_, err := db.Exec("DELETE FROM emails WHERE id = ?", id)
	if err != nil {
		t.Logf("cleanup failed for id %s: %v", id, err)
	}
}

func TestOutboxComponentWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("component tests are skipped in short mode")
	}

	db := getTestDB(t)
	defer db.Close()

	sut := NewDatabase(db, 30)

	ctx := context.TODO()

	// insert two records
	firstId := uuid.NewString()
	defer cleanupEmail(t, db, firstId)

	err := sut.Insert(ctx, firstId, "/payload/path1.json")
	require.NoErrorf(t, err, "failed inserting id %s, error: %v", firstId, err)

	secondId := uuid.NewString()
	defer cleanupEmail(t, db, secondId)

	err = sut.Insert(ctx, secondId, "/payload/path2.json")
	require.NoErrorf(t, err, "failed inserting id %s, error: %v", secondId, err)

	// verify records exist with ACCEPTED status
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM emails WHERE status = 'ACCEPTED' AND id IN (?, ?)", firstId, secondId).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	// verify email_statuses records were created
	err = db.QueryRow("SELECT COUNT(*) FROM email_statuses WHERE email_id IN (?, ?)", firstId, secondId).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	// should not be able to insert again same id
	err = sut.Insert(ctx, firstId, "/")
	require.Errorf(t, err, "inserted id %s, but it should have not because it's duplicated", firstId)
	require.True(t, IsDuplicateEntryError(err))
}

func TestGetStaleEmails(t *testing.T) {
	if testing.Short() {
		t.Skip("component tests are skipped in short mode")
	}

	db := getTestDB(t)
	defer db.Close()

	sut := NewDatabase(db, 30)
	ctx := context.TODO()

	// Insert a stale email directly (bypassing Insert to set custom updated_at)
	staleId := uuid.NewString()
	defer cleanupEmail(t, db, staleId)

	staleTime := time.Now().Add(-35 * time.Minute)
	_, err := db.Exec(
		`INSERT INTO emails (id, status, payload_file_path, version, updated_at) VALUES (?, ?, ?, 1, ?)`,
		staleId, StatusIntaking, "/payload/stale.json", staleTime,
	)
	require.NoError(t, err)

	// Insert a recent email
	recentId := uuid.NewString()
	defer cleanupEmail(t, db, recentId)

	err = sut.Insert(ctx, recentId, "/payload/recent.json")
	require.NoError(t, err)

	// Get stale emails
	staleEmails, err := sut.GetStaleEmails(ctx)
	require.NoError(t, err)

	// Should only find the stale one
	found := false
	for _, e := range staleEmails {
		if e.Id == staleId {
			found = true
			require.Equal(t, StatusIntaking, e.Status)
		}
		require.NotEqual(t, recentId, e.Id, "recent email should not be in stale list")
	}
	require.True(t, found, "stale email should be found")
}

func TestGetInvalidEmails(t *testing.T) {
	if testing.Short() {
		t.Skip("component tests are skipped in short mode")
	}

	db := getTestDB(t)
	defer db.Close()

	sut := NewDatabase(db, 30)
	ctx := context.TODO()

	// Insert an invalid email directly
	invalidId := uuid.NewString()
	defer cleanupEmail(t, db, invalidId)

	_, err := db.Exec(
		`INSERT INTO emails (id, status, payload_file_path, reason, version) VALUES (?, ?, ?, ?, 1)`,
		invalidId, StatusInvalid, "/payload/invalid.json", "Invalid email format",
	)
	require.NoError(t, err)

	// Get invalid emails
	invalidEmails, err := sut.GetInvalidEmails(ctx)
	require.NoError(t, err)

	// Should find the invalid one
	found := false
	for _, e := range invalidEmails {
		if e.Id == invalidId {
			found = true
			require.Equal(t, StatusInvalid, e.Status)
			require.Equal(t, "Invalid email format", e.ErrorMessage)
		}
	}
	require.True(t, found, "invalid email should be found")
}

func TestRequeueEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("component tests are skipped in short mode")
	}

	db := getTestDB(t)
	defer db.Close()

	sut := NewDatabase(db, 30)
	ctx := context.TODO()

	// Test cases: states that can be requeued and their expected new status
	testCases := map[string]string{
		StatusIntaking:              StatusAccepted,
		StatusProcessing:            StatusReady,
		StatusCallingSentCallback:   StatusSent,
		StatusCallingFailedCallback: StatusFailed,
	}

	for currentStatus, expectedNewStatus := range testCases {
		t.Run(currentStatus, func(t *testing.T) {
			id := uuid.NewString()
			defer cleanupEmail(t, db, id)

			// Insert with specific status
			_, err := db.Exec(
				`INSERT INTO emails (id, status, payload_file_path, version) VALUES (?, ?, ?, 1)`,
				id, currentStatus, "/payload/test.json",
			)
			require.NoError(t, err)

			// Requeue
			err = sut.RequeueEmail(ctx, id)
			require.NoError(t, err)

			// Verify new status
			var newStatus string
			var version int
			err = db.QueryRow("SELECT status, version FROM emails WHERE id = ?", id).Scan(&newStatus, &version)
			require.NoError(t, err)
			require.Equal(t, expectedNewStatus, newStatus)
			require.Equal(t, 2, version, "version should be incremented")

			// Verify status history
			var historyCount int
			err = db.QueryRow("SELECT COUNT(*) FROM email_statuses WHERE email_id = ?", id).Scan(&historyCount)
			require.NoError(t, err)
			require.Equal(t, 1, historyCount, "should have 1 status history entry")
		})
	}
}

func TestRequeueEmailErrorsForNonRequeuableStatuses(t *testing.T) {
	if testing.Short() {
		t.Skip("component tests are skipped in short mode")
	}

	db := getTestDB(t)
	defer db.Close()

	sut := NewDatabase(db, 30)
	ctx := context.TODO()

	// States that should NOT be requeuable
	nonRequeuableStatuses := []string{
		StatusAccepted,
		StatusReady,
		StatusSent,
		StatusFailed,
		StatusInvalid,
		StatusSentAcknowledged,
		StatusFailedAcknowledged,
	}

	for _, status := range nonRequeuableStatuses {
		t.Run(status, func(t *testing.T) {
			id := uuid.NewString()
			defer cleanupEmail(t, db, id)

			// Insert with specific status
			_, err := db.Exec(
				`INSERT INTO emails (id, status, payload_file_path, version) VALUES (?, ?, ?, 1)`,
				id, status, "/payload/test.json",
			)
			require.NoError(t, err)

			// Requeue should fail
			err = sut.RequeueEmail(ctx, id)
			require.Error(t, err)
			require.Contains(t, err.Error(), "cannot requeue email with status")

			// Verify status unchanged
			var currentStatus string
			err = db.QueryRow("SELECT status FROM emails WHERE id = ?", id).Scan(&currentStatus)
			require.NoError(t, err)
			require.Equal(t, status, currentStatus)
		})
	}
}

func TestRequeueEmailNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("component tests are skipped in short mode")
	}

	db := getTestDB(t)
	defer db.Close()

	sut := NewDatabase(db, 30)
	ctx := context.TODO()

	err := sut.RequeueEmail(ctx, "non-existent-id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}
