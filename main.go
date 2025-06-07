// main.go
package main

import (
    "database/sql" // Import sql for sql.NullInt64
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
)

func main() {
    // Define command-line flags and arguments
    parser := NewParser("todo", "A simple CLI todo application.")

    // Global flag for database path
    dbPath := parser.String("db-path", "", &Options{Help: "Custom path and name for the database file (e.g., /path/to/my/todo.db)"})

    // Add command
    addCmd := parser.NewCommand("add", "Add a new todo task.")
    addTitle := addCmd.String("title", "t", &Options{Required: true, Help: "Title of the task"})
    addDesc := addCmd.String("description", "d", &Options{Help: "Description of the task"})
    addProject := addCmd.String("project", "p", &Options{Help: "Project name (will be created if not exists)"})
    addStart := addCmd.String("start-date", "s", &Options{Help: "Start date (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time."})
    addDue := addCmd.String("due-date", "D", &Options{Help: "Due date (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time."})
    addRecurrence := addCmd.String("recurrence", "r", &Options{Help: "Recurrence pattern (daily, weekly, monthly, yearly)"})
    addRecurrenceInterval := addCmd.Int("recurrence-interval", "ri", &Options{Default: 1, Help: "Interval for recurrence (e.g., 2 for every 2 days)"})
    addContexts := addCmd.StringList("contexts", "c", &Options{Help: "Comma-separated list of contexts (e.g., 'work,home')"})
    addTags := addCmd.StringList("tags", "T", &Options{Help: "Comma-separated list of tags (e.g., 'urgent,bug')"})
    addStartWaiting := addCmd.String("start-waiting", "sw", &Options{Help: "Start date of waiting period (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time."})
    addEndWaiting := addCmd.String("end-waiting", "ew", &Options{Help: "End date of waiting period (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time."})
    addStatus := addCmd.String("status", "st", &Options{Default: "pending", Help: "Initial status of the task (pending, completed, cancelled, waiting)"})

    // Delete command
    delCmd := parser.NewCommand("del", "Delete a task by ID.")
    delIDs := delCmd.String("ids", "", &Options{Help: "Comma-separated IDs or ID ranges of tasks to delete (e.g., '1,2,3-5,10')"})
    delID := delCmd.Int("id", "i", &Options{Help: "ID of a single task to delete (use -ids for multiple or ranges)"})
    delComplete := delCmd.Flag("complete", "C", &Options{Help: "Mark task as completed instead of deleting (for recurring tasks)"})

    // Update command
    updateCmd := parser.NewCommand("update", "Update an existing task.")
    updateIDs := updateCmd.String("ids", "", &Options{Help: "Comma-separated IDs or ID ranges of tasks to update (e.g., '1,2,3-5,10')"})
    updateID := updateCmd.Int("id", "i", &Options{Help: "ID of a single task to update (use -ids for multiple or ranges)"})
    updateTitle := updateCmd.String("title", "t", &Options{Help: "New title of the task"})
    updateDesc := updateCmd.String("description", "d", &Options{Help: "New description of the task"})
    updateProject := updateCmd.String("project", "p", &Options{Help: "New project name"})
    updateStart := updateCmd.String("start-date", "s", &Options{Help: "New start date (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time."})
    updateDue := updateCmd.String("due-date", "D", &Options{Help: "New due date (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time."})
    updateEnd := updateCmd.String("end-date", "E", &Options{Help: "New end date (completion date) (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time."})
    updateStatus := updateCmd.String("status", "st", &Options{Help: "New status (pending, completed, cancelled, waiting)"}) // Unified flag
    updateRecurrence := updateCmd.String("recurrence", "r", &Options{Help: "New recurrence pattern"})
    updateRecurrenceInterval := updateCmd.Int("recurrence-interval", "ri", &Options{Help: "New interval for recurrence"})
    updateContexts := updateCmd.StringList("contexts", "c", &Options{Help: "Comma-separated list of contexts (replaces existing)"})
    updateTags := updateCmd.StringList("tags", "T", &Options{Help: "Comma-separated list of tags (replaces existing)"})
    updateStartWaiting := updateCmd.String("start-waiting", "sw", &Options{Help: "New start date of waiting period (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time."})
    updateEndWaiting := updateCmd.String("end-waiting", "ew", &Options{Help: "New end date of waiting period (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time."})

    // Clear flags for update command
    updateClearProject := updateCmd.Flag("clear-p", "", &Options{Help: "Clear project association"})
    updateClearContexts := updateCmd.Flag("clear-c", "", &Options{Help: "Clear all context associations"})
    updateClearTags := updateCmd.Flag("clear-T", "", &Options{Help: "Clear all tag associations"})
    updateClearStart := updateCmd.Flag("clear-s", "", &Options{Help: "Clear start date"})
    updateClearDue := updateCmd.Flag("clear-D", "", &Options{Help: "Clear due date"})
    updateClearEnd := updateCmd.Flag("clear-E", "", &Options{Help: "Clear end date"})
    updateClearRecurrence := updateCmd.Flag("clear-r", "", &Options{Help: "Clear recurrence"})
    updateClearWaiting := updateCmd.Flag("clear-wait", "", &Options{Help: "Clear waiting period"})

    // Add Note command
    addNoteCmd := parser.NewCommand("add-note", "Add a new note to a task.")
    addNoteTaskID := addNoteCmd.Int("task-id", "i", &Options{Required: true, Help: "ID of the task to add a note to"})
    addNoteDescription := addNoteCmd.String("description", "d", &Options{Required: true, Help: "Description of the note"})

    // Update Note command - MODIFIED TO USE ACTUAL NOTE ID
    updateNoteCmd := parser.NewCommand("update-note", "Update an existing note by its permanent database ID.")
    updateNoteID := updateNoteCmd.Int("id", "n", &Options{Required: true, Help: "Permanent database ID of the note to update (as shown in 'list' command)"})
    updateNoteDescription := updateNoteCmd.String("description", "d", &Options{Help: "New description for the note"})
    updateNoteTimestamp := updateNoteCmd.String("timestamp", "ts", &Options{Help: "New timestamp for the note (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time."})

    // Delete Note command
    deleteNoteCmd := parser.NewCommand("delete-note", "Delete one or more notes by ID.")
    deleteNoteIDs := deleteNoteCmd.String("ids", "", &Options{Help: "Comma-separated IDs or ID ranges of notes to delete (e.g., '1,2,3-5,10')"})
    deleteNoteAll := deleteNoteCmd.Flag("all", "", &Options{Help: "Delete all notes"})
    // New flags for task-specific note deletion
    deleteNoteTaskID := deleteNoteCmd.Int("task-id", "ti", &Options{Help: "ID of the task whose notes should be deleted"})
    deleteNoteAllForTask := deleteNoteCmd.Flag("all-for-task", "", &Options{Help: "Delete all notes associated with the specified task ID"})


    // List command
    listCmd := parser.NewCommand("list", "List tasks.")
    listProject := listCmd.String("project", "p", &Options{Help: "Filter by project name"})
    listContext := listCmd.String("context", "c", &Options{Help: "Filter by context name"})
    listTag := listCmd.String("tag", "T", &Options{Help: "Filter by tag name"})
    listStatus := listCmd.String("status", "st", &Options{Default: "pending", Help: "Filter by status (pending, completed, cancelled, waiting, all)"})
    listStartBefore := listCmd.String("start-before", "", &Options{Help: "Filter by start date before (YYYY-MM-DD HH:MM:SS)"})
    listStartAfter := listCmd.String("start-after", "", &Options{Help: "Filter by start date after (YYYY-MM-DD HH:MM:SS)"})
    listDueBefore := listCmd.String("due-before", "", &Options{Help: "Filter by due date before (YYYY-MM-DD HH:MM:SS)"})
    listDueAfter := listCmd.String("due-after", "", &Options{Help: "Filter by due date after (YYYY-MM-DD HH:MM:SS)"})
    listSortBy := listCmd.String("sort-by", "", &Options{Default: "due_date", Help: "Sort by field (id, title, start_date, due_date, status, project)"})
    listOrder := listCmd.String("order", "", &Options{Default: "asc", Help: "Sort order (asc, desc)"})
    listFormat := listCmd.Int("format", "f", &Options{Default: DisplayFull, Help: "Output format: 0=Full, 1=Condensed, 2=Minimal"})
    listNotes := listCmd.String("notes", "n", &Options{Default: "none", Help: "Display notes: 'none', 'all', or a number (e.g., '1', '2' for last N notes)"})

    // Holiday commands
    holidayCmd := parser.NewCommand("holiday", "Manage holidays.")
    holidayAddCmd := holidayCmd.NewCommand("add", "Add a new holiday.")
    holidayAddDate := holidayAddCmd.String("date", "d", &Options{Required: true, Help: "Date of the holiday (YYYY-MM-DD)"})
    holidayAddName := holidayAddCmd.String("name", "n", &Options{Required: true, Help: "Name of the holiday"})
    holidayListCmd := holidayCmd.NewCommand("list", "List all holidays.")
    holidayDelCmd := holidayCmd.NewCommand("del", "Delete one or more holidays by ID or delete all.") // Modified help text
    holidayDelIDs := holidayDelCmd.String("ids", "", &Options{Help: "Comma-separated IDs or ID ranges of holidays to delete (e.g., '1,2,3-5,10')"})
    holidayDelAll := holidayDelCmd.Flag("all", "", &Options{Help: "Delete all holidays"})


    // Working hours commands
    workhoursCmd := parser.NewCommand("workhours", "Manage working hours.")
    workhoursSetCmd := workhoursCmd.NewCommand("set", "Set working hours for a day of the week.")
    workhoursSetDay := workhoursSetCmd.Int("day", "d", &Options{Required: true, Help: "Day of week (0=Sunday, 1=Monday, ..., 6=Saturday)"})
    workhoursSetStartHour := workhoursSetCmd.Int("start-hour", "sh", &Options{Required: true, Help: "Start hour (0-23)"})
    workhoursSetStartMinute := workhoursSetCmd.Int("start-minute", "sM", &Options{Default: 0, Help: "Start minute (0-59)"})
    workhoursSetEndHour := workhoursSetCmd.Int("end-hour", "eh", &Options{Required: true, Help: "End hour (0-24)"})
    workhoursSetEndMinute := workhoursSetCmd.Int("end-minute", "eM", &Options{Default: 0, Help: "End minute (0-59)"})
    workhoursSetBreakMinutes := workhoursSetCmd.Int("break-minutes", "b", &Options{Default: 0, Help: "Break duration in minutes for this day"})
    workhoursListCmd := workhoursCmd.NewCommand("list", "List all defined working hours.")
    workhoursDelCmd := workhoursCmd.NewCommand("del", "Delete working hours for one or more days or delete all.") // Modified help text
    workhoursDelDays := workhoursDelCmd.String("days", "", &Options{Help: "Comma-separated day of week numbers or ranges to delete working hours for (e.g., '1,2,3-5')"})
    workhoursDelAll := workhoursDelCmd.Flag("all", "", &Options{Help: "Delete all working hours"})


    // List projects command
    listProjectsCmd := parser.NewCommand("projects", "List all projects.")

    // List contexts command
    listContextsCmd := parser.NewCommand("contexts", "List all contexts.")

    // List tags command
    listTagsCmd := parser.NewCommand("tags", "List all tags.")

    err := parser.Parse(os.Args)
    if err != nil {
        fmt.Println(parser.Usage(err))
        return
    }

    // Initialize TodoManager with the determined database path
    tm := NewTodoManager(*dbPath) // Correctly instantiate tm
    defer tm.Close()

    switch {
    case addCmd.Parsed:
        tm.AddTask(
            nil, // Pass nil for the transaction when adding a new task directly
            *addTitle,
            *addDesc,
            *addProject,
            *addStart,
            addCmd.GetFlag("start-date").IsSet, // Pass IsSet status for start-date
            *addDue,
            addCmd.GetFlag("due-date").IsSet, // Pass IsSet status for due-date
            *addRecurrence,
            *addRecurrenceInterval,
            *addContexts,
            *addTags,
            *addStartWaiting,
            addCmd.GetFlag("start-waiting").IsSet,
            *addEndWaiting,
            addCmd.GetFlag("end-waiting").IsSet,
            *addStatus,
            sql.NullInt64{}, // Pass empty sql.NullInt64 for originalTaskID for new tasks
        )
    case delCmd.Parsed:
        var targetIDs []int64
        if *delIDs != "" {
            var parseErr error
            targetIDs, parseErr = parseIDs(*delIDs) // Use generic parseIDs
            if parseErr != nil {
                fmt.Printf("Error parsing task IDs: %v\n", parseErr)
                fmt.Println(parser.Usage(nil))
                os.Exit(1)
            }
        } else if *delID != 0 {
            targetIDs = []int64{int64(*delID)} // Cast int to int64
        } else {
            fmt.Println("At least one Task ID is required for 'del' command using -id or -ids.")
            fmt.Println(parser.Usage(nil))
            os.Exit(1)
        }
        // Call the new DeleteTasks method
        tm.DeleteTasks(targetIDs, *delComplete)
    case updateCmd.Parsed:
        var targetIDs []int64
        if *updateIDs != "" {
            var parseErr error
            targetIDs, parseErr = parseIDs(*updateIDs) // Use generic parseIDs
            if parseErr != nil {
                fmt.Printf("Error parsing task IDs: %v\n", parseErr)
                fmt.Println(parser.Usage(nil))
                os.Exit(1)
            }
        } else if *updateID != 0 {
            targetIDs = []int64{int64(*updateID)} // Cast int to int64
        } else {
            fmt.Println("At least one Task ID is required for 'update' command using -id or -ids.")
            fmt.Println(parser.Usage(nil)) // Corrected: Call parser.Usage
            os.Exit(1)
        }

        err := tm.UpdateTasks(targetIDs,
            *updateTitle, *updateDesc, *updateProject,
            *updateStart, updateCmd.GetFlag("start-date").IsSet,
            *updateDue, updateCmd.GetFlag("due-date").IsSet,
            *updateEnd, updateCmd.GetFlag("end-date").IsSet,
            *updateStatus,
            *updateRecurrence, *updateRecurrenceInterval,
            *updateContexts, *updateTags,
            *updateStartWaiting, updateCmd.GetFlag("start-waiting").IsSet,
            *updateEndWaiting, updateCmd.GetFlag("end-waiting").IsSet,
            *updateClearProject, *updateClearContexts, *updateClearTags,
            *updateClearStart, *updateClearDue, *updateClearEnd, *updateClearRecurrence, *updateClearWaiting)
        if err != nil {
            log.Fatalf("Error updating tasks: %v", err)
        }
    case addNoteCmd.Parsed:
        tm.AddNoteToTask(int64(*addNoteTaskID), *addNoteDescription)
    case updateNoteCmd.Parsed:
        // Check if at least one of description or timestamp is provided
        if *updateNoteDescription == "" && !updateNoteCmd.GetFlag("timestamp").IsSet {
            fmt.Println("At least one of --description or --timestamp must be provided for 'update-note' command.")
            fmt.Println(parser.Usage(nil))
            os.Exit(1)
        }
        // Pass the directly provided note ID for update
        tm.UpdateNote(int64(*updateNoteID), *updateNoteDescription, *updateNoteTimestamp, updateNoteCmd.GetFlag("timestamp").IsSet)
    case deleteNoteCmd.Parsed:
        // Prioritize specific task notes deletion, then global all, then specific note IDs
        if *deleteNoteTaskID != 0 && *deleteNoteAllForTask {
            tm.DeleteAllNotesForTask(int64(*deleteNoteTaskID))
        } else if *deleteNoteAll {
            tm.DeleteAllNotes()
        } else if *deleteNoteIDs != "" {
            noteIDsToDelete, parseErr := parseIDs(*deleteNoteIDs) // Use generic parseIDs for notes
            if parseErr != nil {
                fmt.Printf("Error parsing note IDs: %v\n", parseErr)
                fmt.Println(parser.Usage(nil))
                os.Exit(1)
            }
            tm.DeleteNotes(noteIDsToDelete)
        } else {
            fmt.Println("At least one of --ids, --all, or (--task-id and --all-for-task) is required for 'delete-note' command.")
            fmt.Println(parser.Usage(nil))
            os.Exit(1)
        }
    case listCmd.Parsed:
        ListTasks(tm, *listProject, *listContext, *listTag, *listStatus,
            *listStartBefore, *listStartAfter, *listDueBefore, *listDueAfter,
            *listSortBy, *listOrder, *listFormat, *listNotes)

    case holidayAddCmd.Parsed:
        tm.AddHoliday(*holidayAddDate, *holidayAddName)
    case holidayListCmd.Parsed:
        ListHolidays(tm)
    case holidayDelCmd.Parsed: // New case for deleting holidays
        if *holidayDelAll {
            tm.DeleteAllHolidays()
        } else if *holidayDelIDs != "" {
            idsToDelete, parseErr := parseIDs(*holidayDelIDs)
            if parseErr != nil {
                fmt.Printf("Error parsing holiday IDs: %v\n", parseErr)
                fmt.Println(parser.Usage(nil))
                os.Exit(1)
            }
            tm.DeleteHolidays(idsToDelete)
        } else {
            fmt.Println("At least one of --ids or --all is required for 'holiday del' command.")
            fmt.Println(parser.Usage(nil))
            os.Exit(1)
        }
    case workhoursSetCmd.Parsed:
        tm.SetWorkingHours(*workhoursSetDay, *workhoursSetStartHour, *workhoursSetStartMinute, *workhoursSetEndHour, *workhoursSetEndMinute, *workhoursSetBreakMinutes)
    case workhoursListCmd.Parsed:
        ListWorkingHours(tm)
    case workhoursDelCmd.Parsed: // New case for deleting working hours
        if *workhoursDelAll {
            tm.DeleteAllWorkingHours()
        } else if *workhoursDelDays != "" {
            daysToDelete, parseErr := parseIDs(*workhoursDelDays) // parseIDs works for int64, need to convert to int
            if parseErr != nil {
                fmt.Printf("Error parsing day IDs for working hours: %v\n", parseErr)
                fmt.Println(parser.Usage(nil))
                os.Exit(1)
            }
            var intDaysToDelete []int
            for _, id := range daysToDelete {
                intDaysToDelete = append(intDaysToDelete, int(id))
            }
            tm.DeleteWorkingHoursByDays(intDaysToDelete)
        } else {
            fmt.Println("At least one of --days or --all is required for 'workhours del' command.")
            fmt.Println(parser.Usage(nil))
            os.Exit(1)
        }
    case listProjectsCmd.Parsed:
        ListProjects(tm)
    case listContextsCmd.Parsed:
        ListContexts(tm)
    case listTagsCmd.Parsed:
        ListTags(tm)
    default:
        fmt.Println(parser.Usage(nil))
    }
}

