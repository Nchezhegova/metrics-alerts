package storage

import (
	"context"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBStorage_CountStorage(t *testing.T) {
	type fields struct {
		Name       string
		MetricType string
		Value      float64
		Delta      int64
	}
	type args struct {
		c context.Context
		k string
		v int64
	}
	uniqName := uuid.New().String()
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "1",
			fields: fields{
				Name:       uniqName,
				MetricType: "counter",
			},
			args: args{
				c: context.Background(),
				k: uniqName,
				v: 2,
			},
		},
	}
	OpenDB(config.DATEBASE)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DBStorage{
				Name:       tt.fields.Name,
				MetricType: tt.fields.MetricType,
				Value:      tt.fields.Value,
				Delta:      tt.fields.Delta,
			}
			d.CountStorage(tt.args.c, tt.args.k, tt.args.v)
			new_value, _ := d.GetCount(tt.args.c, tt.args.k)
			assert.Equal(t, tt.args.v, new_value)
			d.CountStorage(tt.args.c, tt.args.k, tt.args.v)
			new_value, _ = d.GetCount(tt.args.c, tt.args.k)
			assert.Equal(t, tt.args.v*2, new_value)
		})
	}
}

func TestDBStorage_GaugeStorage(t *testing.T) {
	type fields struct {
		Name       string
		MetricType string
		Value      float64
		Delta      int64
	}
	type args struct {
		c context.Context
		k string
		v float64
	}
	uniqName := uuid.New().String()
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "1",
			fields: fields{
				Name:       uniqName,
				MetricType: "gauge",
				Delta:      0,
			},
			args: args{
				c: context.Background(),
				k: uniqName,
				v: 5.5,
			},
		},
	}
	OpenDB(config.DATEBASE)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DBStorage{
				Name:       tt.fields.Name,
				MetricType: tt.fields.MetricType,
				Value:      tt.fields.Value,
				Delta:      tt.fields.Delta,
			}
			d.GaugeStorage(tt.args.c, tt.args.k, tt.args.v)
			new_value, _ := d.GetGauge(tt.args.c, tt.args.k)
			assert.Equal(t, tt.args.v, new_value)
			d.GaugeStorage(tt.args.c, tt.args.k, tt.args.v*2)
			new_value, _ = d.GetGauge(tt.args.c, tt.args.k)
			assert.Equal(t, tt.args.v*2, new_value)
		})
	}
}

func TestDBStorage_GetStorage(t *testing.T) {
	type fields struct {
		Name       string
		MetricType string
		Value      float64
		Delta      int64
	}
	type args struct {
		c context.Context
		k string
		v float64
	}
	uniqName := uuid.New().String()
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "1",
			fields: fields{
				Name:       uniqName,
				MetricType: "gauge",
				Delta:      0,
			},
			args: args{
				c: context.Background(),
				k: uniqName,
				v: 5.5,
			},
		},
	}
	OpenDB(config.DATEBASE)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DBStorage{
				Name:       tt.fields.Name,
				MetricType: tt.fields.MetricType,
				Value:      tt.fields.Value,
				Delta:      tt.fields.Delta,
			}
			d.GaugeStorage(tt.args.c, tt.args.k, tt.args.v)
			res := d.GetStorage(tt.args.c)
			arr, _ := res.([]DBStorage)
			assert.NotEqual(t, 0, len(arr))
		})
	}
}

func TestDBStorage_UpdateBatch(t *testing.T) {
	type fields struct {
		Name       string
		MetricType string
		Value      float64
		Delta      int64
	}
	type args struct {
		c    context.Context
		list []Metrics
	}
	uniqNameGauge := uuid.New().String()
	uniqNameCounter := uuid.New().String()
	var kg int64 = 5
	var kc float64 = 5.5
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "1",
			fields: fields{
				Name:       uniqNameCounter,
				MetricType: "counter",
				Delta:      0,
			},
			args: args{
				c: context.Background(),
				list: []Metrics{
					{
						ID:    uniqNameCounter,
						MType: "counter",
						Delta: &kg,
					},
					{
						ID:    uniqNameGauge,
						MType: "gauge",
						Value: &kc,
					},
				},
			},
		},
	}
	OpenDB(config.DATEBASE)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DBStorage{
				Name:       tt.fields.Name,
				MetricType: tt.fields.MetricType,
				Value:      tt.fields.Value,
				Delta:      tt.fields.Delta,
			}
			d.UpdateBatch(tt.args.c, tt.args.list)
			new_value_counter, _ := d.GetCount(tt.args.c, tt.args.list[0].ID)
			assert.Equal(t, kg, new_value_counter)
			new_value_gauge, _ := d.GetGauge(tt.args.c, tt.args.list[1].ID)
			assert.Equal(t, kc, new_value_gauge)
		})
	}
}
