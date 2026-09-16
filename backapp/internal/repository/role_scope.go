package repository

// 大会共通ロールと、現在操作している大会のロールだけを有効にする。
// 呼び出し側のユーザーロール別名は ur に統一する。
const currentRoleAssignmentCondition = `(ur.event_id IS NULL OR ur.event_id = (SELECT event_id FROM active_event WHERE id = 1))`
