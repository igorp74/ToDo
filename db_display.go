package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
//
)

// Display formats
const (
	DisplayFull = iota
	DisplayCondensed
	DisplayMinimal

	style_reset     = "\033[0m"
	style_bold      = "\033[1m"
	style_italic    = "\033[3m"
	style_underline = "\033[4m"

	fg_black   = "\033[30m"
	fg_red     = "\033[31m"
	fg_green   = "\033[32m"
	fg_yellow  = "\033[33m"
	fg_blue    = "\033[34m"
	fg_magenta = "\033[35m"
	fg_cyan    = "\033[36m"
	fg_white   = "\033[37m"

	bg_black   = "\033[40m"
	bg_red     = "\033[41m"
	bg_green   = "\033[42m"
	bg_yellow  = "\033[43m"
	bg_blue    = "\033[44m"
	bg_magenta = "\033[45m"
	bg_cyan    = "\033[46m"
	bg_white   = "\033[47m"
)

// ListTasks fetches and displays tasks based on filters and sorting.
func ListTasks(tm *TodoManager, projectFilter, contextFilter, tagFilter, statusFilter, startBefore, startAfter, dueBefore, dueAfter, sortBy, order string, format int, displayNotes string) {
	query := `
        SELECT
            t.id, t.title, t.description, p.name, t.start_date, t.due_date, t.end_date, t.status,
            t.recurrence, t.recurrence_interval, t.start_waiting_date, t.end_waiting_date, t.original_task_id
        FROM tasks t
        LEFT JOIN projects p ON t.project_id = p.id
        WHERE 1=1
    `
	args := []any{}
	whereClauses := []string{}

	if projectFilter != "" {
		whereClauses = append(whereClauses, "p.name = ?")
		args = append(args, projectFilter)
	}
	if statusFilter != "" && statusFilter != "all" {
		whereClauses = append(whereClauses, "t.status = ?")
		args = append(args, statusFilter)
	}

	// Date filters
	if startBefore != "" {
		whereClauses = append(whereClauses, "t.start_date <= ?")
		parsed, err := ParseDateTime(startBefore)
		if err != nil {
			log.Printf("Warning: Invalid start-before date format: %v", err)
		} else {
			sqlParsed, _ := parsed.Value() // Get sql.NullTime
			args = append(args, sqlParsed)
		}
	}
	if startAfter != "" {
		whereClauses = append(whereClauses, "t.start_date >= ?")
		parsed, err := ParseDateTime(startAfter)
		if err != nil {
			log.Printf("Warning: Invalid start-after date format: %v", err)
		} else {
			sqlParsed, _ := parsed.Value() // Get sql.NullTime
			args = append(args, sqlParsed)
		}
	}
	if dueBefore != "" {
		whereClauses = append(whereClauses, "t.due_date <= ?")
		parsed, err := ParseDateTime(dueBefore)
		if err != nil {
			log.Printf("Warning: Invalid due-before date format: %v", err)
		} else {
			sqlParsed, _ := parsed.Value() // Get sql.NullTime
			args = append(args, sqlParsed)
		}
	}
	if dueAfter != "" {
		whereClauses = append(whereClauses, "t.due_date >= ?")
		parsed, err := ParseDateTime(dueAfter)
		if err != nil {
			log.Printf("Warning: Invalid due-after date format: %v", err)
		} else {
			sqlParsed, _ := parsed.Value() // Get sql.NullTime
			args = append(args, sqlParsed)
		}
	}

	// Context and Tag filters (require JOINs and GROUP BY or EXISTS subqueries)
	if contextFilter != "" {
		whereClauses = append(whereClauses, `EXISTS (SELECT 1 FROM task_contexts tc JOIN contexts c ON tc.context_id = c.id WHERE tc.task_id = t.id AND c.name = ?)`)
		args = append(args, contextFilter)
	}
	if tagFilter != "" {
		whereClauses = append(whereClauses, `EXISTS (SELECT 1 FROM task_tags tt JOIN tags tg ON tt.tag_id = tg.id WHERE tt.task_id = t.id AND tg.name = ?)`)
		args = append(args, tagFilter)
	}

	if len(whereClauses) > 0 {
		query += " AND " + strings.Join(whereClauses, " AND ")
	}

	// Order by
	orderByMap := map[string]string{
		"id":          "t.id",
		"title":       "t.title",
		"start_date":  "t.start_date",
		"due_date":    "t.due_date",
		"status":      "t.status",
		"project":     "p.name",
		"end_date":    "t.end_date",
	}
	actualSortBy := orderByMap[sortBy]
	if actualSortBy == "" {
		actualSortBy = orderByMap["due_date"] // Default
	}
	if order != "asc" && order != "desc" {
		order = "asc" // Default
	}
	query += fmt.Sprintf(" ORDER BY %s %s", actualSortBy, order)

	rows, err := tm.db.Query(query, args...)
	if err != nil {
		log.Fatalf("Error querying tasks: %v", err)
	}
	defer rows.Close()

	// Load working hours and holidays once for all calculations
	workingHours, err := tm.GetWorkingHours()
	if err != nil {
		log.Fatalf("Error loading working hours: %v", err)
	}
	holidays, err := tm.GetHolidays()
	if err != nil {
		log.Fatalf("Error loading holidays: %v", err)
	}

	// // Print header based on format
	switch format {
	// case DisplayFull:
	// 	fmt.Println("----------------------------------------------------------------------------------------------------------------")
	// 	fmt.Printf("%-5s %-20s %-15s %-12s %-12s %-12s %-10s %-10s %s\n", "ID", "Title", "Project", "Due Date", "Start Date", "End Date", "Status", "Working", "Duration")
	// 	fmt.Println("----------------------------------------------------------------------------------------------------------------")
	// case DisplayCondensed:
	// 	fmt.Println("----------------------------------------------------------------------------------------------------------------")
	// 	fmt.Printf("%-5s %-20s %-15s | %-10s %-15s %-15s | %-12s %-12s %-12s | %-10s %-10s %-10s | %-10s %-10s %s\n",
	// 		"ID", "Title", "Project", "Status", "Tags", "Contexts", "Start Date", "Due Date", "End Date", "Duration", "Working", "Waiting")
	// 	fmt.Println("----------------------------------------------------------------------------------------------------------------")
	case DisplayMinimal:
		fmt.Println("----------------------------------------------------------------------------------------------------------------")
		fmt.Printf("%-5s %-30s %-20s %-20s %-15s %-10s\n", "ID", "Title", "Project", "Tags", "Contexst", "Status")
		fmt.Println("----------------------------------------------------------------------------------------------------------------")
	}


	for rows.Next() {
		var task Task // Assuming Task struct is defined in db_helpers.go
		var project_name sql.NullString
		var desc, recurrence sql.NullString
		var recurrenceInterval sql.NullInt64
		var startDate, dueDate, endDate, startWaitingDate, endWaitingDate sql.NullTime
		var originalTaskID sql.NullInt64

		err := rows.Scan(&task.ID, &task.Title, &desc, &project_name, &startDate, &dueDate, &endDate, &task.Status,
			&recurrence, &recurrenceInterval, &startWaitingDate, &endWaitingDate, &originalTaskID)
		if err != nil {
			log.Printf("Error scanning task: %v", err)
			continue
		}

		task.Description = desc
		task.ProjectName = project_name // Set the project name
		task.StartDate = NullableTime{Time: startDate.Time, Valid: startDate.Valid}
		task.DueDate = NullableTime{Time: dueDate.Time, Valid: dueDate.Valid}
		task.EndDate = NullableTime{Time: endDate.Time, Valid: endDate.Valid}
		task.Recurrence = recurrence
		task.RecurrenceInterval = recurrenceInterval
		task.StartWaitingDate = NullableTime{Time: startWaitingDate.Time, Valid: startWaitingDate.Valid}
		task.EndWaitingDate = NullableTime{Time: endWaitingDate.Time, Valid: endWaitingDate.Valid}
		task.OriginalTaskID = originalTaskID


		// Fetch contexts and tags for the current task using TodoManager method
		task.Contexts = tm.GetTaskNames(int64(task.ID), "task_contexts", "contexts")
		task.Tags = tm.GetTaskNames(int64(task.ID), "task_tags", "tags")

		// Fetch notes based on displayNotes parameter
		if displayNotes != "none" {
			allNotes := tm.GetNotesForTask(task.ID) // GetNotesForTask now returns ASC order
			if displayNotes == "all" {
				task.Notes = allNotes
			} else {
				numNotes, err := strconv.Atoi(displayNotes)
				if err == nil && numNotes > 0 {
					if numNotes > len(allNotes) {
						task.Notes = allNotes
					} else {
						// Slice from the end to get the 'numNotes' most recent notes
						// Since allNotes is ASC, the last `numNotes` elements are the newest
						task.Notes = allNotes[len(allNotes)-numNotes:]
					}
				}
			}
		}

		// Calculate Duration and Working Hours Duration
		totalDurationStr := "N/A"
		workingDurationStr := "N/A"
		waitingDurationStr := "N/A"
		waitingWorkingDurationStr := "N/A" // Initialize new string for waiting working duration


		if task.StartDate.Valid {
			// For completed tasks, calculate from start to end
			if task.Status == "completed" && task.EndDate.Valid {
				totalDuration := CalculateCalendarDuration(task) // Use CalculateCalendarDuration from dateutils
				totalDurationStr = FormatDuration(totalDuration)

				workingDuration := tm.CalculateWorkingDuration(task.StartDate, task.EndDate, workingHours, holidays)
				workingDurationStr = FormatWorkingHoursDisplay(workingDuration) // Use FormatWorkingHoursDisplay from dateutils
			} else if task.Status != "completed" {
				// For in-progress tasks, calculate from start to now
				tempTask := task // Create a temporary task to pass to CalculateCalendarDuration
				tempTask.EndDate = NullableTime{Time: time.Now(), Valid: true}
				totalDuration := CalculateCalendarDuration(tempTask)
				totalDurationStr = FormatDuration(totalDuration)

				// Calculate working duration up to now
				workingDuration := tm.CalculateWorkingDuration(task.StartDate, NullableTime{Time: time.Now(), Valid: true}, workingHours, holidays)
				workingDurationStr = FormatWorkingHoursDisplay(workingDuration) // Use FormatWorkingHoursDisplay from dateutils
			}
		}

		// Calculate waiting duration (calendar time)
		waitingDuration := CalculateWaitingDuration(task)
		waitingDurationStr = FormatDuration(waitingDuration)

		// Calculate working hours within the waiting period
		if task.StartWaitingDate.Valid && task.EndWaitingDate.Valid {
			waitingWorkingDuration := tm.CalculateWorkingDuration(task.StartWaitingDate, task.EndWaitingDate, workingHours, holidays)
			waitingWorkingDurationStr = FormatWorkingHoursDisplay(waitingWorkingDuration)
		}


		switch format {
		case DisplayFull:
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("\n%s%-5d%s", fg_red, task.ID, style_reset))


			titleParts := []string{style_bold + task.Title + style_reset}

			status_str := ""
			switch task.Status {
			case "pending":
				status_str = style_bold + fg_yellow + "pending" + style_reset
			case "completed":
				status_str = style_bold + fg_green + "completed" + style_reset
			case "cancelled":
				status_str = style_bold + fg_red + "canceled" + style_reset
			case "waiting":
				status_str = style_bold + fg_blue + "waiting" + style_reset
			}
			titleParts = append(titleParts, status_str)

			sb.WriteString(fmt.Sprintf(" %s\n", strings.Join(titleParts, " | ")))


			if task.Description.Valid && task.Description.String != "" {
				sb.WriteString(fmt.Sprintf("      📜 %s%s%s%s\n", style_italic, fg_yellow, task.Description.String, style_reset))
			}


			projectParts := []string{}

			if len(task.ProjectName.String) > 0 {
				projectParts = append(projectParts, "📌 Project: " + fg_green + task.ProjectName.String + style_reset)
			}
			if len(task.Tags) > 0 {
				projectParts = append(projectParts, "🏷️ Tags: " + fg_blue + strings.Join(task.Tags, ", ") + style_reset)
			}
			if len(task.Contexts) > 0 {
				projectParts = append(projectParts, "🔖 Context: " + fg_magenta + strings.Join(task.Contexts, ", ") + style_reset)
			}
			if len(projectParts) > 0 {
				sb.WriteString(fmt.Sprintf("      %s\n", strings.Join(projectParts, " | ")))
			}



			dateParts := []string{}

			if task.StartDate.Valid {
				dateParts = append( dateParts, "🚀 Start: " + FormatDisplayDateTime(task.StartDate) )
			}
			if task.EndDate.Valid {
				dateParts = append( dateParts,  "🏁 End: " + FormatDisplayDateTime(task.EndDate) )
			}
			if task.DueDate.Valid {
				dateParts = append( dateParts, "⏱️ Due: " + FormatDisplayDateTime(task.DueDate) )
			}
			if task.Recurrence.Valid {
				interval := ""
				if task.RecurrenceInterval.Valid {
					interval = fmt.Sprintf(" every %d", task.RecurrenceInterval.Int64)
				}
				dateParts = append( dateParts, "🔄 Recurrence: " + task.Recurrence.String + interval)
			}

			if len(dateParts) > 0 {
				sb.WriteString(fmt.Sprintf("      %s\n", strings.Join(dateParts, " | ")))
			}


			waitingParts := []string{}

			if task.StartWaitingDate.Valid {
				waitingParts = append( waitingParts, "⏸️ Pause: " + FormatDisplayDateTime(task.StartWaitingDate) )
			}
			if task.EndWaitingDate.Valid {
				waitingParts = append( waitingParts,  "▶️ End: " + FormatDisplayDateTime(task.EndWaitingDate) )
			}

			if len(waitingParts) > 0 {
				sb.WriteString(fmt.Sprintf("      %s\n", strings.Join(waitingParts, " | ")))
			}



			durationParts := []string{}
			if len(totalDurationStr) > 0 {
				durationParts = append(durationParts, "⌛ Duration: " + totalDurationStr)
			}
			if len(workingDurationStr) > 0 {
				durationParts = append(durationParts, "⌚ Working: " + workingDurationStr)
			}
			if waitingDurationStr != "0s" { // Only add if there's a non-zero waiting calendar duration
				durationParts = append(durationParts, "⏳ Waiting (Calendar): " + waitingDurationStr)
			}
			if waitingWorkingDurationStr != "0s" && waitingWorkingDurationStr != "N/A" { // Only add if there's a non-zero waiting working duration
				durationParts = append(durationParts, "🚧 Waiting (Working): " + waitingWorkingDurationStr)
			}


			if len(durationParts) > 0 {
				sb.WriteString(fmt.Sprintf("      %s\n", strings.Join(durationParts, " | ")))
			}


			// Display Notes
			if len(task.Notes) > 0 {
				sb.WriteString(fmt.Sprintf("      📝 %sNotes:%s\n", style_bold, style_reset))
				// Iterate backwards to display newest (largest ID) first
				for j := len(task.Notes) - 1; j >= 0; j-- {
					note := task.Notes[j]
					if note.Timestamp.Valid && note.Description.Valid {
						// The display ID is 1-based index from the ASC sorted list
						displayID := j + 1
						sb.WriteString(fmt.Sprintf("         %-2d %s%s%s%s: %s%s%s\n", displayID, style_italic, fg_green, FormatDisplayDateTime(note.Timestamp), style_reset, fg_yellow, note.Description.String, style_reset))
					}
				}
			}

			fmt.Printf("%s", sb.String())



		case DisplayCondensed:

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("\n%s%-5d%s", fg_red, task.ID, style_reset))


			titleParts := []string{style_bold + task.Title + style_reset}

			if len(task.ProjectName.String) > 0 {
				titleParts = append(titleParts, fg_green + task.ProjectName.String + style_reset)
			}
			if len(task.Tags) > 0 {
				titleParts = append(titleParts, fg_blue + strings.Join(task.Tags, ", ") + style_reset)
			}
			if len(task.Contexts) > 0 {
				titleParts = append(titleParts, fg_magenta + strings.Join(task.Contexts, ", ") + style_reset)
			}
			if len(titleParts) > 0 {
				sb.WriteString(fmt.Sprintf(" %s\n", strings.Join(titleParts, " | ")))
			}


			if task.Description.Valid && task.Description.String != "" {
				sb.WriteString(fmt.Sprintf("      %s%s%s%s\n", style_italic, fg_yellow, task.Description.String, style_reset))
			}

			// Display Notes
			if len(task.Notes) > 0 {
				sb.WriteString(fmt.Sprintf("      %s%sNotes:%s\n", style_bold, fg_green, style_reset))
				// Iterate backwards to display newest (largest ID) first
				for j := len(task.Notes) - 1; j >= 0; j-- {
					note := task.Notes[j]
					if note.Timestamp.Valid && note.Description.Valid {
						// The display ID is 1-based index from the ASC sorted list
						displayID := j + 1
						sb.WriteString(fmt.Sprintf("         %-2d %s%s%s%s: %s%s%s\n", displayID, style_italic, fg_green, FormatDisplayDateTime(note.Timestamp), style_reset, fg_yellow, note.Description.String, style_reset))
					}
				}
			}

			fmt.Printf("%s", sb.String())



		case DisplayMinimal:
			status_str := ""
			switch task.Status {
			case "pending":
				status_str = style_bold + fg_yellow + "pending" + style_reset
			case "completed":
				status_str = bg_green + fg_black + "completed" + style_reset
			case "cancelled":
				status_str = style_bold + fg_red + "canceled" + style_reset
			case "waiting":
				status_str = style_underline + fg_blue + "waiting" + style_reset
			}

			fmt.Printf("%-5d %s%-30s%s %s%-20s%s %s%-20s%s %s%-15s%s %-10s\n",
				task.ID,
				style_bold, task.Title, style_reset,
				fg_green, task.ProjectName.String, style_reset,
				fg_blue, strings.Join(task.Tags, ", "), style_reset,
				fg_magenta, strings.Join(task.Contexts, ", "), style_reset,
				status_str)
		}
	}
	fmt.Println("----------------------------------------------------------------------------------------------------------------")
}

