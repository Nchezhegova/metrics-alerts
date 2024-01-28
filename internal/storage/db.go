package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/log"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
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

func (d *DBStorage) CountStorage(c context.Context, k string, v int64) {
	var count int
	_, err := withRetriesRow(func() (*sql.Row, error) {
		err := DB.QueryRowContext(c, "SELECT COUNT(*) FROM counter WHERE name = $1", k).Scan(&count)
		return nil, err
	})
	if err != nil {
		log.Logger.Info("Error DB:", zap.Error(err))
	}
	if count > 0 {
		_, err := withRetriesRow(func() (*sql.Row, error) {
			_, err := DB.ExecContext(c, "UPDATE counter SET name=$1, delta=$2 WHERE name=$1", k, d.Delta+v)
			return nil, err
		})
		if err != nil {
			log.Logger.Info("Error DB:", zap.Error(err))
		}
	} else {
		_, err := withRetriesRow(func() (*sql.Row, error) {
			_, err := DB.ExecContext(c, "INSERT INTO counter (name, delta) VALUES ($1, $2)", k, v)
			return nil, err
		})
		if err != nil {
			log.Logger.Info("Error DB:", zap.Error(err))
		}
	}
}

func (d *DBStorage) GaugeStorage(c context.Context, k string, v float64) {
	var count int
	_, err := withRetriesRow(func() (*sql.Row, error) {
		err := DB.QueryRowContext(c, "SELECT COUNT(*) FROM gauge WHERE name = $1", k).Scan(&count)
		return nil, err
	})
	if err != nil {
		log.Logger.Info("Error DB:", zap.Error(err))
	}
	if count > 0 {
		_, err := withRetriesRow(func() (*sql.Row, error) {
			_, err := DB.ExecContext(c, "UPDATE gauge SET name=$1, value=$2 WHERE name=$1", k, v)
			return nil, err
		})
		if err != nil {
			log.Logger.Info("Error DB:", zap.Error(err))
		}
	} else {
		_, err := withRetriesRow(func() (*sql.Row, error) {
			_, err := DB.ExecContext(c, "INSERT INTO gauge (name, value) VALUES ($1, $2)", k, v)
			return nil, err
		})
		if err != nil {
			log.Logger.Info("Error DB:", zap.Error(err))
		}
	}
}

func (d *DBStorage) GetStorage(c context.Context) interface{} {
	arrd := []DBStorage{}
	rows, err := withRetriesRows(func() (*sql.Rows, error) {
		rows, err := DB.QueryContext(c, "SELECT name, value FROM gauge")
		return rows, err
	})

	if err != nil {
		log.Logger.Info("Error DB:", zap.Error(err))
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&d.Name, &d.Value); err != nil {
			log.Logger.Info("Error DB:", zap.Error(err))
		}
		d.MetricType = config.Gauge
		arrd = append(arrd, *d)
	}
	row, _ := withRetriesRow(func() (*sql.Row, error) {
		row := DB.QueryRowContext(c, "SELECT name, delta FROM counter")
		return row, nil
	})
	if err := row.Scan(&d.Name, &d.Delta); err != nil {
		log.Logger.Info("Error DB:", zap.Error(err))
	} else {
		d.MetricType = config.Counter
		arrd = append(arrd, *d)
	}
	if err := rows.Err(); err != nil {
		log.Logger.Info("Error DB:", zap.Error(err))
	}
	return arrd
}

func (d *DBStorage) SetStartData(storage MemStorage) {

}

func (d *DBStorage) GetGauge(c context.Context, key string) (float64, bool) {
	var exists bool
	var count int
	_, err := withRetriesRow(func() (*sql.Row, error) {
		err := DB.QueryRowContext(c, "SELECT COUNT(*) FROM gauge WHERE name = $1", key).Scan(&count)
		return nil, err
	})
	if err != nil {
		log.Logger.Info("Error DB:", zap.Error(err))
	}
	if count > 0 {
		row, _ := withRetriesRow(func() (*sql.Row, error) {
			row := DB.QueryRowContext(c, "SELECT value FROM gauge WHERE name = $1", key)
			return row, nil
		})
		err := row.Scan(&d.Value)
		if err != nil {
			log.Logger.Info("Error DB:", zap.Error(err))
		}
		exists = true
	}
	return d.Value, exists
}

func (d *DBStorage) GetCount(c context.Context, key string) (int64, bool) {
	var exists bool
	var count int
	_, err := withRetriesRow(func() (*sql.Row, error) {
		err := DB.QueryRowContext(c, "SELECT COUNT(*) FROM counter WHERE name = $1", key).Scan(&count)
		return nil, err
	})
	if err != nil {
		log.Logger.Info("Error DB:", zap.Error(err))
	}
	if count > 0 {
		row, _ := withRetriesRow(func() (*sql.Row, error) {
			row := DB.QueryRowContext(c, "SELECT delta FROM counter WHERE name = $1", key)
			return row, nil
		})
		err := row.Scan(&d.Delta)
		if err != nil {
			log.Logger.Info("Error DB:", zap.Error(err))
		}
		if d.Delta > 0 {
			exists = true
		}
	}
	return d.Delta, exists
}

func (d *DBStorage) UpdateBatch(c context.Context, list []Metrics) error {
	tx, err := DB.Begin()
	if err != nil {
		log.Logger.Info("Error DB:", zap.Error(err))
	}
	for _, metric := range list {
		switch metric.MType {
		case config.Gauge:
			k := metric.ID
			v := metric.Value
			d.GaugeStorage(c, k, *v)

		case config.Counter:
			k := metric.ID
			v := metric.Delta
			d.CountStorage(c, k, *v)
			vNew, _ := d.GetCount(c, metric.ID)
			metric.Delta = &vNew
		default:
			tx.Rollback()
			err = fmt.Errorf("unknowning metric type")
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Logger.Info("Error DB:", zap.Error(err))
	}
	return nil
}

func withRetriesRow(operation func() (*sql.Row, error)) (*sql.Row, error) {
	selectedErr := []string{pgerrcode.UniqueViolation, pgerrcode.ConnectionException, pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure, pgerrcode.SQLClientUnableToEstablishSQLConnection,
		pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
		pgerrcode.TransactionResolutionUnknown, pgerrcode.ProtocolViolation}

	for i := 0; i < config.MaxRetries; i++ {
		row, err := operation()
		if err == nil {
			return row, nil
		}
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if slices.Contains(selectedErr, pgErr.Code) {
				time.Sleep(retryDelays[i])
				continue
			}
		}
		return nil, err
	}
	return nil, fmt.Errorf("max retries")
}

// Отдельная функция чтобы избежать ошибки "rows.Err must be checked" в go vet. Возникает когда не должно быть rows
func withRetriesRows(operation func() (*sql.Rows, error)) (*sql.Rows, error) {
	selectedErr := []string{pgerrcode.UniqueViolation, pgerrcode.ConnectionException, pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure, pgerrcode.SQLClientUnableToEstablishSQLConnection,
		pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
		pgerrcode.TransactionResolutionUnknown, pgerrcode.ProtocolViolation}

	for i := 0; i < config.MaxRetries; i++ {
		rows, err := operation()
		if err == nil {
			return rows, nil
		}
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if slices.Contains(selectedErr, pgErr.Code) {
				time.Sleep(retryDelays[i])
				continue
			}
		}
		return nil, err
	}
	return nil, fmt.Errorf("max retries")
}
