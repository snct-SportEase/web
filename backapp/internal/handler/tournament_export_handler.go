package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"backapp/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type tournamentExportStyles struct {
	title       int
	subtitle    int
	roundHeader int
	team        int
	winnerTeam  int
	score       int
	winnerScore int
	hLine       int
	vLineTop    int
	vLineMid    int
	vLineBottom int
	message     int
}

func (h *TournamentHandler) ExportTournamentsExcelHandler(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	tournaments, err := h.tournRepo.GetTournamentsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tournaments"})
		return
	}
	if len(tournaments) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "保存済みトーナメントがありません"})
		return
	}

	eventName := fmt.Sprintf("Event %d", eventID)
	event, err := h.eventRepo.GetEventByID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		return
	}
	if event != nil && event.Name != "" {
		eventName = event.Name
	}

	file, err := buildTournamentExportWorkbook(eventName, tournaments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Excel workbook"})
		return
	}

	var buf bytes.Buffer
	if err := file.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write Excel workbook"})
		return
	}

	filename := fmt.Sprintf("event_%d_tournaments.xlsx", eventID)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func buildTournamentExportWorkbook(eventName string, tournaments []*models.Tournament) (*excelize.File, error) {
	file := excelize.NewFile()

	styles, err := newTournamentExportStyles(file)
	if err != nil {
		return nil, err
	}

	defaultSheet := file.GetSheetName(0)
	usedSheetNames := make(map[string]int)

	for index, tournament := range tournaments {
		sheetName := uniqueSheetName(tournament.Name, usedSheetNames)
		if index == 0 {
			file.SetSheetName(defaultSheet, sheetName)
		} else {
			file.NewSheet(sheetName)
		}

		if err := renderTournamentExportSheet(file, sheetName, eventName, tournament, styles); err != nil {
			return nil, err
		}
	}

	file.SetActiveSheet(0)
	return file, nil
}

func newTournamentExportStyles(file *excelize.File) (tournamentExportStyles, error) {
	styles := tournamentExportStyles{}
	var err error

	styles.title, err = file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "#0F172A"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#DBEAFE"}},
	})
	if err != nil {
		return styles, err
	}

	styles.subtitle, err = file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#475569"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return styles, err
	}

	styles.roundHeader, err = file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#1E293B"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#E2E8F0"}},
		Border: []excelize.Border{
			{Type: "left", Color: "#64748B", Style: 1},
			{Type: "right", Color: "#64748B", Style: 1},
			{Type: "top", Color: "#64748B", Style: 1},
			{Type: "bottom", Color: "#64748B", Style: 1},
		},
	})
	if err != nil {
		return styles, err
	}

	styles.team, err = file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Color: "#0F172A"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FFFFFF"}},
		Border: []excelize.Border{
			{Type: "left", Color: "#CBD5E1", Style: 1},
			{Type: "right", Color: "#CBD5E1", Style: 1},
			{Type: "top", Color: "#CBD5E1", Style: 1},
			{Type: "bottom", Color: "#CBD5E1", Style: 1},
		},
	})
	if err != nil {
		return styles, err
	}

	styles.winnerTeam, err = file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#14532D"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#DCFCE7"}},
		Border: []excelize.Border{
			{Type: "left", Color: "#16A34A", Style: 1},
			{Type: "right", Color: "#16A34A", Style: 1},
			{Type: "top", Color: "#16A34A", Style: 1},
			{Type: "bottom", Color: "#16A34A", Style: 1},
		},
	})
	if err != nil {
		return styles, err
	}

	styles.score, err = file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Color: "#0F172A"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#CBD5E1", Style: 1},
			{Type: "right", Color: "#CBD5E1", Style: 1},
			{Type: "top", Color: "#CBD5E1", Style: 1},
			{Type: "bottom", Color: "#CBD5E1", Style: 1},
		},
	})
	if err != nil {
		return styles, err
	}

	styles.winnerScore, err = file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#14532D"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#DCFCE7"}},
		Border: []excelize.Border{
			{Type: "left", Color: "#16A34A", Style: 1},
			{Type: "right", Color: "#16A34A", Style: 1},
			{Type: "top", Color: "#16A34A", Style: 1},
			{Type: "bottom", Color: "#16A34A", Style: 1},
		},
	})
	if err != nil {
		return styles, err
	}

	styles.hLine, err = file.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "bottom", Color: "#475569", Style: 2},
		},
	})
	if err != nil {
		return styles, err
	}

	styles.vLineTop, err = file.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "right", Color: "#475569", Style: 2},
			{Type: "bottom", Color: "#475569", Style: 2},
		},
	})
	if err != nil {
		return styles, err
	}

	styles.vLineMid, err = file.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "right", Color: "#475569", Style: 2},
		},
	})
	if err != nil {
		return styles, err
	}

	styles.vLineBottom, err = file.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "right", Color: "#475569", Style: 2},
			{Type: "top", Color: "#475569", Style: 2},
		},
	})
	if err != nil {
		return styles, err
	}

	styles.message, err = file.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Italic: true, Color: "#64748B"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	if err != nil {
		return styles, err
	}

	return styles, nil
}