// parseIDs parses a comma-separated string of IDs and ID ranges
// (e.g., "1,3-5,8") into a unique slice of int64 IDs.
// This function is now generic and can be used for tasks, notes, etc.
func parseIDs(idStr string) ([]int64, error) {
    uniqueIDs := make(map[int64]bool)
    parts := strings.Split(idStr, ",")

    for _, part := range parts {
        part = strings.TrimSpace(part)
        if part == "" {
            continue
        }

        if strings.Contains(part, "-") {
            // It's a range
            rangeParts := strings.Split(part, "-")
            if len(rangeParts) != 2 {
                return nil, fmt.Errorf("invalid ID range format: %s. Expected 'start-end'", part)
            }
            start, err := strconv.ParseInt(strings.TrimSpace(rangeParts[0]), 10, 64)
            if err != nil {
                return nil, fmt.Errorf("invalid start ID in range '%s': %w", part, err)
            }
            end, err := strconv.ParseInt(strings.TrimSpace(rangeParts[1]), 10, 64)
            if err != nil {
                return nil, fmt.Errorf("invalid end ID in range '%s': %w", part, err)
            }
            if start > end {
                return nil, fmt.Errorf("start ID (%d) cannot be greater than end ID (%d) in range '%s'", start, end, part)
            }
            for i := start; i <= end; i++ {
                uniqueIDs[i] = true
            }
        } else {
            // It's a single ID
            id, err := strconv.ParseInt(part, 10, 64)
            if err != nil {
                return nil, fmt.Errorf("invalid single ID '%s': %w", part, err)
            }
            uniqueIDs[id] = true
        }
    }

    var ids []int64
    for id := range uniqueIDs {
        ids = append(ids, id)
    }
    // Optionally sort the IDs for consistent behavior, though not strictly necessary for functionality
    // sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
    return ids, nil
}

