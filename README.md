# ToDo
Simple CLI app for task management written in Golang with SQLite as a backend, task recurrence, unlimited notes and working hours duration.

![image](https://github.com/user-attachments/assets/7d167e7c-cc48-409c-8bee-375c4ee1e96b)


# 📢 Features
* 🛢️ **SQLite database** as a backend - I wanted one central place for all my tasks.
  * Easier synching between machines and operating systems - **portable**
  * No need for archiving/moving many files - **compact**
  * With [DB Browser for SQLIte](https://sqlitebrowser.org) or other RDBMS tools like [DBeaver](https://dbeaver.io/download/) you may analyse (or edit) all of your data
* 🔄 **Task recurrence**
  * Make your work with repetitive tasks fun
  * Keep track of your events duration
* 📜 **Unlimited notes** per task(s)
  * Every note has its own timestamp. You may add, delete or update notes & timestamps.
  * I am using notes as milestones during the longer tasks
* 🏖️ **Holidays manager**
  * You may add, edit or delete any holidays and non-working days you want.
  * Essentially, this is for more precise calculating of working hours
* ⌛ **Working hours manager**
  * Define your working days in the week. As you want...
* 🕶️ **3 different display formats**: Default, Compact and Minimal
  * It is not always importnat to see every tiny detail on your task or project, thus compact view
  * Want to check how many time you spent on daily meetings ? See overall tasks on projects ? Then minimal view is for you.
* 🗂️ **Projects, Tags, Contexts**
  * For grouping tasks, enriching organisation and make search more logical
  * Built-in manager for adding, deleting, updating and listing projects, tags and contexts
* 📌 **Filter**
  * You may filter your tasks per status, projects, tags and/or contexts. One in a time or all at once.
  * You may filter tasks before or after start, due or end times    
* 🔎 **Search**
  * Search for any text in task titles, descriptions and notes (case-insensitive)
 
for all features and commands, please read the documentation

# Documentation
* [⚙️ Build](https://github.com/igorp74/ToDo/wiki/%E2%9A%99%EF%B8%8F-BUILD)
* [🚀 Usage](https://github.com/igorp74/ToDo/wiki/%F0%9F%9A%80-Usage)
* [🎬 Examples](https://github.com/igorp74/ToDo/wiki/%F0%9F%8E%AC-Examples)
