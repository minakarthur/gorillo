package database

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS owners (
			id INT PRIMARY KEY AUTO_INCREMENT,
			callsign VARCHAR(20) NOT NULL DEFAULT '',
			name VARCHAR(75) NOT NULL DEFAULT '',
			email VARCHAR(255) NOT NULL DEFAULT '',
			grid_locator VARCHAR(8) NOT NULL DEFAULT '',
			location VARCHAR(50) NOT NULL DEFAULT '',
			club VARCHAR(75) NOT NULL DEFAULT '',
			address TEXT,
			operators VARCHAR(255) NOT NULL DEFAULT '',
			contest VARCHAR(32) NOT NULL DEFAULT '',
			category_operator VARCHAR(20) NOT NULL DEFAULT '',
			category_band VARCHAR(20) NOT NULL DEFAULT '',
			category_mode VARCHAR(10) NOT NULL DEFAULT '',
			category_power VARCHAR(10) NOT NULL DEFAULT '',
			category_station VARCHAR(20) NOT NULL DEFAULT '',
			category_assisted VARCHAR(20) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS qso_logs (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			freq VARCHAR(10) NOT NULL,
			mode VARCHAR(5) NOT NULL,
			date DATE NOT NULL,
			time VARCHAR(4) NOT NULL,
			sent_call VARCHAR(20) NOT NULL,
			sent_rst VARCHAR(5) NOT NULL DEFAULT '59',
			sent_exch VARCHAR(50) NOT NULL DEFAULT '',
			rcvd_call VARCHAR(20) NOT NULL,
			rcvd_rst VARCHAR(5) NOT NULL DEFAULT '59',
			rcvd_exch VARCHAR(50) NOT NULL DEFAULT '',
			transmitter_id TINYINT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_date_time (date, time),
			INDEX idx_sent_call (sent_call),
			INDEX idx_rcvd_call (rcvd_call)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`INSERT IGNORE INTO owners (id, callsign) VALUES (1, '')`,

		`CREATE TABLE IF NOT EXISTS contests (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL DEFAULT '',
			sent_exch VARCHAR(50) NOT NULL DEFAULT '',
			rcvd_exch VARCHAR(50) NOT NULL DEFAULT '',
			freq VARCHAR(10) NOT NULL DEFAULT '',
			mode VARCHAR(5) NOT NULL DEFAULT '',
			category_operator VARCHAR(20) NOT NULL DEFAULT '',
			category_band VARCHAR(20) NOT NULL DEFAULT '',
			category_mode VARCHAR(10) NOT NULL DEFAULT '',
			category_power VARCHAR(10) NOT NULL DEFAULT '',
			category_station VARCHAR(20) NOT NULL DEFAULT '',
			category_assisted VARCHAR(20) NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	alters := []string{
		`ALTER TABLE qso_logs ADD COLUMN contest_id INT DEFAULT NULL`,
		`ALTER TABLE qso_logs ADD INDEX idx_contest_id (contest_id)`,
		`ALTER TABLE owners ADD COLUMN timezone VARCHAR(50) NOT NULL DEFAULT 'UTC'`,
		`ALTER TABLE owners ADD COLUMN language VARCHAR(5) NOT NULL DEFAULT 'en'`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}

	for _, q := range alters {
		db.Exec(q) // ignore "duplicate column" errors
	}

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
