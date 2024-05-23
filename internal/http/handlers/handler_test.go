package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"github.com/Nchezhegova/metrics-alerts/internal/helpers"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testreq struct {
	url    string
	method string
	body   string
}

func createContext(req testreq, ms *storage.MemStorage) (storage.MemStorage, *gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	t, _ := http.NewRequest(req.method, req.url, bytes.NewBuffer([]byte(req.body)))
	c.Request = t
	r.POST("/update/:type/:name/:value", func(c *gin.Context) {
		updateMetrics(c, ms, false, "")
	})
	r.GET("/value/:type/:name/", func(c *gin.Context) {
		getMetric(c, ms)
	})
	r.GET("/", func(c *gin.Context) {
		printMetrics(c, ms)
	})
	r.POST("/update/", func(c *gin.Context) {
		updateMetricsFromBody(c, ms, false, "", "")
	})
	r.POST("/updates/", func(c *gin.Context) {
		updateBatchMetricsFromBody(c, ms, false, "", "")
	})
	r.POST("/value/", func(c *gin.Context) {
		getMetricFromBody(c, ms, "")
	})
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
		{
			name: "3 not valid counter",
			value: testreq{
				url:    "/update/counter/qwe/dsf",
				method: "POST",
			},
			want: storage.MemStorage{
				Gauge:   map[string]float64{},
				Counter: map[string]int64{},
			},
		},
		{
			name: "4 not valid gauge",
			value: testreq{
				url:    "/update/gauge/qwe/dsf",
				method: "POST",
			},
			want: storage.MemStorage{
				Gauge:   map[string]float64{},
				Counter: map[string]int64{},
			},
		},
		{
			name: "5 not valid type",
			value: testreq{
				url:    "/update/no_valid/qwe/dsf",
				method: "POST",
			},
			want: storage.MemStorage{
				Gauge:   map[string]float64{},
				Counter: map[string]int64{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := storage.MemStorage{
				Gauge:   make(map[string]float64),
				Counter: make(map[string]int64),
			}
			m, _, _ := createContext(tt.value, &ms)
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
	}, {
		name: "3 not valid",
		value: testreq{
			url:    "/value/no_valid/q/",
			method: "GET",
		},
		want: "",
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
func Test_printMetricsWithGzip(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	m := storage.MemStorage{
		Gauge:   map[string]float64{"w": 36},
		Counter: map[string]int64{"q": 54},
	}
	c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Accept-Encoding", "gzip")

	printMetrics(c, &m)
	assert.Equal(t, http.StatusOK, c.Writer.Status())

	contentEncoding := c.Writer.Header().Get("Content-Encoding")
	assert.Equal(t, "gzip", contentEncoding)
}

func TestStartServ(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := &http.Client{}
	m := storage.MemStorage{
		Gauge:   map[string]float64{"w": 36},
		Counter: map[string]int64{"q": 54},
	}

	go StartServ(&m, "localhost:8099", 1, "", false, "", "", "")

	time.Sleep(1000 * time.Millisecond)

	req, _ := http.NewRequest("GET", "http://localhost:8099/value/gauge/w/", nil)
	resp, _ := client.Do(req)
	_ = resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func Test_updateMetricsFromBody(t *testing.T) {
	tests := []struct {
		name  string
		value testreq
		want  storage.MemStorage
	}{
		{
			name: "1 gauge",
			value: testreq{
				url:    "/update/",
				method: "POST",
				body:   `{"ID":"qwe", "Type":"gauge", "Value":54}`,
			},
			want: storage.MemStorage{
				Gauge:   map[string]float64{"qwe": 54},
				Counter: map[string]int64{},
			},
		},
		{
			name: "2 counter",
			value: testreq{
				url:    "/update/",
				method: "POST",
				body:   `{"ID":"qwe", "Type":"counter", "Delta":32}`,
			},
			want: storage.MemStorage{
				Gauge:   map[string]float64{},
				Counter: map[string]int64{"qwe": 32},
			},
		},
		{
			name: "empty",
			value: testreq{
				url:    "/update/",
				method: "POST",
				body:   `{}`,
			},
			want: storage.MemStorage{
				Gauge:   map[string]float64{},
				Counter: map[string]int64{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := storage.MemStorage{
				Gauge:   make(map[string]float64),
				Counter: make(map[string]int64),
			}
			m, _, _ := createContext(tt.value, &ms)
			assert.Equal(t, m.Gauge["qwe"], tt.want.Gauge["qwe"])
			assert.Equal(t, m.Counter["qwe"], tt.want.Counter["qwe"])
		})
	}
}

func Test_updateBatchMetricsFromBody(t *testing.T) {
	tests := []struct {
		name  string
		value testreq
		want  int
	}{
		{
			name: "Valid JSON",
			value: testreq{
				url:    "/updates/",
				method: "POST",
				body:   `[{"ID":"qwe", "Type":"gauge", "Value":54},{"ID":"qwe", "Type":"counter", "Delta":32}]`,
			},
			want: http.StatusOK,
		},
		{
			name: "Invalid JSON",
			value: testreq{
				url:    "/updates/",
				method: "POST",
				body:   `invalid_json`,
			},
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := storage.MemStorage{
				Gauge:   make(map[string]float64),
				Counter: make(map[string]int64),
			}
			_, _, w := createContext(tt.value, &ms)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func Test_updateMetricsFromBodyWitnZip(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	dataByte := []byte(`{"ID":"qwe", "Type":"gauge", "Value":54}`)
	var compressBody io.ReadWriter = &bytes.Buffer{}

	gzipWriter := gzip.NewWriter(compressBody)
	_, _ = gzipWriter.Write(dataByte)
	_ = gzipWriter.Close()

	req, _ := http.NewRequest("POST", "/", compressBody)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	c.Request = req
	m := storage.MemStorage{
		Gauge:   map[string]float64{"w": 36},
		Counter: map[string]int64{"q": 54},
	}
	updateMetricsFromBody(c, &m, false, "", "")
	if c.Writer.Header().Get("Accept-Encoding") != "gzip" {
		t.Errorf("Expected Accept-Encoding header to be 'gzip'; got %s", c.Writer.Header().Get("Accept-Encoding"))
	}
}

func Test_getMetricFromBody(t *testing.T) {
	tests := []struct {
		name  string
		value testreq
		want  int
	}{
		{
			name: "get counter",
			value: testreq{
				url:    "/value/",
				method: "POST",
				body:   `{"ID":"qwe", "Type":"counter"}`,
			},
			want: http.StatusNotFound,
		},
		{
			name: "Invalid JSON",
			value: testreq{
				url:    "/value/",
				method: "POST",
				body:   `invalid_json`,
			},
			want: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := storage.MemStorage{
				Gauge:   make(map[string]float64),
				Counter: make(map[string]int64),
			}
			_, _, w := createContext(tt.value, &ms)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func Test_getMetricFromBodyWithGzip(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	m := storage.MemStorage{
		Gauge:   map[string]float64{"metric_id": 36},
		Counter: map[string]int64{"q": 54},
	}

	data := storage.Metrics{
		MType: "gauge",
		ID:    "metric_id",
	}
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", "/", bytes.NewReader(body))
	req.Header.Set("Accept-Encoding", "gzip")
	c.Request = req

	getMetricFromBody(c, &m, "")
	assert.Equal(t, http.StatusOK, c.Writer.Status())
	contentEncoding := c.Writer.Header().Get("Content-Encoding")
	assert.Equal(t, "gzip", contentEncoding)
}

func TestCheckHash(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	hashKey := "your_hash_key"
	bodyContent := "example body content"
	c.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(bodyContent)))
	c.Request.Header.Set("HashSHA256", base64.StdEncoding.EncodeToString(helpers.CalculateHash([]byte(bodyContent), hashKey)))

	assert.True(t, checkHash(c, hashKey), "Expected true, got false")
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, c.Request.Body)
	assert.NoError(t, err, "Error copying request body")
	assert.Equal(t, bodyContent, buf.String(), "Expected request body to remain unchanged")

	c.Request.Header.Set("HashSHA256", "incorrect_hash")
	assert.False(t, checkHash(c, hashKey), "Expected false, got true")

	c.Request.Header.Set("HashSHA256", "")
	assert.True(t, checkHash(c, hashKey), "Expected true, got false")
}
