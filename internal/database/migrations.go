package database

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS owners (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			callsign TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			grid_locator TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			club TEXT NOT NULL DEFAULT '',
			address TEXT,
			operators TEXT NOT NULL DEFAULT '',
			contest TEXT NOT NULL DEFAULT '',
			category_operator TEXT NOT NULL DEFAULT '',
			category_band TEXT NOT NULL DEFAULT '',
			category_mode TEXT NOT NULL DEFAULT '',
			category_power TEXT NOT NULL DEFAULT '',
			category_station TEXT NOT NULL DEFAULT '',
			category_assisted TEXT NOT NULL DEFAULT ''
		)`,

		`CREATE TABLE IF NOT EXISTS qso_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			freq TEXT NOT NULL,
			mode TEXT NOT NULL,
			date TEXT NOT NULL,
			time TEXT NOT NULL,
			sent_call TEXT NOT NULL,
			sent_rst TEXT NOT NULL DEFAULT '59',
			sent_exch TEXT NOT NULL DEFAULT '',
			rcvd_call TEXT NOT NULL,
			rcvd_rst TEXT NOT NULL DEFAULT '59',
			rcvd_exch TEXT NOT NULL DEFAULT '',
			transmitter_id INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE INDEX IF NOT EXISTS idx_date_time ON qso_logs (date, time)`,
		`CREATE INDEX IF NOT EXISTS idx_sent_call ON qso_logs (sent_call)`,
		`CREATE INDEX IF NOT EXISTS idx_rcvd_call ON qso_logs (rcvd_call)`,

		`INSERT OR IGNORE INTO owners (id, callsign) VALUES (1, '')`,

		`CREATE TABLE IF NOT EXISTS contests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT '',
			sent_exch TEXT NOT NULL DEFAULT '',
			rcvd_exch TEXT NOT NULL DEFAULT '',
			freq TEXT NOT NULL DEFAULT '',
			mode TEXT NOT NULL DEFAULT '',
			category_operator TEXT NOT NULL DEFAULT '',
			category_band TEXT NOT NULL DEFAULT '',
			category_mode TEXT NOT NULL DEFAULT '',
			category_power TEXT NOT NULL DEFAULT '',
			category_station TEXT NOT NULL DEFAULT '',
			category_assisted TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	alters := []string{
		`ALTER TABLE qso_logs ADD COLUMN contest_id INTEGER DEFAULT NULL`,
		`ALTER TABLE owners ADD COLUMN timezone TEXT NOT NULL DEFAULT 'UTC'`,
		`ALTER TABLE owners ADD COLUMN language TEXT NOT NULL DEFAULT 'en'`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}

	for _, q := range alters {
		db.Exec(q) // ignore "duplicate column" errors
	}

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_contest_id ON qso_logs (contest_id)`)

	migrateContestData(db)

	return nil
}

func migrateContestData(db *sqlx.DB) {
	var count int
	db.Get(&count, "SELECT COUNT(*) FROM contests")
	if count > 0 {
		return
	}

	var contestName sql.NullString
	db.Get(&contestName, "SELECT contest FROM owners WHERE id = 1")
	if !contestName.Valid || contestName.String == "" {
		return
	}

	type ownerContest struct {
		Contest          string `db:"contest"`
		CategoryOperator string `db:"category_operator"`
		CategoryBand     string `db:"category_band"`
		CategoryMode     string `db:"category_mode"`
		CategoryPower    string `db:"category_power"`
		CategoryStation  string `db:"category_station"`
		CategoryAssisted string `db:"category_assisted"`
	}
	var oc ownerContest
	if err := db.Get(&oc, "SELECT contest, category_operator, category_band, category_mode, category_power, category_station, category_assisted FROM owners WHERE id = 1"); err != nil {
		return
	}

	result, err := db.Exec(`INSERT INTO contests (name, category_operator, category_band, category_mode, category_power, category_station, category_assisted) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		oc.Contest, oc.CategoryOperator, oc.CategoryBand, oc.CategoryMode, oc.CategoryPower, oc.CategoryStation, oc.CategoryAssisted)
	if err != nil {
		return
	}

	contestID, _ := result.LastInsertId()
	db.Exec("UPDATE qso_logs SET contest_id = ? WHERE contest_id IS NULL", contestID)
}
