package domain

import (
	"fmt"
	"testing"
)

func TestCountToScore(t *testing.T) {
	tests := []struct {
		in  int32
		out int32
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 1},
		{4, 1},
		{5, 1},
		{6, 2},
		{7, 2},
		{8, 2},
		{9, 2},
		{10, 2},
		{11, 3},
		{12, 3},
		{13, 3},
		{14, 3},
		{15, 3},
		{16, 4},
		{17, 4},
		{18, 4},
		{19, 4},
		{20, 4},
		{21, 5},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.in), func(t *testing.T) {
			out := CountToScore(test.in)
			if out != test.out {
				t.Errorf("got %v, want %v", out, test.out)
			}
		})
	}
}
