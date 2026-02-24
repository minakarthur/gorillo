package models

type Contest struct {
	ID               int    `db:"id"`
	Name             string `db:"name"`
	SentExch         string `db:"sent_exch"`
	RcvdExch         string `db:"rcvd_exch"`
	Freq             string `db:"freq"`
	Mode             string `db:"mode"`
	CategoryOperator string `db:"category_operator"`
	CategoryBand     string `db:"category_band"`
	CategoryMode     string `db:"category_mode"`
	CategoryPower    string `db:"category_power"`
	CategoryStation  string `db:"category_station"`
	CategoryAssisted string `db:"category_assisted"`
}
