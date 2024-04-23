package helpers

import (
	"bytes"
	"compress/gzip"
	"io"
	"reflect"
	"testing"
)

func compressString(input string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write([]byte(input))
	w.Close()
	return b.Bytes()
}

func TestCompressResp(t *testing.T) {
	tests := []struct {
		name        string
		metricsByte []byte
		want        []byte
	}{
		{
			name:        "Test case 1",
			metricsByte: []byte("hello"),
			want:        compressString("hello"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := CompressResp(tt.metricsByte)
			expected := bytes.NewBuffer(tt.want)

			actualGzip, _ := gzip.NewReader(&actual)
			expectedGzip, _ := gzip.NewReader(expected)

			actualBytes, _ := io.ReadAll(actualGzip)
			expectedBytes, _ := io.ReadAll(expectedGzip)
			if !reflect.DeepEqual(actualBytes, expectedBytes) {
				t.Errorf("Test case %s failed. Expected: %v, Got: %v", tt.name, expectedBytes, actualBytes)
			}
		})
	}
}
