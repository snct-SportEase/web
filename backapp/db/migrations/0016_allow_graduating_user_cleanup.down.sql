ALTER TABLE board_game_rankings
    DROP FOREIGN KEY fk_board_game_ranking_recorder;

ALTER TABLE board_game_rankings
    MODIFY COLUMN recorded_by CHAR(36) NOT NULL,
    ADD CONSTRAINT fk_board_game_ranking_recorder FOREIGN KEY (recorded_by)
        REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE board_game_runs
    DROP FOREIGN KEY fk_board_game_run_creator;

ALTER TABLE board_game_runs
    MODIFY COLUMN created_by CHAR(36) NOT NULL,
    ADD CONSTRAINT fk_board_game_run_creator FOREIGN KEY (created_by)
        REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE board_game_entry_members
    DROP FOREIGN KEY fk_board_game_member_user;

ALTER TABLE board_game_entry_members
    ADD CONSTRAINT fk_board_game_member_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE notifications
    DROP FOREIGN KEY fk_notifications_creator,
    ADD CONSTRAINT notifications_ibfk_1 FOREIGN KEY (created_by)
        REFERENCES users(id);

ALTER TABLE mic_votes
    DROP FOREIGN KEY fk_mic_votes_user,
    ADD CONSTRAINT mic_votes_ibfk_2 FOREIGN KEY (voter_user_id)
        REFERENCES users(id);

ALTER TABLE check_ins
    DROP FOREIGN KEY fk_check_ins_user,
    ADD CONSTRAINT check_ins_ibfk_1 FOREIGN KEY (user_id)
        REFERENCES users(id);
