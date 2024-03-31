package helpers

import "testing"

func BenchmarkCompressResp(b *testing.B) {
	data := make([]byte, 1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompressResp(data)
	}
}
