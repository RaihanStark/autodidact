package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"autodidact-backend/internal/model"
	"github.com/google/uuid"
)

type SyllabusRepo struct {
	db *sql.DB
}

func NewSyllabusRepo(db *sql.DB) *SyllabusRepo {
	return &SyllabusRepo{db: db}
}

func (r *SyllabusRepo) ListSyllabi() ([]model.Syllabus, error) {
	query := `
		SELECT 
			s.id, s.title, s.slug, s.description, s.is_template, s.created_at, s.updated_at,
			COUNT(st.id) AS total_steps,
			COUNT(CASE WHEN sp.status = 'completed' THEN 1 END) AS completed_steps
		FROM syllabi s
		LEFT JOIN modules m ON m.syllabus_id = s.id
		LEFT JOIN steps st ON st.module_id = m.id
		LEFT JOIN step_progress sp ON sp.step_id = st.id
		GROUP BY s.id
		ORDER BY s.updated_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query syllabi: %w", err)
	}
	defer rows.Close()

	var syllabi []model.Syllabus
	for rows.Next() {
		var s model.Syllabus
		var createdAt, updatedAt string
		err := rows.Scan(
			&s.ID, &s.Title, &s.Slug, &s.Description, &s.IsTemplate,
			&createdAt, &updatedAt, &s.TotalSteps, &s.CompletedSteps,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan syllabus: %w", err)
		}
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		syllabi = append(syllabi, s)
	}

	if syllabi == nil {
		syllabi = []model.Syllabus{}
	}
	return syllabi, nil
}

func (r *SyllabusRepo) GetSyllabusByID(id string) (*model.Syllabus, error) {
	var s model.Syllabus
	var createdAt, updatedAt string
	err := r.db.QueryRow(`
		SELECT id, title, slug, description, is_template, created_at, updated_at
		FROM syllabi WHERE id = ?
	`, id).Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.IsTemplate, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get syllabus: %w", err)
	}
	s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	// Fetch modules
	modRows, err := r.db.Query(`
		SELECT id, syllabus_id, title, description, order_index, created_at
		FROM modules
		WHERE syllabus_id = ?
		ORDER BY order_index ASC, created_at ASC
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get modules: %w", err)
	}
	defer modRows.Close()

	modules := []model.Module{}
	moduleMap := make(map[string]*model.Module)

	for modRows.Next() {
		var m model.Module
		var mCreatedAt string
		err := modRows.Scan(&m.ID, &m.SyllabusID, &m.Title, &m.Description, &m.OrderIndex, &mCreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan module: %w", err)
		}
		m.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", mCreatedAt)
		m.Steps = []model.Step{}
		modules = append(modules, m)
	}

	for i := range modules {
		moduleMap[modules[i].ID] = &modules[i]
	}

	// Fetch steps with progress
	stepRows, err := r.db.Query(`
		SELECT 
			st.id, st.module_id, st.title, st.type, st.content, st.resource_url,
			COALESCE(st.resources, '[]'),
			st.acceptance_criteria, st.estimated_minutes, st.order_index, st.created_at,
			COALESCE(sp.status, 'pending'),
			COALESCE(sp.notes, ''),
			COALESCE(sp.submission_url, ''),
			sp.completed_at
		FROM steps st
		JOIN modules m ON m.id = st.module_id
		LEFT JOIN step_progress sp ON sp.step_id = st.id
		WHERE m.syllabus_id = ?
		ORDER BY st.order_index ASC, st.created_at ASC
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get steps: %w", err)
	}
	defer stepRows.Close()

	totalSteps := 0
	completedSteps := 0

	for stepRows.Next() {
		var st model.Step
		var sCreatedAt, resourcesJSON string
		var completedAt sql.NullString
		err := stepRows.Scan(
			&st.ID, &st.ModuleID, &st.Title, &st.Type, &st.Content, &st.ResourceURL,
			&resourcesJSON,
			&st.AcceptanceCriteria, &st.EstimatedMinutes, &st.OrderIndex, &sCreatedAt,
			&st.Status, &st.Notes, &st.SubmissionURL, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan step: %w", err)
		}
		st.Resources = parseResources(resourcesJSON, st.ResourceURL)
		st.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", sCreatedAt)
		if completedAt.Valid && completedAt.String != "" {
			t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
			st.CompletedAt = &t
		}

		totalSteps++
		if st.Status == model.StepStatusCompleted {
			completedSteps++
		}

		if mod, exists := moduleMap[st.ModuleID]; exists {
			mod.Steps = append(mod.Steps, st)
		}
	}

	s.Modules = modules
	s.TotalSteps = totalSteps
	s.CompletedSteps = completedSteps

	return &s, nil
}

func (r *SyllabusRepo) CreateSyllabus(title, description string) (*model.Syllabus, error) {
	id := uuid.New().String()
	slug := slugify(title) + "-" + id[:8]

	_, err := r.db.Exec(`
		INSERT INTO syllabi (id, title, slug, description, is_template)
		VALUES (?, ?, ?, ?, 0)
	`, id, title, slug, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create syllabus: %w", err)
	}

	return r.GetSyllabusByID(id)
}

func (r *SyllabusRepo) UpdateSyllabus(id, title, description string) (*model.Syllabus, error) {
	_, err := r.db.Exec(`
		UPDATE syllabi
		SET title = ?, description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, title, description, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update syllabus: %w", err)
	}
	return r.GetSyllabusByID(id)
}