// A simple argparse-like library for Go (for demonstration purposes)
// In a real project, consider using a more robust library like "cobra" or "urfave/cli"

type Options struct {
    Required bool
    Default  any
    Help     string
}

type Flag struct {
    Name    string
    Short   string
    Value   any
    Options *Options
    IsSet   bool // Added to track if the flag was explicitly set
}

type Command struct {
    Name     string
    Help     string
    Flags    []*Flag
    Commands []*Command // Added to support subcommands
    Parsed   bool
    parent   *Parser    // Parent is a Parser for top-level commands
}

// NewCommand method for Command struct to create subcommands
func (c *Command) NewCommand(name, help string) *Command {
    subCmd := &Command{Name: name, Help: help}
    c.Commands = append(c.Commands, subCmd)
    return subCmd
}

// String method for Command struct to define string flags
func (c *Command) String(name, short string, opts *Options) *string {
    var val string
    if opts != nil && opts.Default != nil {
        val = opts.Default.(string)
    }
    flag := &Flag{Name: name, Short: short, Value: &val, Options: opts}
    c.Flags = append(c.Flags, flag)
    return &val
}

// StringList method for Command struct to define string list flags
func (c *Command) StringList(name, short string, opts *Options) *[]string {
    var val []string
    if opts != nil && opts.Default != nil {
        val = opts.Default.([]string)
    }
    flag := &Flag{Name: name, Short: short, Value: &val, Options: opts}
    c.Flags = append(c.Flags, flag)
    return &val
}

