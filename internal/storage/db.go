package storage

import (
	"database/sql"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/lib/pq"
	"slices"
	"time"
)

type DBStorage struct {
	Name       string  `json:"id"`
	MetricType string  `json:"type"`
	Value      float64 `json:"delta,omitempty"`
	Delta      int64   `json:"value,omitempty"`
}

var DB *sql.DB
var retryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func OpenDB(addr string) {
	var err error

	DB, err = sql.Open("postgres", addr)
	if err != nil {
		panic(err)
	}
	createTableQuery :=
		`CREATE TABLE IF NOT EXISTS gauge (
                      id SERIAL PRIMARY KEY,
                      name VARCHAR(255) NOT NULL,
                      value DOUBLE PRECISION NOT NULL);
		CREATE TABLE IF NOT EXISTS counter (
                         id SERIAL PRIMARY KEY,
                         name VARCHAR(255) NOT NULL,
                         delta BIGINT NOT NULL);`
	_, err = DB.Exec(createTableQuery)
	if err != nil {
		panic(err)
	}
}

func CheckConnect(db *sql.DB) error {
	err := db.Ping()
	return err
}

func (d *DBStorage) CountStorage(k string, v int64) {
	var count int
	_, _, err := withRetries(func() (*sql.Rows, *sql.Row, error) {
		err := DB.QueryRow("SELECT COUNT(*) FROM counter WHERE name = $1", k).Scan(&count)
		return nil, nil, err
	})
	if err != nil {
		panic(err)
	}
	if count > 0 {
		_, _, err := withRetries(func() (*sql.Rows, *sql.Row, error) {
			_, err := DB.Exec("UPDATE counter SET name=$1, delta=$2 WHERE name=$1", k, d.Delta+v)
			return nil, nil, err
		})
		if err != nil {
			panic(err)
		}
	} else {
		_, _, err := withRetries(func() (*sql.Rows, *sql.Row, error) {
			_, err := DB.Exec("INSERT INTO counter (name, delta) VALUES ($1, $2)", k, v)
			return nil, nil, err
		})
		if err != nil {
			panic(err)
		}
	}
}

func (d *DBStorage) GaugeStorage(k string, v float64) {
	var count int
	_, _, err := withRetries(func() (*sql.Rows, *sql.Row, error) {
		err := DB.QueryRow("SELECT COUNT(*) FROM gauge WHERE name = $1", k).Scan(&count)
		return nil, nil, err
	})
	if err != nil {
		panic(err)
	}
	if count > 0 {
		_, _, err := withRetries(func() (*sql.Rows, *sql.Row, error) {
			_, err := DB.Exec("UPDATE gauge SET name=$1, value=$2 WHERE name=$1", k, v)
			return nil, nil, err
		})
		if err != nil {
			panic(err)
		}
	} else {
		_, _, err := withRetries(func() (*sql.Rows, *sql.Row, error) {
			_, err := DB.Exec("INSERT INTO gauge (name, value) VALUES ($1, $2)", k, v)
			return nil, nil, err
		})
		if err != nil {
			panic(err)
		}
	}
}

func (d *DBStorage) GetStorage() interface{} {
	arrd := []DBStorage{}
	rows, _, err := withRetries(func() (*sql.Rows, *sql.Row, error) {
		rows, err := DB.Query("SELECT name, value FROM gauge")
		return rows, nil, err
	})

	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&d.Name, &d.Value); err != nil {
			panic(err)
		}
		d.MetricType = config.Gauge
		arrd = append(arrd, *d)
	}
	_, row, _ := withRetries(func() (*sql.Rows, *sql.Row, error) {
		row := DB.QueryRow("SELECT name, delta FROM counter")
		return nil, row, nil
	})
	if err := row.Scan(&d.Name, &d.Delta); err != nil {
		panic(err)
	} else {
		d.MetricType = config.Counter
		arrd = append(arrd, *d)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return arrd
}

func (d *DBStorage) SetStartData(storage MemStorage) {

}

func (d *DBStorage) GetGauge(key string) (float64, bool) {
	var exists bool
	var count int
	_, _, err := withRetries(func() (*sql.Rows, *sql.Row, error) {
		err := DB.QueryRow("SELECT COUNT(*) FROM gauge WHERE name = $1", key).Scan(&count)
		return nil, nil, err
	})
	if err != nil {
		panic(err)
	}
	if count > 0 {
		_, row, _ := withRetries(func() (*sql.Rows, *sql.Row, error) {
			row := DB.QueryRow("SELECT value FROM gauge WHERE name = $1", key)
			return nil, row, nil
		})
		err := row.Scan(&d.Value)
		if err != nil {
			panic(err)
		}
		exists = true
	}
	return d.Value, exists
}

func (d *DBStorage) GetCount(key string) (int64, bool) {
	var exists bool
	var count int
	_, _, err := withRetries(func() (*sql.Rows, *sql.Row, error) {
		err := DB.QueryRow("SELECT COUNT(*) FROM counter WHERE name = $1", key).Scan(&count)
		return nil, nil, err
	})
	if err != nil {
		panic(err)
	}
	if count > 0 {
		_, row, _ := withRetries(func() (*sql.Rows, *sql.Row, error) {
			row := DB.QueryRow("SELECT delta FROM counter WHERE name = $1", key)
			return nil, row, nil
		})
		err := row.Scan(&d.Delta)
		if err != nil {
			panic(err)
		}
		if d.Delta > 0 {
			exists = true
		}
	}
	return d.Delta, exists
}

func (d *DBStorage) UpdateBatch(list []Metrics) error {
	tx, err := DB.Begin()
	if err != nil {
		panic(err)
	}
	for _, metric := range list {
		switch metric.MType {
		case config.Gauge:
			k := metric.ID
			v := metric.Value
			d.GaugeStorage(k, *v)

		case config.Counter:
			k := metric.ID
			v := metric.Delta
			d.CountStorage(k, *v)
			vNew, _ := d.GetCount(metric.ID)
			metric.Delta = &vNew
		default:
			tx.Rollback()
			err = fmt.Errorf("unknowning metric type")
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
	return nil
}

func withRetries(operation func() (*sql.Rows, *sql.Row, error)) (*sql.Rows, *sql.Row, error) {
	selectedErr := []string{pgerrcode.UniqueViolation, pgerrcode.ConnectionException, pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure, pgerrcode.SQLClientUnableToEstablishSQLConnection,
		pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
		pgerrcode.TransactionResolutionUnknown, pgerrcode.ProtocolViolation}

	for i := 0; i < config.MaxRetries; i++ {
		rows, row, err := operation()
		if err == nil {
			return rows, row, nil
		}
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if slices.Contains(selectedErr, pgErr.Code) {
				time.Sleep(retryDelays[i])
				continue
			}
		}
		return nil, nil, err
	}
	return nil, nil, fmt.Errorf("max retries")
}
