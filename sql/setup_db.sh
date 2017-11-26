#!/usr/bin/env bash

# User as whom to create the database.
user="travis"

# Create database if requested.  This created database is a test
# database named `flow`.
if [ "$1" = "" ]; then
    echo Specify either '-t' for test database, or an existing database name
    exit 1
elif [ "$1" = "-t" ]; then
    mysql -u $user < ./sql/wf_database.sql > err.log 2>&1
    db="flow"
else
    db=$1
fi

# Create document-related masters.
mysql -u $user $db < ./sql/wf_doctypes_master.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_docstates_master.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_docactions_master.sql >> err.log 2>&1

# Create a local users master, if in test mode.
if [ "$1" = "-t" ]; then
    mysql -u $user $db < ./sql/users_master.sql >> err.log 2>&1
fi

# Users, groups, roles and permissions.
mysql -u $user $db < ./sql/wf_users_master.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_groups_master.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_roles_master.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_group_users.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_role_docactions.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_access_contexts.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_ac_group_roles.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_ac_group_hierarchy.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_ac_perms_v.sql >> err.log 2>&1

# Workflow related.
mysql -u $user $db < ./sql/wf_documents.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_docstate_transitions.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_docevents.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_docevent_application.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_workflows.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_workflow_nodes.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_messages.sql >> err.log 2>&1
mysql -u $user $db < ./sql/wf_mailboxes.sql >> err.log 2>&1
