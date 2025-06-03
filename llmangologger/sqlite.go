package llmangologger

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver

	"github.com/llmang/llmango/llmango"
)

// SetupDB creates the logging table if it doesn't exist
func setupSQLiteDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS mango_logs (
			timestamp INTEGER DEFAULT (strftime('%s', 'now')),
			goal_uid TEXT NOT NULL DEFAULT '',
			prompt_uid TEXT NOT NULL DEFAULT '',
			raw_request TEXT NOT NULL DEFAULT '',
			input_object TEXT NOT NULL DEFAULT '',
			raw_response TEXT NOT NULL DEFAULT '',
			output_object TEXT NOT NULL DEFAULT '',
			input_tokens INTEGER NOT NULL DEFAULT 0,
			output_tokens INTEGER NOT NULL DEFAULT 0,
			cost REAL NOT NULL DEFAULT 0.0,
			request_time REAL NOT NULL DEFAULT 0.0,
			generation_time REAL NOT NULL DEFAULT 0.0,
			error TEXT NOT NULL DEFAULT ''
		);
	`)
	return err
}

// LogObject inserts a LogObject into the database
func sqlite3LogObject(db *sql.DB, logObj *llmango.LLMangoLog) error {
	_, err := db.Exec(`
		INSERT INTO mango_logs (
			timestamp, goal_uid, prompt_uid, raw_request, input_object,
			raw_response, output_object, input_tokens, output_tokens,
			cost, request_time, generation_time, error
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		logObj.Timestamp,
		logObj.GoalUID,
		logObj.PromptUID,
		logObj.RawRequest,
		logObj.InputObject,
		logObj.RawResponse,
		logObj.OutputObject,
		logObj.InputTokens,
		logObj.OutputTokens,
		logObj.Cost,
		logObj.RequestTime,
		logObj.GenerationTime,
		logObj.Error,
	)
	return err
}

// GetLogs retrieves logs from the database based on the provided filters
func sqlite3GetLogs(db *sql.DB, filter *llmango.LLmangoLogFilter) ([]llmango.LLMangoLog, int, error) {
	// Start building the query using snake_case columns
	query := "SELECT timestamp, goal_uid, prompt_uid"

	countQuery := "SELECT count(*) FROM mango_logs WHERE 1=1"

	// Only include raw fields if requested, using snake_case columns
	if filter.IncludeRaw {
		query += ", raw_request, input_object, raw_response, output_object"
	} else {
		query += ", '' as raw_request, input_object, '' as raw_response, output_object"
	}

	// Add remaining fields using snake_case columns
	query += ", input_tokens, output_tokens, cost, request_time, generation_time, error FROM mango_logs WHERE 1=1"

	// Add filter conditions using snake_case columns
	var args []interface{}
	var countArgs []interface{}

	if filter.MinTimestamp != nil {
		query += " AND timestamp >= ?"
		countQuery += " AND timestamp >= ?"
		args = append(args, *filter.MinTimestamp)
		countArgs = append(countArgs, *filter.MinTimestamp)
	}

	if filter.MaxTimestamp != nil {
		query += " AND timestamp <= ?"
		countQuery += " AND timestamp <= ?"
		args = append(args, *filter.MaxTimestamp)
		countArgs = append(countArgs, *filter.MaxTimestamp)
	}

	if filter.GoalUID != nil {
		query += " AND goal_uid = ?"
		countQuery += " AND goal_uid = ?"
		args = append(args, *filter.GoalUID)
		countArgs = append(countArgs, *filter.GoalUID)
	}

	if filter.PromptUID != nil {
		query += " AND prompt_uid = ?"
		countQuery += " AND prompt_uid = ?"
		args = append(args, *filter.PromptUID)
		countArgs = append(countArgs, *filter.PromptUID)
	}

	// Add order by, limit and offset
	query += " ORDER BY timestamp DESC"

	if filter.Limit != nil {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset != nil {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Process results - Scan targets remain the same (Go struct fields)
	var logs []llmango.LLMangoLog
	for rows.Next() {
		var log llmango.LLMangoLog
		err := rows.Scan(
			&log.Timestamp,
			&log.GoalUID,
			&log.PromptUID,
			&log.RawRequest,
			&log.InputObject,
			&log.RawResponse,
			&log.OutputObject,
			&log.InputTokens,
			&log.OutputTokens,
			&log.Cost,
			&log.RequestTime,
			&log.GenerationTime,
			&log.Error,
		)
		if err != nil {
			return logs, 0, fmt.Errorf("error scanning log row: %w", err)
		}
		logs = append(logs, log)
	}
	if err = rows.Err(); err != nil {
		return logs, 0, fmt.Errorf("error iterating log rows: %w", err)
	}

	// Get total count matching filters (using snake_case columns)
	var totalCount int
	err = db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return logs, 0, fmt.Errorf("error getting total log count: %w", err)
	}

	return logs, totalCount, nil
}

func UseSQLiteLogging(m *llmango.LLMangoManager, db *sql.DB, opts *MangoLoggingOptions) error {
	if m == nil {
		return errors.New("failed to setup logging as the llmangomanger was nil")
	}
	if db == nil {
		return errors.New("failed to setup logging as the database connection was nil")
	}

	err := setupSQLiteDB(db)
	if err != nil {
		return fmt.Errorf("failed to create logging table: %w", err)
	}

	if m.Logging == nil {
		m.Logging = &llmango.Logging{}
	}

	m.Logging.LogResponse = func(mangolog *llmango.LLMangoLog) error {
		logToInsert := *mangolog
		if opts == nil || !opts.LogRawRequestResponse {
			logToInsert.RawRequest = ""
			logToInsert.RawResponse = ""
		}
		return sqlite3LogObject(db, &logToInsert)
	}

	m.Logging.GetLogs = func(filter *llmango.LLmangoLogFilter) ([]llmango.LLMangoLog, int, error) {
		return sqlite3GetLogs(db, filter)
	}

	return nil
}

// CreateSQLiteLogger creates a SQLite logger that can be used with WithLogging()
func CreateSQLiteLogger(db *sql.DB, logFullRequests bool) (*llmango.Logging, error) {
	if db == nil {
		return nil, errors.New("database connection cannot be nil")
	}

	err := setupSQLiteDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging table: %w", err)
	}

	return &llmango.Logging{
		LogResponse: func(mangolog *llmango.LLMangoLog) error {
			logToInsert := *mangolog
			if !logFullRequests {
				logToInsert.RawRequest = ""
				logToInsert.RawResponse = ""
			}
			return sqlite3LogObject(db, &logToInsert)
		},
		GetLogs: func(filter *llmango.LLmangoLogFilter) ([]llmango.LLMangoLog, int, error) {
			return sqlite3GetLogs(db, filter)
		},
	}, nil
}
