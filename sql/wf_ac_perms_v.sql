CREATE OR REPLACE VIEW wf_ac_perms_v AS
SELECT ac.ns_id, ac.group_id, gu.user_id, ac.role_id, rdas.doctype_id, rdas.docaction_id
FROM wf_access_contexts ac, wf_group_users gu, wf_role_docactions rdas
JOIN ON ac.group_id = gu.group_id
JOIN ON ac.role_id = rdas.role_id;
