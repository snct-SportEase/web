-- 専教は学生の在籍人数を持たず、アンケート得点の対象にも含めない。
UPDATE classes
SET student_count = 0
WHERE name = '専教';

DELETE score_log
FROM score_logs score_log
JOIN classes special_class ON special_class.id = score_log.class_id
WHERE special_class.name = '専教'
  AND score_log.reason = 'survey_points';
