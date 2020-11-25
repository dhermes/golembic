// Go MySQL Driver - A MySQL-Driver for Go's database/sql package
//
// Copyright 2012 The Go-MySQL-Driver Authors. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package mysql

import (
	"crypto/tls"
	"sync"
)

// Registry for custom tls.Configs
var (
	tlsConfigLock     sync.RWMutex
	tlsConfigRegistry map[string]*tls.Config
)

func getTLSConfigClone(key string) (config *tls.Config) {
	tlsConfigLock.RLock()
	if v, ok := tlsConfigRegistry[key]; ok {
		config = v.Clone()
	}
	tlsConfigLock.RUnlock()
	return
}

// Returns the bool value of the input.
// The 2nd return value indicates if the input was a valid bool value
func readBool(input string) (value bool, valid bool) {
	switch input {
	case "1", "true", "TRUE", "True":
		return true, true
	case "0", "false", "FALSE", "False":
		return false, true
	}

	// Not a valid bool value
	return
}
