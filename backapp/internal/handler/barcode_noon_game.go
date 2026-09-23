package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *BarcodeHandler) checkInNoonGame(c *gin.Context, req models.BarcodeCheckInRequest, user *models.User) {
	if h.noonRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "昼競技チェックインを初期化できません"})
		return
	}
	session, err := h.noonRepo.GetSessionByID(req.NoonGameSessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "昼競技セッションの取得に失敗しました"})
		return
	}
	if session == nil || session.EventID != req.EventID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "選択した昼競技がこの大会に存在しません"})
		return
	}
	match, err := h.noonRepo.GetMatchByID(req.MatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "昼競技の試合確認に失敗しました"})
		return
	}
	if match == nil || match.SessionID != session.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "選択した試合がこの昼競技に存在しません"})
		return
	}
	classIDs, err := h.noonGameMatchClassIDs(match)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "昼競技の参加クラス確認に失敗しました"})
		return
	}
	if user.ClassID == nil || !containsNoonGameClassID(classIDs, *user.ClassID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "このユーザーはこの昼競技の試合に参加しません"})
		return
	}
	if err := h.noonRepo.CheckInMember(req.EventID, session.ID, match.ID, user.ID); err != nil {
		if errors.Is(err, repository.ErrNoonGameAlreadyCheckedIn) {
			c.JSON(http.StatusConflict, gin.H{"error": "チェックイン済みです", "already_checked_in": true})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "昼競技のチェックインに失敗しました"})
		return
	}
	displayName := ""
	if user.DisplayName != nil {
		displayName = *user.DisplayName
	}
	c.JSON(http.StatusOK, gin.H{
		"valid": true, "confirmed": true, "checked_in": true,
		"event_id": req.EventID, "sport_id": req.SportID, "sport_name": session.Name,
		"round": 1, "match_id": match.ID, "match_ids": []int{match.ID},
		"user_id": user.ID, "display_name": displayName, "barcode_data": req.BarcodeData,
	})
}

func (h *BarcodeHandler) noonGameMatchClassIDs(match *models.NoonGameMatchWithResult) ([]int, error) {
	classIDs := make([]int, 0)
	seen := make(map[int]bool)
	addClassID := func(classID int) {
		if classID > 0 && !seen[classID] {
			seen[classID] = true
			classIDs = append(classIDs, classID)
		}
	}
	for _, entry := range match.Entries {
		if entry == nil {
			continue
		}
		if entry.ClassID != nil {
			addClassID(*entry.ClassID)
		}
		if entry.GroupID != nil {
			members, err := h.noonRepo.GetGroupMembers(*entry.GroupID)
			if err != nil {
				return nil, err
			}
			for _, member := range members {
				if member != nil {
					addClassID(member.ClassID)
				}
			}
		}
	}
	return classIDs, nil
}

func containsNoonGameClassID(classIDs []int, target int) bool {
	for _, classID := range classIDs {
		if classID == target {
			return true
		}
	}
	return false
}

func (h *BarcodeHandler) getNoonGameMatchCheckIns(c *gin.Context, eventID, sportID, sessionID, matchID int) {
	if h.noonRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "昼競技チェックインを初期化できません"})
		return
	}
	session, err := h.noonRepo.GetSessionByID(sessionID)
	if err != nil || session == nil || session.EventID != eventID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "選択した昼競技がこの大会に存在しません"})
		return
	}
	match, err := h.noonRepo.GetMatchByID(matchID)
	if err != nil || match == nil || match.SessionID != sessionID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "選択した試合がこの昼競技に存在しません"})
		return
	}
	classIDs, err := h.noonGameMatchClassIDs(match)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "昼競技の参加クラス確認に失敗しました"})
		return
	}
	checkIns, err := h.noonRepo.GetMatchCheckIns(eventID, sessionID, matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "昼競技のチェックイン一覧取得に失敗しました"})
		return
	}
	checkedAt := make(map[string]time.Time, len(checkIns))
	for _, checkIn := range checkIns {
		if checkIn != nil {
			checkedAt[checkIn.UserID] = checkIn.CheckedInAt
		}
	}
	checked := make([]*models.MatchCheckInMember, 0)
	unchecked := make([]*models.MatchCheckInMember, 0)
	for _, classID := range classIDs {
		class, err := h.classRepo.GetClassByID(classID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "クラス情報の取得に失敗しました"})
			return
		}
		members, err := h.classRepo.GetClassMembers(classID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "クラスメンバーの取得に失敗しました"})
			return
		}
		className := ""
		if class != nil {
			className = class.Name
		}
		for _, member := range members {
			if member == nil {
				continue
			}
			item := &models.MatchCheckInMember{
				UserID: member.ID, Email: member.Email, DisplayName: member.DisplayName,
				ClassID: classID, ClassName: className, TeamName: session.Name,
				EventID: eventID, SportID: sportID, MatchID: matchID, Round: 1,
			}
			if checkedInAt, ok := checkedAt[member.ID]; ok {
				item.CheckedInAt = checkedInAt
				checked = append(checked, item)
			} else {
				unchecked = append(unchecked, item)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"members": checked, "count": len(checked),
		"checked_in_members": checked, "checked_in_count": len(checked),
		"unchecked_members": unchecked, "unchecked_count": len(unchecked),
	})
}
