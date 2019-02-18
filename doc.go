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

	U "github.com/js-ojus/goutils"
)

const (
	// DefACRoleCount is the default number of roles a group can have
	// in an access context.
	DefACRoleCount = 1
)

var db *sql.DB
var blobsDir string

// RegisterDB provides an already initialised database handle to `flow`.
//
// N.B. This method **MUST** be called before anything else in `flow`.
func RegisterDB(sdb *sql.DB) *U.Error {
	if sdb == nil {
		log.Fatal("given database handle is `nil`")
	}
	db = sdb

	return nil
}

// SetBlobsDir specifies the base directory inside which blob files
// should be stored.
//
// Inside this base directory, 256 subdirectories are created as hex
// `00` through `ff`.  A blob is stored in the subdirectory whose name
// matches the first two hex digits of its SHA1 sum.
//
// N.B. Once set, this MUST NOT change between runs.  Doing so will
// result in loss of all previously stored blobs.  In addition,
// corresponding documents get corrupted.
func SetBlobsDir(base string) *U.Error {
	if base == "" {
		log.Fatal("given base directory path is empty")
	}
	blobsDir = base

	return nil
}
