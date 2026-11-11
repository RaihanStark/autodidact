package repository

import (
	"database/sql"
	"fmt"
	"time"

	"autodidact-backend/internal/model"
	"github.com/google/uuid"
)

type FocusRepo struct {
	db *sql.DB
}

func NewFocusRepo(db *sql.DB) *FocusRepo {
	return &FocusRepo{db: db}
}

func (r *FocusRepo) ListFocusItems(period string) ([]model.FocusItem, error) {
	baseQuery := `
		SELECT 
			f.id,
			COALESCE(f.step_id, ''),
			f.custom_title,
			f.period,
			f.status,
			f.notes,
			f.order_index,
			f.completed_at,
			f.created_at,
			f.updated_at,
			COALESCE(s.title, f.custom_title) AS title,
			COALESCE(s.resource_url, f.resource_url) AS resource_url,
			COALESCE(NULLIF(s.resources, '[]'), NULLIF(f.resources, '[]'), '[]') AS resources,
			COALESCE(m.title, '') AS module_title,
			COALESCE(syl.id, '') AS syllabus_id,
			COALESCE(syl.title, '') AS syllabus_title
		FROM focus_items f
		LEFT JOIN steps s ON f.step_id = s.id
		LEFT JOIN modules m ON s.module_id = m.id
		LEFT JOIN syllabi syl ON m.syllabus_id = syl.id
	`
	var rows *sql.Rows
	var err error

	if period != "" && period != "all" {
		query := baseQuery + ` WHERE f.period = ? ORDER BY 
			CASE f.status 
				WHEN 'in_progress' THEN 1 
				WHEN 'todo' THEN 2 
				WHEN 'completed' THEN 3 
			END,
			f.order_index ASC,
			f.created_at DESC`
		rows, err = r.db.Query(query, period)
	} else {
		query := baseQuery + ` ORDER BY 
			CASE f.status 
				WHEN 'in_progress' THEN 1 
				WHEN 'todo' THEN 2 
				WHEN 'completed' THEN 3 
			END,
			f.order_index ASC,
			f.created_at DESC`
		rows, err = r.db.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query focus items: %w", err)
	}
	defer rows.Close()

	var items []model.FocusItem
	for rows.Next() {
		var item model.FocusItem
		var completedAt sql.NullString
		var createdAt, updatedAt, resourcesJSON string

		err := rows.Scan(
			&item.ID,
			&item.StepID,
			&item.CustomTitle,
			&item.Period,
			&item.Status,
			&item.Notes,
			&item.OrderIndex,
			&completedAt,
			&createdAt,
			&updatedAt,
			&item.Title,
			&item.ResourceURL,
			&resourcesJSON,
			&item.ModuleTitle,
			&item.SyllabusID,
			&item.SyllabusTitle,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan focus item: %w", err)
		}

		item.Resources = parseResources(resourcesJSON, item.ResourceURL)
		item.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		item.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		if completedAt.Valid && completedAt.String != "" {
			t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
			item.CompletedAt = &t
		}

		items = append(items, item)
	}

	if items == nil {
		items = []model.FocusItem{}
	}
	return items, nil
}

func (r *FocusRepo) GetFocusItemByID(id string) (*model.FocusItem, error) {
	query := `
		SELECT 
			f.id,
			COALESCE(f.step_id, ''),
			f.custom_title,
			f.period,
			f.status,
			f.notes,
			f.order_index,
			f.completed_at,
			f.created_at,
			f.updated_at,
			COALESCE(s.title, f.custom_title) AS title,
			COALESCE(s.resource_url, f.resource_url) AS resource_url,
			COALESCE(NULLIF(s.resources, '[]'), NULLIF(f.resources, '[]'), '[]') AS resources,
			COALESCE(m.title, '') AS module_title,
			COALESCE(syl.id, '') AS syllabus_id,
			COALESCE(syl.title, '') AS syllabus_title
		FROM focus_items f
		LEFT JOIN steps s ON f.step_id = s.id
		LEFT JOIN modules m ON s.module_id = m.id
		LEFT JOIN syllabi syl ON m.syllabus_id = syl.id
		WHERE f.id = ?
	`
	var item model.FocusItem
	var completedAt sql.NullString
	var createdAt, updatedAt, resourcesJSON string

	err := r.db.QueryRow(query, id).Scan(
		&item.ID,
		&item.StepID,
		&item.CustomTitle,
		&item.Period,
		&item.Status,
		&item.Notes,
		&item.OrderIndex,
		&completedAt,
		&createdAt,
		&updatedAt,
		&item.Title,
		&item.ResourceURL,
		&resourcesJSON,
		&item.ModuleTitle,
		&item.SyllabusID,
		&item.SyllabusTitle,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get focus item: %w", err)
	}

	item.Resources = parseResources(resourcesJSON, item.ResourceURL)
	item.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	item.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	if completedAt.Valid && completedAt.String != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
		item.CompletedAt = &t
	}

	return &item, nil
}

