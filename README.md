# ToDo
Simple CLI app for task management with SQLite as a backend

## ⚙️ BUILD:

### Prerequisites:
1. Installed GO

### Process
1. Create a folder for your Golang workspace (i.e. **ws_todo**)
2. Create a folder for source data (i.e. **todo**)
3. Copy files to **todo** folder
4. Open shell in **ws_todo** folder and type:
   
   `go mod init todo/app`
     
   `go mod tidy`
6. Enter todo folder and type:

   `go build`
8. Depends on environment, you will get the todo.exe on Windows or todo on Linux

   `todo add -t "Test"`   (on Windows)
    
   `./todo add -t "Test"` (on Linux)

    This last step will create **todo.db** in your home folder
10. Save the executable/binary in folder that is on system path
    * `sudo cp todo /usr/local/bin` (on Linux)
  
    * Or add your path to the environment variable (Windows)
    * `setx /M PATH "%PATH%;<your-new-path>"`
    


## 🎒 HELP


Usage: `todo [global options] <command> [command options]`
  A simple CLI todo application.

**Global Options**:

  `--db-path`     Custom path and name for the database file (e.g., /path/to/my/todo.db)

**Commands:**

  `add`  Add a new todo task.
  
    -t, --title Title of the task (required)
    -d, --description   Description of the task
    -p, --project       Project name (will be created if not exists)
    -s, --start-date    Start date (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.
    -D, --due-date      Due date (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.
    -r, --recurrence    Recurrence pattern (daily, weekly, monthly, yearly)
    -ri, --recurrence-interval  Interval for recurrence (e.g., 2 for every 2 days) (default: 1)
    -c, --contexts      Comma-separated list of contexts (e.g., 'work,home')
    -T, --tags  Comma-separated list of tags (e.g., 'urgent,bug')
    -sw, --start-waiting        Start date of waiting period (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.
    -ew, --end-waiting  End date of waiting period (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.
    -st, --status       Initial status of the task (pending, completed, cancelled, waiting) (default: pending)

  `del`   Delete a task by ID.
  
    --ids       Comma-separated IDs or ID ranges of tasks to delete (e.g., '1,2,3-5,10')
    -i, --id    ID of a single task to delete (use -ids for multiple or ranges)
    -C, --complete      Mark task as completed instead of deleting (for recurring tasks)

  `update`        Update an existing task.
  
    --ids       Comma-separated IDs or ID ranges of tasks to update (e.g., '1,2,3-5,10')
    -i, --id    ID of a single task to update (use -ids for multiple or ranges)
    -t, --title New title of the task
    -d, --description   New description of the task
    -p, --project       New project name
    -s, --start-date    New start date (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.
    -D, --due-date      New due date (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.
    -E, --end-date      New end date (completion date) (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.
    -st, --status       New status (pending, completed, cancelled, waiting)
    -r, --recurrence    New recurrence pattern
    -ri, --recurrence-interval  New interval for recurrence
    -c, --contexts      Comma-separated list of contexts (replaces existing)
    -T, --tags  Comma-separated list of tags (replaces existing)
    -sw, --start-waiting        New start date of waiting period (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.
    -ew, --end-waiting  New end date of waiting period (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.
    --clear-p   Clear project association
    --clear-c   Clear all context associations
    --clear-T   Clear all tag associations
    --clear-s   Clear start date
    --clear-D   Clear due date
    --clear-E   Clear end date
    --clear-r   Clear recurrence
    --clear-wait        Clear waiting period

  `add-note`      Add a new note to a task.
  
    -i, --task-id       ID of the task to add a note to (required)
    -d, --description   Description of the note (required)

  `update-note`   Update an existing note by its permanent database ID.
  
    -n, --id    Permanent database ID of the note to update (as shown in 'list' command) (required)
    -d, --description   New description for the note
    -ts, --timestamp    New timestamp for the note (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.

  `delete-note`   Delete one or more notes by ID.
  
    --ids       Comma-separated IDs or ID ranges of notes to delete (e.g., '1,2,3-5,10')
    --all       Delete all notes
    -ti, --task-id      ID of the task whose notes should be deleted
    --all-for-task      Delete all notes associated with the specified task ID


  `list`  List tasks.
  
    -p, --project       Filter by project name
    -c, --context       Filter by context name
    -T, --tag   Filter by tag name
    -st, --status       Filter by status (pending, completed, cancelled, waiting, all) (default: pending)
    --start-before      Filter by start date before (YYYY-MM-DD HH:MM:SS)
    --start-after       Filter by start date after (YYYY-MM-DD HH:MM:SS)
    --due-before        Filter by due date before (YYYY-MM-DD HH:MM:SS)
    --due-after Filter by due date after (YYYY-MM-DD HH:MM:SS)
    --sort-by   Sort by field (id, title, start_date, due_date, status, project) (default: due_date)
    --order     Sort order (asc, desc) (default: asc)
    -f, --format        Output format: 0=Full, 1=Condensed, 2=Minimal (default: 0)
    -n, --notes Display notes: 'none', 'all', or a number (e.g., '1', '2' for last N notes) (default: none)

  `holiday`       Manage holidays.
  
    Subcommands for holiday:
      holiday add       Add a new holiday.
          -d, --date    Date of the holiday (YYYY-MM-DD) (required)
          -n, --name    Name of the holiday (required)
      holiday list      List all holidays.
      holiday del       Delete one or more holidays by ID or delete all.
          --ids Comma-separated IDs or ID ranges of holidays to delete (e.g., '1,2,3-5,10')
          --all Delete all holidays

  `workhours`     Manage working hours.
  
    Subcommands for workhours:
      workhours set     Set working hours for a day of the week.
          -d, --day     Day of week (0=Sunday, 1=Monday, ..., 6=Saturday) (required)
          -sh, --start-hour     Start hour (0-23) (required)
          -sM, --start-minute   Start minute (0-59) (default: 0)
          -eh, --end-hour       End hour (0-24) (required)
          -eM, --end-minute     End minute (0-59) (default: 0)
          -b, --break-minutes   Break duration in minutes for this day (default: 0)
      workhours list    List all defined working hours.
      workhours del     Delete working hours for one or more days or delete all.
          --days        Comma-separated day of week numbers or ranges to delete working hours for (e.g., '1,2,3-5')
          --all Delete all working hours


  `projects`      List all projects.

  `contexts`      List all contexts.

  `tags`  List all tags.


# 📽️ Examples
This example shows very basic task creations and then updating projects over task range.



https://github.com/user-attachments/assets/dcd0e6b8-5d7d-4565-8ea8-eb2c8a1df01c

