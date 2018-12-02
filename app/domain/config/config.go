// Copyright 2018 murosan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import "go.uber.org/zap"

type Config interface {
	GetEnginePath(string) string
	GetEngineNames() []string
	GetLogConf() zap.Config
}