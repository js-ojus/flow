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
[![Go Report Card](https://goreportcard.com/badge/github.com/js-ojus/flow)](https://goreportcard.com/report/github.com/js-ojus/flow)
[![GoDoc](https://godoc.org/github.com/js-ojus/flow?status.svg)](https://godoc.org/github.com/js-ojus/flow)

## STATUS

`flow` is usable in its current state, even though it hasn't acquired all the desired functionality.

**N.B.** Those who intend to use `flow` should always use the most recent release.  DO NOT use `master`!

## `flow`

`flow` is a tiny open source (Apache 2-licensed) workflow engine written in Go (golang).

### What `flow` is

As a workflow engine, `flow` intends to help in defining and driving "front office" <---> "back office" document flows.  Examples of such flows are:

- customer registration and verification of details,
- assignment of work to field personnel, and its follow-up,
- review and approval of documents, and
- trouble ticket life cycle.

`flow` provides value in scenarios of type "programming in the large", though it is not (currently) distributed in nature.  As such, it addresses only orchestration, not choreography!

`flow` - at least currently - aims to support only graph-like regimes (not hierarchical).

### What `flow` is not

`flow` is a library, not a full-stack solution.  Accordingly, it cannot be downloaded and deployed as a reday-to-use service.  It has to be used by an application that programs workflow definitions and processing.  The only "programming in the small" language supported is Go!  Of course, you could employ `flow` in a microservice architecture by wrapping it in a thin service.  That can enable you to use your favourite programming language to drive `flow`.

### Express non-goals

`flow` is intended to be small!  It is expressly **not** intended to be an enterprise-grade workflow engine.  Accordingly, import from - and export to - workflow modelling formats like BPMN/XPDL are not supported.  Similarly, executable specifications like BPEL and Wf-XML are not supported.  True enterprise-grade engines already exist for addressing such complex workflows and interoperability as require high-end engines.

## `flow` Concepts

Let us familiarise ourselves with the most important concepts and moving parts of `flow`.  Even as you read the following, it is highly recommended that you read the database table definitions in `sql` directory, as well as the corresponding object definitions in their respective `*.go` files.  That should help you in forming a mental model of `flow` faster.

### Users

`flow` neither creates nor manages users.  It assumes that an external identity provider is in charge of user management.  Similarly, `flow` does not deal with user authentication, nor does it manage authorisation of tasks not flowing (pun intended) though it!

Accordingly, users in `flow` are represented only by their unique IDs, names and unique e-mail addresses.  The identity provider has to also provide the status of the user: active vs. inactive.

### Groups

`flow`, similar to most UNIXes, implicitly creates a group for each defined user.  This is a singleton group: its only member is the corresponding user.

In addition, arbitrary general (non-singleton) groups can be defined by including one or more users as members.  The relationship between users and groups is M:N.

### Roles

Sets of document actions permitted for a given document type, can be grouped into roles.  For instance, some users should be able to raise a request, but not approve it.  Some others should be able to, in turn, approve or reject it.  Roles make it convenient to group logically related sets of permissions.

See `Document Types` and `Document Actions` for more details.

### Access Contexts

An access context is a namespace that defines jurisdictions of permissions granted to users and groups.  Examples of such jurisdictions include departments, branches, cost centres and projects.

Within an access context, a given user (though the associated singleton group) or group can be assigned one or more roles.  The effective set of permissions available to a user - in this access context - is the union of the permissions granted through all roles assigned to this user, through all groups the user is included in.

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

### Document Actions

Given a document in a particular state, each legal (for that state) action on the document could result in a new state.  `DocAction`s enumerate possible actions that affect documents.

A document action is just a string.  `flow` does not assume anything about the specifics of any document action.

### Documents

A document comprises:

- title,
- body (usually, text),
- enclosures/attachments (optional),
- tags (optional) and
- children documents (optional).

Thus, `Document` is a recursive structure.  As a document transitions from state to state, a child document is created and attached to it.  Thus, documents form a hierarchy, much like an e-mail thread, a discussion forum thread or a trouble ticket thread.

Only root documents have their own titles.  Similarly, only root documents participate in workflows.  Children documents follow their root documents along.

### Workflows

Each `DocType` defined in `flow` should have an associated `Workflow` defined.  This workflow handles the lifecycle of documents of that type.

A workflow is begun when a document of a particular type is created.  This workflow then tracks the document as it transitions from one document state to another, in reponse to either user actions or system events.  These document states and their transitions constitute the defining graph of this workflow.

### Document Events

A `DocEvent` represents a user (or system) `DocAction` on a particular document that is in a particular `DocState`.  They record the agent casuing them (system/which user), as well as the time when the action was performed.

### Nodes

A document in a particular state in its workflow, is sent to a `Node`.  There, it awaits an action to take place that moves it along the workflow, to a different state.  Thus, nodes are receivers of document events.

A `DocAction` on a document in a `DocState` may require some processing to be performed before a possible state transition.  Accordingly, applications can register custom processing functions with nodes.  Such functions are invoked by the respective nodes when an action is received by them.  These `NodeFunc`s generate the messages that are dispatched to applicable mailboxes.

### Mailboxes and Messages

Every defined user has a `Mailbox` of virtually no size limit.  Documents that transition into specific states trigger notifications for specific users.  These notifications are delivered to the mailboxes of such users.

The actual content of a notification constitutes a `Message`.  The most essential details include the document reference, the `Node` in the workflow at which the document is, and the time at which the message was sent.

Upon getting notified, users open the corresponding documents, and take appropriate actions.

## Example Structure

The following is a simple example structure.

```
Document Type : docType1
Document States : [
    docState1, docState2, docState3, docState4 // for example
]
Document Actions : [
    docAction12, docAction23, docAction34 // for the above document states
]
Document Type State Transitions : [
    docState1 --docAction12--> docState2,
    docState2 --docAction23--> docState3,
    docState3 --docAction34--> docState4,
]

Access Contexts : [
    accCtx1, accCtx2 // for example
]

Workflow : {
    Name : wFlow1,
    Initial State : docState1
}
Nodes : [
    node1: {
        Document Type : docType1,
        Workflow : wFlow1,
        Node Type : NodeTypeBegin, // note this
        From State : docState1,
        Access Context : accCtx1,
    },
    node2: {
        Document Type : docType1,
        Workflow : wFlow1,
        Node Type : NodeTypeLinear, // note this
        From State : docState2,
        Access Context : accCtx2, // a different context
    },
    node3: {
        Document Type : docType1,
        Workflow : wFlow1,
        Node Type : NodeTypeEnd, // note this
        From State : docState3,
        Access Context : accCtx1,
    },
]
```

With the above setup, we can dispatch document events to the workflow appropriately.  With each event, the workflow moves along, as defined.