func renderTournamentExportSheet(file *excelize.File, sheetName string, eventName string, tournament *models.Tournament, styles tournamentExportStyles) error {
	maxRoundIndex := 0
	data := models.TournamentData{}
	if len(tournament.Data) > 0 {
		if err := json.Unmarshal(tournament.Data, &data); err != nil {
			return err
		}
	}

	for _, match := range data.Matches {
		if match.RoundIndex > maxRoundIndex {
			maxRoundIndex = match.RoundIndex
		}
	}
	if len(data.Rounds) > 0 && len(data.Rounds)-1 > maxRoundIndex {
		maxRoundIndex = len(data.Rounds) - 1
	}

	totalCols := (maxRoundIndex + 1) * 4
	if totalCols < 4 {
		totalCols = 4
	}
	lastCol, err := excelize.ColumnNumberToName(totalCols)
	if err != nil {
		return err
	}

	titleCell := "A1"
	titleRangeEnd := fmt.Sprintf("%s1", lastCol)
	subtitleRangeEnd := fmt.Sprintf("%s2", lastCol)
	if err := file.MergeCell(sheetName, titleCell, titleRangeEnd); err != nil {
		return err
	}
	if err := file.MergeCell(sheetName, "A2", subtitleRangeEnd); err != nil {
		return err
	}
	file.SetCellValue(sheetName, titleCell, fmt.Sprintf("%s - %s", eventName, tournament.Name))
	file.SetCellStyle(sheetName, titleCell, titleRangeEnd, styles.title)
	file.SetCellValue(sheetName, "A2", "保存済みトーナメントを Excel 形式で出力")
	file.SetCellStyle(sheetName, "A2", subtitleRangeEnd, styles.subtitle)

	for roundIndex := 0; roundIndex <= maxRoundIndex; roundIndex++ {
		baseCol := 1 + roundIndex*4
		headerStart, err := cellRef(baseCol, 3)
		if err != nil {
			return err
		}
		headerEnd, err := cellRef(baseCol+3, 3)
		if err != nil {
			return err
		}
		if err := file.MergeCell(sheetName, headerStart, headerEnd); err != nil {
			return err
		}
		file.SetCellValue(sheetName, headerStart, exportRoundLabel(data.Rounds, roundIndex))
		file.SetCellStyle(sheetName, headerStart, headerEnd, styles.roundHeader)

		if err := setExportColumnWidths(file, sheetName, baseCol); err != nil {
			return err
		}
	}

	if len(data.Matches) == 0 {
		file.SetCellValue(sheetName, "A5", "トーナメントの対戦データがありません。")
		file.SetCellStyle(sheetName, "A5", "A5", styles.message)
		return nil
	}

	groupedMatches := make(map[int][]models.Match)
	for _, match := range data.Matches {
		groupedMatches[match.RoundIndex] = append(groupedMatches[match.RoundIndex], match)
	}
	for roundIndex := range groupedMatches {
		sort.Slice(groupedMatches[roundIndex], func(i, j int) bool {
			return groupedMatches[roundIndex][i].Order < groupedMatches[roundIndex][j].Order
		})
	}

	for roundIndex := 0; roundIndex <= maxRoundIndex; roundIndex++ {
		baseCol := 1 + roundIndex*4
		for _, match := range groupedMatches[roundIndex] {
			centerRow := 5 + ((1 << (roundIndex + 1)) - 1) + match.Order*(1<<(roundIndex+2))
			offset := 1 << roundIndex
			topRow := centerRow - offset
			bottomRow := centerRow + offset

			if err := writeTournamentSide(file, sheetName, baseCol, topRow, match, 0, data.Contestants, styles); err != nil {
				return err
			}
			if err := writeTournamentSide(file, sheetName, baseCol, bottomRow, match, 1, data.Contestants, styles); err != nil {
				return err
			}
			if err := drawTournamentConnector(file, sheetName, baseCol, topRow, bottomRow, styles); err != nil {
				return err
			}
		}
	}

	return nil
}

func setExportColumnWidths(file *excelize.File, sheetName string, baseCol int) error {
	teamCol, err := excelize.ColumnNumberToName(baseCol)
	if err != nil {
		return err
	}
	scoreCol, err := excelize.ColumnNumberToName(baseCol + 1)
	if err != nil {
		return err
	}
	lineCol, err := excelize.ColumnNumberToName(baseCol + 2)
	if err != nil {
		return err
	}
	connectorCol, err := excelize.ColumnNumberToName(baseCol + 3)
	if err != nil {
		return err
	}
	if err := file.SetColWidth(sheetName, teamCol, teamCol, 24); err != nil {
		return err
	}
	if err := file.SetColWidth(sheetName, scoreCol, scoreCol, 6); err != nil {
		return err
	}
	if err := file.SetColWidth(sheetName, lineCol, lineCol, 4); err != nil {
		return err
	}
	if err := file.SetColWidth(sheetName, connectorCol, connectorCol, 4); err != nil {
		return err
	}
	return nil
}

