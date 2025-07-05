# ToDo
Simple CLI app for task management with SQLite as a backend, task recurrence, notes and calendar&working hours duration.

![image](https://github.com/user-attachments/assets/f9a96f35-2792-4a4f-9569-35ab1a7a6889)


* [Build](https://github.com/igorp74/ToDo?tab=readme-ov-file#-build)
* [Help](https://github.com/igorp74/ToDo?tab=readme-ov-file#-help)
* [Examples](https://github.com/igorp74/ToDo?tab=readme-ov-file#-examples)

---

## ⚙️ BUILD:

### Prerequisites:
1. Installed GO

### Process
1. Create a folder for your Golang workspace (i.e. **ws_todo**)
2. Download this code and extract **ToDo-main** to **ws_todo** folder.
3. Open shell in **ws_todo** folder and type:
   
   `go mod init todo/app`
     
   `go mod tidy`
4. Enter todo folder and type:

   `go build`
5. Depends on environment, you will get the todo.exe on Windows or todo on Linux

   `todo add -t "Test"`   (on Windows)
    
   `./todo add -t "Test"` (on Linux)

    This last step will create **todo.db** in your home folder
6. Save the executable/binary in folder that is on system path
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

Adding task could be rich as the first entry, but also light, as the others:

![image](https://github.com/user-attachments/assets/0987b2dd-aeca-4e7b-b83e-7fb8bc56106c)



## Update tasks
Now, we can start grouping tasks into projects, adding tags, which can be useful for tagging a team, a ticket or whatever key word that has some meaning to you. Also, we can add contexts.

![image](https://github.com/user-attachments/assets/f36f5ef0-664f-4221-a98e-b2b2f056a48e)


## Display formats

OK, tasks can be displayed in 3 format: 

**Default** - this format shows all relevant data related to tasks formated to be functional and logical

![image](https://github.com/user-attachments/assets/f8a5bfe5-8617-4e0e-b590-588f96d4e214)


**Compact** - Keeps more relevant details about tasks, and hide details like start - end times. For clearer overview, without titles and emojis...

![image](https://github.com/user-attachments/assets/954ce171-6328-4add-8ac0-0655af542575)


**Minimal** - this one is suitable for large number of tasks overview, only task titles and projects related

![image](https://github.com/user-attachments/assets/2279fbf5-f1d3-4389-8c15-a199ce77fa2e)


## Durations

With the flag `-D` we can set the due date and with `-s` and `-E` we could add/change start and end times.

![image](https://github.com/user-attachments/assets/9f2ce5f1-42d3-4f17-b3ef-9ffc46e4ebc7)


If I want to see all tasks with all statuese, there is `-st` flag for showing the statuses. Default is pending, but we will use all here. Also, I would like to see only tasks 1 and 2:

![image](https://github.com/user-attachments/assets/7682842e-b75d-4df1-9c8e-9142c5ab9143)


But, wait a minute... where is duration in working hours ?

We need to define working hours first. It could be different for every day in the week, including weekends. So it is up to you how you will define it.

![image](https://github.com/user-attachments/assets/d2e0caa2-bb2b-4852-b281-339003347b6d)

Once defined, it should be working...

![image](https://github.com/user-attachments/assets/8d0c4ac8-d2bd-41e2-9d7a-50cd06635c32)


### When I need to wait for something or someone...

What if I have delay in the process ? Let's say I need to wait for someone else to do something before I can continue. Maybe I need to wait for IT department for access to something...
We can define **start waiting** `-sw` and **end waiting** `-ew` times during the task duration:

![image](https://github.com/user-attachments/assets/a67b3fb9-aee8-488f-8c64-42eaaf7ca950)


And you may see the different duration now. But working hours are the same. Yes, I want it that way. I might change it later, though.

## Notes
Notes are not the same as task descriptions. They are similar, but notes have timestamps and descriptions and you may enter them as many as you like. Notes are useful for tracking parts of the tasks and saving your personal remarks along the task journey. You may display them (all or n last ones) or not.
Number of notes is unlimited (well, this is not entirely true, since you may hit SQLite limit of 281 terabytes, who knows... Some people are notoholics. There is nothing wrong in that.)

Entering notes

![image](https://github.com/user-attachments/assets/4e79f399-01de-4c06-bfb8-30b85c8c1232)

Notes will not be displayed by default. You need to enter the `-n all` flag for showing all notes, or `-n 2` to show only last 2 notes, for example.

![image](https://github.com/user-attachments/assets/8c203484-d284-4f79-8778-07f58306177b)


## Recurrence

Recurrence is available with `-r` for  main recurrence type and `-ri` flags for recurrence interval.

In the next example I will show you how to set recurrence on **Task 5** (ID 5) to repeat every 1 week on Monday, Wednesday and Friday.

![image](https://github.com/user-attachments/assets/6407f0b6-acf6-4974-90b3-77bf40e2aa8f)

Once we set original task with recurrence as completed, a new task with the same attributes, but with different ID would be created on the next date per recurrence. If there are due dates or waiting dates, they will be carry-over to be at same relative distance to the new start date as they were to the original date in previous task.

![image](https://github.com/user-attachments/assets/a0965dc5-28c5-4e4f-a3c0-390882b4d839)

Notice, that the **task** with recurrence which started on Monday, once completed will create **a new task** on the next day in recurrence interval (Wednesday) and carry-over due date to be on the same relative distance from start date as the original start-due dates.

![image](https://github.com/user-attachments/assets/ace6cdc6-6140-4c3b-842d-de45d03c274e)