// Int method for Command struct to define integer flags
func (c *Command) Int(name, short string, opts *Options) *int {
    var val int
    if opts != nil && opts.Default != nil {
        val = opts.Default.(int)
    }
    flag := &Flag{Name: name, Short: short, Value: &val, Options: opts}
    c.Flags = append(c.Flags, flag)
    return &val
}

// Flag method for Command struct to define boolean flags
func (c *Command) Flag(name, short string, opts *Options) *bool {
    var val bool
    if opts != nil && opts.Default != nil {
        val = opts.Default.(bool)
    }
    flag := &Flag{Name: name, Short: short, Value: &val, Options: opts}
    c.Flags = append(c.Flags, flag)
    return &val
}

// GetFlag retrieves a flag by its name. Used to check IsSet status.
func (c *Command) GetFlag(name string) *Flag {
    for _, flag := range c.Flags {
        if flag.Name == name || flag.Short == name {
            return flag
        }
    }
    return nil
}

type Parser struct {
    Name     string
    Help     string
    Commands []*Command
    Flags    []*Flag // Global flags for the parser itself
    parsed   bool
}

func NewParser(name, help string) *Parser {
    return &Parser{Name: name, Help: help}
}

// NewCommand method for Parser struct to create top-level commands
func (p *Parser) NewCommand(name, help string) *Command {
    cmd := &Command{Name: name, Help: help, parent: p}
    p.Commands = append(p.Commands, cmd)
    return cmd
}

