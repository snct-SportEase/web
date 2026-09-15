ALTER TABLE notification_recipients
    DROP PRIMARY KEY,
    MODIFY COLUMN user_id CHAR(36) NULL,
    MODIFY COLUMN class_id INT NULL,
    ADD UNIQUE KEY uq_notification_recipients_user (notification_id, user_id),
    ADD UNIQUE KEY uq_notification_recipients_class (notification_id, class_id);
