// Copyright 2018 murosan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import (
	"github.com/murosan/shogi-board-server/app/config"
	"go.uber.org/zap"
)

// Logger is a interface of logger.
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}

// New
func New(c config.Config) Logger {
	l, err := c.Log.Build()

	if err != nil {
		panic(err)
	}

	return l
}