func writeTournamentSide(file *excelize.File, sheetName string, baseCol int, row int, match models.Match, sideIndex int, contestants map[string]models.Contestant, styles tournamentExportStyles) error {
	teamCell, err := cellRef(baseCol, row)
	if err != nil {
		return err
	}
	scoreCell, err := cellRef(baseCol+1, row)
	if err != nil {
		return err
	}
	lineCell, err := cellRef(baseCol+2, row)
	if err != nil {
		return err
	}

	side, label, score, isWinner := exportSideDisplay(match, sideIndex, contestants)
	teamStyle := styles.team
	scoreStyle := styles.score
	if isWinner {
		teamStyle = styles.winnerTeam
		scoreStyle = styles.winnerScore
	}

	file.SetCellValue(sheetName, teamCell, label)
	file.SetCellStyle(sheetName, teamCell, teamCell, teamStyle)
	if score != "" {
		file.SetCellValue(sheetName, scoreCell, score)
	}
	file.SetCellStyle(sheetName, scoreCell, scoreCell, scoreStyle)

	if side != nil {
		file.SetCellStyle(sheetName, lineCell, lineCell, styles.hLine)
	}

	return nil
}

func exportSideDisplay(match models.Match, sideIndex int, contestants map[string]models.Contestant) (*models.Side, string, string, bool) {
	if sideIndex < len(match.Sides) {
		side := &match.Sides[sideIndex]
		label := side.Title
		if label == "" && side.ContestantID != "" {
			if contestant, ok := contestants[side.ContestantID]; ok && len(contestant.Players) > 0 {
				label = contestant.Players[0].Title
			}
		}
		if label == "" && match.RoundIndex == 0 {
			label = "BYE"
		}
		if label == "" {
			label = "TBD"
		}

		score := ""
		if len(side.Scores) > 0 {
			score = strconv.Itoa(int(side.Scores[0].MainScore))
		}

		return side, label, score, side.IsWinner
	}

	if match.RoundIndex == 0 {
		return nil, "BYE", "", false
	}
	return nil, "TBD", "", false
}

func drawTournamentConnector(file *excelize.File, sheetName string, baseCol int, topRow int, bottomRow int, styles tournamentExportStyles) error {
	connectorCol := baseCol + 3
	topCell, err := cellRef(connectorCol, topRow)
	if err != nil {
		return err
	}
	file.SetCellStyle(sheetName, topCell, topCell, styles.vLineTop)

	if bottomRow-topRow > 1 {
		midStart, err := cellRef(connectorCol, topRow+1)
		if err != nil {
			return err
		}
		midEnd, err := cellRef(connectorCol, bottomRow-1)
		if err != nil {
			return err
		}
		file.SetCellStyle(sheetName, midStart, midEnd, styles.vLineMid)
	}

	if bottomRow != topRow {
		bottomCell, err := cellRef(connectorCol, bottomRow)
		if err != nil {
			return err
		}
		file.SetCellStyle(sheetName, bottomCell, bottomCell, styles.vLineBottom)
	}

	return nil
}

func exportRoundLabel(rounds []models.Round, roundIndex int) string {
	if roundIndex >= 0 && roundIndex < len(rounds) && rounds[roundIndex].Name != "" {
		return rounds[roundIndex].Name
	}
	return fmt.Sprintf("Round %d", roundIndex+1)
}

func cellRef(col int, row int) (string, error) {
	colName, err := excelize.ColumnNumberToName(col)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%d", colName, row), nil
}

func uniqueSheetName(base string, used map[string]int) string {
	name := sanitizeSheetName(base)
	if count, exists := used[name]; exists {
		for {
			count++
			suffix := fmt.Sprintf("_%d", count)
			baseRunes := []rune(name)
			maxBaseLen := 31 - len([]rune(suffix))
			if maxBaseLen < 1 {
				maxBaseLen = 1
			}
			if len(baseRunes) > maxBaseLen {
				baseRunes = baseRunes[:maxBaseLen]
			}
			candidate := string(baseRunes) + suffix
			if _, duplicated := used[candidate]; !duplicated {
				used[name] = count
				used[candidate] = 1
				return candidate
			}
		}
	}

	used[name] = 1
	return name
}

func sanitizeSheetName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = "Tournament"
	}

	replacer := strings.NewReplacer("\\", "_", "/", "_", "?", "_", "*", "_", "[", "_", "]", "_", ":", "_")
	sanitized := replacer.Replace(trimmed)
	runes := []rune(sanitized)
	if len(runes) > 31 {
		sanitized = string(runes[:31])
	}

	if strings.TrimSpace(sanitized) == "" {
		return "Tournament"
	}
	return sanitized
}
