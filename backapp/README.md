# Backend


## DB(supabase)

```sql
-- ENUM型 (カスタム型) の定義
CREATE TYPE user_role AS ENUM ('root', 'admin', 'student');
CREATE TYPE event_season AS ENUM ('spring', 'autumn');
CREATE TYPE sport_location AS ENUM ('gym1', 'gym2', 'ground', 'noon_game', 'other');
CREATE TYPE check_in_purpose AS ENUM ('opening_ceremony', 'event_participation');

-- ログインを許可するメールアドレスのホワイトリスト
CREATE TABLE whitelisted_emails (
    email TEXT PRIMARY KEY,
    role user_role NOT NULL
);

-- クラス情報テーブル
CREATE TABLE classes (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    student_count INTEGER NOT NULL DEFAULT 0
);

-- ユーザーテーブル
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    display_name TEXT,
    class_id INTEGER, -- FK
    role user_role NOT NULL,
    is_profile_complete BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 大会イベントテーブル
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    year INTEGER NOT NULL,
    season event_season NOT NULL,
    start_date DATE,
    end_date DATE,
    UNIQUE(year, season)
);

-- 競技（スポーツ）テーブル
CREATE TABLE sports (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    rules TEXT,
    location sport_location NOT NULL
);

-- チームテーブル
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    class_id INTEGER NOT NULL, -- FK
    sport_id INTEGER NOT NULL, -- FK
    event_id INTEGER NOT NULL, -- FK
    UNIQUE(class_id, sport_id, event_id)
);

-- チームメンバーの中間テーブル
CREATE TABLE team_members (
    team_id INTEGER NOT NULL, -- FK
    user_id UUID NOT NULL, -- FK
    PRIMARY KEY (team_id, user_id)
);

-- トーナメントテーブル
CREATE TABLE tournaments (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    event_id INTEGER NOT NULL, -- FK
    sport_id INTEGER NOT NULL -- FK
);

-- 試合テーブル
CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    tournament_id INTEGER, -- FK
    round INTEGER,
    match_number_in_round INTEGER,
    team1_id INTEGER, -- FK
    team2_id INTEGER, -- FK
    team1_score INTEGER,
    team2_score INTEGER,
    winner_team_id INTEGER, -- FK
    next_match_id INTEGER, -- FK
    status VARCHAR(50),
    start_time TIMESTAMPTZ,
    court_number TEXT
);

-- クラスごとの集計得点テーブル
CREATE TABLE class_scores (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL, -- FK
    class_id INTEGER NOT NULL, -- FK
    initial_points INTEGER DEFAULT 0,
    survey_points INTEGER DEFAULT 0,
    attendance_points INTEGER DEFAULT 0,
    gym1_win1_points INTEGER DEFAULT 0,
    gym1_win2_points INTEGER DEFAULT 0,
    gym1_win3_points INTEGER DEFAULT 0,
    gym1_champion_points INTEGER DEFAULT 0,
    gym2_win1_points INTEGER DEFAULT 0,
    gym2_win2_points INTEGER DEFAULT 0,
    gym2_win3_points INTEGER DEFAULT 0,
    gym2_champion_points INTEGER DEFAULT 0,
    ground_win1_points INTEGER DEFAULT 0,
    ground_win2_points INTEGER DEFAULT 0,
    ground_win3_points INTEGER DEFAULT 0,
    ground_champion_points INTEGER DEFAULT 0,
    noon_game_points INTEGER DEFAULT 0,
    mvp_points INTEGER DEFAULT 0,
    total_points_current_event INTEGER,
    rank_current_event INTEGER,
    total_points_overall INTEGER,
    rank_overall INTEGER,
    UNIQUE(event_id, class_id)
);

-- 得点履歴テーブル
CREATE TABLE score_logs (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL, -- FK
    class_id INTEGER NOT NULL, -- FK
    points INTEGER NOT NULL,
    reason TEXT NOT NULL,
    source_match_id INTEGER, -- FK
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 出席チェックインテーブル
CREATE TABLE check_ins (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL, -- FK
    event_id INTEGER NOT NULL, -- FK
    purpose check_in_purpose NOT NULL,
    checked_in_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- MVP投票テーブル
CREATE TABLE mvp_votes (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL, -- FK
    voter_user_id UUID NOT NULL, -- FK
    voted_for_class_id INTEGER NOT NULL, -- FK
    points INTEGER NOT NULL DEFAULT 3,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(event_id, voter_user_id)
);

-- 通知テーブル
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_by UUID, -- FK
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 通知の宛先テーブル
CREATE TABLE notification_recipients (
    notification_id INTEGER NOT NULL, -- FK
    user_id UUID, -- FK
    class_id INTEGER, -- FK
    PRIMARY KEY (notification_id, user_id, class_id)
);

-- users テーブル
ALTER TABLE users ADD CONSTRAINT fk_users_class_id FOREIGN KEY (class_id) REFERENCES classes(id);

-- teams テーブル
ALTER TABLE teams ADD CONSTRAINT fk_teams_class_id FOREIGN KEY (class_id) REFERENCES classes(id);
ALTER TABLE teams ADD CONSTRAINT fk_teams_sport_id FOREIGN KEY (sport_id) REFERENCES sports(id);
ALTER TABLE teams ADD CONSTRAINT fk_teams_event_id FOREIGN KEY (event_id) REFERENCES events(id);

-- team_members テーブル
ALTER TABLE team_members ADD CONSTRAINT fk_team_members_team_id FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
ALTER TABLE team_members ADD CONSTRAINT fk_team_members_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- tournaments テーブル
ALTER TABLE tournaments ADD CONSTRAINT fk_tournaments_event_id FOREIGN KEY (event_id) REFERENCES events(id);
ALTER TABLE tournaments ADD CONSTRAINT fk_tournaments_sport_id FOREIGN KEY (sport_id) REFERENCES sports(id);

-- matches テーブル
ALTER TABLE matches ADD CONSTRAINT fk_matches_tournament_id FOREIGN KEY (tournament_id) REFERENCES tournaments(id);
ALTER TABLE matches ADD CONSTRAINT fk_matches_team1_id FOREIGN KEY (team1_id) REFERENCES teams(id);
ALTER TABLE matches ADD CONSTRAINT fk_matches_team2_id FOREIGN KEY (team2_id) REFERENCES teams(id);
ALTER TABLE matches ADD CONSTRAINT fk_matches_winner_team_id FOREIGN KEY (winner_team_id) REFERENCES teams(id);
ALTER TABLE matches ADD CONSTRAINT fk_matches_next_match_id FOREIGN KEY (next_match_id) REFERENCES matches(id);

-- class_scores テーブル
ALTER TABLE class_scores ADD CONSTRAINT fk_class_scores_event_id FOREIGN KEY (event_id) REFERENCES events(id);
ALTER TABLE class_scores ADD CONSTRAINT fk_class_scores_class_id FOREIGN KEY (class_id) REFERENCES classes(id);

-- score_logs テーブル
ALTER TABLE score_logs ADD CONSTRAINT fk_score_logs_event_id FOREIGN KEY (event_id) REFERENCES events(id);
ALTER TABLE score_logs ADD CONSTRAINT fk_score_logs_class_id FOREIGN KEY (class_id) REFERENCES classes(id);
ALTER TABLE score_logs ADD CONSTRAINT fk_score_logs_source_match_id FOREIGN KEY (source_match_id) REFERENCES matches(id);

-- check_ins テーブル
ALTER TABLE check_ins ADD CONSTRAINT fk_check_ins_user_id FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE check_ins ADD CONSTRAINT fk_check_ins_event_id FOREIGN KEY (event_id) REFERENCES events(id);

-- mvp_votes テーブル
ALTER TABLE mvp_votes ADD CONSTRAINT fk_mvp_votes_event_id FOREIGN KEY (event_id) REFERENCES events(id);
ALTER TABLE mvp_votes ADD CONSTRAINT fk_mvp_votes_voter_user_id FOREIGN KEY (voter_user_id) REFERENCES users(id);
ALTER TABLE mvp_votes ADD CONSTRAINT fk_mvp_votes_voted_for_class_id FOREIGN KEY (voted_for_class_id) REFERENCES classes(id);

-- notifications テーブル
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_created_by FOREIGN KEY (created_by) REFERENCES users(id);

-- notification_recipients テーブル
ALTER TABLE notification_recipients ADD CONSTRAINT fk_recipients_notification_id FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE;
ALTER TABLE notification_recipients ADD CONSTRAINT fk_recipients_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE notification_recipients ADD CONSTRAINT fk_recipients_class_id FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE;
```

