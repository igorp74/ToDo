package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite" // SQLite driver without CGO
)

const (
	dbFileName = "todo.db"
)

// TodoManager handles all todo operations.
type TodoManager struct {
	db *sql.DB
}

// NewTodoManager creates a new TodoManager instance and initializes the database.
func NewTodoManager(dbPath string) *TodoManager {
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting user home directory: %v", err)
		}
		dbPath = fmt.Sprintf("%s%c%s", homeDir, os.PathSeparator, dbFileName)
	}

	// Added _busy_timeout to the connection string
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_busy_timeout=5000", dbPath))
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	tm := &TodoManager{db: db}
	tm.initDB()
	return tm
}

// Close closes the database connection.
func (tm *TodoManager) Close() {
	tm.db.Close()
}

// initDB initializes the database schema.
func (tm *TodoManager) initDB() {
	schema := `
    CREATE TABLE IF NOT EXISTS projects (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE
    );

    CREATE TABLE IF NOT EXISTS contexts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE
    );

    CREATE TABLE IF NOT EXISTS tags (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE
    );

    CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        description TEXT,
        project_id INTEGER,
        start_date DATETIME,
        due_date DATETIME,
        end_date DATETIME,
        status TEXT NOT NULL DEFAULT 'pending', -- pending, completed, cancelled, waiting
        recurrence TEXT, -- daily, weekly, monthly, yearly
        recurrence_interval INTEGER DEFAULT 1,
        start_waiting_date DATETIME,
        end_waiting_date DATETIME,
        original_task_id INTEGER, -- Added original_task_id column
        FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
    );

    CREATE TABLE IF NOT EXISTS task_contexts (
        task_id INTEGER,
        context_id INTEGER,
        PRIMARY KEY (task_id, context_id),
        FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
        FOREIGN KEY (context_id) REFERENCES contexts(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS task_tags (
        task_id INTEGER,
        tag_id INTEGER,
        PRIMARY KEY (task_id, tag_id),
        FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
        FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS holidays (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        date TEXT NOT NULL UNIQUE, --YYYY-MM-DD
        name TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS working_hours (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        day_of_week INTEGER NOT NULL UNIQUE, -- 0=Sunday, 1=Monday, ..., 6=Saturday
        start_hour INTEGER NOT NULL,
        start_minute INTEGER NOT NULL DEFAULT 0,
        end_hour INTEGER NOT NULL,
        end_minute INTEGER NOT NULL DEFAULT 0,
        break_minutes INTEGER NOT NULL DEFAULT 0 -- Added break_minutes
    );

    CREATE TABLE IF NOT EXISTS task_notes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id INTEGER NOT NULL,
        timestamp DATETIME NOT NULL,
        description TEXT NOT NULL,
        FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
    );
    `
	_, err := tm.db.Exec(schema)
	if err != nil {
		log.Fatalf("Error initializing database schema: %v", err)
	}

	// Add original_task_id column if it doesn't exist
	// This handles schema migration for existing databases
	_, err = tm.db.Exec(`
        PRAGMA foreign_keys = OFF;
        ALTER TABLE tasks ADD COLUMN original_task_id INTEGER;
        PRAGMA foreign_keys = ON;
    `)
	if err != nil {
		// Ignore "duplicate column name" error if it already exists
		if !strings.Contains(err.Error(), "duplicate column name: original_task_id") {
			log.Fatalf("Error adding original_task_id column to tasks table: %v", err)
		}
	}

	// Add start_minute and end_minute columns to working_hours if they don't exist
	_, err = tm.db.Exec(`
        ALTER TABLE working_hours ADD COLUMN start_minute INTEGER NOT NULL DEFAULT 0;
    `)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate column name: start_minute") {
			log.Fatalf("Error adding start_minute column to working_hours table: %v", err)
		}
	}

	_, err = tm.db.Exec(`
        ALTER TABLE working_hours ADD COLUMN end_minute INTEGER NOT NULL DEFAULT 0;
    `)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate column name: end_minute") {
			log.Fatalf("Error adding end_minute column to working_hours table: %v", err)
		}
	}

	// Add break_minutes column to working_hours if it doesn't exist
	_, err = tm.db.Exec(`
        ALTER TABLE working_hours ADD COLUMN break_minutes INTEGER NOT NULL DEFAULT 0;
    `)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate column name: break_minutes") {
			log.Fatalf("Error adding break_minutes column to working_hours table: %v", err)
		}
	}
}

