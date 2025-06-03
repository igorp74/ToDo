// db_helpers.go
package main

import (
    "database/sql"
    "time"
)

// NullableTime struct to handle NULL timestamps from the database
type NullableTime struct {
    time.Time
    Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface for NullableTime.
func (nt *NullableTime) Scan(value any) error {
    if value == nil {
        nt.Time, nt.Valid = time.Time{}, false
        return nil
    }
    // Attempt to scan as time.Time
    if t, ok := value.(time.Time); ok {
        nt.Time, nt.Valid = t, true
        return nil
    }
    // Fallback for string/blob types if needed (though time.Time is preferred)
    // This part is less critical if the DB driver consistently returns time.Time for DATETIME
    return nil // Allow nil or other types, but set Valid to false if not time.Time
}

// Value implements the driver.Valuer interface for NullableTime.
func (nt NullableTime) Value() (sql.NullTime, error) {
    if !nt.Valid {
        return sql.NullTime{Valid: false}, nil
    }
    return sql.NullTime{Time: nt.Time, Valid: true}, nil
}

// Task represents a single todo task.
type Task struct {
    ID                 int64
    Title              string
    Description        sql.NullString
    ProjectID          sql.NullInt64
    ProjectName        sql.NullString // For display purposes from JOIN
    StartDate          NullableTime
    DueDate            NullableTime
    EndDate            NullableTime   // Completion date
    Status             string         // e.g., pending, completed, cancelled, waiting
    Recurrence         sql.NullString // e.g., "daily", "weekly"
    RecurrenceInterval sql.NullInt64  // e.g., 1, 2
    StartWaitingDate   NullableTime   // Task cannot be started before this date
    EndWaitingDate     NullableTime   // Task cannot be started after this date
    OriginalTaskID     sql.NullInt64  // Added: ID of the original recurring task
    Contexts           []string       // For display purposes, fetched from join table
    Tags               []string       // For display purposes, fetched from join table
}

// Holiday represents a public or personal holiday.
type Holiday struct {
    ID   int64
    Date NullableTime // Use NullableTime for consistency with other dates
    Name string
}

// WorkingHours represents daily working hours.
type WorkingHours struct {
    ID          int64
    DayOfWeek   int // 0=Sunday, 1=Monday, ... 6=Saturday
    StartHour   int // 0-23
    StartMinute int // 0-59
    EndHour     int // 0-24 (exclusive, e.g., 17 for 5 PM)
    EndMinute   int // 0-59
    BreakMinutes int // Added: Duration of break in minutes
}
