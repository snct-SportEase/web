-- 同年度の春大会で登録済みの在籍人数を、既存の秋大会の同名クラスへ引き継ぐ。
-- 秋大会ですでに登録済みの値は上書きしない。
UPDATE classes autumn_class
JOIN events autumn_event
  ON autumn_event.id = autumn_class.event_id
  AND autumn_event.season = 'autumn'
JOIN (
  SELECT `year`, MAX(id) AS event_id
  FROM events
  WHERE season = 'spring'
  GROUP BY `year`
) latest_spring_event
  ON latest_spring_event.`year` = autumn_event.`year`
JOIN classes spring_class
  ON spring_class.event_id = latest_spring_event.event_id
  AND spring_class.name = autumn_class.name
SET autumn_class.student_count = spring_class.student_count
WHERE autumn_class.student_count = 0
  AND spring_class.student_count > 0;
