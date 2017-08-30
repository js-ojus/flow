<!--
   (c) Copyright 2017 JONNALAGADDA Srinivas

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
-->

[![Build Status](https://travis-ci.org/js-ojus/flow.svg?branch=master)](https://travis-ci.org/js-ojus/flow)

## STATUS

`flow` is inching towards a release, but is not yet usable!

## `flow`

`flow` is a tiny workflow engine written in Go (golang).

### What is `flow`?

`flow` is intended to help in defining and driving "front office" <---> "back office" document flows.  Examples of such flows are:

- customer registration and verification of details,
- assignment of work to field personnel, and its follow-up,
- review and approval of documents, and
- trouble ticket life cycle.

`flow` provides value in scenarios of type "programming in the large", though it is not (currently) distributed in nature.  As such, it addresses only orchestration, not choreography!

`flow` - at least currently - aims to support only graph-like regimes (not hierarchical).

`flow` is a library -- it requires workflow processing to be programmed.  The only "programming in the small" language supported is Go!

### What `flow` is not

`flow` is intended to be small!  It is expressly **not** intended to be an enterprise-grade workflow engine.  Accordingly, import from - and export to - workflow modelling formats like BPMN/XPDL are not supported.  Similarly, executable specifications like BPEL and Wf-XML are not supported.  True enterprise-grade engines already exist for addressing such complex workflows and interoperability as require high-end engines.

## `flow` Concepts

Let us familiarise ourselves with the most important concepts and moving parts of `flow`.

### Document Types

Each type of document, whose workflow is managed by `flow`, has a unique `DocType` that is defined by the consuming application.  Document types are one of the central concepts in `flow`.  They serve as namespaces for several other types of entities.

A document type is just a string.  `flow` does not assume anything about the specifics of any document type.  Nonetheless, it is highly recommended, but not necessary, that document types be defined in a system of hierarchical namespaces.  For example:

    PUR:RFQ

could mean that the department is 'Purchasing', while the document type is 'Request For Quotation'.  As a variant,

    PUR:ORD

could mean that the document type is 'Purchase Order'.

### Document States

Every document has various stages in its life.  The possible stages in the life of a document are encoded as a set of `DocState`s specific to its document type.

A document state is just a string.  `flow` does not assume anything about the specifics of any document state.

### Workflows

Each `DocType` defined in `flow` should have an associated `Workflow` defined.  This workflow handles the lifecycle of documents of that type.

A workflow is begun when a document of a particular type is created.  This workflow then tracks the document as it transitions from one document state to another, in reponse to either user actions or system events.  These document states and their transitions constitute the defining graph of this workflow.

### Document Actions

Given a document in a particular state, each legal (for that state) action on the document could result in a new state.  `DocAction`s enumerate possible actions that affect documents.

A document action is just a string.  `flow` does not assume anything about the specifics of any document action.

### Document Events

A `DocEvent` represents a user (or system) `DocAction` on a particular document that is in a particular `DocState`.  They record the agent casuing them (system/which user), as well as the time when the action was performed.

### Nodes

A document in a particular state in its workflow, is sent to a `Node`.  There, it awaits an action to take place that moves it along the workflow, to a different state.  Thus, nodes are receivers of document events.

A `DocAction` on a document in a `DocState` may require some processing to be performed before a possible state transition.  Accordingly, applications register custom processing functions with nodes.  Such functions are invoked by the respective nodes when an action is received by them.  These `NodeFunc`s determine the next states assumed by the documents that they process.
