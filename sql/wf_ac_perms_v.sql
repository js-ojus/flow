CREATE OR REPLACE VIEW wf_ac_perms_v AS
SELECT ac_grs.ac_id, ac_grs.group_id, gu.user_id, ac_grs.role_id, rdas.doctype_id, rdas.docaction_id
FROM wf_ac_group_roles ac_grs
JOIN wf_group_users gu ON ac_grs.group_id = gu.group_id
JOIN wf_role_docactions rdas ON ac_grs.role_id = rdas.role_id;
