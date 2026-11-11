package service_test

import (
	"context"
	"testing"

	"autodidact-backend/internal/db"
	"autodidact-backend/internal/model"
	"autodidact-backend/internal/repository"
	"autodidact-backend/internal/service"
)

func setupTestDB(t *testing.T) (*service.FocusService, *service.SyllabusService) {
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory db: %v", err)
	}

	sylRepo := repository.NewSyllabusRepo(database)
	focusRepo := repository.NewFocusRepo(database)

	sylSvc := service.NewSyllabusService(sylRepo)
	focusSvc := service.NewFocusService(focusRepo, sylRepo, database)

	return focusSvc, sylSvc
}

func TestFocusService(t *testing.T) {
	ctx := context.Background()
	focusSvc, sylSvc := setupTestDB(t)

	// Verify backlog items exist from seeds
	backlog, err := focusSvc.ListBacklogItems(ctx)
	if err != nil {
		t.Fatalf("failed to list backlog items: %v", err)
	}
	if len(backlog) == 0 {
		t.Fatalf("expected seeded backlog items, got 0")
	}

	firstStep := backlog[0]

	// 1. Add step to What Next focus
	item, err := focusSvc.CreateFocusItem(ctx, &model.CreateFocusItemRequest{
		StepID: firstStep.ID,
		Period: model.FocusPeriodNext,
	})
	if err != nil {
		t.Fatalf("failed to create focus item: %v", err)
	}

	if item.Period != model.FocusPeriodNext {
		t.Errorf("expected period next, got %s", item.Period)
	}
	if item.Status != model.FocusStatusTodo {
		t.Errorf("expected status todo, got %s", item.Status)
	}

	// 2. Update status to in_progress
	inProgress := model.FocusStatusInProgress
	notes := "Read first 20 pages"
	updated, err := focusSvc.UpdateFocusItem(ctx, item.ID, &model.UpdateFocusItemRequest{
		Status: &inProgress,
		Notes:  &notes,
	})
	if err != nil {
		t.Fatalf("failed to update focus item: %v", err)
	}
	if updated.Status != model.FocusStatusInProgress {
		t.Errorf("expected status in_progress, got %s", updated.Status)
	}

	// Verify underlying step is in_progress
	step, err := sylSvc.GetStepByID(firstStep.ID)
	if err != nil {
		t.Fatalf("failed to get step: %v", err)
	}
	if step.Status != model.StepStatusInProgress {
		t.Errorf("expected underlying step status in_progress, got %s", step.Status)
	}

	// 3. Update status to completed
	completed := model.FocusStatusCompleted
	doneNotes := "Finished chapter and made notes"
	doneItem, err := focusSvc.UpdateFocusItem(ctx, item.ID, &model.UpdateFocusItemRequest{
		Status: &completed,
		Notes:  &doneNotes,
	})
	if err != nil {
		t.Fatalf("failed to mark focus item completed: %v", err)
	}
	if doneItem.Status != model.FocusStatusCompleted {
		t.Errorf("expected status completed, got %s", doneItem.Status)
	}
	if doneItem.CompletedAt == nil {
		t.Errorf("expected completed_at to be set")
	}

	// Verify underlying step completed
	step, err = sylSvc.GetStepByID(firstStep.ID)
	if err != nil {
		t.Fatalf("failed to get step: %v", err)
	}
	if step.Status != model.StepStatusCompleted {
		t.Errorf("expected underlying step status completed, got %s", step.Status)
	}

	// 4. Quick ad-hoc focus item
	adhoc, err := focusSvc.CreateFocusItem(ctx, &model.CreateFocusItemRequest{
		CustomTitle: "Implement quick LRU cache",
		Period:      model.FocusPeriodNext,
	})
	if err != nil {
		t.Fatalf("failed to create ad-hoc focus item: %v", err)
	}
	if adhoc.Title != "Implement quick LRU cache" {
		t.Errorf("expected title to match, got %s", adhoc.Title)
	}

	// 5. Test reorder
	item2, err := focusSvc.CreateFocusItem(ctx, &model.CreateFocusItemRequest{
		CustomTitle: "Second item",
		Period:      model.FocusPeriodNext,
	})
	if err != nil {
		t.Fatalf("failed to create second item: %v", err)
	}

	err = focusSvc.ReorderFocusItems(ctx, []string{item2.ID, adhoc.ID})
	if err != nil {
		t.Fatalf("failed to reorder focus items: %v", err)
	}

	list, err := focusSvc.ListFocusItems(ctx, "next")
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}
	if len(list) < 2 {
		t.Fatalf("expected at least 2 items, got %d", len(list))
	}
	if list[0].ID != item2.ID {
		t.Errorf("expected first item to be item2 after reordering, got %s", list[0].ID)
	}

	// 6. Delete adhoc
	if err := focusSvc.DeleteFocusItem(ctx, adhoc.ID); err != nil {
		t.Fatalf("failed to delete focus item: %v", err)
	}
}

