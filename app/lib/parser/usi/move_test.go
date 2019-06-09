package usi

import (
	"reflect"
	"strings"
	"testing"

	"github.com/murosan/shogi-board-server/app/domain/model/shogi"
)

func TestParseMove(t *testing.T) {
	cases := []struct {
		in   string
		want *shogi.Move
		err  error
	}{
		{"7g7f",
			&shogi.Move{
				Source:     &shogi.Point{Row: 6, Column: 6},
				Dest:       &shogi.Point{Row: 5, Column: 6},
				PieceID:    0,
				IsPromoted: false,
			},
			nil,
		},
		{"8h2b+",
			&shogi.Move{
				Source:     &shogi.Point{Row: 7, Column: 7},
				Dest:       &shogi.Point{Row: 1, Column: 1},
				PieceID:    0,
				IsPromoted: true,
			},
			nil},
		{"G*5b",
			&shogi.Move{
				Source:     &shogi.Point{Row: -1, Column: -1},
				Dest:       &shogi.Point{Row: 1, Column: 4},
				PieceID:    5,
				IsPromoted: false,
			},
			nil,
		},
		{
			"s*5b",
			&shogi.Move{
				Source:     &shogi.Point{Row: -1, Column: -1},
				Dest:       &shogi.Point{Row: 1, Column: 4},
				PieceID:    -4,
				IsPromoted: false,
			},
			nil,
		},
		{"", &shogi.Move{}, emptyErr},
		{"7g7z", &shogi.Move{}, emptyErr},
		{"7g7$", &shogi.Move{}, emptyErr},
		{"0g7a", &shogi.Move{}, emptyErr},
		{"1x7a", &shogi.Move{}, emptyErr},
		{"G*vb", &shogi.Move{}, emptyErr},
		{"G*4z", &shogi.Move{}, emptyErr},
		{"A*7a", &shogi.Move{}, emptyErr},
	}

	for i, c := range cases {
		moveHelper(t, i, c.in, c.want, c.err)
	}
}

func moveHelper(t *testing.T, i int, in string, want *shogi.Move, err error) {
	t.Helper()
	res, e := ParseMove(in)
	msg := ""

	if (e == nil && err != nil) || (e != nil && err == nil) {
		msg = "Expected error, but was not as expected."
		moveErrorPrintHelper(t, i, msg, in, err, e)
	}

	// expected same error
	if e != nil && strings.Contains(string(e.Error()), string(err.Error())) {
		return
	}

	// was error but not as expected
	if e != nil && !strings.Contains(string(e.Error()), string(err.Error())) {
		msg = "Expected error, but was not as expected."
		moveErrorPrintHelper(t, i, msg, in, err, e)
	}

	if !reflect.DeepEqual(res, want) {
		msg = "The value was not as expected."
		moveErrorPrintHelper(t, i, msg, in, want, res)
	}
}

func moveErrorPrintHelper(
	t *testing.T,
	i int,
	msg, in string,
	expected, actual interface{},
) {
	t.Helper()
	t.Errorf(`[Parse Move] %s
Index:    %d
Input:    %s
Expected: %v
Actual:   %v
`, msg, i, in, expected, actual)
}
