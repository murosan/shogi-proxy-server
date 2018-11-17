// Copyright 2018 murosan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package to_usi

import (
	"bytes"
	"testing"

	"github.com/murosan/shogi-proxy-server/app/domain/entity/shogi"
)

func TestToUsi_Position(t *testing.T) {
	tu := NewToUsi()
	cases := []struct {
		in   shogi.Position
		want []byte
	}{
		{
			shogi.Position{
				Pos: [9][9]int{
					{-2, -3, -4, -5, -8, 0, -4, -3, -2},
					{0, 0, 0, 0, 0, 0, -5, -6, 0},
					{-1, 0, -1, -1, -1, -1, 0, 0, -1},
					{0, 0, 0, 0, 0, 0, 7, 0, 0},
					{0, 0, 0, 0, 0, 0, 0, 0, 0},
					{0, -7, 1, 0, 0, 0, 0, 0, 0},
					{1, 0, 0, 1, 1, 1, 1, 0, 1},
					{0, 6, 5, 0, 0, 0, 0, 0, 0},
					{2, 3, 4, 0, 8, 5, 4, 3, 2},
				},
				Cap0:      [7]int{3, 0, 0, 0, 0, 0, 0},
				Cap1:      [7]int{2, 0, 0, 0, 0, 0, 0},
				Turn:      1,
				MoveCount: 100,
			},
			[]byte("position sfen lnsgk1snl/6gb1/p1pppp2p/6R2/9/1rP6/P2PPPP1P/1BG6/LNS1KGSNL w 3P2p 100"),
		},
	}

	for i, c := range cases {
		b, e := tu.Position(c.in)
		if e != nil {
			t.Errorf("Error: %v\nindex: %d", e.Error(), i)
		}
		if !bytes.Equal(b, c.want) {
			t.Errorf("index: %d\nExpected: %v\nActual: %v", i, string(c.want), string(b))
		}
	}
}
