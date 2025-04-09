package llmangologger

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/llmang/llmango/llmango"
)

// SetupDB creates the logging table if it doesn't exist
func setupSQLiteDB(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS mango_logs (
			timestamp INTEGER DEFAULT (strftime('%s', 'now')),
			goalUID TEXT NOT NULL DEFAULT '',
			promptUID TEXT NOT NULL DEFAULT '',
			RawInput TEXT NOT NULL DEFAULT '',
			InputObject TEXT NOT NULL DEFAULT '',
			RawOutput TEXT NOT NULL DEFAULT '',
			OutputObject TEXT NOT NULL DEFAULT '',
			InputTokens INTEGER NOT NULL DEFAULT 0,
			OutputTokens INTEGER NOT NULL DEFAULT 0,
			Cost REAL NOT NULL DEFAULT 0.0,
			RequestTime REAL NOT NULL DEFAULT 0.0,
			Error TEXT NOT NULL DEFAULT ''
		);
	`)
	return err
}

// LogObject inserts a LogObject into the database
func sqlite3LogObject(db *sql.DB, logObj *llmango.LLMangoLog) error {
	_, err := db.Exec(`
		INSERT INTO mango_logs (
			timestamp, goalUID, promptUID, rawInput, inputObject, 
			rawOutput, outputObject, inputTokens, outputTokens, 
			cost, requestTime, error
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		logObj.Timestamp,
		logObj.GoalUID,
		logObj.PromptUID,
		logObj.RawInput,
		logObj.InputObject,
		logObj.RawOutput,
		logObj.OutputObject,
		logObj.InputTokens,
		logObj.OutputTokens,
		logObj.Cost,
		logObj.RequestTime,
		logObj.Error,
	)
	return err
}

// GetLogs retrieves logs from the database based on the provided filters
func sqlite3GetLogs(db *sql.DB, filter *llmango.LLmangoLogFilter) ([]llmango.LLMangoLog, int, error) {
	// Start building the query
	query := "SELECT timestamp, goalUID, promptUID"

	countQuery := "SELECT count(*) FROM mango_logs WHERE 1=1"

	// Only include raw fields if requested
	if filter.IncludeRaw {
		query += ", rawInput, inputObject, rawOutput, outputObject"
	} else {
		query += ", '' as rawInput, inputObject, '' as rawOutput, outputObject"
	}

	// Add remaining fields
	query += ", inputTokens, outputTokens, cost, requestTime, error FROM mango_logs WHERE 1=1"

	// Add filter conditions
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
		query += " AND goalUID = ?"
		countQuery += " AND goalUID = ?"
		args = append(args, *filter.GoalUID)
		countArgs = append(countArgs, *filter.GoalUID)
	}

	if filter.PromptUID != nil {
		query += " AND promptUID = ?"
		countQuery += " AND promptUID = ?"
		args = append(args, *filter.PromptUID)
		countArgs = append(countArgs, *filter.PromptUID)
	}

	// Add order by, limit and offset
	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	} else {
		query += " LIMIT 100" // Default limit
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Process results
	var logs []llmango.LLMangoLog
	for rows.Next() {
		var log llmango.LLMangoLog
		err := rows.Scan(
			&log.Timestamp,
			&log.GoalUID,
			&log.PromptUID,
			&log.RawInput,
			&log.InputObject,
			&log.RawOutput,
			&log.OutputObject,
			&log.InputTokens,
			&log.OutputTokens,
			&log.Cost,
			&log.RequestTime,
			&log.Error,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	rows.Close()
	var totalCount int
	err = db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	return logs, totalCount, err
}

func UseSQLiteLogging(m *llmango.LLMangoManager, db *sql.DB, opts *MangoLoggingOptions) error {
	if m == nil {
		return errors.New("failed to setup logging as the llmangomanger was nil")
	}
	err := setupSQLiteDB(db)
	if err != nil {
		return fmt.Errorf("failed to create logging table: %w", err)
	}
	//setup default logger and getter
	if m.Logging == nil {
		m.Logging = &llmango.Logging{}
	}
	m.LogResponse = func(mangolog *llmango.LLMangoLog) error {
		return sqlite3LogObject(db, mangolog)
	}

	m.GetLogs = func(filter *llmango.LLmangoLogFilter) ([]llmango.LLMangoLog, int, error) {
		return sqlite3GetLogs(db, filter)
	}
	m.LogPercentage = opts.LogPercentage
	m.LogFullInputOutputMessages = opts.LogFullInputOutputMessages

	return nil
}
