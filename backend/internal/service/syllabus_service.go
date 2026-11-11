package service

import (
	"fmt"

	"autodidact-backend/internal/model"
	"autodidact-backend/internal/repository"
)

type SyllabusService struct {
	repo *repository.SyllabusRepo
}

func NewSyllabusService(repo *repository.SyllabusRepo) *SyllabusService {
	return &SyllabusService{repo: repo}
}

func (s *SyllabusService) ListSyllabi() ([]model.Syllabus, error) {
	return s.repo.ListSyllabi()
}

func (s *SyllabusService) GetSyllabusByID(id string) (*model.Syllabus, error) {
	return s.repo.GetSyllabusByID(id)
}

func (s *SyllabusService) CreateSyllabus(req model.CreateSyllabusRequest) (*model.Syllabus, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	return s.repo.CreateSyllabus(req.Title, req.Description)
}

func (s *SyllabusService) UpdateSyllabus(id string, req model.UpdateSyllabusRequest) (*model.Syllabus, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	return s.repo.UpdateSyllabus(id, req.Title, req.Description)
}

func (s *SyllabusService) DeleteSyllabus(id string) error {
	return s.repo.DeleteSyllabus(id)
}

func (s *SyllabusService) DuplicateSyllabus(id string) (*model.Syllabus, error) {
	return s.repo.DuplicateSyllabus(id)
}

func (s *SyllabusService) CreateModule(syllabusID string, req model.CreateModuleRequest) (*model.Module, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("module title cannot be empty")
	}
	return s.repo.CreateModule(syllabusID, req.Title, req.Description)
}

func (s *SyllabusService) UpdateModule(id string, req model.UpdateModuleRequest) error {
	if req.Title == "" {
		return fmt.Errorf("module title cannot be empty")
	}
	return s.repo.UpdateModule(id, req.Title, req.Description, req.OrderIndex)
}

func (s *SyllabusService) DeleteModule(id string) error {
	return s.repo.DeleteModule(id)
}

func (s *SyllabusService) CreateStep(moduleID string, req model.CreateStepRequest) (*model.Step, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("step title cannot be empty")
	}
	if req.Type == "" {
		req.Type = model.StepTypeArticle
	}
	if req.EstimatedMinutes <= 0 {
		req.EstimatedMinutes = 30
	}
	return s.repo.CreateStep(moduleID, req)
}

func (s *SyllabusService) UpdateStep(id string, req model.UpdateStepRequest) error {
	if req.Title == "" {
		return fmt.Errorf("step title cannot be empty")
	}
	return s.repo.UpdateStep(id, req)
}

func (s *SyllabusService) DeleteStep(id string) error {
	return s.repo.DeleteStep(id)
}

func (s *SyllabusService) UpdateStepProgress(stepID string, req model.UpdateStepProgressRequest) error {
	if req.Status == "" {
		req.Status = model.StepStatusPending
	}
	return s.repo.UpdateStepProgress(stepID, req)
}

func (s *SyllabusService) GetStepByID(stepID string) (*model.Step, error) {
	return s.repo.GetStepByID(stepID)
}

type ActiveStepResponse struct {
	Step        *model.Step `json:"step"`
	ModuleTitle string      `json:"module_title"`
	ModuleID    string      `json:"module_id"`
	PrevStepID  *string     `json:"prev_step_id"`
	NextStepID  *string     `json:"next_step_id"`
	StepIndex   int         `json:"step_index"`
	TotalSteps  int         `json:"total_steps"`
}

func (s *SyllabusService) GetActiveStep(syllabusID string) (*ActiveStepResponse, error) {
	syl, err := s.repo.GetSyllabusByID(syllabusID)
	if err != nil {
		return nil, err
	}
	if syl == nil {
		return nil, fmt.Errorf("syllabus not found")
	}

	type flatStep struct {
		step        model.Step
		moduleTitle string
		moduleID    string
	}

	var flatSteps []flatStep
	for _, mod := range syl.Modules {
		for _, st := range mod.Steps {
			flatSteps = append(flatSteps, flatStep{
				step:        st,
				moduleTitle: mod.Title,
				moduleID:    mod.ID,
			})
		}
	}

	if len(flatSteps) == 0 {
		return &ActiveStepResponse{
			Step:       nil,
			TotalSteps: 0,
		}, nil
	}

	// 1. Look for in_progress step
	chosenIndex := -1
	for i, fs := range flatSteps {
		if fs.step.Status == model.StepStatusInProgress {
			chosenIndex = i
			break
		}
	}

	// 2. If none in_progress, look for first pending step
	if chosenIndex == -1 {
		for i, fs := range flatSteps {
			if fs.step.Status == model.StepStatusPending {
				chosenIndex = i
				break
			}
		}
	}

	// 3. Fallback to 0 if all completed/skipped
	if chosenIndex == -1 {
		chosenIndex = 0
	}

	var prevID *string
	if chosenIndex > 0 {
		prev := flatSteps[chosenIndex-1].step.ID
		prevID = &prev
	}

	var nextID *string
	if chosenIndex < len(flatSteps)-1 {
		next := flatSteps[chosenIndex+1].step.ID
		nextID = &next
	}

	return &ActiveStepResponse{
		Step:        &flatSteps[chosenIndex].step,
		ModuleTitle: flatSteps[chosenIndex].moduleTitle,
		ModuleID:    flatSteps[chosenIndex].moduleID,
		PrevStepID:  prevID,
		NextStepID:  nextID,
		StepIndex:   chosenIndex + 1,
		TotalSteps:  len(flatSteps),
	}, nil
}