// String method for Parser struct to define global string flags
func (p *Parser) String(name, short string, opts *Options) *string {
    var val string
    if opts != nil && opts.Default != nil {
        val = opts.Default.(string)
    }
    flag := &Flag{Name: name, Short: short, Value: &val, Options: opts}
    p.Flags = append(p.Flags, flag)
    return &val
}

func (p *Parser) Parse(args []string) error {
    if len(args) < 2 {
        return fmt.Errorf("no command provided")
    }

    // Parse global flags first
    cmdArgs := args[1:] // Arguments after the program name
    remainingArgs := []string{}
    for i := 0; i < len(cmdArgs); i++ {
        arg := cmdArgs[i]
        isFlag := false
        var flagName string

        if strings.HasPrefix(arg, "--") {
            flagName = strings.TrimPrefix(arg, "--")
            isFlag = true
        } else if strings.HasPrefix(arg, "-") {
            flagName = strings.TrimPrefix(arg, "-")
            isFlag = true
        }

        if isFlag {
            foundFlag := false
            for _, flag := range p.Flags { // Check global flags
                if flag.Name == flagName || flag.Short == flagName {
                    flag.IsSet = true
                    foundFlag = true
                    switch v := flag.Value.(type) {
                    case *string:
                        // If the next argument is not a flag, it's the value
                        if i+1 < len(cmdArgs) && !strings.HasPrefix(cmdArgs[i+1], "-") {
                            *v = cmdArgs[i+1]
                            i++
                        } else {
                            // Flag is present but no value provided (e.g., `--flag` instead of `--flag value`)
                            *v = "" // Ensure it's an empty string if no value is given
                        }
                    }
                    break
                }
            }
            if !foundFlag { // If not a global flag, it's part of the command arguments
                remainingArgs = append(remainingArgs, arg)
            }
        } else { // Not a flag, so it's a command or subcommand argument
            remainingArgs = append(remainingArgs, arg)
        }
    }

    if len(remainingArgs) == 0 {
        return fmt.Errorf("no command provided after global flags")
    }

    // Now parse commands and their flags
    cmdName := remainingArgs[0]
    var currentCmd *Command
    var argStartIndex int = 1 // Default start index for flags after command name

    for _, cmd := range p.Commands {
        if cmd.Name == cmdName {
            currentCmd = cmd
            break
        }
    }

    if currentCmd == nil {
        return fmt.Errorf("unknown command: %s", cmdName)
    }

    // If a top-level command is found, check if it has subcommands and if a subcommand was provided
    if len(currentCmd.Commands) > 0 && len(remainingArgs) >= 2 {
        subCmdName := remainingArgs[1]
        foundSubcommand := false
        for _, subCmd := range currentCmd.Commands {
            if subCmd.Name == subCmdName {
                currentCmd = subCmd // Switch to the subcommand
                argStartIndex = 2    // Flags start after the subcommand
                foundSubcommand = true
                break
            }
        }
        // If a subcommand was expected but not found (remainingArgs[1] exists and isn't a flag)
        if !foundSubcommand && !strings.HasPrefix(remainingArgs[1], "-") {
            return fmt.Errorf("unknown subcommand '%s' for command '%s'", subCmdName, cmdName)
        }
    }

    currentCmd.Parsed = true
    p.parsed = true

    flagArgs := remainingArgs[argStartIndex:]

    for i := 0; i < len(flagArgs); i++ {
        arg := flagArgs[i]
        isFlag := false
        var flagName string

        if strings.HasPrefix(arg, "--") {
            flagName = strings.TrimPrefix(arg, "--")
            isFlag = true
        } else if strings.HasPrefix(arg, "-") {
            flagName = strings.TrimPrefix(arg, "-")
            isFlag = true
        }

        if isFlag {
            foundFlag := false
            for _, flag := range currentCmd.Flags {
                if flag.Name == flagName || flag.Short == flagName {
                    flag.IsSet = true // Mark the flag as set
                    foundFlag = true

                    switch v := flag.Value.(type) {
                    case *string:
                        // Check if the next argument exists and is not another flag
                        if i+1 < len(flagArgs) && !strings.HasPrefix(flagArgs[i+1], "-") {
                            *v = flagArgs[i+1]
                            i++
                        } else {
                            // Flag is present but no value provided (e.g., `--flag` instead of `--flag value`)
                            *v = "" // Ensure it's an empty string if no value is given
                        }
                    case *[]string:
                        // This is the change: split the single argument by comma
                        if i+1 < len(flagArgs) && !strings.HasPrefix(flagArgs[i+1], "-") {
                            valuesStr := flagArgs[i+1]
                            parts := strings.Split(valuesStr, ",")
                            var cleanedParts []string
                            for _, p := range parts {
                                trimmed := strings.TrimSpace(p)
                                if trimmed != "" {
                                    cleanedParts = append(cleanedParts, trimmed)
                                }
                            }
                            *v = cleanedParts
                            i++ // Consume the value argument
                        } else {
                            // Flag is present but no value provided, set to empty slice
                            *v = []string{}
                        }
                    case *int:
                        if i+1 < len(flagArgs) && !strings.HasPrefix(flagArgs[i+1], "-") {
                            val, err := strconv.Atoi(flagArgs[i+1])
                            if err != nil {
                                return fmt.Errorf("flag --%s requires an integer value", flag.Name)
                            }
                            *v = val
                            i++
                        } else if flag.Options != nil && flag.Options.Required {
                            return fmt.Errorf("flag --%s requires a value", flag.Name)
                        }
                    case *bool:
                        *v = true // For boolean flags, presence means true
                    }
                    break
                }
            }
            if !foundFlag {
                return fmt.Errorf("unknown flag: %s", arg)
            }
        } else {
            return fmt.Errorf("unexpected argument: %s", arg)
        }
    }

    // Check for required flags
    for _, flag := range currentCmd.Flags {
        if flag.Options != nil && flag.Options.Required && !flag.IsSet {
            return fmt.Errorf("required flag --%s is missing", flag.Name)
        }
    }

    return nil
}

