ALTER TABLE check_ins
    DROP FOREIGN KEY check_ins_ibfk_1,
    ADD CONSTRAINT fk_check_ins_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE mic_votes
    DROP FOREIGN KEY mic_votes_ibfk_2,
    ADD CONSTRAINT fk_mic_votes_user FOREIGN KEY (voter_user_id)
        REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE notifications
    DROP FOREIGN KEY notifications_ibfk_1,
    ADD CONSTRAINT fk_notifications_creator FOREIGN KEY (created_by)
        REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE board_game_entry_members
    DROP FOREIGN KEY fk_board_game_member_user;

ALTER TABLE board_game_entry_members
    ADD CONSTRAINT fk_board_game_member_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE board_game_runs
    DROP FOREIGN KEY fk_board_game_run_creator;

ALTER TABLE board_game_runs
    MODIFY COLUMN created_by CHAR(36) NULL,
    ADD CONSTRAINT fk_board_game_run_creator FOREIGN KEY (created_by)
        REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE board_game_rankings
    DROP FOREIGN KEY fk_board_game_ranking_recorder;

ALTER TABLE board_game_rankings
    MODIFY COLUMN recorded_by CHAR(36) NULL,
    ADD CONSTRAINT fk_board_game_ranking_recorder FOREIGN KEY (recorded_by)
        REFERENCES users(id) ON DELETE SET NULL;