// getID inserts a name into a lookup table (contexts, tags, projects) and returns its ID.
// Now accepts a transaction *sql.Tx
func (tm *TodoManager) getID(tx *sql.Tx, tableName, name string) (int64, error) {
	var id int64
	query := fmt.Sprintf("SELECT id FROM %s WHERE name = ?", tableName)
	err := tx.QueryRow(query, name).Scan(&id) // Use tx for query

	if err == sql.ErrNoRows {
		insertQuery := fmt.Sprintf("INSERT INTO %s (name) VALUES (?)", tableName)
		res, err := tx.Exec(insertQuery, name) // Use tx for exec
		if err != nil {
			return 0, fmt.Errorf("failed to insert %s %s: %w", tableName, name, err)
		}
		id, err = res.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to get last insert id for %s %s: %w", tableName, name, err)
		}
		return id, nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to query %s for %s: %w", tableName, name, err)
	}
	return id, nil
}

// GetNameByID gets the name for a given ID in a table.
func (tm *TodoManager) GetNameByID(tableName string, id int64) (string, error) {
	var name string
	err := tm.db.QueryRow(fmt.Sprintf("SELECT name FROM %s WHERE id = ?", tableName), id).Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil // Not found
	} else if err != nil {
		return "", fmt.Errorf("failed to query %s by ID %d: %w", tableName, id, err)
	}
	return name, nil
}

// GetTaskNames fetches associated names (contexts or tags) for a given task.
func (tm *TodoManager) GetTaskNames(taskID int64, joinTable, nameTable string) []string {
	names := []string{}
	query := fmt.Sprintf(`
        SELECT t.name FROM %s jt
        JOIN %s t ON jt.%s_id = t.id
        WHERE jt.task_id = ?
    `, joinTable, nameTable, strings.TrimSuffix(nameTable, "s")) // context_id or tag_id
	rows, err := tm.db.Query(query, taskID)
	if err != nil {
		log.Printf("Error getting %s for task %d: %v", nameTable, taskID, err)
		return names
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Printf("Error scanning %s name for task %d: %v", nameTable, taskID, err)
			continue
		}
		names = append(names, name)
	}
	return names
}

// associateTaskWithNames handles linking tasks to contexts or tags.
// Now accepts a transaction *sql.Tx and does not manage its own transaction.
func (tm *TodoManager) associateTaskWithNames(tx *sql.Tx, taskID int64, nameIDs []int64, joinTable, foreignKey string) error {
	// Clear existing associations for the task if a new set is provided or explicitly cleared
	clearQuery := fmt.Sprintf("DELETE FROM %s WHERE task_id = ?", joinTable)
	_, err := tx.Exec(clearQuery, taskID) // Use tx for exec
	if err != nil {
		return fmt.Errorf("failed to clear old associations for task %d in %s: %w", taskID, joinTable, err)
	}

	for _, nameID := range nameIDs {
		insertQuery := fmt.Sprintf("INSERT INTO %s (task_id, %s) VALUES (?, ?)", joinTable, foreignKey)
		_, err := tx.Exec(insertQuery, taskID, nameID) // Use tx for exec
		if err != nil {
			return fmt.Errorf("failed to associate task %d with ID %d in %s: %w", taskID, nameID, joinTable, err)
		}
	}

	return nil // No commit here, caller's transaction handles it
}