func (p *Parser) Usage(err error) string {
    var sb strings.Builder
    if err != nil {
        sb.WriteString(fmt.Sprintf("Error: %s\n\n", err))
    }
    sb.WriteString(fmt.Sprintf("Usage: %s [global options] <command> [command options]\n", p.Name))
    sb.WriteString(fmt.Sprintf("  %s\n\n", p.Help))

    // Global flags usage
    if len(p.Flags) > 0 {
        sb.WriteString("Global Options:\n")
        for _, flag := range p.Flags {
            short := ""
            if flag.Short != "" {
                short = fmt.Sprintf("-%s, ", flag.Short)
            }
            defaultValue := ""
            if flag.Options != nil && flag.Options.Default != nil {
                defaultValue = fmt.Sprintf(" (default: %v)", flag.Options.Default)
            }
            sb.WriteString(fmt.Sprintf("  %s--%s\t%s%s\n", short, flag.Name, flag.Options.Help, defaultValue))
        }
        sb.WriteString("\n")
    }

    sb.WriteString("Commands:\n")
    for _, cmd := range p.Commands {
        sb.WriteString(fmt.Sprintf("\n  %s%s%s\t%s%s\n", style_bold, fg_green, cmd.Name, style_reset, cmd.Help))
        if len(cmd.Flags) > 0 {
            for _, flag := range cmd.Flags {
                short := ""
                if flag.Short != "" {
                    short = fmt.Sprintf("-%s, ", flag.Short)
                }
                required := ""
                if flag.Options != nil && flag.Options.Required {
                    required = " (required)"
                }
                defaultValue := ""
                if flag.Options != nil && flag.Options.Default != nil {
                    defaultValue = fmt.Sprintf(" (default: %v)", flag.Options.Default)
                }
                sb.WriteString(fmt.Sprintf("    %s--%s\t%s%s%s\n", short, flag.Name, flag.Options.Help, required, defaultValue))
            }
        }
        // List subcommands
        if len(cmd.Commands) > 0 {
            sb.WriteString(fmt.Sprintf("    Subcommands for %s:\n", cmd.Name))
            for _, subCmd := range cmd.Commands {
                sb.WriteString(fmt.Sprintf("      %s %s\t%s\n", cmd.Name, subCmd.Name, subCmd.Help))
                for _, flag := range subCmd.Flags {
                    short := ""
                    if flag.Short != "" {
                        short = fmt.Sprintf("-%s, ", flag.Short)
                    }
                    required := ""
                    if flag.Options != nil && flag.Options.Required {
                        required = " (required)"
                    }
                    defaultValue := ""
                    if flag.Options != nil && flag.Options.Default != nil {
                        defaultValue = fmt.Sprintf(" (default: %v)", flag.Options.Default)
                    }
                    // Changed \\n to \n to correctly render newlines
                    sb.WriteString(fmt.Sprintf("          %s--%s\t%s%s%s\n", short, flag.Name, flag.Options.Help, required, defaultValue))
                }
            }
        }
    }
    return sb.String()
}
