package models

import "time"

type QSO struct {
	ID            int64     `db:"id"`
	Freq          string    `db:"freq"`
	Mode          string    `db:"mode"`
	Date          time.Time `db:"date"`
	Time          string    `db:"time"`
	SentCall      string    `db:"sent_call"`
	SentRST       string    `db:"sent_rst"`
	SentExch      string    `db:"sent_exch"`
	RcvdCall      string    `db:"rcvd_call"`
	RcvdRST       string    `db:"rcvd_rst"`
	RcvdExch      string    `db:"rcvd_exch"`
	TransmitterID int  `db:"transmitter_id"`
	ContestID     *int `db:"contest_id"`
}
