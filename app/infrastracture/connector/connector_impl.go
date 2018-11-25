// Copyright 2018 murosan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package connector

import (
	"bytes"
	"regexp"
	"time"

	"github.com/murosan/shogi-proxy-server/app/domain/entity/config"
	"github.com/murosan/shogi-proxy-server/app/domain/entity/converter/from_usi"
	"github.com/murosan/shogi-proxy-server/app/domain/entity/engine/option"
	"github.com/murosan/shogi-proxy-server/app/domain/entity/engine/state"
	"github.com/murosan/shogi-proxy-server/app/domain/entity/exception"
	"github.com/murosan/shogi-proxy-server/app/domain/entity/usi"
	conn "github.com/murosan/shogi-proxy-server/app/domain/infrastracture/connector"
	"github.com/murosan/shogi-proxy-server/app/service/logger"
	"go.uber.org/zap"
)

var (
	idRegex  = regexp.MustCompile(`id.*`)
	optRegex = regexp.MustCompile(`option.*`)
	name     = "name"
	author   = "author"
)

type connector struct {
	conf config.Config
	pool conn.ConnectionPool
	fu   *from_usi.FromUsi

	// TODO: 2つのエンジンから受け取る方法を考える
	egnOut chan []byte
}

func NewConnector(c config.Config, p conn.ConnectionPool, fu *from_usi.FromUsi) conn.Connector {
	return &connector{c, p, fu, make(chan []byte)}
}

func (c *connector) Connect() error {
	if c.pool.NamedEngine() != nil {
		logger.Use().Debug(exception.EngineIsAlreadyRunning.Error() + " Ignore request...")
		return nil
	}

	c.pool.Initialize() // TODO: Nameを渡して2つのエンジンを使えるように
	egn := c.pool.NamedEngine()
	stt := egn.GetState()
	logger.Use().Debug("Connect", zap.Any("EngineState", stt))

	if e := egn.Start(); e != nil {
		return e
	}

	egn.Lock()
	egn.SetState(state.Connected)
	go c.catchOutput(c.egnOut)
	if e := c.Exec(usi.CmdUsi); e != nil {
		logger.Use().Error("ExecUsiError", zap.Error(e))
		return e
	}
	if e := c.waitFor(usi.ResOk, true); e != nil {
		logger.Use().Error("WaitUsiOkError", zap.Error(e))
		return e
	}
	if e := c.Exec(usi.CmdIsReady); e != nil {
		logger.Use().Error("ExecIsReadyError", zap.Error(e))
		return e
	}
	if e := c.waitFor(usi.ResReady, false); e != nil {
		logger.Use().Error("WaitReadyOkError", zap.Error(e))
		return e
	}
	egn.Unlock()
	// TODO: 出力を受け取り続けるやつ
	return nil
}

// TODO: エンジンに接続済か確認する処理はどうにか共通化して綺麗にしたい
func (c *connector) Close() error {
	defer c.pool.Remove()
	egn := c.pool.NamedEngine()
	if egn == nil || egn.GetState() == state.NotConnected {
		logger.Use().Debug("Close", zap.Any("EngineState", state.NotConnected))
		return nil
	}

	// TODO: timeout
	return egn.Close()
}

func (c *connector) Exec(b []byte) error {
	egn := c.pool.NamedEngine()
	if egn == nil || egn.GetState() == state.NotConnected {
		logger.Use().Debug("ExecFail", zap.Any("EngineState", state.NotConnected))
		return exception.EngineIsNotRunning
	}
	if err := egn.Exec(b); err != nil {
		logger.Use().Error(exception.FailedToExecCommand.Error(), zap.Error(err))
		return err
	}
	return nil
}

func (c *connector) GetOptions() option.OptMap {
	egn := c.pool.NamedEngine()
	if egn == nil || egn.GetState() == state.NotConnected {
		logger.Use().Debug("GetOptions", zap.Any("EngineState", state.NotConnected))
		return *option.EmptyOptMap()
	}
	return egn.GetOptions()
}

func (c *connector) Start() error {
	egn := c.pool.NamedEngine()
	if egn == nil || egn.GetState() == state.NotConnected {
		logger.Use().Debug("Start", zap.Any("EngineState", state.NotConnected))
		return exception.EngineIsNotRunning
	}
	if egn.GetState() == state.Thinking {
		logger.Use().Debug("Start. Engine is thinking. Nothing to do.")
		return nil
	}
	egn.SetState(state.Thinking)
	go c.waitInf()
	return c.Exec(usi.CmdGoInf)
}

func (c *connector) SetNewOptionValue(v option.UpdateOptionValue) error {
	egn := c.pool.NamedEngine()
	if egn == nil || egn.GetState() == state.NotConnected {
		logger.Use().Debug("SetNewOptionValue", zap.Any("EngineState", state.NotConnected))
		return exception.EngineIsNotRunning
	}
	return egn.UpdateOption(v)
}

func (c *connector) StateEquals(s state.State) bool {
	egn := c.pool.NamedEngine()
	return egn != nil && egn.GetState() == s
}

func (c *connector) catchOutput(ch chan []byte) {
	egn := c.pool.NamedEngine()
	s := egn.GetScanner()

	for s.Scan() {
		b := s.Bytes()
		logger.Use().Info("StdoutPipe", zap.ByteString("EngineOut", b))
		ch <- b
	}

	if s.Err() != nil {
		logger.Use().Debug("scan:" + s.Err().Error())
	}
}

func (c *connector) waitInf() {
	for {
		select {
		case b := <-c.egnOut:
			logger.Use().Info("receive", zap.String("msg", string(b)))
		}
	}
}

func (c *connector) waitFor(exitWord []byte, parseOpt bool) error {
	timeout := make(chan struct{})
	go func() {
		time.Sleep(time.Second * 10)
		timeout <- struct{}{}
		close(timeout)
	}()
	for {
		select {
		case b := <-c.egnOut:
			if len(b) == 0 {
				continue
			}

			if bytes.Equal(b, exitWord) {
				return nil
			}

			if parseOpt && idRegex.Match(b) {
				k, v, e := c.fu.EngineID(string(b))
				if e == nil {
					c.setId(k, v)
					continue
				}
				return e
			}

			if parseOpt && optRegex.Match(b) {
				o, e := c.fu.Option(string(b))
				if e == nil {
					c.appendOption(o)
					continue
				}
				return e
			}
		case <-timeout:
			logger.Use().Error(exception.ConnectionTimeout.Error())
			return exception.ConnectionTimeout
		}
	}
}

func (c *connector) setId(k, v string) {
	egn := c.pool.NamedEngine()
	switch k {
	case name:
		egn.SetName(v)
	case author:
		egn.SetAuthor(v)
	default:
		panic("unknown id name")
	}
}

func (c *connector) appendOption(o option.Option) {
	c.pool.NamedEngine().SetOption(string(o.GetName()), o)
}