func (r *FocusRepo) GetFocusItemByStepID(stepID string) (*model.FocusItem, error) {
	var id string
	err := r.db.QueryRow(`SELECT id FROM focus_items WHERE step_id = ? AND status != 'completed' LIMIT 1`, stepID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return r.GetFocusItemByID(id)
}

func (r *FocusRepo) CreateFocusItem(req *model.CreateFocusItemRequest) (*model.FocusItem, error) {
	id := uuid.New().String()
	if req.CustomTitle != "" {
		req.CustomTitle = cleanTitleText(req.CustomTitle)
	}
	period := req.Period
	if period == "" {
		period = model.FocusPeriodNext
	}

	var stepIDParam sql.NullString
	if req.StepID != "" {
		stepIDParam = sql.NullString{String: req.StepID, Valid: true}
	}

	orderIndex := req.OrderIndex
	if orderIndex == 0 {
		_ = r.db.QueryRow(`SELECT COALESCE(MAX(order_index) + 1, 0) FROM focus_items WHERE status != 'completed'`).Scan(&orderIndex)
	}

	resources := req.Resources
	if len(resources) == 0 && req.ResourceURL != "" {
		resources = []string{req.ResourceURL}
	} else if len(resources) > 0 && req.ResourceURL == "" {
		req.ResourceURL = resources[0]
	}
	if resources == nil {
		resources = []string{}
	}
	resourcesJSON := serializeResources(resources, req.ResourceURL)

	query := `
		INSERT INTO focus_items (id, step_id, custom_title, resource_url, resources, period, status, notes, order_index, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 'todo', ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := r.db.Exec(query, id, stepIDParam, req.CustomTitle, req.ResourceURL, resourcesJSON, period, req.Notes, orderIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to create focus item: %w", err)
	}

	return r.GetFocusItemByID(id)
}

func (r *FocusRepo) UpdateFocusItem(id string, req *model.UpdateFocusItemRequest) (*model.FocusItem, error) {
	existing, err := r.GetFocusItemByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("focus item not found")
	}

	status := existing.Status
	period := existing.Period
	notes := existing.Notes
	orderIndex := existing.OrderIndex
	var completedAtQuery string

	if req.Status != nil {
		status = *req.Status
		if status == model.FocusStatusCompleted {
			completedAtQuery = ", completed_at = CURRENT_TIMESTAMP"
		} else {
			completedAtQuery = ", completed_at = NULL"
		}
	}

	if req.Period != nil {
		period = *req.Period
	}

	if req.Notes != nil {
		notes = *req.Notes
	}

	if req.OrderIndex != nil {
		orderIndex = *req.OrderIndex
	}

	var extraUpdates string
	var extraArgs []interface{}
	if req.Resources != nil {
		res := *req.Resources
		firstURL := ""
		if len(res) > 0 {
			firstURL = res[0]
		}
		resourcesJSON := serializeResources(res, firstURL)
		extraUpdates = ", resources = ?, resource_url = ?"
		extraArgs = append(extraArgs, resourcesJSON, firstURL)
	}

	query := fmt.Sprintf(`
		UPDATE focus_items
		SET status = ?, period = ?, notes = ?, order_index = ?, updated_at = CURRENT_TIMESTAMP %s %s
		WHERE id = ?
	`, extraUpdates, completedAtQuery)

	args := append([]interface{}{status, period, notes, orderIndex}, extraArgs...)
	args = append(args, id)

	_, err = r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update focus item: %w", err)
	}

	return r.GetFocusItemByID(id)
}

func (r *FocusRepo) ReorderFocusItems(ids []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE focus_items SET order_index = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for idx, id := range ids {
		if _, err := stmt.Exec(idx, id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *FocusRepo) DeleteFocusItem(id string) error {
	_, err := r.db.Exec(`DELETE FROM focus_items WHERE id = ?`, id)
	return err
}

func (r *FocusRepo) ListBacklogItems() ([]model.BacklogStepItem, error) {
	query := `
		SELECT 
			s.id,
			s.title,
			s.resource_url,
			COALESCE(s.resources, '[]') AS resources,
			sp.status,
			m.title AS module_title,
			syl.id AS syllabus_id,
			syl.title AS syllabus_title,
			EXISTS(SELECT 1 FROM focus_items f WHERE f.step_id = s.id AND f.status != 'completed') AS in_focus
		FROM steps s
		JOIN modules m ON s.module_id = m.id
		JOIN syllabi syl ON m.syllabus_id = syl.id
		JOIN step_progress sp ON sp.step_id = s.id
		ORDER BY syl.title ASC, m.order_index ASC, s.order_index ASC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query backlog items: %w", err)
	}
	defer rows.Close()

	var items []model.BacklogStepItem
	for rows.Next() {
		var item model.BacklogStepItem
		var resourcesJSON string
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.ResourceURL,
			&resourcesJSON,
			&item.Status,
			&item.ModuleTitle,
			&item.SyllabusID,
			&item.SyllabusTitle,
			&item.InFocus,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan backlog item: %w", err)
		}
		item.Resources = parseResources(resourcesJSON, item.ResourceURL)
		items = append(items, item)
	}

	if items == nil {
		items = []model.BacklogStepItem{}
	}
	return items, nil
}
