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

package flow

// NodeType enumerates the possible types of workflow nodes.
type NodeType string

// The following constants are represented **identically** as part of
// an enumeration in the database.  DO NOT ALTER THESE WITHOUT ALSO
// ALTERING THE DATABASE; ELSE DATA COULD GET CORRUPTED!
const (
	// NodeTypeBegin : none incoming, one outgoing
	NodeTypeBegin NodeType = "begin"
	// NodeTypeEnd : one incoming, none outgoing
	NodeTypeEnd = "end"
	// NodeTypeLinear : one incoming, one outgoing
	NodeTypeLinear = "linear"
	// NodeTypeBranch : one incoming, two or more outgoing
	NodeTypeBranch = "branch"
	// NodeTypeJoinAny : two or more incoming, one outgoing
	NodeTypeJoinAny = "joinany"
	// NodeTypeJoinAll : two or more incoming, one outgoing
	NodeTypeJoinAll = "joinall"
)

// IsValidNodeType answers `true` if the given node type is a
// recognised node type in the system.
func IsValidNodeType(ntype string) bool {
	nt := NodeType(ntype)
	switch nt {
	case NodeTypeBegin, NodeTypeEnd, NodeTypeLinear, NodeTypeBranch, NodeTypeJoinAny, NodeTypeJoinAll:
		return true

	default:
		return false
	}
}
