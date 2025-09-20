package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// import (
// 	"bytes"
// 	"testing"
// )

func TestGetInOut(t *testing.T) {
	tests := []struct {
		name string
		v    string
		in   string
		out  string
	}{
		{
			name: "normal",
			v:    "8090:8080",
			in:   "8090",
			out:  "8080",
		},
	}

	for _, item := range tests {
		t.Run(item.name, func(t *testing.T) {
			in, out := getInOut(item.v)
			assert.Equal(t, in, item.in, "they should be equal")
			assert.Equal(t, out, item.out, "they should be equal")
		})
	}
}
