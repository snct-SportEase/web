package repository

import (
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrSportLocationInUse        = errors.New("sport location is already in use")
	ErrInvalidSportLocation      = errors.New("invalid sport location")
	ErrSportManagedInOtherScreen = errors.New("sport is managed in another screen")
)

func validRegularSportLocation(location string) bool {
	if utf8.RuneCountInString(location) > 255 {
		return false
	}
	switch location {
	case "gym1", "gym2", "ground", "other":
		return true
	}
	return strings.HasPrefix(location, "other:") && strings.TrimSpace(strings.TrimPrefix(location, "other:")) != ""
}

// UpdateSportLocationAndDescription keeps the event/sport identity intact, so
// existing teams, members, tournament matches, and class roles are preserved.
func (r *sportRepository) UpdateSportLocationAndDescription(eventID, sportID int, location string, description *string) error {
	defer invalidateReads(r.db, "sports")
	if !validRegularSportLocation(location) {
		return ErrInvalidSportLocation
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Serialize location changes and new assignments for the same event.
	var lockedEventID int
	if err := tx.QueryRow("SELECT id FROM events WHERE id = ? FOR UPDATE", eventID).Scan(&lockedEventID); err != nil {
		return err
	}
	var currentLocation, sportName string
	var templateKey sql.NullString
	if err := tx.QueryRow(`SELECT es.location, es.template_key, s.name
		FROM event_sports es JOIN sports s ON s.id = es.sport_id
		WHERE es.event_id = ? AND es.sport_id = ? FOR UPDATE`, eventID, sportID).Scan(&currentLocation, &templateKey, &sportName); err != nil {
		return err
	}
	if currentLocation == "noon_game" || templateKey.String == "board_game_tournament" || sportName == "将棋" || sportName == "オセロ" {
		return ErrSportManagedInOtherScreen
	}
	if location != currentLocation && !isOtherLocation(location) {
		var count int
		if err := tx.QueryRow("SELECT COUNT(*) FROM event_sports WHERE event_id = ? AND location = ? AND sport_id <> ?", eventID, location, sportID).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return ErrSportLocationInUse
		}
	}
	if _, err := tx.Exec("UPDATE event_sports SET location = ?, description = ? WHERE event_id = ? AND sport_id = ?", location, description, eventID, sportID); err != nil {
		return err
	}
	return tx.Commit()
}