// AddTask adds a new task to the database.
// It now accepts an optional *sql.Tx to allow participation in an existing transaction.
func (tm *TodoManager) AddTask(tx *sql.Tx, title, description, project string, startDateStr string, isStartDateSet bool, dueDateStr string, isDueDateSet bool,
	recurrence string, recurrenceInterval int, contexts, tags []string, startWaitingStr string, isStartWaitingSet bool, endWaitingStr string, isEndWaitingSet bool, status string, originalTaskID sql.NullInt64) { // Added originalTaskID

	// If no transaction is provided, start a new one.
	shouldCommit := false
	if tx == nil {
		var err error
		tx, err = tm.db.Begin()
		if err != nil {
			log.Fatalf("Error starting transaction: %v", err)
		}
		shouldCommit = true
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				panic(r) // Re-throw panic after rollback
			} else if err != nil { // Check for error from the function itself
				tx.Rollback()
			}
		}()
	}

	var projectID sql.NullInt64
	if project != "" {
		id, err := tm.getID(tx, "projects", project) // Pass tx
		if err != nil {
			log.Fatalf("Error getting project ID: %v", err)
		}
		projectID = sql.NullInt64{Int64: id, Valid: true}
	}

	var startDate, dueDate, startWaitingDate, endWaitingDate NullableTime

	// Handle start date
	if isStartDateSet {
		if startDateStr == "" {
			startDate = NullableTime{Time: time.Now(), Valid: true} // Default to current time if flag is present but value is empty
		} else {
			parsed, err := ParseDateTime(startDateStr)
			if err != nil {
				log.Fatalf("Invalid start date format: %v", err)
			}
			startDate = parsed
		}
	} else {
		// If not explicitly set, set to current time (original default behavior)
		startDate = NullableTime{Time: time.Now(), Valid: true}
	}

	// Handle due date
	if isDueDateSet {
		if dueDateStr == "" {
			dueDate = NullableTime{Time: time.Now(), Valid: true} // Default to current time if flag is present but value is empty
		} else {
			parsed, err := ParseDateTime(dueDateStr)
			if err != nil {
				log.Fatalf("Invalid due date format: %v", err)
			}
			dueDate = parsed
		}
	}

	// Handle start waiting date
	if isStartWaitingSet {
		if startWaitingStr == "" {
			startWaitingDate = NullableTime{Time: time.Now(), Valid: true} // Default to current time if flag is present but value is empty
		} else {
			parsed, err := ParseDateTime(startWaitingStr)
			if err != nil {
				log.Fatalf("Invalid start waiting date format: %v", err)
			}
			startWaitingDate = parsed
		}
	}

	// Handle end waiting date
	if isEndWaitingSet {
		if endWaitingStr == "" {
			endWaitingDate = NullableTime{Time: time.Now(), Valid: true} // Default to current time if flag is present but value is empty
		} else {
			parsed, err := ParseDateTime(endWaitingStr)
			if err != nil {
				log.Fatalf("Invalid end waiting date format: %v", err)
			}
			endWaitingDate = parsed
		}
	}

	finalStatus := status // Use provided status
	if startWaitingDate.Valid && !endWaitingDate.Valid {
		finalStatus = "waiting"
	} else if endWaitingDate.Valid {
		finalStatus = "pending"
	}

	// Get sql.NullTime values from NullableTime for database insertion
	sqlStartDate, _ := startDate.Value()
	sqlDueDate, _ := dueDate.Value()
	sqlStartWaitingDate, _ := startWaitingDate.Value()
	sqlEndWaitingDate, _ := endWaitingDate.Value()

	insertQuery := `
        INSERT INTO tasks (title, description, project_id, start_date, due_date, recurrence, recurrence_interval, status, start_waiting_date, end_waiting_date, original_task_id)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	res, err := tx.Exec(insertQuery,
		title,
		sql.NullString{String: description, Valid: description != ""},
		projectID,
		sqlStartDate,
		sqlDueDate,
		sql.NullString{String: recurrence, Valid: recurrence != ""},
		sql.NullInt64{Int64: int64(recurrenceInterval), Valid: recurrenceInterval != 0},
		finalStatus,
		sqlStartWaitingDate,
		sqlEndWaitingDate,
		originalTaskID, // Pass originalTaskID
	)
	if err != nil {
		log.Fatalf("Error adding task: %v", err)
	}

	taskID, err := res.LastInsertId()
	if err != nil {
		log.Fatalf("Error getting last insert ID: %v", err)
	}

	contextIDs := []int64{}
	for _, c := range contexts {
		id, err := tm.getID(tx, "contexts", c) // Pass tx
		if err != nil {
			log.Fatalf("Error getting context ID: %v", err)
		}
		contextIDs = append(contextIDs, id)
	}
	if err := tm.associateTaskWithNames(tx, taskID, contextIDs, "task_contexts", "context_id"); err != nil { // Pass tx
		log.Fatalf("Error associating contexts: %v", err)
	}

	tagIDs := []int64{}
	for _, t := range tags {
		id, err := tm.getID(tx, "tags", t) // Pass tx
		if err != nil {
			log.Fatalf("Error getting tag ID: %v", err)
		}
		tagIDs = append(tagIDs, id)
	}
	if err := tm.associateTaskWithNames(tx, taskID, tagIDs, "task_tags", "tag_id"); err != nil { // Pass tx
		log.Fatalf("Error associating tags: %v", err)
	}

	if shouldCommit {
		if err := tx.Commit(); err != nil {
			log.Fatalf("Error committing transaction: %v", err)
		}
	}
	fmt.Printf("Task '%s' added successfully with ID: %d\n", title, taskID)
}

// DeleteTask deletes a single task by ID.
func (tm *TodoManager) DeleteTask(id int64, completeInstead bool) {
	if completeInstead {
		_, err := tm.db.Exec("UPDATE tasks SET status = 'completed', end_date = ? WHERE id = ?", time.Now(), id)
		if err != nil {
			log.Fatalf("Error completing task %d: %v", id, err)
		}
		fmt.Printf("Task %d marked as completed.\n", id)
	} else {
		_, err := tm.db.Exec("DELETE FROM tasks WHERE id = ?", id)
		if err != nil {
			log.Fatalf("Error deleting task %d: %v", id, err)
		}
		fmt.Printf("Task %d deleted successfully.\n", id)
	}
}

// DeleteTasks deletes multiple tasks by a slice of IDs.
func (tm *TodoManager) DeleteTasks(ids []int64, completeInstead bool) {
	if len(ids) == 0 {
		fmt.Println("No task IDs provided for deletion.")
		return
	}

	for _, id := range ids {
		tm.DeleteTask(id, completeInstead) // Reuse the single task delete logic
	}
}

// UpdateTasks updates one or more tasks.
func (tm *TodoManager) UpdateTasks(ids []int64, title, description, project, startDateStr string, isStartDateSet bool, dueDateStr string, isDueDateSet bool,
	endDateStr string, isEndDateSet bool, status string, recurrence string, recurrenceInterval int, contexts, tags []string, startWaitingStr string, isStartWaitingSet bool, endWaitingStr string, isEndWaitingSet bool,
	clearProject, clearContexts, clearTags, clearStart, clearDue, clearEnd, clearRecurrence, clearWaiting bool) error {

	if len(ids) == 0 {
		return fmt.Errorf("no task IDs provided for update")
	}

	tx, err := tm.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback() // Ensure rollback if commit fails

	for _, id := range ids {
		// Fetch current task state to apply conditional updates and recurrence logic
		var currentTask Task
		row := tx.QueryRow(`
            SELECT id, title, description, project_id, start_date, due_date, end_date, status,
                   recurrence, recurrence_interval, start_waiting_date, end_waiting_date, original_task_id
            FROM tasks
            WHERE id = ?`, id)
		var currentDesc sql.NullString
		var currentProjectID sql.NullInt64
		var currentStartDate, currentDueDate, currentEndDate, currentStartWaitingDate, currentEndWaitingDate sql.NullTime
		var currentRecurrence sql.NullString
		var currentRecurrenceInterval sql.NullInt64
		var currentOriginalTaskID sql.NullInt64

		err := row.Scan(&currentTask.ID, &currentTask.Title, &currentDesc, &currentProjectID,
			&currentStartDate, &currentDueDate, &currentEndDate, &currentTask.Status,
			&currentRecurrence, &currentRecurrenceInterval, &currentStartWaitingDate, &currentEndWaitingDate, &currentOriginalTaskID)
		if err == sql.ErrNoRows {
			fmt.Printf("Task ID %d not found, skipping update.\n", id)
			continue
		} else if err != nil {
			return fmt.Errorf("error fetching current task state for ID %d: %w", id, err)
		}
		currentTask.Description = currentDesc
		currentTask.ProjectID = currentProjectID
		currentTask.StartDate = NullableTime{Time: currentStartDate.Time, Valid: currentStartDate.Valid}
		currentTask.DueDate = NullableTime{Time: currentDueDate.Time, Valid: currentDueDate.Valid}
		currentTask.EndDate = NullableTime{Time: currentEndDate.Time, Valid: currentEndDate.Valid}
		currentTask.Recurrence = currentRecurrence
		currentTask.RecurrenceInterval = currentRecurrenceInterval
		currentTask.StartWaitingDate = NullableTime{Time: currentStartWaitingDate.Time, Valid: currentStartWaitingDate.Valid}
		currentTask.EndWaitingDate = NullableTime{Time: currentEndWaitingDate.Time, Valid: currentEndWaitingDate.Valid}
		currentTask.OriginalTaskID = currentOriginalTaskID

		updates := []string{}
		args := []any{}

		// Title
		if title != "" {
			updates = append(updates, "title = ?")
			args = append(args, title)
		}
		// Description
		if description != "" {
			updates = append(updates, "description = ?")
			args = append(args, sql.NullString{String: description, Valid: true})
		} else if description == "" && strings.Contains(strings.Join(os.Args, " "), "-d") { // Check if -d was provided explicitly as empty
			updates = append(updates, "description = NULL")
		}

		// Project
		if project != "" {
			projectID, err := tm.getID(tx, "projects", project) // Pass tx
			if err != nil {
				return fmt.Errorf("error getting project ID for task %d: %w", id, err)
			}
			updates = append(updates, "project_id = ?")
			args = append(args, projectID)
		} else if clearProject {
			updates = append(updates, "project_id = NULL")
		}

		// Start Date
		if isStartDateSet { // Only update if the flag was explicitly provided
			var parsedDate NullableTime
			if startDateStr == "" {
				parsedDate = NullableTime{Time: time.Now(), Valid: true} // Default to current time
			} else {
				parsedDate, err = ParseDateTime(startDateStr)
				if err != nil {
					return fmt.Errorf("invalid start date format for task %d: %w", id, err)
				}
			}
			sqlParsedDate, _ := parsedDate.Value()
			updates = append(updates, "start_date = ?")
			args = append(args, sqlParsedDate)
		} else if clearStart {
			updates = append(updates, "start_date = NULL")
		}

		// Due Date
		if isDueDateSet { // Only update if the flag was explicitly provided
			var parsedDate NullableTime
			if dueDateStr == "" {
				parsedDate = NullableTime{Time: time.Now(), Valid: true} // Default to current time
			} else {
				parsedDate, err = ParseDateTime(dueDateStr)
				if err != nil {
					return fmt.Errorf("invalid due date format for task %d: %w", id, err)
				}
			}
			sqlParsedDate, _ := parsedDate.Value()
			updates = append(updates, "due_date = ?")
			args = append(args, sqlParsedDate)
		} else if clearDue {
			updates = append(updates, "due_date = NULL")
		}

		// End Date (Completion Date) and Status interaction
		endUpdateApplied := false
		oldStatus := currentTask.Status // Store old status before potential update

		if clearEnd { // Explicitly clear end date (e.g., -clear-E)
			updates = append(updates, "end_date = NULL")
			endUpdateApplied = true
		} else if isEndDateSet { // Explicitly set end date (e.g., -E "2025-01-01", -E "now", or -E)
			var parsedDate NullableTime
			if endDateStr == "" {
				parsedDate = NullableTime{Time: time.Now(), Valid: true} // Default to current time
			} else {
				parsedDate, err = ParseDateTime(endDateStr)
				if err != nil {
					return fmt.Errorf("invalid end date format for task %d: %w", id, err)
				}
			}
			sqlParsedDate, _ := parsedDate.Value()
			updates = append(updates, "end_date = ?")
			args = append(args, sqlParsedDate)
			endUpdateApplied = true

			// If end_date is set via -E, and status is not explicitly provided, set status to 'completed'
			if status == "" && oldStatus != "completed" { // Only change if not already completed
				status = "completed" // Set the status variable, which will be used below
			}
		}

		// Handle waiting period status updates
		if isStartWaitingSet && !isEndWaitingSet {
			// If -sw is used without -ew, set status to 'waiting' if not explicitly overridden
			if status == "" && currentTask.Status != "waiting" {
				status = "waiting"
			}
		} else if isEndWaitingSet {
			// If -ew is used, set status to 'pending' if it was 'waiting' and not explicitly overridden
			if status == "" && currentTask.Status == "waiting" {
				status = "pending"
			}
		}

		// Handle status update (after potential auto-update from -E and waiting flags)
		if status != "" {
			updates = append(updates, "status = ?")
			args = append(args, status)
			// If status is explicitly set to 'completed' and end_date was NOT already handled by -E or -clear-E
			if status == "completed" && !endUpdateApplied && !currentTask.EndDate.Valid {
				updates = append(updates, "end_date = ?")
				args = append(args, time.Now())
			}
		}

		// Recurrence
		if recurrence != "" {
			updates = append(updates, "recurrence = ?")
			args = append(args, recurrence)
		} else if clearRecurrence {
			updates = append(updates, "recurrence = NULL")
		}
		if recurrenceInterval != 0 { // Only update if a positive interval is given
			updates = append(updates, "recurrence_interval = ?")
			args = append(args, recurrenceInterval)
		}

		// Start Waiting Date & End Waiting Date
		if clearWaiting {
			updates = append(updates, "start_waiting_date = NULL, end_waiting_date = NULL")
		} else {
			if isStartWaitingSet {
				var parsed NullableTime
				if startWaitingStr == "" {
					parsed = NullableTime{Time: time.Now(), Valid: true} // Default to current time
				} else {
					parsed, err = ParseDateTime(startWaitingStr)
					if err != nil {
						return fmt.Errorf("invalid start waiting date format for task %d: %w", id, err)
					}
				}
				sqlParsed, _ := parsed.Value()
				updates = append(updates, "start_waiting_date = ?")
				args = append(args, sqlParsed)
			}
			if isEndWaitingSet {
				var parsed NullableTime
				if endWaitingStr == "" {
					parsed = NullableTime{Time: time.Now(), Valid: true} // Default to current time
				} else {
					parsed, err = ParseDateTime(endWaitingStr)
					if err != nil {
						return fmt.Errorf("invalid end waiting date format for task %d: %w", id, err)
					}
				}
				sqlParsed, _ := parsed.Value()
				updates = append(updates, "end_waiting_date = ?")
				args = append(args, sqlParsed)
			}
		}

		if len(updates) == 0 && len(contexts) == 0 && len(tags) == 0 && !clearContexts && !clearTags {
			fmt.Printf("No update parameters provided for task ID %d.\n", id)
			continue
		}

		// Build and execute the UPDATE query for task details
		if len(updates) > 0 {
			updateQuery := fmt.Sprintf("UPDATE tasks SET %s WHERE id = ?", strings.Join(updates, ", "))
			args = append(args, id)
			_, err := tx.Exec(updateQuery, args...)
			if err != nil {
				return fmt.Errorf("error updating task %d: %w", id, err)
			}
		}

		// Update contexts
		if len(contexts) > 0 || clearContexts {
			contextIDs := []int64{}
			if !clearContexts { // Only get IDs if not clearing
				for _, c := range contexts {
					cID, err := tm.getID(tx, "contexts", c) // Pass tx
					if err != nil {
						return fmt.Errorf("error getting context ID for task %d: %w", id, err)
					}
					contextIDs = append(contextIDs, cID)
				}
			}
			if err := tm.associateTaskWithNames(tx, id, contextIDs, "task_contexts", "context_id"); err != nil { // Pass tx
				return fmt.Errorf("error associating contexts for task %d: %w", id, err)
			}
		}

		// Update tags
		if len(tags) > 0 || clearTags {
			tagIDs := []int64{}
			if !clearTags { // Only get IDs if not clearing
				for _, t := range tags {
					tID, err := tm.getID(tx, "tags", t) // Pass tx
					if err != nil {
						return fmt.Errorf("error getting tag ID for task %d: %w", id, err)
					}
					tagIDs = append(tagIDs, tID)
				}
			}
			if err := tm.associateTaskWithNames(tx, id, tagIDs, "task_tags", "tag_id"); err != nil { // Pass tx
				return fmt.Errorf("error associating tags for task %d: %w", id, err)
			}
		}
		fmt.Printf("Task %d updated successfully.\n", id)

		// --- Recurrence Logic: Create next task if completed and recurring ---
		// Check if the status was just changed to "completed" and it's a recurring task
		if status == "completed" && oldStatus != "completed" && currentTask.Recurrence.Valid {
			nextStartDate := currentTask.StartDate.Time
			nextDueDate := currentTask.DueDate.Time

			interval := currentTask.RecurrenceInterval.Int64
			if interval == 0 { // Default to 1 if not set
				interval = 1
			}

			switch currentTask.Recurrence.String {
			case "daily":
				nextStartDate = nextStartDate.AddDate(0, 0, int(interval))
				if currentTask.DueDate.Valid {
					nextDueDate = nextDueDate.AddDate(0, 0, int(interval))
				}
			case "weekly":
				nextStartDate = nextStartDate.AddDate(0, 0, int(interval)*7)
				if currentTask.DueDate.Valid {
					nextDueDate = nextDueDate.AddDate(0, 0, int(interval)*7)
				}
			case "monthly":
				nextStartDate = nextStartDate.AddDate(0, int(interval), 0)
				if currentTask.DueDate.Valid {
					nextDueDate = nextDueDate.AddDate(0, int(interval), 0)
				}
			case "yearly":
				nextStartDate = nextStartDate.AddDate(int(interval), 0, 0)
				if currentTask.DueDate.Valid {
					nextDueDate = nextDueDate.AddDate(int(interval), 0, 0)
				}
			default:
				log.Printf("Warning: Unknown recurrence pattern '%s' for task %d. Not creating next task.", currentTask.Recurrence.String, id)
				continue // Skip to the next task ID in the loop
			}

			// Determine the original_task_id for the new recurring task
			newOriginalTaskID := currentTask.OriginalTaskID
			if !newOriginalTaskID.Valid {
				newOriginalTaskID = sql.NullInt64{Int64: id, Valid: true} // If this is the first instance, set itself as original
			}

			// Get current contexts and tags to pass to the new task
			currentContexts := tm.GetTaskNames(id, "task_contexts", "contexts")
			currentTags := tm.GetTaskNames(id, "task_tags", "tags")

			// Create the next task using the AddTask method, passing the existing transaction
			tm.AddTask(
				tx, // Pass the existing transaction
				currentTask.Title,
				currentTask.Description.String,
				func() string { // Get project name from ID
					if currentTask.ProjectID.Valid {
						name, _ := tm.GetNameByID("projects", currentTask.ProjectID.Int64)
						return name
					}
					return ""
				}(),
				nextStartDate.Format("2006-01-02 15:04:05"),
				true, // isStartDateSet (force setting the new start date)
				nextDueDate.Format("2006-01-02 15:04:05"),
				currentTask.DueDate.Valid, // isDueDateSet (only set if original had a due date)
				currentTask.Recurrence.String,
				int(currentTask.RecurrenceInterval.Int64),
				currentContexts,
				currentTags,
				currentTask.StartWaitingDate.Time.Format("2006-01-02 15:04:05"),
				currentTask.StartWaitingDate.Valid, // isStartWaitingSet
				currentTask.EndWaitingDate.Time.Format("2006-01-02 15:04:05"),
				currentTask.EndWaitingDate.Valid, // isEndWaitingSet
				"pending", // New task is always pending
				newOriginalTaskID, // Pass the calculated originalTaskID
			)
		}
	}

	return tx.Commit()
}

// AddHoliday adds a new holiday.
func (tm *TodoManager) AddHoliday(date, name string) {
	_, err := tm.db.Exec("INSERT INTO holidays (date, name) VALUES (?, ?)", date, name)
	if err != nil {
		log.Fatalf("Error adding holiday: %v", err)
	}
	fmt.Printf("Holiday '%s' on %s added successfully.\n", name, date)
}

// SetWorkingHours sets working hours for a specific day of the week, including minutes and break.
func (tm *TodoManager) SetWorkingHours(dayOfWeek, startHour, startMinute, endHour, endMinute, breakMinutes int) {
	if dayOfWeek < 0 || dayOfWeek > 6 {
		log.Fatalf("Invalid day of week. Must be 0-6 (Sunday-Saturday).")
	}
	if startHour < 0 || startHour > 23 || endHour < 0 || endHour > 24 {
		log.Fatalf("Invalid hour. Must be 0-23 for start, 0-24 for end.")
	}
	if startMinute < 0 || startMinute > 59 || endMinute < 0 || endMinute > 59 {
		log.Fatalf("Invalid minute. Must be 0-59.")
	}
	if breakMinutes < 0 {
		log.Fatalf("Break minutes cannot be negative.")
	}
	// Check if start time is before end time
	if startHour*60+startMinute >= endHour*60+endMinute {
		log.Fatalf("Invalid working hours. Start time must be before end time.")
	}

	// UPSERT: try to update, if no row exists, insert
	res, err := tm.db.Exec("UPDATE working_hours SET start_hour = ?, start_minute = ?, end_hour = ?, end_minute = ?, break_minutes = ? WHERE day_of_week = ?", startHour, startMinute, endHour, endMinute, breakMinutes, dayOfWeek)
	if err != nil {
		log.Fatalf("Error updating working hours: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error checking rows affected for working hours update: %v", err)
	}

	if rowsAffected == 0 {
		_, err := tm.db.Exec("INSERT INTO working_hours (day_of_week, start_hour, start_minute, end_hour, end_minute, break_minutes) VALUES (?, ?, ?, ?, ?, ?)", dayOfWeek, startHour, startMinute, endHour, endMinute, breakMinutes)
		if err != nil {
			log.Fatalf("Error inserting working hours: %v", err)
		}
		fmt.Printf("Working hours set for day %d (%s) from %02d:%02d to %02d:%02d with a %d minute break.\n", dayOfWeek, time.Weekday(dayOfWeek).String(), startHour, startMinute, endHour, endMinute, breakMinutes)
	} else {
		fmt.Printf("Working hours updated for day %d (%s) from %02d:%02d to %02d:%02d with a %d minute break.\n", dayOfWeek, time.Weekday(dayOfWeek).String(), startHour, startMinute, endHour, endMinute, breakMinutes)
	}
}

// GetWorkingHours fetches all defined working hours from the database.
func (tm *TodoManager) GetWorkingHours() (map[time.Weekday]WorkingHours, error) {
	hours := make(map[time.Weekday]WorkingHours)
	rows, err := tm.db.Query("SELECT day_of_week, start_hour, start_minute, end_hour, end_minute, break_minutes FROM working_hours")
	if err != nil {
		return nil, fmt.Errorf("failed to query working hours: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var wh WorkingHours
		if err := rows.Scan(&wh.DayOfWeek, &wh.StartHour, &wh.StartMinute, &wh.EndHour, &wh.EndMinute, &wh.BreakMinutes); err != nil {
			return nil, fmt.Errorf("failed to scan working hours: %v", err)
		}
		hours[time.Weekday(wh.DayOfWeek)] = wh
	}
	return hours, nil
}

// GetHolidays fetches all defined holidays from the database.
func (tm *TodoManager) GetHolidays() (map[string]Holiday, error) { // Use map[YYYY-MM-DD]Holiday for easy lookup
	holidays := make(map[string]Holiday)
	rows, err := tm.db.Query("SELECT date, name FROM holidays")
	if err != nil {
		return nil, fmt.Errorf("failed to query holidays: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dateStr string
		var name string
		if err := rows.Scan(&dateStr, &name); err != nil {
			return nil, fmt.Errorf("failed to scan holiday: %v", err)
		}
		parsedDate, err := time.Parse("2006-01-02", dateStr) // Assuming "YYYY-MM-DD" format
		if err != nil {
			log.Printf("Warning: Could not parse holiday date '%s': %v", dateStr, err)
			continue
		}
		holidays[dateStr] = Holiday{Date: NullableTime{Time: parsedDate, Valid: true}, Name: name}
	}
	return holidays, nil
}

// CalculateWorkingDuration calculates the actual working time between start and end dates,
// considering defined working hours and holidays.
// It returns the duration in minutes.
func (tm *TodoManager) CalculateWorkingDuration(start, end NullableTime, workingHours map[time.Weekday]WorkingHours, holidays map[string]Holiday) time.Duration {
	// Delegate to the utility function in dateutils, passing the *sql.DB for holiday/working hour lookups if needed there.
	// However, since workingHours and holidays maps are already fetched, pass them directly.
	return CalculateWorkingHoursDuration(tm.db, start, end, workingHours, holidays)
}

// AddNoteToTask adds a new note to a specific task.
func (tm *TodoManager) AddNoteToTask(taskID int64, description string) {
	insertQuery := `
        INSERT INTO task_notes (task_id, timestamp, description)
        VALUES (?, ?, ?)
    `
	_, err := tm.db.Exec(insertQuery, taskID, time.Now(), description)
	if err != nil {
		log.Fatalf("Error adding note to task %d: %v", taskID, err)
	}
	fmt.Printf("Note added to task %d successfully.\n", taskID)
}

// GetNotesForTask fetches notes for a given task, ordered by timestamp.
// This now orders notes by timestamp in ascending order to facilitate 1-based indexing
// where 1 is the oldest note, and N is the newest.
func (tm *TodoManager) GetNotesForTask(taskID int64) []Note {
	notes := []Note{}
	query := `
        SELECT id, timestamp, description FROM task_notes
        WHERE task_id = ?
        ORDER BY timestamp ASC
    `
	rows, err := tm.db.Query(query, taskID)
	if err != nil {
		log.Printf("Error getting notes for task %d: %v", taskID, err)
		return notes
	}
	defer rows.Close()

	for rows.Next() {
		var note Note
		var timestamp sql.NullTime
		var desc sql.NullString
		if err := rows.Scan(&note.ID, &timestamp, &desc); err != nil {
			log.Printf("Error scanning note for task %d: %v", taskID, err)
			continue
		}
		note.Timestamp = NullableTime{Time: timestamp.Time, Valid: timestamp.Valid}
		note.Description = desc
		notes = append(notes, note)
	}
	return notes
}

// GetNoteIDByTaskAndIndex fetches the actual database ID of a note
// given its task ID and its 1-based display index within that task.
// Notes are considered ordered by timestamp ASC for this indexing.
func (tm *TodoManager) GetNoteIDByTaskAndIndex(taskID int64, noteIndex int) (int64, error) {
	if noteIndex <= 0 {
		return 0, fmt.Errorf("note index must be a positive integer (1-based)")
	}

	// Get notes ordered by timestamp ASC, so index 0 is oldest, index N-1 is newest.
	notes := tm.GetNotesForTask(taskID)

	if noteIndex > len(notes) {
		return 0, fmt.Errorf("note index %d out of bounds for task %d (total notes: %d)", noteIndex, taskID, len(notes))
	}

	// noteIndex is 1-based, slice index is 0-based
	return notes[noteIndex-1].ID, nil
}


// UpdateNote updates the description and/or timestamp of an existing note.
func (tm *TodoManager) UpdateNote(noteID int64, description string, timestampStr string, isTimestampSet bool) {
	updates := []string{}
	args := []any{}

	if description != "" {
		updates = append(updates, "description = ?")
		args = append(args, description)
	}

	if isTimestampSet {
		var parsedTime NullableTime
		if timestampStr == "" {
			parsedTime = NullableTime{Time: time.Now(), Valid: true} // Default to current time if flag is present but value is empty
		} else {
			var err error
			parsedTime, err = ParseDateTime(timestampStr)
			if err != nil {
				log.Fatalf("Invalid timestamp format for note %d: %v", noteID, err)
			}
		}
		sqlParsedTime, _ := parsedTime.Value()
		updates = append(updates, "timestamp = ?")
		args = append(args, sqlParsedTime)
	}

	if len(updates) == 0 {
		fmt.Printf("No update parameters provided for note ID %d.\n", noteID)
		return
	}

	updateQuery := fmt.Sprintf("UPDATE task_notes SET %s WHERE id = ?", strings.Join(updates, ", "))
	args = append(args, noteID)

	res, err := tm.db.Exec(updateQuery, args...)
	if err != nil {
		log.Fatalf("Error updating note %d: %v", noteID, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error checking rows affected for note update: %v", err)
	}
	if rowsAffected == 0 {
		fmt.Printf("Note %d not found or values were not changed.\n", noteID)
	} else {
		fmt.Printf("Note %d updated successfully.\n", noteID)
	}
}

// DeleteNotes deletes one or more notes by their IDs.
func (tm *TodoManager) DeleteNotes(noteIDs []int64) {
	if len(noteIDs) == 0 {
		fmt.Println("No note IDs provided for deletion.")
		return
	}

	// Start a transaction for multiple deletions
	tx, err := tm.db.Begin()
	if err != nil {
		log.Fatalf("Error starting transaction for note deletion: %v", err)
	}
	defer tx.Rollback() // Ensure rollback if any deletion fails

	deleteQuery := `DELETE FROM task_notes WHERE id = ?`
	stmt, err := tx.Prepare(deleteQuery)
	if err != nil {
		log.Fatalf("Error preparing delete statement for notes: %v", err)
	}
	defer stmt.Close()

	for _, id := range noteIDs {
		res, err := stmt.Exec(id)
		if err != nil {
			log.Printf("Error deleting note %d: %v", id, err)
			continue // Continue to next note even if one fails
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			log.Printf("Error checking rows affected for note %d deletion: %v", id, err)
		}
		if rowsAffected == 0 {
			fmt.Printf("Note %d not found.\n", id)
		} else {
			fmt.Printf("Note %d deleted successfully.\n", id)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Error committing note deletion transaction: %v", err)
	}
}

