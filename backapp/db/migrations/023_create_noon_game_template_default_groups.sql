SET NAMES utf8mb4;
SET time_zone = '+09:00';

-- テンプレートのデフォルトグループ設定を保存するテーブル
CREATE TABLE noon_game_template_default_groups (
    id INT PRIMARY KEY AUTO_INCREMENT,
    template_key VARCHAR(50) NOT NULL COMMENT 'テンプレートキー (year_relay, course_relay, tug_of_war)',
    group_index INT NOT NULL COMMENT 'グループの順序',
    group_name VARCHAR(255) NOT NULL COMMENT 'グループ名',
    class_names JSON NOT NULL COMMENT 'クラス名のリスト (JSON配列)',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_template_group_index (template_key, group_index),
    INDEX idx_template_key (template_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='昼競技テンプレートのデフォルトグループ設定';

-- デフォルトデータを挿入
-- 学年対抗リレー
INSERT INTO noon_game_template_default_groups (template_key, group_index, group_name, class_names) VALUES
('year_relay', 1, '1年生', JSON_ARRAY('1-1', '1-2', '1-3')),
('year_relay', 2, '2年生', JSON_ARRAY('IS2', 'IT2', 'IE2')),
('year_relay', 3, '3年生', JSON_ARRAY('IS3', 'IT3', 'IE3')),
('year_relay', 4, '4年生', JSON_ARRAY('IS4', 'IT4', 'IE4')),
('year_relay', 5, '5年生', JSON_ARRAY('IS5', 'IT5', 'IE5')),
('year_relay', 6, '専教', JSON_ARRAY('専教'));

-- コース対抗リレー
INSERT INTO noon_game_template_default_groups (template_key, group_index, group_name, class_names) VALUES
('course_relay', 1, '1-1 & IEコース', JSON_ARRAY('1-1', 'IE2', 'IE3', 'IE4', 'IE5')),
('course_relay', 2, '1-2 & ISコース', JSON_ARRAY('1-2', 'IS2', 'IS3', 'IS4', 'IS5')),
('course_relay', 3, '1-3 & ITコース', JSON_ARRAY('1-3', 'IT2', 'IT3', 'IT4', 'IT5')),
('course_relay', 4, '専攻科・教員', JSON_ARRAY('専教'));

-- 綱引き
INSERT INTO noon_game_template_default_groups (template_key, group_index, group_name, class_names) VALUES
('tug_of_war', 1, '1-1 & ISコース', JSON_ARRAY('1-1', 'IS2', 'IS3', 'IS4', 'IS5')),
('tug_of_war', 2, '1-2 & ITコース', JSON_ARRAY('1-2', 'IT2', 'IT3', 'IT4', 'IT5')),
('tug_of_war', 3, '1-3 & IEコース', JSON_ARRAY('1-3', 'IE2', 'IE3', 'IE4', 'IE5')),
('tug_of_war', 4, '専攻科・教員', JSON_ARRAY('専教'));

