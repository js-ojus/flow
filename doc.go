// (c) Copyright 2015-2017 JONNALAGADDA Srinivas
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package flow is a tiny workflow engine written in Go (golang).
package flow

import (
	"database/sql"
	"log"
)

const (
	// DefACRoleCount is the default number of roles a group can have
	// in an access context.
	DefACRoleCount = 1
)

var db *sql.DB

// RegisterDB provides an already initialised database handle to `flow`.
//
// N.B. This method **MUST** be called before anything else in `flow`.
func RegisterDB(sdb *sql.DB) error {
	if sdb == nil {
		log.Fatalln("given database handle is `nil`")
	}
	db = sdb

	return nil
}
