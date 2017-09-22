#!/usr/bin/env bash

# Create database if requested.  This created database is a test
# database named `flow`.
if [ "$1" = "-t" ]; then
    mysql < wf_database.sql >> err.log 2>&1
    db="flow"
else
    db=$1
fi

# Create document-related masters.
mysql $db < wf_doctypes_master.sql >> err.log 2>&1
mysql $db < wf_docstates_master.sql >> err.log 2>&1
mysql $db < wf_docactions_master.sql >> err.log 2>&1

# Create a local users master, if in test mode.
if [ "$1" = "-t" ]; then
    mysql $db < users_master.sql >> err.log 2>&1
fi

# Users, groups, roles and permissions.
mysql $db < wf_users_master.sql >> err.log 2>&1
mysql $db < wf_groups_master.sql >> err.log 2>&1
mysql $db < wf_roles_master.sql >> err.log 2>&1
mysql $db < wf_access_contexts.sql >> err.log 2>&1
mysql $db < wf_group_users.sql >> err.log 2>&1
mysql $db < wf_role_docactions.sql >> err.log 2>&1
mysql $db < wf_ac_perms_v.sql >> err.log 2>&1

# Workflow related.
mysql $db < wf_documents.sql >> err.log 2>&1
mysql $db < wf_docstate_transitions.sql >> err.log 2>&1
mysql $db < wf_docevents.sql >> err.log 2>&1
mysql $db < wf_docevent_application.sql >> err.log 2>&1
mysql $db < wf_workflows.sql >> err.log 2>&1
mysql $db < wf_workflow_nodes.sql >> err.log 2>&1
mysql $db < wf_messages.sql >> err.log 2>&1
mysql $db < wf_mailboxes.sql >> err.log 2>&1
