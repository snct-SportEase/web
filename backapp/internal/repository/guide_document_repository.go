package repository

import (
	"backapp/internal/models"
	"database/sql"
	"errors"
)

type GuideDocumentRepository interface {
	ListGuideDocuments(eventID int) ([]*models.GuideDocument, error)
	CreateGuideDocument(doc *models.GuideDocument) (int64, error)
	DeleteGuideDocument(id int) error
}

type guideDocumentRepository struct {
	db *sql.DB
}

func NewGuideDocumentRepository(db *sql.DB) GuideDocumentRepository {
	return &guideDocumentRepository{db: db}
}

func (r *guideDocumentRepository) ListGuideDocuments(eventID int) ([]*models.GuideDocument, error) {
	rows, err := r.db.Query(`
		SELECT id, event_id, title, description, pdf_url, created_at, updated_at
		FROM guide_documents
		WHERE event_id = ?
		ORDER BY created_at DESC, id DESC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*models.GuideDocument
	for rows.Next() {
		doc := &models.GuideDocument{}
		var description sql.NullString
		if err := rows.Scan(&doc.ID, &doc.EventID, &doc.Title, &description, &doc.PdfURL, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, err
		}
		if description.Valid {
			doc.Description = &description.String
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

func (r *guideDocumentRepository) CreateGuideDocument(doc *models.GuideDocument) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO guide_documents (event_id, title, description, pdf_url)
		VALUES (?, ?, ?, ?)
	`, doc.EventID, doc.Title, doc.Description, doc.PdfURL)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *guideDocumentRepository) DeleteGuideDocument(id int) error {
	result, err := r.db.Exec("DELETE FROM guide_documents WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func IsGuideDocumentNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
