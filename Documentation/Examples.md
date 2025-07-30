# 🎬 Examples

## Adding tasks
Adding task could be rich as the first entry, but also light, as the others:

![image](https://github.com/user-attachments/assets/0987b2dd-aeca-4e7b-b83e-7fb8bc56106c)

## Update tasks
Now, we can start grouping tasks into projects, adding tags, which can be useful for tagging a team, a ticket or whatever key word that has some meaning to you. Also, we can add contexts.

![image](https://github.com/user-attachments/assets/f36f5ef0-664f-4221-a98e-b2b2f056a48e)

## Display formats

OK, tasks can be displayed in 3 format: 

### Default

this format shows all relevant data related to tasks and formated to be functional and logical

![image](https://github.com/user-attachments/assets/f8a5bfe5-8617-4e0e-b590-588f96d4e214)

### Compact

Keeps more relevant details about tasks, and hide details like start - end times. For clearer overview, without titles and emojis...

![image](https://github.com/user-attachments/assets/954ce171-6328-4add-8ac0-0655af542575)

### Minimal
this one is suitable for big overwiev of large number of tasks, with start time, duration, project and task titles.

<img width="914" height="290" alt="image" src="https://github.com/user-attachments/assets/e61df692-d227-4877-ace4-a9ffb380e820" />


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