```json
[
  {
    "テーブル名": "check_ins",
    "ポリシー名": "Allow admin full access on check_ins",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "check_ins",
    "ポリシー名": "Students can insert their own check_in",
    "コマンド": "INSERT",
    "条件式 (USING)": null,
    "チェック式 (WITH CHECK)": "(user_id = auth.uid())"
  },
  {
    "テーブル名": "check_ins",
    "ポリシー名": "Students can view their own check_ins",
    "コマンド": "SELECT",
    "条件式 (USING)": "(user_id = auth.uid())",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "class_scores",
    "ポリシー名": "Allow admin full access on class_scores",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "class_scores",
    "ポリシー名": "Allow authenticated users to read class_scores",
    "コマンド": "SELECT",
    "条件式 (USING)": "(auth.role() = 'authenticated'::text)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "classes",
    "ポリシー名": "Allow admin full access on classes",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "classes",
    "ポリシー名": "Allow authenticated users to read classes",
    "コマンド": "SELECT",
    "条件式 (USING)": "(auth.role() = 'authenticated'::text)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "events",
    "ポリシー名": "Allow admin full access on events",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "events",
    "ポリシー名": "Allow authenticated users to read events",
    "コマンド": "SELECT",
    "条件式 (USING)": "(auth.role() = 'authenticated'::text)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "matches",
    "ポリシー名": "Allow admin full access on matches",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "matches",
    "ポリシー名": "Allow authenticated users to read matches",
    "コマンド": "SELECT",
    "条件式 (USING)": "(auth.role() = 'authenticated'::text)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "mvp_votes",
    "ポリシー名": "Allow admin full access on mvp_votes",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "notification_recipients",
    "ポリシー名": "Allow admin full access on notification_recipients",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "notification_recipients",
    "ポリシー名": "Users can view their own recipient status",
    "コマンド": "SELECT",
    "条件式 (USING)": "((user_id = auth.uid()) OR (class_id = ( SELECT u.class_id\n   FROM users u\n  WHERE (u.id = auth.uid()))))",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "notifications",
    "ポリシー名": "Allow admin full access on notifications",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "notifications",
    "ポリシー名": "Users can view notifications sent to them",
    "コマンド": "SELECT",
    "条件式 (USING)": "(EXISTS ( SELECT 1\n   FROM notification_recipients\n  WHERE ((notification_recipients.notification_id = notifications.id) AND ((notification_recipients.user_id = auth.uid()) OR (notification_recipients.class_id = ( SELECT u.class_id\n           FROM users u\n          WHERE (u.id = auth.uid())))))))",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "score_logs",
    "ポリシー名": "Allow admin full access on score_logs",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "score_logs",
    "ポリシー名": "Allow authenticated users to read score_logs",
    "コマンド": "SELECT",
    "条件式 (USING)": "(auth.role() = 'authenticated'::text)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "sports",
    "ポリシー名": "Allow admin full access on sports",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "sports",
    "ポリシー名": "Allow authenticated users to read sports",
    "コマンド": "SELECT",
    "条件式 (USING)": "(auth.role() = 'authenticated'::text)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "team_members",
    "ポリシー名": "Allow admin full access on team_members",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "team_members",
    "ポリシー名": "Allow authenticated users to read team_members",
    "コマンド": "SELECT",
    "条件式 (USING)": "(auth.role() = 'authenticated'::text)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "teams",
    "ポリシー名": "Allow admin full access on teams",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "teams",
    "ポリシー名": "Allow authenticated users to read teams",
    "コマンド": "SELECT",
    "条件式 (USING)": "(auth.role() = 'authenticated'::text)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "tournaments",
    "ポリシー名": "Allow admin full access on tournaments",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "tournaments",
    "ポリシー名": "Allow authenticated users to read tournaments",
    "コマンド": "SELECT",
    "条件式 (USING)": "(auth.role() = 'authenticated'::text)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "users",
    "ポリシー名": "Allow root and admin full access on users",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  },
  {
    "テーブル名": "users",
    "ポリシー名": "Students can update their own user info",
    "コマンド": "UPDATE",
    "条件式 (USING)": "(auth.uid() = id)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "users",
    "ポリシー名": "Students can view their own user info",
    "コマンド": "SELECT",
    "条件式 (USING)": "(auth.uid() = id)",
    "チェック式 (WITH CHECK)": null
  },
  {
    "テーブル名": "whitelisted_emails",
    "ポリシー名": "Allow admin full access on whitelisted_emails",
    "コマンド": "ALL",
    "条件式 (USING)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))",
    "チェック式 (WITH CHECK)": "(get_my_role() = ANY (ARRAY['root'::text, 'admin'::text]))"
  }
]
```