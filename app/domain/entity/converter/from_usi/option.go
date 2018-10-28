// Copyright 2018 murosan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package from_usi

import (
	"bytes"
	"strconv"

	"github.com/murosan/shogi-proxy-server/app/domain/entity/engine/option"
	"github.com/murosan/shogi-proxy-server/app/domain/entity/exception"
	"github.com/murosan/shogi-proxy-server/app/lib/byteutil"
)

var (
	space  = []byte(" ")
	id     = []byte("id")
	opt    = []byte("option")
	name   = []byte("name")
	author = []byte("author")
	tpe    = []byte("type")
	deflt  = []byte("default")
	min    = []byte("min")
	max    = []byte("max")
	selOpt = []byte("var")
)

// id name <EngineName>
// id author <AuthorName> をEngineにセットする
// EngineNameやAuthorNameにスペースが入る場合もあるので最後にJoinしている
// TODO: 正規表現でやるか検討
func (fu *FromUsi) EngineID(b []byte) ([]byte, []byte, error) {
	s := bytes.Split(bytes.TrimSpace(b), space)
	if len(s) < 3 || !bytes.Equal(s[0], id) {
		return nil, nil, exception.InvalidIdSyntax
	}

	if bytes.Equal(s[1], name) {
		return name, bytes.Join(s[2:], space), nil
	}

	if bytes.Equal(s[1], author) {
		return author, bytes.Join(s[2:], space), nil
	}

	return nil, nil, exception.UnknownOption
}

// 一行受け取って、EngineのOptionMapにセットする
// パースできなかったらエラーを返す
func (fu *FromUsi) Option(b []byte) (option.Option, error) {
	s := bytes.Split(bytes.TrimSpace(b), space)
	if len(s) < 5 || !bytes.Equal(s[0], opt) || !bytes.Equal(s[1], name) || !bytes.Equal(s[3], tpe) || len(s[4]) == 0 {
		return nil, exception.InvalidOptionSyntax
	}

	switch string(s[4]) {
	case "check":
		return fu.parseCheck(s)
	case "spin":
		return fu.parseSpin(s)
	case "combo":
		return fu.parseSelect(s)
	case "button":
		return fu.parseButton(s)
	case "string":
		return fu.parseString(s)
	case "filename":
		return fu.parseFileName(s)
	default:
		return nil, exception.UnknownOptionType
	}
}

// check type を Egn の Options にセットする
// option name <string> type check default <bool>
// このフォーマット以外は許容しない
// default がなかったり、bool ではない時はエラー
func (fu *FromUsi) parseCheck(b [][]byte) (option.Option, error) {
	if len(b) != 7 || !bytes.Equal(b[5], deflt) || len(b[2]) == 0 || len(b[6]) == 0 {
		return nil, exception.InvalidOptionSyntax.WithMsg("Received option type was 'check', but malformed. The format must be [option name <string> type check default <bool>]")
	}

	n, d := b[2], b[6]
	if bytes.Equal(d, []byte("true")) {
		return option.Check{Name: n, Val: true, Default: true}, nil
	}
	if bytes.Equal(d, []byte("false")) {
		return option.Check{Name: n, Val: false, Default: false}, nil
	}
	return nil, exception.InvalidOptionSyntax.WithMsg("Default want of 'check' type was not bool. Received: " + string(d))
}

// spin type を Egn の Options にセットする
// option name <string> type spin default <int> min <int> max <int>
// このフォーマット以外は許容しない
// 各値がなかったり、int ではない時、min > max の時はエラー
func (fu *FromUsi) parseSpin(b [][]byte) (option.Spin, error) {
	if len(b) != 11 || !bytes.Equal(b[5], deflt) || !bytes.Equal(b[7], min) || !bytes.Equal(b[9], max) || len(b[2]) == 0 {
		return option.Spin{}, exception.InvalidOptionSyntax.WithMsg("Received option type was 'spin', but malformed. The format must be [option name <string> type spin default <int> min <int> max <int>]")
	}

	df, err := strconv.Atoi(string(b[6]))
	if err != nil {
		return option.Spin{}, exception.InvalidOptionSyntax.WithMsg("Default want of 'spin' type was not int. Received: " + string(min))
	}
	mi, err := strconv.Atoi(string(b[8]))
	if err != nil {
		return option.Spin{}, exception.InvalidOptionSyntax.WithMsg("Min want of 'spin' type was not int. Received: " + string(min))
	}
	ma, err := strconv.Atoi(string(b[10]))
	if err != nil {
		return option.Spin{}, exception.InvalidOptionSyntax.WithMsg("Max want of 'spin' type was not int. Received: " + string(min))
	}

	return option.Spin{Name: b[2], Val: df, Default: df, Min: mi, Max: ma}, nil
}

// select type を Egn の Options にセットする
// option name <string> type combo default <string> rep(var <string>)
// このフォーマット以外は許容しない
// Default がない、var がない、default が var にない時はエラー
func (fu *FromUsi) parseSelect(b [][]byte) (option.Select, error) {
	if len(b) < 9 || len(b[2]) == 0 || len(b[6]) == 0 {
		return option.Select{}, exception.InvalidOptionSyntax.WithMsg("Received option type was 'combo', but malformed. The format must be [option name <string> type combo default <string> rep(var <string>)]")
	}

	s := option.Select{Name: b[2]}

	i := 8
	for i < len(b) && bytes.Equal(b[i-1], selOpt) {
		s.Vars = append(s.Vars, b[i])
		i += 2
	}

	s.Index = byteutil.IndexOfBytes(s.Vars, b[6])
	if s.Index == -1 {
		return option.Select{}, exception.InvalidOptionSyntax.WithMsg("Default want of 'combo' type was not found in vars.")
	}

	return s, nil
}

// button type を Egn の Options にセットする
// option name <string> type button
func (fu *FromUsi) parseButton(b [][]byte) (option.Button, error) {
	if len(b) != 5 || len(b[2]) == 0 {
		return option.Button{}, exception.InvalidOptionSyntax.WithMsg("Received option type was 'button', but malformed. The format must be [option name <string> type button]")
	}
	return option.Button{Name: b[2]}, nil
}

// string type を Egn の Options にセットする
// option name <string> type string default <string>
func (fu *FromUsi) parseString(b [][]byte) (option.String, error) {
	if len(b) != 7 || len(b[2]) == 0 || len(b[6]) == 0 {
		return option.String{}, exception.InvalidOptionSyntax.WithMsg("Received option type was 'string', but malformed. The format must be [option name <string> type string default <string>]")
	}
	return option.String{Name: b[2], Val: b[6], Default: b[6]}, nil
}

// option name <string> type filename default <string>
func (fu *FromUsi) parseFileName(b [][]byte) (option.FileName, error) {
	if len(b) != 7 || len(b[2]) == 0 || len(b[6]) == 0 {
		return option.FileName{}, exception.InvalidOptionSyntax.WithMsg("Received option type was 'filename', but malformed. The format must be [option name <string> type filename default <string>]")
	}
	return option.FileName{Name: b[2], Val: b[6], Default: b[6]}, nil
}