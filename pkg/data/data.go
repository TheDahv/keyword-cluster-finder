package data

import (
	"database/sql"
	"fmt"

	// lib/pg lets us communicate with Postgres databases
	_ "github.com/lib/pq"
)

const query = `
	WITH domain_params AS (
		SELECT *, keywords.name AS keyword
		FROM v_serp_params
		JOIN keywords USING (keyword_id)
		WHERE domain_id = $1
		AND date IS NOT NULL
		AND name = $2
    -- TODO add filters for markets
	), serp_competitors AS (
		SELECT
		keyword,
		domain AS competitor,
		max(date) AS date,
		round(avg(avg_rank), 1) AS avg_rank,
		f_wilson(array_remove(array_agg(avg_rank), NULL)) AS wilson,
		array_agg(avg_rank ORDER BY keyword_id, market_id) AS ranks
		FROM domain_params
		JOIN multisample_rankings mr USING (domain_id, keyword_id, market_id, date)
		WHERE avg_rank <= 20
		GROUP BY keyword, competitor
		ORDER BY wilson DESC NULLS LAST
		LIMIT 20
	)
	SELECT
		keyword,
		(row_number() OVER w)::INT AS prominence,
		a.competitor
	FROM serp_competitors a
	WINDOW w AS (ORDER BY a.wilson DESC NULLS LAST)
`

// Driver manages connection to the data store
type Driver struct {
	User        string
	Password    string
	Host        string
	Database    string
	db          *sql.DB
	maxInFlight int
}

// Option configures a Driver
type Option func(*Driver)

// WithUserAndPass configures the driver with the credentials for the connection
func WithUserAndPass(user string, pass string) Option {
	return func(d *Driver) {
		d.User = user
		d.Password = pass
	}
}

// WithHost configures the driver with the connection host
func WithHost(host string) Option {
	return func(d *Driver) {
		d.Host = host
	}
}

// WithDatabase configures the driver with the database host
func WithDatabase(database string) Option {
	return func(d *Driver) {
		d.Database = database
	}
}

// WithMaxInFlight configures the maximum amount of queries to run at once when collecting data
func WithMaxInFlight(max int) Option {
	return func(d *Driver) {
		d.maxInFlight = max
	}
}

// New sets up a database driver
func New(options ...Option) (*Driver, error) {
	d := &Driver{maxInFlight: 5}
	for _, opt := range options {
		opt(d)
	}

	return d, nil
}

// withConn opens a connection to the database for the given operation and
// closes it when it is finished
func (d Driver) withConn(operation func(*sql.DB) error) error {
	db, err := sql.Open("postgres", d.connString())
	if err != nil {
		return fmt.Errorf("could not open database: %v", err)
	}
	defer db.Close()

	return operation(db)
}

// connString creates the connection string from the driver configuration
func (d Driver) connString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		d.User,
		d.Password,
		d.Host,
		d.Database,
	)
}

// FetchKeywords loads the keywords for a given domain
func (d Driver) FetchKeywords(domainID int) ([]string, error) {
	var keywords []string

	err := d.withConn(func(db *sql.DB) error {
		rows, err := db.Query(`
			SELECT DISTINCT name AS keyword
			FROM v_serp_params
			JOIN keywords USING (keyword_id)
			WHERE domain_id = $1
		`, domainID)
		if err != nil {
			return fmt.Errorf("could not query the database: %v", err)
		}

		for rows.Next() {
			var kw string
			err := rows.Scan(&kw)
			if err != nil {
				return fmt.Errorf("could not parse keyword result: %v", err)
			}
			keywords = append(keywords, kw)
		}

		return nil
	})

	return keywords, err
}

// FetchSERP loads prominent SERP members for a given keyword
func (d Driver) FetchSERP(domainID int, keyword string, eachRow func(*sql.Rows) error) error {
	return d.withConn(func(db *sql.DB) error {
		rows, err := db.Query(query, domainID, keyword)
		if err != nil {
			return fmt.Errorf("could not query database: %v", err)
		}

		for rows.Next() {
			err := eachRow(rows)
			if err != nil {
				return fmt.Errorf("error parsing row: %v", err)
			}
		}

		return nil
	})
}
