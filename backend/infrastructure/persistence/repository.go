// Package persistence provides database implementations of the domain repository interface.
package persistence

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/simon0-o/offline_me/backend/domain"
)

// SQLiteStore handles database operations
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store and initializes the database
// Returns domain.Repository interface for dependency inversion
func NewSQLiteStore(dbPath string) (domain.Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	store := &SQLiteStore{db: db}
	if err := store.initTables(); err != nil {
		return nil, err
	}

	return store, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// initTables creates the required database tables
func (s *SQLiteStore) initTables() error {
	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS work_sessions (
		id TEXT PRIMARY KEY,
		date TEXT NOT NULL,
		check_in DATETIME NOT NULL,
		check_out DATETIME,
		work_hours INTEGER NOT NULL
	);`

	createConfigTable := `
	CREATE TABLE IF NOT EXISTS work_config (
		id TEXT PRIMARY KEY,
		default_work_hours INTEGER NOT NULL,
		check_in_api_url TEXT DEFAULT '',
		auto_fetch_enabled BOOLEAN DEFAULT FALSE,
		authorization TEXT DEFAULT '',
		p_auth TEXT DEFAULT '',
		p_rtoken TEXT DEFAULT '',
		check_in_webhook_url TEXT DEFAULT '',
		check_out_webhook_url TEXT DEFAULT ''
	);`

	if _, err := s.db.Exec(createSessionsTable); err != nil {
		return err
	}

	if _, err := s.db.Exec(createConfigTable); err != nil {
		return err
	}

	// Migrate existing table to add new columns if they don't exist
	if err := s.migrateConfigTable(); err != nil {
		return err
	}

	// Insert default config if not exists
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO work_config (
			id, default_work_hours, check_in_api_url, auto_fetch_enabled,
			authorization, p_auth, p_rtoken, check_in_webhook_url, check_out_webhook_url
		) VALUES ('default', ?, '', FALSE, '', '', '', '', '')
	`, domain.StandardWorkMinutes)

	return err
}

// migrateConfigTable adds missing columns to the config table
func (s *SQLiteStore) migrateConfigTable() error {
	rows, err := s.db.Query("PRAGMA table_info(work_config)")
	if err != nil {
		return err
	}
	defer rows.Close()

	existingColumns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue *string
		var pk int

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		existingColumns[name] = true
	}

	// Add missing columns
	if !existingColumns["check_in_webhook_url"] {
		if _, err := s.db.Exec("ALTER TABLE work_config ADD COLUMN check_in_webhook_url TEXT DEFAULT ''"); err != nil {
			return err
		}
	}

	if !existingColumns["check_out_webhook_url"] {
		if _, err := s.db.Exec("ALTER TABLE work_config ADD COLUMN check_out_webhook_url TEXT DEFAULT ''"); err != nil {
			return err
		}
	}

	return nil
}

// GetTodaySession retrieves the work session for a specific date
func (s *SQLiteStore) GetTodaySession(date string) *domain.WorkSession {
	var session domain.WorkSession
	var checkOut sql.NullTime

	row := s.db.QueryRow(`
		SELECT id, date, check_in, check_out, work_hours
		FROM work_sessions
		WHERE date = ?
	`, date)

	err := row.Scan(&session.ID, &session.Date, &session.CheckIn, &checkOut, &session.WorkHours)
	if err != nil {
		return nil
	}

	if checkOut.Valid {
		session.CheckOut = &checkOut.Time
	}

	return &session
}

// GetSessionsByMonth retrieves all work sessions for a specific month (YYYY-MM format)
func (s *SQLiteStore) GetSessionsByMonth(yearMonth string) ([]*domain.WorkSession, error) {
	rows, err := s.db.Query(`
		SELECT id, date, check_in, check_out, work_hours
		FROM work_sessions
		WHERE date LIKE ?
		ORDER BY date ASC
	`, yearMonth+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.WorkSession
	for rows.Next() {
		var session domain.WorkSession
		var checkOut sql.NullTime

		err := rows.Scan(&session.ID, &session.Date, &session.CheckIn, &checkOut, &session.WorkHours)
		if err != nil {
			return nil, err
		}

		if checkOut.Valid {
			session.CheckOut = &checkOut.Time
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// SaveSession saves or updates a work session
func (s *SQLiteStore) SaveSession(session *domain.WorkSession) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO work_sessions (id, date, check_in, check_out, work_hours)
		VALUES (?, ?, ?, ?, ?)
	`, session.ID, session.Date, session.CheckIn, session.CheckOut, session.WorkHours)
	return err
}

// GetConfig retrieves the work configuration
func (s *SQLiteStore) GetConfig() (*domain.WorkConfig, error) {
	var config domain.WorkConfig
	row := s.db.QueryRow(`
		SELECT id, default_work_hours, check_in_api_url, auto_fetch_enabled,
		       p_auth, p_rtoken, check_in_webhook_url, check_out_webhook_url
		FROM work_config
		WHERE id = 'default'
	`)

	err := row.Scan(
		&config.ID,
		&config.DefaultWorkHours,
		&config.CheckInAPIURL,
		&config.AutoFetchEnabled,
		&config.PAuth,
		&config.PRToken,
		&config.CheckInWebhookURL,
		&config.CheckOutWebhookURL,
	)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves or updates the work configuration
func (s *SQLiteStore) SaveConfig(config *domain.WorkConfig) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO work_config (
			id, default_work_hours, check_in_api_url, auto_fetch_enabled,
			p_auth, p_rtoken, check_in_webhook_url, check_out_webhook_url
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		config.ID,
		config.DefaultWorkHours,
		config.CheckInAPIURL,
		config.AutoFetchEnabled,
		config.PAuth,
		config.PRToken,
		config.CheckInWebhookURL,
		config.CheckOutWebhookURL,
	)
	return err
}
