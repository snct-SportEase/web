-- 旧主キーへ戻すと春秋どちらかの割り当てが失われるため、自動ロールバックしない。
SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'Event-scoped roles cannot be collapsed without losing historical assignments';
