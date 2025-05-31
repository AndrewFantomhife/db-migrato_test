package goose

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"time"
)

const (
	redColor         = "\033[31m"
	grayColor        = "\033[90m"
	resetColor       = "\033[0m"
	migrationTimeout = 30 * time.Second
	statementTimeout = 5 * time.Second
)

// Run a migration specified in raw SQL.
//
// Sections of the script can be annotated with a special comment,
// starting with "-- +goose" to specify whether the section should
// be applied during an Up or Down migration
//
// All statements following an Up or Down annotation are grouped together
// until another direction annotation is found.

func runSQLMigration(
	ctx context.Context,
	db *sql.DB,
	statements []string,
	useTx bool,
	v int64,
	direction bool,
	noVersioning bool,
	timeout time.Duration,
) error {
	ctx = withTimeoutContext(ctx, timeout)
	if useTx {
		return runSQLMigrationInTransaction(ctx, db, statements, v, direction, noVersioning)
	}
	return runSQLMigrationNoTransaction(ctx, db, statements, v, direction, noVersioning)
}

// withTimeoutContext creates a context with a timeout for the migration.
func withTimeoutContext(parent context.Context, timeout time.Duration) context.Context {
	if timeout == 0 {
		timeout = migrationTimeout
	}
	ctx, _ := context.WithTimeout(parent, timeout)
	return ctx
}

func runSQLMigrationInTransaction(
	ctx context.Context,
	db *sql.DB,
	statements []string,
	v int64,
	direction bool,
	noVersioning bool,
) error {
	verboseInfo("Begin transaction")
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logError("Failed to begin transaction: %v", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	if err := executeSQLStatements(ctx, tx, statements); err != nil {
		logError("Failed to execute SQL statements: %v", err)
		verboseInfo("Rollback transaction")
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logError("Failed to rollback transaction: %v", rollbackErr)
		}
		return err
	}
	if !noVersioning {
		if direction {
			err = store.InsertVersion(ctx, tx, TableName(), v)
		} else {
			err = store.DeleteVersion(ctx, tx, TableName(), v)
		}
		if err != nil {
			msg := fmt.Sprintf("Failed to update migration version: %v", err)
			logError(msg)
			verboseInfo("Rollback transaction")

			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logError("Failed to rollback transaction: %v", rollbackErr)
			}
			return fmt.Errorf("%s: %w", msg, err)
		}
	}
	verboseInfo("Commit transaction")
	if err := tx.Commit(); err != nil {
		logError("Failed to commit transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
func executeSQLStatements(ctx context.Context, execer interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}, statements []string) error {
	for _, stmt := range statements {
		cleaned := clearStatement(stmt)
		verboseInfo("Executing SQL statement: %s", cleaned)

		err := func() error {
			stmtCtx, cancel := context.WithTimeout(ctx, statementTimeout)
			defer cancel()

			_, err := execer.ExecContext(stmtCtx, stmt)
			return err
		}()

		if err != nil {
			msg := fmt.Sprintf("failed to execute SQL query %q: %v", cleaned, err)
			if err == context.DeadlineExceeded {
				logError("SQL statement execution timed out: %s", cleaned)
			} else {
				logError(msg)
			}
			return fmt.Errorf("%s: %w", msg, err)
		}
	}
	return nil
}
func runSQLMigrationNoTransaction(
	ctx context.Context,
	db *sql.DB,
	statements []string,
	v int64,
	direction bool,
	noVersioning bool,
) error {
	if err := executeStatements(ctx, db, statements); err != nil {
		logError("Failed to execute non-transactional SQL statements: %v", err)
		return err
	}

	if !noVersioning {
		var err error
		if direction {
			err = store.InsertVersionNoTx(ctx, db, TableName(), v)
		} else {
			err = store.DeleteVersionNoTx(ctx, db, TableName(), v)
		}
		if err != nil {
			return fmt.Errorf("failed to update version: %w", err)
		}
	}

	return nil
}

func verboseInfo(s string, args ...interface{}) {
	if verbose {
		if noColor {
			log.Printf(s, args...)
		} else {
			log.Printf(grayColor+s+resetColor, args...)
		}
	}
}

func logError(s string, args ...interface{}) {
	msg := fmt.Sprintf(s, args...)
	log.Printf(redColor + "ERROR " + resetColor + msg)
}

var (
	matchSQLComments = regexp.MustCompile(`(?m)^--.*$[\r\n]*`)
	matchEmptyEOL    = regexp.MustCompile(`(?m)^$[\r\n]*`) // TODO: Duplicate
)

func clearStatement(s string) string {
	s = matchSQLComments.ReplaceAllString(s, ``)
	return matchEmptyEOL.ReplaceAllString(s, ``)
}