// ListHolidays lists all configured holidays.
// It now accepts *TodoManager.
func ListHolidays(tm *TodoManager) {
	rows, err := tm.db.Query("SELECT date, name FROM holidays ORDER BY date ASC")
	if err != nil {
		log.Fatalf("Error listing holidays: %v", err)
	}
	defer rows.Close()

	fmt.Println("--- Holidays ---")
	found := false
	for rows.Next() {
		found = true
		var dateStr, name string
		if err := rows.Scan(&dateStr, &name); err != nil {
			log.Printf("Error scanning holiday: %v", err)
			continue
		}
		fmt.Printf("  %s: %s\n", dateStr, name)
	}
	if !found {
		fmt.Println("No holidays configured.")
	}
	if err = rows.Err(); err != nil {
		log.Fatalf("Error after listing holidays: %v", err)
	}
}

// ListWorkingHours lists all configured working hours.
// It now accepts *TodoManager.
func ListWorkingHours(tm *TodoManager) {
	// Query all columns relevant to working hours, including minutes and break duration.
	rows, err := tm.db.Query("SELECT day_of_week, start_hour, start_minute, end_hour, end_minute, break_minutes FROM working_hours ORDER BY day_of_week ASC")
	if err != nil {
		log.Fatalf("Error listing working hours: %v", err)
	}
	defer rows.Close()

	fmt.Println("--- Working Hours ---")
	found := false
	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	for rows.Next() {
		found = true
		var day, startHour, startMinute, endHour, endMinute, breakMinutes int
		// Scan all fetched columns into respective variables.
		if err := rows.Scan(&day, &startHour, &startMinute, &endHour, &endMinute, &breakMinutes); err != nil {
			log.Printf("Error scanning working hours: %v", err)
			continue
		}
		if day >= 0 && day < len(dayNames) {
			// Print working hours including minutes and break duration.
			fmt.Printf("  %-10s %02d:%02d - %02d:%02d (Break: %d minutes)\n", dayNames[day], startHour, startMinute, endHour, endMinute, breakMinutes)
		} else {
			fmt.Printf("  Day %d: %02d:%02d - %02d:%02d (Break: %d minutes) (Invalid Day Index)\n", day, startHour, startMinute, endHour, endMinute, breakMinutes)
		}
	}
	if !found {
		fmt.Println("No working hours configured.")
	}
	if err = rows.Err(); err != nil {
		log.Fatalf("Error after listing working hours: %v", err)
	}
}

