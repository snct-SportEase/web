ALTER TABLE teams
    ADD COLUMN entry_key VARCHAR(64) NOT NULL DEFAULT 'default' AFTER sport_id,
    DROP INDEX class_id,
    ADD CONSTRAINT uq_teams_class_sport_entry UNIQUE (class_id, sport_id, entry_key);

ALTER TABLE event_sports
    ADD COLUMN template_key VARCHAR(64) NULL AFTER location,
    ADD INDEX idx_event_sports_template_key (event_id, template_key);

CREATE TABLE board_game_runs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL,
    sport_id INT NOT NULL,
    game_type ENUM('shogi', 'othello') NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    location VARCHAR(255) NOT NULL DEFAULT 'ICTメディア室',
    rules_pdf_url VARCHAR(255) NULL,
    scheduled_date DATE NULL,
    win_points INT NOT NULL,
    rank_points JSON NOT NULL,
    regular_minutes INT NOT NULL,
    final_minutes INT NOT NULL,
    players_per_class INT NOT NULL,
    substitutes_per_class INT NOT NULL,
    status ENUM('draft', 'published', 'completed') NOT NULL DEFAULT 'draft',
    created_by CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_board_game_run_event_type (event_id, game_type),
    CONSTRAINT fk_board_game_run_event_sport FOREIGN KEY (event_id, sport_id)
        REFERENCES event_sports(event_id, sport_id) ON DELETE CASCADE,
    CONSTRAINT fk_board_game_run_creator FOREIGN KEY (created_by)
        REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT chk_board_game_win_points CHECK (win_points >= 0),
    CONSTRAINT chk_board_game_regular_minutes CHECK (regular_minutes > 0),
    CONSTRAINT chk_board_game_final_minutes CHECK (final_minutes > 0),
    CONSTRAINT chk_board_game_players CHECK (players_per_class > 0),
    CONSTRAINT chk_board_game_substitutes CHECK (substitutes_per_class >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE board_game_entries (
    id INT PRIMARY KEY AUTO_INCREMENT,
    run_id INT NOT NULL,
    tournament_id INT NOT NULL,
    team_id INT NOT NULL,
    class_id INT NOT NULL,
    slot_key VARCHAR(64) NOT NULL,
    seed_number INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_board_game_entry_slot (run_id, slot_key, class_id),
    UNIQUE KEY uq_board_game_entry_team (team_id),
    CONSTRAINT fk_board_game_entry_run FOREIGN KEY (run_id)
        REFERENCES board_game_runs(id) ON DELETE CASCADE,
    CONSTRAINT fk_board_game_entry_tournament FOREIGN KEY (tournament_id)
        REFERENCES tournaments(id) ON DELETE CASCADE,
    CONSTRAINT fk_board_game_entry_team FOREIGN KEY (team_id)
        REFERENCES teams(id) ON DELETE CASCADE,
    CONSTRAINT fk_board_game_entry_class FOREIGN KEY (class_id)
        REFERENCES classes(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE board_game_entry_members (
    entry_id INT NOT NULL,
    user_id CHAR(36) NOT NULL,
    member_order INT NOT NULL,
    is_substitute BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (entry_id, user_id),
    UNIQUE KEY uq_board_game_entry_member_order (entry_id, member_order),
    CONSTRAINT fk_board_game_member_entry FOREIGN KEY (entry_id)
        REFERENCES board_game_entries(id) ON DELETE CASCADE,
    CONSTRAINT fk_board_game_member_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE board_game_rankings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    run_id INT NOT NULL,
    tournament_id INT NOT NULL,
    entry_id INT NOT NULL,
    rank_number INT NOT NULL,
    win_count INT NOT NULL,
    win_points INT NOT NULL,
    rank_points INT NOT NULL,
    total_points INT NOT NULL,
    recorded_by CHAR(36) NOT NULL,
    recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_board_game_ranking_entry (run_id, tournament_id, entry_id),
    UNIQUE KEY uq_board_game_ranking_rank (run_id, tournament_id, rank_number),
    CONSTRAINT fk_board_game_ranking_run FOREIGN KEY (run_id)
        REFERENCES board_game_runs(id) ON DELETE CASCADE,
    CONSTRAINT fk_board_game_ranking_tournament FOREIGN KEY (tournament_id)
        REFERENCES tournaments(id) ON DELETE CASCADE,
    CONSTRAINT fk_board_game_ranking_entry FOREIGN KEY (entry_id)
        REFERENCES board_game_entries(id) ON DELETE CASCADE,
    CONSTRAINT fk_board_game_ranking_recorder FOREIGN KEY (recorded_by)
        REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT chk_board_game_rank CHECK (rank_number BETWEEN 1 AND 4),
    CONSTRAINT chk_board_game_wins CHECK (win_count >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE score_logs
    DROP CHECK chk_score_logs_reason,
    ADD COLUMN board_game_run_id INT NULL AFTER source_match_id,
    ADD INDEX idx_score_logs_board_game_run (board_game_run_id),
    ADD CONSTRAINT fk_score_logs_board_game_run FOREIGN KEY (board_game_run_id)
        REFERENCES board_game_runs(id) ON DELETE CASCADE,
    ADD CONSTRAINT chk_score_logs_reason CHECK (reason IN (
        'attendance_points', 'initial_points', 'survey_points', 'mic_points',
        'gym1_win1_points', 'gym1_win2_points', 'gym1_win3_points', 'gym1_champion_points',
        'gym2_win1_points', 'gym2_win2_points', 'gym2_win3_points', 'gym2_champion_points',
        'gym2_loser_bracket_champion_points',
        'ground_win1_points', 'ground_win2_points', 'ground_win3_points', 'ground_champion_points',
        'noon_game_points', 'board_game_win_points', 'board_game_rank_points'
    ));
