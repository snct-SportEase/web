-- 同年度の別の季節に残った所属を、現在の大会の同名クラスに引き継ぐ。
-- 表示名・プロフィール完了状態・他年度の所属は変更しない。
UPDATE users u
JOIN classes previous_class ON previous_class.id = u.class_id
JOIN events previous_event ON previous_event.id = previous_class.event_id
JOIN active_event ae ON ae.id = 1
JOIN events current_event ON current_event.id = ae.event_id
    AND current_event.year = previous_event.year
JOIN classes current_class ON current_class.event_id = current_event.id
    AND current_class.name = previous_class.name
SET u.class_id = current_class.id
WHERE previous_event.id <> current_event.id;
