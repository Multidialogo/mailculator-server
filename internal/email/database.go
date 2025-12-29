package email

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	StatusMeta                  = "_META"
	StatusAccepted              = "ACCEPTED"
	StatusIntaking              = "INTAKING"
	StatusProcessing            = "PROCESSING"
	StatusCallingSentCallback   = "CALLING-SENT-CALLBACK"
	StatusCallingFailedCallback = "CALLING-FAILED-CALLBACK"
	StatusReady                 = "READY"
	StatusSent                  = "SENT"
	StatusFailed                = "FAILED"
	StatusInvalid               = "INVALID"
	StatusSentAcknowledged      = "SENT-ACKNOWLEDGED"
	StatusFailedAcknowledged    = "FAILED-ACKNOWLEDGED"
)

const (
	statusInitial               = StatusAccepted
	statusIntaking              = StatusIntaking
	statusProcessing            = StatusProcessing
	statusCallingSentCallback   = StatusCallingSentCallback
	statusCallingFailedCallback = StatusCallingFailedCallback
	statusReady                 = StatusReady
	statusSent                  = StatusSent
	statusFailed                = StatusFailed
)

// MySQL error codes
const (
	mysqlDuplicateEntryCode = 1062
)

type Database struct {
	db                          *sql.DB
	staleEmailsThresholdMinutes int
}

func NewDatabase(db *sql.DB, staleEmailsThresholdMinutes int) *Database {
	return &Database{
		db:                          db,
		staleEmailsThresholdMinutes: staleEmailsThresholdMinutes,
	}
}

func (d *Database) Insert(ctx context.Context, id string, payloadFilePath string) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert into emails table
	_, err = tx.ExecContext(ctx,
		`INSERT INTO emails (id, status, payload_file_path, version) VALUES (?, ?, ?, 1)`,
		id, statusInitial, payloadFilePath,
	)
	if err != nil {
		return err
	}

	// Insert initial status into email_statuses
	_, err = tx.ExecContext(ctx,
		`INSERT INTO email_statuses (email_id, status) VALUES (?, ?)`,
		id, statusInitial,
	)
	if err != nil {
		return fmt.Errorf("failed to insert status history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *Database) GetStaleEmails(ctx context.Context) ([]Email, error) {
	thresholdTime := time.Now().Add(-time.Duration(d.staleEmailsThresholdMinutes) * time.Minute)

	rows, err := d.db.QueryContext(ctx,
		`SELECT id, status, created_at, updated_at 
		FROM emails 
		WHERE status IN (?, ?, ?, ?) 
		AND updated_at < ?`,
		statusIntaking,
		statusProcessing,
		statusCallingSentCallback,
		statusCallingFailedCallback,
		thresholdTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query stale emails: %w", err)
	}
	defer rows.Close()

	var emails []Email
	for rows.Next() {
		var e Email
		if err := rows.Scan(&e.Id, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan email row: %w", err)
		}
		emails = append(emails, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating email rows: %w", err)
	}

	return emails, nil
}

func (d *Database) GetInvalidEmails(ctx context.Context) ([]Email, error) {
	rows, err := d.db.QueryContext(ctx,
		`SELECT id, status, reason, created_at, updated_at 
		FROM emails 
		WHERE status = ?`,
		StatusInvalid,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query invalid emails: %w", err)
	}
	defer rows.Close()

	var emails []Email
	for rows.Next() {
		var e Email
		var reason sql.NullString
		if err := rows.Scan(&e.Id, &e.Status, &reason, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan email row: %w", err)
		}
		if reason.Valid {
			e.ErrorMessage = reason.String
		}
		emails = append(emails, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating email rows: %w", err)
	}

	return emails, nil
}

func (d *Database) RequeueEmail(ctx context.Context, id string) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current status and version
	var currentStatus string
	var version int
	err = tx.QueryRowContext(ctx,
		`SELECT status, version FROM emails WHERE id = ? FOR UPDATE`,
		id,
	).Scan(&currentStatus, &version)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("email with id %s not found", id)
		}
		return fmt.Errorf("failed to get email: %w", err)
	}

	// Map current status to new status
	var newStatus string
	switch currentStatus {
	case statusIntaking:
		newStatus = statusInitial // ACCEPTED
	case statusProcessing:
		newStatus = statusReady
	case statusCallingSentCallback:
		newStatus = statusSent
	case statusCallingFailedCallback:
		newStatus = statusFailed
	default:
		return fmt.Errorf("cannot requeue email with status: %s", currentStatus)
	}

	// Update email status with optimistic locking
	result, err := tx.ExecContext(ctx,
		`UPDATE emails SET status = ?, version = version + 1 WHERE id = ? AND version = ?`,
		newStatus, id, version,
	)
	if err != nil {
		return fmt.Errorf("failed to update email status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("email was modified by another process")
	}

	// Insert status change into history
	_, err = tx.ExecContext(ctx,
		`INSERT INTO email_statuses (email_id, status, reason) VALUES (?, ?, ?)`,
		id, newStatus, fmt.Sprintf("Requeued from %s", currentStatus),
	)
	if err != nil {
		return fmt.Errorf("failed to insert status history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// IsDuplicateEntryError checks if the error is a MySQL duplicate entry error
func IsDuplicateEntryError(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		return mysqlErr.Number == mysqlDuplicateEntryCode
	}
	return false
}
