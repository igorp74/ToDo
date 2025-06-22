# ToDo
Simple CLI app for task management with SQLite as a backend, task recurrence, notes and calendar&working hours duration.

![Screenshot_20250620_005744](https://github.com/user-attachments/assets/78671f1d-1f55-4041-848a-530dd4899c5c)


* [Build](https://github.com/igorp74/ToDo?tab=readme-ov-file#-build)
* [Help](https://github.com/igorp74/ToDo?tab=readme-ov-file#-help)
* [Examples](https://github.com/igorp74/ToDo?tab=readme-ov-file#-examples)

---

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
    -E, --end-date      End date (completion date) (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.
    -r, --recurrence    Recurrence pattern (daily, weekly, monthly, yearly, or comma-separated days of week for weekly, e.g., 'weekly:Tue,Thu')
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
    -ac, --add-contexts Comma-separated list of contexts to add (e.g., 'new_work,urgent_call'). Will append to existing.
    -rc, --remove-contexts      Comma-separated list of contexts to remove (e.g., 'old_context'). Will remove from existing.
    -at, --add-tags     Comma-separated list of tags to add (e.g., 'new_feature,high_priority'). Will append to existing.
    -rt, --remove-tags  Comma-separated list of tags to remove (e.g., 'bug_fix'). Will remove from existing.
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
    -ts, --timestamp    Timestamp for the note (YYYY-MM-DD HH:MM:SS orYYYY-MM-DD). Use empty string with flag to set current time.


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
    --end-before        Filter by end date before (YYYY-MM-DD HH:MM:SS)
    --end-after Filter by end date after (YYYY-MM-DD HH:MM:SS)
    --sort-by   Sort by field (id, title, start_date, due_date, status, project, end_date) (default: due_date)
    --order     Sort order (asc, desc) (default: asc)
    -f, --format        Output format: 0=Full, 1=Condensed, 2=Minimal (default: 0)
    -n, --notes Display notes: 'none', 'all', or a number (e.g., '1', '2' for last N notes) (default: none)
    -i, --ids   Comma-separated IDs or ID ranges of tasks to list (e.g., '1,2,3-5,10')
    -S, --search        Search for text in task titles, descriptions and notes (case-insensitive)



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


# 🎬 Examples

Adding task could be rich:

![Screenshot_20250620_002419](https://github.com/user-attachments/assets/5d1696b0-a849-47d9-85ff-33a9224c1e4e)


But also could be very fast:

![Screenshot_20250620_002517](https://github.com/user-attachments/assets/b09c5840-7aaf-4e65-aeee-a8eec426ffb0)


## Update tasks
Now, we can start grouping tasks into projects, adding tags, which can be useful for tagging a team, a ticket or whatever key word that has some meaning to you. Also, we can add contexts.

![Screenshot_20250620_002905](https://github.com/user-attachments/assets/734ad7da-7a27-488b-9846-45d21b0f90f4)


## Display formats

OK, tasks can be displayed in 3 format: 

**Default** - this format shows all relevant data related to tasks formated to be functional and logical

![Screenshot_20250620_002947](https://github.com/user-attachments/assets/4b86ed30-8673-4da3-b293-45a1849cd89c)


**Compact** - Keeps more relevant details about tasks, and hide details like start - end times. For clearer overview, without titles and emojis...

![Screenshot_20250620_003022](https://github.com/user-attachments/assets/b1ef7bcf-2d20-4244-90c7-20182790adda)


**Minimal** - this one is suitable for large number of tasks overview, only task titles and projects related

![Screenshot_20250620_003039](https://github.com/user-attachments/assets/5066ec7c-4bf4-48c5-ae5c-e8a31645cf81)


## Durations

With flag `-E` we will end tasks with the current timestamp as the end date, and `-s` will set up the start time:

![Screenshot_20250620_003259](https://github.com/user-attachments/assets/f50f36cb-6eb7-4167-9610-de138270a798)


But, wait a minute... where is duration in working hours ?

![image](https://github.com/user-attachments/assets/f3e937da-544e-4901-a93c-83eb248bb635)

We need to define working hours first. It could be different for every day in the week, including weekends. So it is up to you how you will define it.

Once defined, it should be working...

![image](https://github.com/user-attachments/assets/a6223cfc-b14b-4bbd-b9f3-584b3ca84271)


### When I need to wait for something or someone...

What if I have delay in the process ? Let's say I need to wait for someone else to do something before I can continue. Maybe I need to wait for IT department for access to something...
We can define **start waiting** `-sw` and **end waiting** `-ew` times during the task duration:

![Screenshot_20250622_090727](https://github.com/user-attachments/assets/2a962eac-1632-4f09-866e-87f47124231b)


And you may see the different duration now. But working hours are the same. Yes, I want it that way. I might change it later, though.

## Notes
Notes are not the same as task descriptions. They are similar, but notes have timestamps and descriptions and you may enter them as many as you like. Notes are useful for tracking parts of the tasks and saving your personal remarks along the task journey. You may display them (all or n last ones) or not.
Number of notes is unlimited (well, this is not entirely true, since you may hit SQLite limit of 281 terabytes, who knows... Some people are notoholics. There is nothing wrong in that.)

![Screenshot_20250620_005332](https://github.com/user-attachments/assets/e885343e-e647-4505-a4ca-69e0ac73ca9f)

Notes will not be displayed by default. You need to enter the `-n all` flag for showing all notes, or `-n 2` to show only last 2 notes, for example.
