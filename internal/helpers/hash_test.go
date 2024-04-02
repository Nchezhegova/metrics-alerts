package helpers

import (
	"reflect"
	"testing"
)

func TestCalculateHash(t *testing.T) {
	type args struct {
		body []byte
		key  string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Test case 1",
			args: struct {
				body []byte
				key  string
			}{body: []byte("hello"), key: "secret"},
			want: []byte{136, 170, 179, 237, 232, 211, 173, 249, 77, 38, 171, 144, 211, 186, 253, 74, 32, 131, 7, 12, 59, 204, 233, 192, 20, 238, 4, 164, 67, 132, 124, 11},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateHash(tt.args.body, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
