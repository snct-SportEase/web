package models

import "time"

type NoonGameSession struct {
	ID                  int                         `json:"id"`
	EventID             int                         `json:"event_id"`
	Name                string                      `json:"name"`
	Description         *string                     `json:"description,omitempty"`
	Mode                string                      `json:"mode"`
	WinPoints           int                         `json:"win_points"`
	LossPoints          int                         `json:"loss_points"`
	DrawPoints          int                         `json:"draw_points"`
	ParticipationPoints int                         `json:"participation_points"`
	AllowManualPoints   bool                        `json:"allow_manual_points"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
	Groups              []*NoonGameGroupWithMembers `json:"groups,omitempty"`
	Matches             []*NoonGameMatchWithResult  `json:"matches,omitempty"`
	PointsSummary       []*NoonGamePointsSummary    `json:"points_summary,omitempty"`
}

type NoonGameGroup struct {
	ID          int       `json:"id"`
	SessionID   int       `json:"session_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type NoonGameGroupMember struct {
	ID      int     `json:"id"`
	GroupID int     `json:"group_id"`
	ClassID int     `json:"class_id"`
	Weight  float64 `json:"weight"`
	Class   *Class  `json:"class,omitempty"`
}

type NoonGameGroupWithMembers struct {
	*NoonGameGroup
	Members []*NoonGameGroupMember `json:"members"`
}

type NoonGameMatch struct {
	ID           int                   `json:"id"`
	SessionID    int                   `json:"session_id"`
	Title        *string               `json:"title,omitempty"`
	ScheduledAt  *time.Time            `json:"scheduled_at,omitempty"`
	Location     *string               `json:"location,omitempty"`
	Format       *string               `json:"format,omitempty"`
	Status       string                `json:"status"`
	Memo         *string               `json:"memo,omitempty"`
	HomeSideType string                `json:"home_side_type"`
	HomeClassID  *int                  `json:"home_class_id,omitempty"`
	HomeGroupID  *int                  `json:"home_group_id,omitempty"`
	AwaySideType string                `json:"away_side_type"`
	AwayClassID  *int                  `json:"away_class_id,omitempty"`
	AwayGroupID  *int                  `json:"away_group_id,omitempty"`
	AllowDraw    bool                  `json:"allow_draw"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	Entries      []*NoonGameMatchEntry `json:"entries,omitempty"`
}

type NoonGameResult struct {
	ID         int                     `json:"id"`
	MatchID    int                     `json:"match_id"`
	Winner     string                  `json:"winner"`
	RecordedBy string                  `json:"recorded_by"`
	RecordedAt time.Time               `json:"recorded_at"`
	Note       *string                 `json:"note,omitempty"`
	Details    []*NoonGameResultDetail `json:"details,omitempty"`
}

type NoonGameMatchWithResult struct {
	*NoonGameMatch
	Result          *NoonGameResult       `json:"result,omitempty"`
	HomeDisplayName string                `json:"home_display_name"`
	AwayDisplayName string                `json:"away_display_name"`
	WinnerDisplay   *string               `json:"winner_display,omitempty"`
	HomeClassIDs    []int                 `json:"home_class_ids"`
	AwayClassIDs    []int                 `json:"away_class_ids"`
	Entries         []*NoonGameMatchEntry `json:"entries"`
}

type NoonGamePoint struct {
	ID        int       `json:"id"`
	SessionID int       `json:"session_id"`
	MatchID   *int      `json:"match_id,omitempty"`
	ClassID   int       `json:"class_id"`
	Points    int       `json:"points"`
	Reason    *string   `json:"reason,omitempty"`
	Source    string    `json:"source"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type NoonGamePointsSummary struct {
	ClassID   int    `json:"class_id"`
	ClassName string `json:"class_name"`
	Points    int    `json:"points"`
}

type NoonGameMatchEntry struct {
	ID           int     `json:"id"`
	MatchID      int     `json:"match_id"`
	EntryIndex   int     `json:"entry_index"`
	SideType     string  `json:"side_type"`
	ClassID      *int    `json:"class_id,omitempty"`
	GroupID      *int    `json:"group_id,omitempty"`
	DisplayName  *string `json:"display_name,omitempty"`
	ResolvedName string  `json:"resolved_name"`
	ClassIDs     []int   `json:"class_ids"`
}

type NoonGameResultDetail struct {
	ID                int     `json:"id"`
	EntryID           int     `json:"entry_id"`
	Rank              *int    `json:"rank,omitempty"`
	Points            int     `json:"points"`
	Note              *string `json:"note,omitempty"`
	EntryResolvedName string  `json:"entry_resolved_name"`
}
