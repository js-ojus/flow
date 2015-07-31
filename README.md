<!--
   (c) Copyright 2015 JONNALAGADDA Srinivas

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

## flow
A tiny workflow engine written in Go (golang).

`flow` is intended to help in defining and driving "front office" <---> "back office" document flows.  Examples of such flows are:

- customer registration and verification of details,
- assignment of work to field personnel, and its follow-up,
- review and approval of documents, and
- trouble ticket life cycle.

`flow` provides value in scenarios of type "programming in the large", though it is not (currently) distributed in nature.  As such, it addresses only orchestration, not choreography!

`flow` - at least currently - aims to support only graph-like regimes (not hierarchical).

`flow` is a library -- it requires workflows to be programmed.  The only "programming in the small" language supported is Go!

`flow` is intended to be small!  It is expressly **not** intended to be an enterprise-grade workflow engine.  Accordingly, import from - and export to - workflow modelling formats like BPMN/XPDL are not supported.  Similarly, executable specifications like BPEL and Wf-XML are not supported.  True enterprise-grade engines already exist for addressing such complex workflows and interoperability as require high-end engines.
