package model

import "time"

type StepType string

const (
	StepTypeArticle StepType = "article"
	StepTypeVideo   StepType = "video"
	StepTypeProject StepType = "project"
	StepTypeQuiz    StepType = "quiz"
	StepTypeNote    StepType = "note"
)

type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusCompleted  StepStatus = "completed"
	StepStatusSkipped    StepStatus = "skipped"
)

type Syllabus struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	IsTemplate     bool      `json:"is_template"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TotalSteps     int       `json:"total_steps"`
	CompletedSteps int       `json:"completed_steps"`
	Modules        []Module  `json:"modules,omitempty"`
}

type Module struct {
	ID          string    `json:"id"`
	SyllabusID  string    `json:"syllabus_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	OrderIndex  int       `json:"order_index"`
	CreatedAt   time.Time `json:"created_at"`
	Steps       []Step    `json:"steps,omitempty"`
}

type Step struct {
	ID                 string     `json:"id"`
	ModuleID           string     `json:"module_id"`
	Title              string     `json:"title"`
	Type               StepType   `json:"type"`
	Content            string     `json:"content"`
	ResourceURL        string     `json:"resource_url"`
	Resources          []string   `json:"resources"`
	AcceptanceCriteria string     `json:"acceptance_criteria"`
	EstimatedMinutes   int        `json:"estimated_minutes"`
	OrderIndex         int        `json:"order_index"`
	Status             StepStatus `json:"status"`
	Notes              string     `json:"notes"`
	SubmissionURL      string     `json:"submission_url"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

type CreateSyllabusRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateSyllabusRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateModuleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateModuleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	OrderIndex  int    `json:"order_index"`
}

type CreateStepRequest struct {
	Title              string   `json:"title"`
	Type               StepType `json:"type"`
	Content            string   `json:"content"`
	ResourceURL        string   `json:"resource_url"`
	Resources          []string `json:"resources,omitempty"`
	AcceptanceCriteria string   `json:"acceptance_criteria"`
	EstimatedMinutes   int      `json:"estimated_minutes"`
}

type UpdateStepRequest struct {
	Title              string   `json:"title"`
	Type               StepType `json:"type"`
	Content            string   `json:"content"`
	ResourceURL        string   `json:"resource_url"`
	Resources          []string `json:"resources,omitempty"`
	AcceptanceCriteria string   `json:"acceptance_criteria"`
	EstimatedMinutes   int      `json:"estimated_minutes"`
	OrderIndex         int      `json:"order_index"`
}

type UpdateStepProgressRequest struct {
	Status        StepStatus `json:"status"`
	Notes         string     `json:"notes"`
	SubmissionURL string     `json:"submission_url"`
}

type VerifyAuthRequest struct {
	Token string `json:"token"`
}

type FocusPeriod string

const (
	FocusPeriodNext  FocusPeriod = "next"
	FocusPeriodToday FocusPeriod = "today"
)

type FocusStatus string

const (
	FocusStatusTodo       FocusStatus = "todo"
	FocusStatusInProgress FocusStatus = "in_progress"
	FocusStatusCompleted  FocusStatus = "completed"
)

type FocusItem struct {
	ID            string      `json:"id"`
	StepID        string      `json:"step_id,omitempty"`
	CustomTitle   string      `json:"custom_title,omitempty"`
	Title         string      `json:"title"`
	ResourceURL   string      `json:"resource_url"`
	Resources     []string    `json:"resources"`
	ModuleTitle   string      `json:"module_title,omitempty"`
	SyllabusID    string      `json:"syllabus_id,omitempty"`
	SyllabusTitle string      `json:"syllabus_title,omitempty"`
	Period        FocusPeriod `json:"period"`
	Status        FocusStatus `json:"status"`
	Notes         string      `json:"notes"`
	OrderIndex    int         `json:"order_index"`
	CompletedAt   *time.Time  `json:"completed_at,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type CreateFocusItemRequest struct {
	StepID      string      `json:"step_id,omitempty"`
	CustomTitle string      `json:"custom_title,omitempty"`
	ResourceURL string      `json:"resource_url,omitempty"`
	Resources   []string    `json:"resources,omitempty"`
	Period      FocusPeriod `json:"period"`
	Notes       string      `json:"notes,omitempty"`
	OrderIndex  int         `json:"order_index,omitempty"`
}

type UpdateFocusItemRequest struct {
	Status     *FocusStatus `json:"status,omitempty"`
	Period     *FocusPeriod `json:"period,omitempty"`
	Notes      *string      `json:"notes,omitempty"`
	OrderIndex *int         `json:"order_index,omitempty"`
	Resources  *[]string    `json:"resources,omitempty"`
}

type ReorderFocusRequest struct {
	IDs []string `json:"ids"`
}

type BacklogStepItem struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	ResourceURL   string     `json:"resource_url"`
	Resources     []string   `json:"resources"`
	Status        StepStatus `json:"status"`
	ModuleTitle   string     `json:"module_title"`
	SyllabusID    string     `json:"syllabus_id"`
	SyllabusTitle string     `json:"syllabus_title"`
	InFocus       bool       `json:"in_focus"`
}
