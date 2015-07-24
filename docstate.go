// (c) Copyright 2015 JONNALAGADDA Srinivas
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

// DocState is one of a set of enumerated states for a document, as
// defined by the consuming application.
//
// `flow`, therefore, does not assume anything about the specifics of
// any state.
type DocState struct {
	dtype      DocType // for namespace purposes
	name       string
	successors []*DocState // possible next states
}