// ListProjects lists all projects.
// It now accepts *TodoManager.
func ListProjects(tm *TodoManager) {
	rows, err := tm.db.Query("SELECT id, name FROM projects ORDER BY name ASC")
	if err != nil {
		log.Fatalf("Error listing projects: %v", err)
	}
	defer rows.Close()

	fmt.Println("  ID    Project")
	fmt.Println("----------------------------")
	found := false
	for rows.Next() {
		found = true
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("Error scanning project: %v", err)
			continue
		}
		fmt.Printf("  %-5d %s\n", id, name)
	}
	if !found {
		fmt.Println("No projects found.")
	}
	if err = rows.Err(); err != nil {
		log.Fatalf("Error after listing projects: %v", err)
	}
}

// ListContexts lists all contexts.
// It now accepts *TodoManager.
func ListContexts(tm *TodoManager) {
	rows, err := tm.db.Query("SELECT id, name FROM contexts ORDER BY name ASC")
	if err != nil {
		log.Fatalf("Error listing contexts: %v", err)
	}
	defer rows.Close()

	fmt.Println("  ID    Context")
	fmt.Println("----------------------------")
	found := false
	for rows.Next() {
		found = true
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("Error scanning context: %v", err)
			continue
		}
		fmt.Printf("  %-5d %s\n", id, name)
	}
	if !found {
		fmt.Println("No contexts found.")
	}
	if err = rows.Err(); err != nil {
		log.Fatalf("Error after listing contexts: %v", err)
	}
}

// ListTags lists all tags.
// It now accepts *TodoManager.
func ListTags(tm *TodoManager) {
	rows, err := tm.db.Query("SELECT id, name FROM tags ORDER BY name ASC")
	if err != nil {
		log.Fatalf("Error listing tags: %v", err)
	}
	defer rows.Close()

	fmt.Println("  ID    Tag")
	fmt.Println("----------------------------")
	found := false
	for rows.Next() {
		found = true
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("Error scanning tag: %v", err)
			continue
		}
		fmt.Printf("  %-5d %s\n", id, name)
	}
	if !found {
		fmt.Println("No tags found.")
	}
	if err = rows.Err(); err != nil {
		log.Fatalf("Error after listing tags: %v", err)
	}
}

