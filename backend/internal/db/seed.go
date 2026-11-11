package db

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type seedStep struct {
	title       string
	stepType    string
	resourceURL string
	content     string
}

type seedModule struct {
	title string
	steps []seedStep
}

func seedDefaultData(database *sql.DB) error {
	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM syllabi").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	syllabusID := uuid.New().String()
	_, err = database.Exec(`
		INSERT INTO syllabi (id, title, slug, description, is_template)
		VALUES (?, ?, ?, ?, ?)
	`, syllabusID, "Backend & Systems Queue", "backend-systems-queue",
		"Books, papers, and courses I want to work through in order.", 0)
	if err != nil {
		return fmt.Errorf("failed to insert default syllabus: %w", err)
	}

	modules := []seedModule{
		{
			title: "Currently Reading & Doing",
			steps: []seedStep{
				{
					title:       "Designing Data-Intensive Applications (DDIA)",
					stepType:    "article",
					resourceURL: "https://dataintensive.net/",
					content:     "Working through Part II on distributed data: replication, partitioning, transactions, and consensus.",
				},
				{
					title:       "Stanford CS144: Introduction to Computer Networking",
					stepType:    "video",
					resourceURL: "https://cs144.github.io/",
					content:     "Following course labs: building a TCP stack from scratch in user space.",
				},
				{
					title:       "Paper: In Search of an Understandable Consensus Algorithm (Raft)",
					stepType:    "article",
					resourceURL: "https://raft.github.io/raft.pdf",
					content:     "Classic Diego Ongaro and John Ousterhout paper on Raft consensus.",
				},
			},
		},
		{
			title: "Next Up in the Queue",
			steps: []seedStep{
				{
					title:       "Book: Crafting Interpreters (Bob Nystrom)",
					stepType:    "article",
					resourceURL: "https://craftinginterpreters.com/",
					content:     "Building a tree-walk interpreter in Java/Go, then bytecode VM in C.",
				},
				{
					title:       "Watch: Bryan Mills - Rethinking Concurrency Patterns",
					stepType:    "video",
					resourceURL: "https://www.youtube.com/watch?v=5zXAHh5tJqQ",
					content:     "Channel ownership, avoiding goroutine leaks, and clean cancellation in Go.",
				},
				{
					title:       "The Go Memory Model Specification",
					stepType:    "article",
					resourceURL: "https://go.dev/ref/mem",
					content:     "Detailed read on happens-before relationships and synchronization guarantees in Go runtime.",
				},
			},
		},
	}

	for mIdx, mod := range modules {
		modID := uuid.New().String()
		_, err := database.Exec(`
			INSERT INTO modules (id, syllabus_id, title, description, order_index)
			VALUES (?, ?, ?, ?, ?)
		`, modID, syllabusID, mod.title, "", mIdx)
		if err != nil {
			return fmt.Errorf("failed to insert module %d: %w", mIdx, err)
		}

		for sIdx, st := range mod.steps {
			stepID := uuid.New().String()
			_, err := database.Exec(`
				INSERT INTO steps (id, module_id, title, type, content, resource_url, acceptance_criteria, estimated_minutes, order_index)
				VALUES (?, ?, ?, ?, ?, ?, '', 0, ?)
			`, stepID, modID, st.title, st.stepType, st.content, st.resourceURL, sIdx)
			if err != nil {
				return fmt.Errorf("failed to insert step %d: %w", sIdx, err)
			}

			status := "pending"
			if mIdx == 0 && sIdx == 0 {
				status = "in_progress"
			}
			progID := uuid.New().String()
			_, err = database.Exec(`
				INSERT INTO step_progress (id, step_id, status, notes, submission_url)
				VALUES (?, ?, ?, '', '')
			`, progID, stepID, status)
			if err != nil {
				return fmt.Errorf("failed to insert step progress: %w", err)
			}
		}
	}

	return nil
}
