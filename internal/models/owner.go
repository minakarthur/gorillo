package models

type Owner struct {
	ID                int    `db:"id"`
	Callsign          string `db:"callsign"`
	Name              string `db:"name"`
	Email             string `db:"email"`
	GridLocator       string `db:"grid_locator"`
	Location          string `db:"location"`
	Club              string `db:"club"`
	Address           string `db:"address"`
	Operators string `db:"operators"`
	Timezone  string `db:"timezone"`
	Language  string `db:"language"`
}
