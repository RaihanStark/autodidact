package service

import (
	"context"
	"database/sql"
	"fmt"

	"autodidact-backend/internal/model"
	"autodidact-backend/internal/repository"
)

type FocusService struct {
	repo         *repository.FocusRepo
	syllabusRepo *repository.SyllabusRepo
	db           *sql.DB
}

func NewFocusService(repo *repository.FocusRepo, syllabusRepo *repository.SyllabusRepo, db *sql.DB) *FocusService {
	return &FocusService{
		repo:         repo,
		syllabusRepo: syllabusRepo,
		db:           db,
	}
}

func (s *FocusService) ListFocusItems(ctx context.Context, period string) ([]model.FocusItem, error) {
	return s.repo.ListFocusItems(period)
}

func (s *FocusService) CreateFocusItem(ctx context.Context, req *model.CreateFocusItemRequest) (*model.FocusItem, error) {
	if req.StepID != "" {
		existing, err := s.repo.GetFocusItemByStepID(req.StepID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			// Update period if requested
			if req.Period != "" && req.Period != existing.Period {
				period := req.Period
				return s.repo.UpdateFocusItem(existing.ID, &model.UpdateFocusItemRequest{
					Period: &period,
				})
			}
			return existing, nil
		}
	} else if req.CustomTitle == "" {
		return nil, fmt.Errorf("title or step_id is required")
	}

	item, err := s.repo.CreateFocusItem(req)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *FocusService) UpdateFocusItem(ctx context.Context, id string, req *model.UpdateFocusItemRequest) (*model.FocusItem, error) {
	item, err := s.repo.UpdateFocusItem(id, req)
	if err != nil {
		return nil, err
	}

	// Sync back to step_progress if linked to a step
	if item.StepID != "" {
		if req.Status != nil {
			var stepStatus model.StepStatus
			switch *req.Status {
			case model.FocusStatusCompleted:
				stepStatus = model.StepStatusCompleted
			case model.FocusStatusInProgress:
				stepStatus = model.StepStatusInProgress
			default:
				stepStatus = model.StepStatusPending
			}

			// Update step progress
			_ = s.syllabusRepo.UpdateStepProgress(item.StepID, model.UpdateStepProgressRequest{
				Status: stepStatus,
				Notes:  item.Notes,
			})
		} else if req.Notes != nil {
			_ = s.syllabusRepo.UpdateStepProgress(item.StepID, model.UpdateStepProgressRequest{
				Status: model.StepStatus(item.Status),
				Notes:  *req.Notes,
			})
		}
	}

	return item, nil
}

func (s *FocusService) ReorderFocusItems(ctx context.Context, ids []string) error {
	return s.repo.ReorderFocusItems(ids)
}

func (s *FocusService) DeleteFocusItem(ctx context.Context, id string) error {
	return s.repo.DeleteFocusItem(id)
}

func (s *FocusService) ListBacklogItems(ctx context.Context) ([]model.BacklogStepItem, error) {
	return s.repo.ListBacklogItems()
}