func (r *SyllabusRepo) DeleteSyllabus(id string) error {
	_, err := r.db.Exec("DELETE FROM syllabi WHERE id = ?", id)
	return err
}

func (r *SyllabusRepo) DuplicateSyllabus(id string) (*model.Syllabus, error) {
	orig, err := r.GetSyllabusByID(id)
	if err != nil || orig == nil {
		return nil, fmt.Errorf("original syllabus not found")
	}

	newID := uuid.New().String()
	newTitle := orig.Title + " (Copy)"
	newSlug := slugify(newTitle) + "-" + newID[:8]

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO syllabi (id, title, slug, description, is_template)
		VALUES (?, ?, ?, ?, 0)
	`, newID, newTitle, newSlug, orig.Description)
	if err != nil {
		return nil, err
	}

	for _, mod := range orig.Modules {
		newModID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO modules (id, syllabus_id, title, description, order_index)
			VALUES (?, ?, ?, ?, ?)
		`, newModID, newID, mod.Title, mod.Description, mod.OrderIndex)
		if err != nil {
			return nil, err
		}

		for _, st := range mod.Steps {
			newStepID := uuid.New().String()
			resourcesJSON := serializeResources(st.Resources, st.ResourceURL)
			_, err = tx.Exec(`
				INSERT INTO steps (id, module_id, title, type, content, resource_url, resources, acceptance_criteria, estimated_minutes, order_index)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, newStepID, newModID, st.Title, st.Type, st.Content, st.ResourceURL, resourcesJSON, st.AcceptanceCriteria, st.EstimatedMinutes, st.OrderIndex)
			if err != nil {
				return nil, err
			}

			newProgID := uuid.New().String()
			_, err = tx.Exec(`
				INSERT INTO step_progress (id, step_id, status, notes, submission_url)
				VALUES (?, ?, 'pending', '', '')
			`, newProgID, newStepID)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetSyllabusByID(newID)
}

func (r *SyllabusRepo) CreateModule(syllabusID, title, description string) (*model.Module, error) {
	id := uuid.New().String()

	var maxOrder int
	err := r.db.QueryRow("SELECT COALESCE(MAX(order_index), -1) FROM modules WHERE syllabus_id = ?", syllabusID).Scan(&maxOrder)
	if err != nil {
		maxOrder = -1
	}

	_, err = r.db.Exec(`
		INSERT INTO modules (id, syllabus_id, title, description, order_index)
		VALUES (?, ?, ?, ?, ?)
	`, id, syllabusID, title, description, maxOrder+1)
	if err != nil {
		return nil, fmt.Errorf("failed to create module: %w", err)
	}

	return &model.Module{
		ID:          id,
		SyllabusID:  syllabusID,
		Title:       title,
		Description: description,
		OrderIndex:  maxOrder + 1,
		CreatedAt:   time.Now(),
		Steps:       []model.Step{},
	}, nil
}

func (r *SyllabusRepo) UpdateModule(id, title, description string, orderIndex int) error {
	_, err := r.db.Exec(`
		UPDATE modules
		SET title = ?, description = ?, order_index = ?
		WHERE id = ?
	`, title, description, orderIndex, id)
	return err
}

func (r *SyllabusRepo) DeleteModule(id string) error {
	_, err := r.db.Exec("DELETE FROM modules WHERE id = ?", id)
	return err
}

func (r *SyllabusRepo) CreateStep(moduleID string, req model.CreateStepRequest) (*model.Step, error) {
	id := uuid.New().String()
	req.Title = cleanTitleText(req.Title)

	var maxOrder int
	err := r.db.QueryRow("SELECT COALESCE(MAX(order_index), -1) FROM steps WHERE module_id = ?", moduleID).Scan(&maxOrder)
	if err != nil {
		maxOrder = -1
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

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO steps (id, module_id, title, type, content, resource_url, resources, acceptance_criteria, estimated_minutes, order_index)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, moduleID, req.Title, req.Type, req.Content, req.ResourceURL, resourcesJSON, req.AcceptanceCriteria, req.EstimatedMinutes, maxOrder+1)
	if err != nil {
		return nil, fmt.Errorf("failed to create step: %w", err)
	}

	progID := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO step_progress (id, step_id, status, notes, submission_url)
		VALUES (?, ?, 'pending', '', '')
	`, progID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to init step progress: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &model.Step{
		ID:                 id,
		ModuleID:           moduleID,
		Title:              req.Title,
		Type:               req.Type,
		Content:            req.Content,
		ResourceURL:        req.ResourceURL,
		Resources:          resources,
		AcceptanceCriteria: req.AcceptanceCriteria,
		EstimatedMinutes:   req.EstimatedMinutes,
		OrderIndex:         maxOrder + 1,
		Status:             model.StepStatusPending,
		CreatedAt:          time.Now(),
	}, nil
}

func (r *SyllabusRepo) UpdateStep(id string, req model.UpdateStepRequest) error {
	req.Title = cleanTitleText(req.Title)
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

	_, err := r.db.Exec(`
		UPDATE steps
		SET title = ?, type = ?, content = ?, resource_url = ?, resources = ?, acceptance_criteria = ?, estimated_minutes = ?, order_index = ?
		WHERE id = ?
	`, req.Title, req.Type, req.Content, req.ResourceURL, resourcesJSON, req.AcceptanceCriteria, req.EstimatedMinutes, req.OrderIndex, id)
	return err
}

func (r *SyllabusRepo) DeleteStep(id string) error {
	_, err := r.db.Exec("DELETE FROM steps WHERE id = ?", id)
	return err
}

func (r *SyllabusRepo) UpdateStepProgress(stepID string, req model.UpdateStepProgressRequest) error {
	var completedAt *string
	if req.Status == model.StepStatusCompleted {
		nowStr := time.Now().UTC().Format("2006-01-02 15:04:05")
		completedAt = &nowStr
	}

	_, err := r.db.Exec(`
		INSERT INTO step_progress (id, step_id, status, notes, submission_url, completed_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(step_id) DO UPDATE SET
			status = excluded.status,
			notes = excluded.notes,
			submission_url = excluded.submission_url,
			completed_at = CASE 
				WHEN excluded.status = 'completed' THEN COALESCE(step_progress.completed_at, CURRENT_TIMESTAMP)
				ELSE NULL 
			END
	`, uuid.New().String(), stepID, req.Status, req.Notes, req.SubmissionURL, completedAt)
	return err
}

func (r *SyllabusRepo) GetStepByID(stepID string) (*model.Step, error) {
	var st model.Step
	var sCreatedAt, resourcesJSON string
	var completedAt sql.NullString
	err := r.db.QueryRow(`
		SELECT 
			st.id, st.module_id, st.title, st.type, st.content, st.resource_url,
			COALESCE(st.resources, '[]'),
			st.acceptance_criteria, st.estimated_minutes, st.order_index, st.created_at,
			COALESCE(sp.status, 'pending'),
			COALESCE(sp.notes, ''),
			COALESCE(sp.submission_url, ''),
			sp.completed_at
		FROM steps st
		LEFT JOIN step_progress sp ON sp.step_id = st.id
		WHERE st.id = ?
	`, stepID).Scan(
		&st.ID, &st.ModuleID, &st.Title, &st.Type, &st.Content, &st.ResourceURL,
		&resourcesJSON,
		&st.AcceptanceCriteria, &st.EstimatedMinutes, &st.OrderIndex, &sCreatedAt,
		&st.Status, &st.Notes, &st.SubmissionURL, &completedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	st.Resources = parseResources(resourcesJSON, st.ResourceURL)
	st.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", sCreatedAt)
	if completedAt.Valid && completedAt.String != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
		st.CompletedAt = &t
	}
	return &st, nil
}

func parseResources(jsonStr, fallbackURL string) []string {
	var res []string
	if strings.TrimSpace(jsonStr) != "" && strings.TrimSpace(jsonStr) != "[]" {
		_ = json.Unmarshal([]byte(jsonStr), &res)
	}
	if len(res) == 0 && strings.TrimSpace(fallbackURL) != "" {
		res = []string{strings.TrimSpace(fallbackURL)}
	}
	if res == nil {
		res = []string{}
	}
	return res
}

func serializeResources(res []string, fallbackURL string) string {
	var cleaned []string
	for _, item := range res {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	if len(cleaned) == 0 && strings.TrimSpace(fallbackURL) != "" {
		cleaned = []string{strings.TrimSpace(fallbackURL)}
	}
	if len(cleaned) == 0 {
		return "[]"
	}
	data, err := json.Marshal(cleaned)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func cleanTitleText(t string) string {
	t = strings.TrimSpace(t)
	t = strings.TrimLeft(t, "-—–*• ")
	t = strings.TrimRight(t, "-—–*• ")
	for strings.Contains(t, "--") {
		t = strings.ReplaceAll(t, "--", "-")
	}
	return strings.TrimSpace(t)
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var out []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out = append(out, r)
		} else if r == ' ' || r == '-' || r == '_' {
			out = append(out, '-')
		}
	}
	result := string(out)
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return strings.Trim(result, "-")
}
