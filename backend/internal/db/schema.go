package db

import (
	"database/sql"
	"strings"
)

func runSchema(database *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS syllabi (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			is_template BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS modules (
			id TEXT PRIMARY KEY,
			syllabus_id TEXT NOT NULL REFERENCES syllabi(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			order_index INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS steps (
			id TEXT PRIMARY KEY,
			module_id TEXT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			type TEXT NOT NULL CHECK (type IN ('article', 'video', 'project', 'quiz', 'note')),
			content TEXT NOT NULL DEFAULT '',
			resource_url TEXT NOT NULL DEFAULT '',
			resources TEXT NOT NULL DEFAULT '[]',
			acceptance_criteria TEXT NOT NULL DEFAULT '',
			estimated_minutes INTEGER NOT NULL DEFAULT 30,
			order_index INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS step_progress (
			id TEXT PRIMARY KEY,
			step_id TEXT NOT NULL UNIQUE REFERENCES steps(id) ON DELETE CASCADE,
			status TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'completed', 'skipped')) DEFAULT 'pending',
			notes TEXT NOT NULL DEFAULT '',
			submission_url TEXT NOT NULL DEFAULT '',
			completed_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_modules_syllabus_id ON modules(syllabus_id);`,
		`CREATE INDEX IF NOT EXISTS idx_steps_module_id ON steps(module_id);`,
		`CREATE INDEX IF NOT EXISTS idx_step_progress_step_id ON step_progress(step_id);`,
		`CREATE TABLE IF NOT EXISTS focus_items (
			id TEXT PRIMARY KEY,
			step_id TEXT REFERENCES steps(id) ON DELETE CASCADE,
			custom_title TEXT NOT NULL DEFAULT '',
			resource_url TEXT NOT NULL DEFAULT '',
			resources TEXT NOT NULL DEFAULT '[]',
			period TEXT NOT NULL DEFAULT 'next',
			status TEXT NOT NULL CHECK (status IN ('todo', 'in_progress', 'completed')) DEFAULT 'todo',
			notes TEXT NOT NULL DEFAULT '',
			order_index INTEGER NOT NULL DEFAULT 0,
			completed_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range queries {
		if _, err := database.Exec(q); err != nil {
			return err
		}
	}

	// Migrate existing database table if order_index doesn't exist
	_, _ = database.Exec("ALTER TABLE focus_items ADD COLUMN order_index INTEGER NOT NULL DEFAULT 0;")

	// Migrate existing database tables for resources JSON column
	_, _ = database.Exec("ALTER TABLE steps ADD COLUMN resources TEXT NOT NULL DEFAULT '[]';")
	_, _ = database.Exec("ALTER TABLE focus_items ADD COLUMN resources TEXT NOT NULL DEFAULT '[]';")

	// Migrate existing database table if it has the old ('today', 'week') CHECK constraint
	var existingSQL string
	_ = database.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='focus_items'`).Scan(&existingSQL)
	if strings.Contains(existingSQL, "period IN ('today', 'week')") {
		migration := `
			PRAGMA foreign_keys=OFF;
			CREATE TABLE focus_items_v2 (
				id TEXT PRIMARY KEY,
				step_id TEXT REFERENCES steps(id) ON DELETE CASCADE,
				custom_title TEXT NOT NULL DEFAULT '',
				resource_url TEXT NOT NULL DEFAULT '',
				period TEXT NOT NULL DEFAULT 'next',
				status TEXT NOT NULL CHECK (status IN ('todo', 'in_progress', 'completed')) DEFAULT 'todo',
				notes TEXT NOT NULL DEFAULT '',
				order_index INTEGER NOT NULL DEFAULT 0,
				completed_at DATETIME,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
			INSERT INTO focus_items_v2 (id, step_id, custom_title, resource_url, period, status, notes, order_index, completed_at, created_at, updated_at)
			SELECT id, step_id, custom_title, resource_url, 'next', status, notes, order_index, completed_at, created_at, updated_at FROM focus_items;
			DROP TABLE focus_items;
			ALTER TABLE focus_items_v2 RENAME TO focus_items;
			PRAGMA foreign_keys=ON;
		`
		if _, err := database.Exec(migration); err != nil {
			return err
		}
	}

	// Create indexes now that table and columns are guaranteed to exist
	_, _ = database.Exec("CREATE INDEX IF NOT EXISTS idx_focus_items_step_id ON focus_items(step_id);")
	_, _ = database.Exec("CREATE INDEX IF NOT EXISTS idx_focus_items_period ON focus_items(period);")
	_, _ = database.Exec("CREATE INDEX IF NOT EXISTS idx_focus_items_order_index ON focus_items(order_index);")

	return nil
}
