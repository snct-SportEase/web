package models

type ClassScore struct {
	ID                      int               `json:"id"`
	EventID                 int               `json:"event_id"`
	ClassID                 int               `json:"class_id"`
	ClassName               string            `json:"class_name"`
	Season                  string            `json:"season"`
	InitialPoints           int               `json:"initial_points"`
	SurveyPoints            int               `json:"survey_points"`
	AttendancePoints        int               `json:"attendance_points"`
	Gym1Win1Points          int               `json:"gym1_win1_points"`
	Gym1Win2Points          int               `json:"gym1_win2_points"`
	Gym1Win3Points          int               `json:"gym1_win3_points"`
	Gym1ChampionPoints      int               `json:"gym1_champion_points"`
	Gym2Win1Points          int               `json:"gym2_win1_points"`
	Gym2Win2Points          int               `json:"gym2_win2_points"`
	Gym2Win3Points          int               `json:"gym2_win3_points"`
	Gym2ChampionPoints      int               `json:"gym2_champion_points"`
	GroundWin1Points        int               `json:"ground_win1_points"`
	GroundWin2Points        int               `json:"ground_win2_points"`
	GroundWin3Points        int               `json:"ground_win3_points"`
	GroundChampionPoints    int               `json:"ground_champion_points"`
	NoonGamePoints          int               `json:"noon_game_points"`
	TotalPointsCurrentEvent int               `json:"total_points_current_event"`
	RankCurrentEvent        int               `json:"rank_current_event"`
	TotalPointsOverall      int               `json:"total_points_overall"`
	RankOverall             int               `json:"rank_overall"`
	SportNames              map[string]string `json:"sport_names"`
}
