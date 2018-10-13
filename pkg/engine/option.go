package engine

import (
	"bytes"
	"github.com/murosan/shogi-proxy-server/pkg/lib"
	"github.com/murosan/shogi-proxy-server/pkg/msg"
	"strconv"
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
func (e *Client) ParseId(b []byte) error {
	s := bytes.Split(bytes.TrimSpace(b), space)
	if len(s) < 3 || !bytes.Equal(s[0], id) {
		return msg.InvalidIdSyntax
	}

	if bytes.Equal(s[1], name) {
		e.Name = bytes.Join(s[2:], space)
		return nil
	}

	if bytes.Equal(s[1], author) {
		e.Author = bytes.Join(s[2:], space)
		return nil
	}

	return msg.UnknownOption
}

// 一行受け取って、EngineのOptionMapにセットする
// パースできなかったらエラーを返す
func (e *Client) ParseOpt(b []byte) error {
	s := bytes.Split(bytes.TrimSpace(b), space)
	if len(s) < 5 || !bytes.Equal(s[0], opt) || !bytes.Equal(s[1], name) || !bytes.Equal(s[3], tpe) || len(s[4]) == 0 {
		return msg.InvalidOptionSyntax
	}

	switch string(s[4]) {
	case "check":
		return e.parseCheck(s)
	case "spin":
		return e.parseSpin(s)
	case "combo":
		return e.parseSelect(s)
	case "button":
		return e.parseButton(s)
	case "string":
		return e.parseString(s)
	case "filename":
		return e.parseFileName(s)
	default:
		return msg.UnknownOptionType
	}
}

// check type を Engine の Options にセットする
// option name <string> type check default <bool>
// このフォーマット以外は許容しない
// default がなかったり、bool ではない時はエラー
func (e *Client) parseCheck(b [][]byte) error {
	n, d := b[2], b[6]
	if len(b) != 7 || !bytes.Equal(b[5], deflt) || len(n) == 0 || len(d) == 0 {
		return msg.InvalidOptionSyntax.WithMsg("Received option type was 'check', but malformed. The format must be [option name <string> type check default <bool>]")
	}
	if bytes.Equal(d, []byte("true")) {
		e.Options[string(n)] = Check{Name: n, Val: true, Default: true}
		return nil
	}
	if bytes.Equal(d, []byte("false")) {
		e.Options[string(n)] = Check{n, false, false}
		return nil
	}
	return msg.InvalidOptionSyntax.WithMsg("Default want of 'check' type was not bool. Received: " + string(d))
}

// spin type を Engine の Options にセットする
// option name <string> type spin default <int> min <int> max <int>
// このフォーマット以外は許容しない
// 各値がなかったり、int ではない時、min > max の時はエラー
func (e *Client) parseSpin(b [][]byte) error {
	n, d, mi, ma := b[2], b[6], b[8], b[10]
	if len(b) != 11 || !bytes.Equal(b[5], deflt) || !bytes.Equal(b[7], min) || !bytes.Equal(b[9], max) || len(n) == 0 {
		return msg.InvalidOptionSyntax.WithMsg("Received option type was 'spin', but malformed. The format must be [option name <string> type spin default <int> min <int> max <int>]")
	}

	df, err := strconv.Atoi(string(d))
	if err != nil {
		return msg.InvalidOptionSyntax.WithMsg("Default want of 'spin' type was not int. Received: " + string(min))
	}
	imi, err := strconv.Atoi(string(mi))
	if err != nil {
		return msg.InvalidOptionSyntax.WithMsg("Min want of 'spin' type was not int. Received: " + string(min))
	}
	ima, err := strconv.Atoi(string(ma))
	if err != nil {
		return msg.InvalidOptionSyntax.WithMsg("Max want of 'spin' type was not int. Received: " + string(min))
	}

	e.Options[string(n)] = Spin{n, df, df, imi, ima}
	return nil
}

// select type を Engine の Options にセットする
// option name <string> type combo default <string> rep(var <string>)
// このフォーマット以外は許容しない
// Default がない、var がない、default が var にない時はエラー
func (e *Client) parseSelect(b [][]byte) error {
	n, d := b[2], b[6]
	if len(b) < 9 || len(n) == 0 || len(d) == 0 {
		return msg.InvalidOptionSyntax.WithMsg("Received option type was 'combo', but malformed. The format must be [option name <string> type combo default <string> rep(var <string>)]")
	}

	s := Select{Name: n}

	i := 8
	for i < len(b) && bytes.Equal(b[i-1], selOpt) {
		s.Vars = append(s.Vars, b[i])
		i += 2
	}

	s.Index = lib.IndexOfBytes(b, d)
	if s.Index == -1 {
		return msg.InvalidOptionSyntax.WithMsg("Default want of 'combo' type was not found in vars.")
	}

	return nil
}

// button type を Engine の Options にセットする
// option name <string> type button
func (e *Client) parseButton(b [][]byte) error {
	n := b[2]
	if len(b) != 5 || len(n) == 0 {
		return msg.InvalidOptionSyntax.WithMsg("Received option type was 'button', but malformed. The format must be [option name <string> type button]")
	}
	e.Options[string(n)] = Button{n}
	return nil
}

// string type を Engine の Options にセットする
// option name <string> type string default <string>
func (e *Client) parseString(b [][]byte) error {
	n, d := b[2], b[6]
	if len(b) != 7 || len(n) == 0 || len(d) == 0 {
		return msg.InvalidOptionSyntax.WithMsg("Received option type was 'string', but malformed. The format must be [option name <string> type string default <string>]")
	}
	e.Options[string(n)] = String{n, d, d}
	return nil
}

// option name <string> type filename default <string>
func (e *Client) parseFileName(b [][]byte) error {
	n, d := b[2], b[6]
	if len(b) != 7 || len(n) == 0 || len(d) == 0 {
		return msg.InvalidOptionSyntax.WithMsg("Received option type was 'filename', but malformed. The format must be [option name <string> type filename default <string>]")
	}
	e.Options[string(n)] = FileName{n, d, d}
	return nil
}