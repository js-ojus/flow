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

// DocType enumerates the types of documents in the system, as defined
// by the consuming application.  Each document type has an associated
// workflow definition that drives its life cycle.
//
// Accordingly, `flow` does not assume anything about the specifics of
// the any document type.  Instead, it treats document types as plain,
// but controlled, vocabulary.  Nonetheless, it is highly recommended,
// but not necessary, that document types be defined in a system of
// hierarchical namespaces. For example:
//
//     PUR:RFQ
//
// could mean that the department is 'Purchasing', while the document
// type is 'Request For Quotation'.  As a variant,
//
//     PUR:ORD
//
// could mean that the document type is 'Purchase Order'.
//
// N.B. All document types must be defined as constant strings.
type DocType string
