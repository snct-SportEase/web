DELETE FROM notification_recipients
WHERE user_id IS NULL OR class_id IS NULL;

ALTER TABLE notification_recipients
    DROP INDEX uq_notification_recipients_user,
    DROP INDEX uq_notification_recipients_class,
    MODIFY COLUMN user_id CHAR(36) NOT NULL,
    MODIFY COLUMN class_id INT NOT NULL,
    ADD PRIMARY KEY (notification_id, user_id, class_id);
