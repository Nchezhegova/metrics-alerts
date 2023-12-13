package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testreq struct {
	url    string
	method string
}

func createContext(req testreq, ms *MemStorage) (MemStorage, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/update/:type/:name/:value", ms.UpdateMetrics)
	r.GET("/value/:type/:name/", ms.GetMetric)
	r.GET("/", ms.PrintMetrics)

	t, _ := http.NewRequest(req.method, req.url, nil)
	r.ServeHTTP(w, t)
	return *ms, w
}

func TestMemStorage_UpdateMetrics(t *testing.T) {

	tests := []struct {
		name  string
		value testreq
		want  MemStorage
	}{{
		name: "1 gauge",
		value: testreq{
			url:    "/update/gauge/qwe/54",
			method: "POST",
		},
		want: MemStorage{
			Gauge:   map[string]float64{"qwe": 54},
			Counter: map[string]int64{},
		},
	}, {
		name: "2 counter",
		value: testreq{
			url:    "/update/counter/qwe/54",
			method: "POST",
		},
		want: MemStorage{
			Gauge:   map[string]float64{},
			Counter: map[string]int64{"qwe": 54},
		},
	},
	}

	for _, test := range tests { // цикл по всем тестам
		t.Run(test.name, func(t *testing.T) {
			ms := MemStorage{
				Gauge:   make(map[string]float64),
				Counter: make(map[string]int64),
			}
			m, _ := createContext(test.value, &ms)
			assert.Equal(t, m.Gauge["qwe"], test.want.Gauge["qwe"])
			assert.Equal(t, m.Counter["qwe"], test.want.Counter["qwe"])
		})
	}
}

func TestMemStorage_GetMetric(t *testing.T) {

	tests := []struct {
		name  string
		value testreq
		want  string
	}{{
		name: "1 gauge",
		value: testreq{
			url:    "/value/gauge/w/",
			method: "GET",
		},
		want: "36",
	}, {
		name: "2 counter",
		value: testreq{
			url:    "/value/counter/q/",
			method: "GET",
		},
		want: "54",
	},
	}

	for _, test := range tests { // цикл по всем тестам
		t.Run(test.name, func(t *testing.T) {
			ms := MemStorage{
				Gauge:   map[string]float64{"w": 36},
				Counter: map[string]int64{"q": 54},
			}
			_, w := createContext(test.value, &ms)
			assert.Equal(t, w.Body.String(), test.want)
		})
	}
}

func TestMemStorage_PrintMetrics(t *testing.T) {
	tests := []struct {
		name  string
		value testreq
		want  int
	}{{
		name: "print MemStorage",
		value: testreq{
			url:    "/",
			method: "GET",
		},
		want: http.StatusOK,
	},
	}

	for _, test := range tests { // цикл по всем тестам
		t.Run(test.name, func(t *testing.T) {
			ms := MemStorage{
				Gauge:   map[string]float64{"w": 36},
				Counter: map[string]int64{"q": 54},
			}
			_, w := createContext(test.value, &ms)
			assert.Equal(t, test.want, w.Code)
		})
	}
}
