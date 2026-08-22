ALTER TABLE noon_game_points
    MODIFY COLUMN source ENUM('result', 'manual', 'typing_system') NOT NULL;

CREATE TABLE noon_game_typing_system_imports (
    id INT PRIMARY KEY AUTO_INCREMENT,
    session_id INT NOT NULL,
    export_id CHAR(36) NOT NULL,
    sha256 CHAR(64) NOT NULL,
    status ENUM('success', 'failed') NOT NULL,
    action ENUM('import', 'replace') NOT NULL,
    replaced_export_id CHAR(36) NULL,
    requested_by CHAR(36) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    filename VARCHAR(255) NULL,
    payload_size INT NOT NULL,
    message VARCHAR(255) NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_noon_game_typing_import_session FOREIGN KEY (session_id) REFERENCES noon_game_sessions(id) ON DELETE CASCADE,
    INDEX idx_noon_game_typing_import_session (session_id),
    INDEX idx_noon_game_typing_import_export (session_id, export_id),
    INDEX idx_noon_game_typing_import_active (session_id, is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
