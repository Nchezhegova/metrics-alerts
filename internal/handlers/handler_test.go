package handlers

import (
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
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

func createContext(req testreq, ms *storage.MemStorage) (storage.MemStorage, *gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/update/:type/:name/:value", func(c *gin.Context) {
		updateMetrics(c, ms, false, "")
	})
	r.GET("/value/:type/:name/", func(c *gin.Context) {
		getMetric(c, ms)
	})
	r.GET("/", func(c *gin.Context) {
		printMetrics(c, ms)
	})

	t, _ := http.NewRequest(req.method, req.url, nil)
	r.ServeHTTP(w, t)
	return *ms, c, w
}

func Test_updateMetrics(t *testing.T) {

	tests := []struct {
		name  string
		value testreq
		want  storage.MemStorage
	}{{
		name: "1 gauge",
		value: testreq{
			url:    "/update/gauge/qwe/54",
			method: "POST",
		},
		want: storage.MemStorage{
			Gauge:   map[string]float64{"qwe": 54},
			Counter: map[string]int64{},
		},
	},
		{
			name: "2 counter",
			value: testreq{
				url:    "/update/counter/qwe/54",
				method: "POST",
			},
			want: storage.MemStorage{
				Gauge:   map[string]float64{},
				Counter: map[string]int64{"qwe": 54},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := storage.MemStorage{
				Gauge:   make(map[string]float64),
				Counter: make(map[string]int64),
			}
			m, c, _ := createContext(tt.value, &ms)
			updateMetrics(c, &m, false, "")
			assert.Equal(t, m.Gauge["qwe"], tt.want.Gauge["qwe"])
			assert.Equal(t, m.Counter["qwe"], tt.want.Counter["qwe"])
		})
	}
}

func Test_getMetric(t *testing.T) {

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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := storage.MemStorage{
				Gauge:   map[string]float64{"w": 36},
				Counter: map[string]int64{"q": 54},
			}
			_, _, w := createContext(test.value, &ms)
			assert.Equal(t, w.Body.String(), test.want)
		})
	}
}

func Test_printMetrics(t *testing.T) {
	value := testreq{
		url:    "/",
		method: "GET",
	}
	want := http.StatusOK
	ms := storage.MemStorage{
		Gauge:   map[string]float64{"w": 36},
		Counter: map[string]int64{"q": 54},
	}
	_, _, w := createContext(value, &ms)
	assert.Equal(t, want, w.Code)

}