func TestMultipleResources(t *testing.T) {
	ctx := context.Background()
	focusSvc, sylSvc := setupTestDB(t)

	syl, err := sylSvc.CreateSyllabus(model.CreateSyllabusRequest{
		Title:       "Algorithms Queue",
		Description: "Learning DSA",
	})
	if err != nil {
		t.Fatalf("failed to create syllabus: %v", err)
	}

	mod, err := sylSvc.CreateModule(syl.ID, model.CreateModuleRequest{
		Title:       "Graphs",
		Description: "Graph theory notes",
	})
	if err != nil {
		t.Fatalf("failed to create module: %v", err)
	}

	links := []string{
		"https://example.com/graphs-intro",
		"https://example.com/dfs-bfs-visualizer",
		"https://example.com/tarjan-scc-paper",
	}

	// 1. Create step with multiple resources
	step, err := sylSvc.CreateStep(mod.ID, model.CreateStepRequest{
		Title:       "Depth First Search Deep Dive",
		Type:        model.StepTypeArticle,
		Resources:   links,
		ResourceURL: links[0],
	})
	if err != nil {
		t.Fatalf("failed to create step with resources: %v", err)
	}

	if len(step.Resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(step.Resources))
	}
	if step.Resources[1] != links[1] || step.Resources[2] != links[2] {
		t.Errorf("resources mismatch: got %v", step.Resources)
	}

	// 2. Fetch step by ID
	fetchedStep, err := sylSvc.GetStepByID(step.ID)
	if err != nil {
		t.Fatalf("failed to fetch step: %v", err)
	}
	if len(fetchedStep.Resources) != 3 {
		t.Fatalf("expected 3 resources on fetched step, got %d", len(fetchedStep.Resources))
	}

	// 3. Add to What Next focus
	fItem, err := focusSvc.CreateFocusItem(ctx, &model.CreateFocusItemRequest{
		StepID: step.ID,
		Period: model.FocusPeriodNext,
	})
	if err != nil {
		t.Fatalf("failed to add step to focus: %v", err)
	}
	if len(fItem.Resources) != 3 {
		t.Fatalf("expected focus item to have 3 resources inherited from step, got %d: %v", len(fItem.Resources), fItem.Resources)
	}
	if fItem.Resources[2] != links[2] {
		t.Errorf("expected 3rd resource %s, got %s", links[2], fItem.Resources[2])
	}

	// 4. Create ad-hoc focus item with multiple links
	adhocLinks := []string{"https://adhoc.org/1", "https://adhoc.org/2"}
	adhocItem, err := focusSvc.CreateFocusItem(ctx, &model.CreateFocusItemRequest{
		CustomTitle: "Ad-hoc task with multiple links",
		Period:      model.FocusPeriodNext,
		Resources:   adhocLinks,
		ResourceURL: adhocLinks[0],
	})
	if err != nil {
		t.Fatalf("failed to create adhoc focus item: %v", err)
	}
	if len(adhocItem.Resources) != 2 {
		t.Fatalf("expected adhoc item to have 2 resources, got %d: %v", len(adhocItem.Resources), adhocItem.Resources)
	}
}
