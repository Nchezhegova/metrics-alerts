package storage

import (
	"database/sql"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	_ "github.com/lib/pq"
)

type DBStorage struct {
	Name       string  `json:"id"`
	MetricType string  `json:"type"`
	Value      float64 `json:"delta,omitempty"`
	Delta      int64   `json:"value,omitempty"`
}

var DB *sql.DB

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
                         delta INT NOT NULL);`
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
	err := DB.QueryRow("SELECT COUNT(*) FROM counter WHERE name = $1", k).Scan(&count)
	if err != nil {
		panic(err)
	}
	if count > 0 {
		_, err := DB.Exec("UPDATE counter SET name=$1, delta=$2 WHERE name=$1", k, d.Delta+v)
		if err != nil {
			panic(err)
		}
	} else {
		_, err := DB.Exec("INSERT INTO counter (name, delta) VALUES ($1, $2)", k, v)
		if err != nil {
			panic(err)
		}
	}
}

func (d *DBStorage) GaugeStorage(k string, v float64) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM gauge WHERE name = $1", k).Scan(&count)
	if err != nil {
		panic(err)
	}
	if count > 0 {
		_, err := DB.Exec("UPDATE gauge SET name=$1, value=$2 WHERE name=$1", k, v)
		if err != nil {
			panic(err)
		}
	} else {
		_, err := DB.Exec("INSERT INTO gauge (name, value) VALUES ($1, $2)", k, v)
		if err != nil {
			panic(err)
		}
	}
}

// TODO убрать функцию из интерфейса
func (d *DBStorage) GetStorage() interface{} {
	arrd := []DBStorage{}

	rows, err := DB.Query("SELECT name, value FROM gauge")
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
	row := DB.QueryRow("SELECT name, delta FROM counter")
	if err := row.Scan(&d.Name, &d.Delta); err != nil {
		//panic(err)
	} else {
		d.MetricType = config.Counter
		arrd = append(arrd, *d)
	}

	return arrd
}

func (d *DBStorage) SetStartData(storage MemStorage) {

}

func (d *DBStorage) GetGauge(key string) (float64, bool) {
	var exists bool
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM gauge WHERE name = $1", key).Scan(&count)
	if err != nil {
		panic(err)
	}
	if count > 0 {
		row := DB.QueryRow("SELECT value FROM gauge WHERE name = $1", key)
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
	err := DB.QueryRow("SELECT COUNT(*) FROM counter WHERE name = $1", key).Scan(&count)
	if err != nil {
		panic(err)
	}

	if count > 0 {
		row := DB.QueryRow("SELECT delta FROM counter WHERE name = $1", key)
		err := row.Scan(&d.Delta)
		if err != nil {
			panic(err)
		}
		if d.Value > 0 {
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
			err = fmt.Errorf("unknowning metric type", err, err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
	return nil
}
