package main

import (
    "database/sql"
    "fmt"
    "strings"
    "time"
)

// ParseDateTime parses a date/time string into NullableTime.
// It attempts to parse various formats:YYYY-MM-DD HH:MM:SS,YYYY-MM-DD, MM-DD-YYYY, DD-MM-YYYY.
// The parsed time is always converted to UTC before being returned, for consistent storage.
// If a location is provided, the string is parsed relative to that location; otherwise, UTC is assumed for parsing.
func ParseDateTime(dateTimeStr string, loc *time.Location) (NullableTime, error) {
    if dateTimeStr == "" {
        return NullableTime{Valid: false}, nil
    }

    if loc == nil {
        // Default to UTC if no location is provided for parsing input string
        loc = time.UTC
    }

    layouts := []string{
        "2006-01-02 15:04:05", // YYYY-MM-DD HH:MM:SS
        "2006-01-02",          // YYYY-MM-DD
        "01-02-2006",          // MM-DD-YYYY
        "02-01-2006",          // DD-MM-YYYY
    }

    for _, layout := range layouts {
        t, err := time.ParseInLocation(layout, dateTimeStr, loc)
        if err == nil {
            // Always convert to UTC for storage
            return NullableTime{Time: t.UTC(), Valid: true}, nil
        }
    }

    return NullableTime{Valid: false}, fmt.Errorf("could not parse date/time '%s'. Please use YYYY-MM-DD HH:MM:SS, YYYY-MM-DD, MM-DD-YYYY, or DD-MM-YYYY format", dateTimeStr)
}

// FormatDuration formats a time.Duration into a human-readable string (days, hours, minutes, seconds),
// skipping any components that are zero. This is for general calendar duration.
func FormatDuration(d time.Duration) string { // Renamed to FormatDuration
    totalSeconds := int(d.Seconds())
    if totalSeconds == 0 {
        return "0s" // Changed to a shorter format for minimal display
    }

    // Make duration positive for display
    if totalSeconds < 0 {
        totalSeconds = -totalSeconds
    }

    days := totalSeconds / (24 * 3600)
    remainingSecondsAfterDays := totalSeconds % (24 * 3600)
    hours := remainingSecondsAfterDays / 3600
    minutes := (remainingSecondsAfterDays % 3600) / 60
    seconds := remainingSecondsAfterDays % 60

    parts := []string{}
    if days > 0 {
        parts = append(parts, fmt.Sprintf("%dd", days))
    }
    if hours > 0 {
        parts = append(parts, fmt.Sprintf("%dh", hours))
    }
    if minutes > 0 {
        parts = append(parts, fmt.Sprintf("%dm", minutes))
    }
    if seconds > 0 {
        parts = append(parts, fmt.Sprintf("%ds", seconds))
    }

    return strings.Join(parts, " ") // Join with space for shorter output
}

// FormatWorkingHoursDisplay formats a time.Duration into a human-readable string
// specifically for working hours, assuming an 8-hour working day for 'days' calculation.
// It skips any components that are zero.
func FormatWorkingHoursDisplay(d time.Duration) string { // Renamed to FormatWorkingHoursDisplay
    totalSeconds := int(d.Seconds())
    if totalSeconds == 0 {
        return "0s" // Changed to a shorter format for minimal display
    }

    // Assume 8 hours per working day for display purposes when converting to "working days"
    workingHoursPerDay := 8
    secondsPerWorkingDay := workingHoursPerDay * 3600

    workingDays := totalSeconds / secondsPerWorkingDay
    remainingSeconds := totalSeconds % secondsPerWorkingDay

    hours := remainingSeconds / 3600
    minutes := (remainingSeconds % 3600) / 60
    seconds := remainingSeconds % 60

    parts := []string{}
    if workingDays > 0 {
        parts = append(parts, fmt.Sprintf("%dwd", workingDays)) // "wd" for working days
    }
    if hours > 0 {
        parts = append(parts, fmt.Sprintf("%dh", hours))
    }
    if minutes > 0 {
        parts = append(parts, fmt.Sprintf("%dm", minutes))
    }
    if seconds > 0 {
        parts = append(parts, fmt.Sprintf("%ds", seconds))
    }

    if len(parts) == 0 {
        return "0s" // Fallback if all components are zero after calculation
    }
    return strings.Join(parts, " ") // Join with space for shorter output
}

// FormatDisplayDateTime formats a NullableTime into the desired "Day Mon-DD-YYYY HH:MM:SS" format.
// It converts the stored UTC time to the local timezone for display.
// Returns "N/A" if the NullableTime is not valid.
func FormatDisplayDateTime(nt NullableTime) string { // Changed to accept NullableTime
    if !nt.Valid {
        return "N/A" // Consistent with other "N/A" for invalid dates
    }
    // Convert UTC time to local time for display
    t := nt.Time.Local()
    dayAbbr := t.Format("Mon")
    formattedTime := t.Format("2006-01-02 15:04:05")
    return fmt.Sprintf("%s %s", dayAbbr, formattedTime)
}

// CalculateCalendarDuration calculates the duration of a task in calendar time, excluding waiting time.
// It ensures startDate is before or equal to endDate by swapping if necessary.
func CalculateCalendarDuration(task Task) time.Duration { // Returns time.Duration
    if !task.StartDate.Valid {
        return 0 // Return zero duration if start date is missing
    }

    startDate := task.StartDate.Time.Local() // Convert to local for calculation
    var endDate time.Time
    if task.EndDate.Valid {
        endDate = task.EndDate.Time.Local() // Convert to local for calculation
    } else {
        endDate = time.Now() // If not completed, duration to today (local time)
    }

    // Swap dates if start date is after end date to ensure positive duration
    if startDate.After(endDate) {
        startDate, endDate = endDate, startDate
    }

    totalDuration := endDate.Sub(startDate)
    waitingDuration := time.Duration(0)

    if task.StartWaitingDate.Valid && task.EndWaitingDate.Valid {
        startWaiting := task.StartWaitingDate.Time.Local() // Convert to local for calculation
        endWaiting := task.EndWaitingDate.Time.Local()     // Convert to local for calculation

        if startWaiting.Before(endWaiting) {
            // Calculate intersection of task duration and waiting period
            actualWaitStart := MaxTime(startDate, startWaiting)
            actualWaitEnd := MinTime(endDate, endWaiting)

            if actualWaitStart.Before(actualWaitEnd) {
                waitingDuration = actualWaitEnd.Sub(actualWaitStart)
            }
        }
    }

    return totalDuration - waitingDuration
}

// CalculateDurationToDueDate calculates the remaining time until the due date.
func CalculateDurationToDueDate(task Task) time.Duration { // Returns time.Duration
    if !task.DueDate.Valid {
        return 0 // Return zero duration if due date is missing
    }

    dueDate := task.DueDate.Time.Local() // Convert to local for comparison
    now := time.Now()                    // Local time

    // If due date is in the past, return a negative duration or 0, depending on desired behavior.
    // For now, let's return 0 if due date is in the past, as it's "no remaining time".
    if dueDate.Before(now) {
        return 0
    }

    return dueDate.Sub(now)
}

// CalculateTimeDifference calculates the time difference between a given NullableTime and now.
// It returns the absolute duration and a boolean indicating if the given time is in the past (true for overdue).
func CalculateTimeDifference(targetDate NullableTime) (time.Duration, bool) {
    if !targetDate.Valid {
        return 0, false // No valid date, no difference
    }

    now := time.Now() // Local time
    target := targetDate.Time.Local() // Convert to local for comparison

    if target.Before(now) {
        return now.Sub(target), true // Target is in the past, return positive duration and true for overdue
    }
    return target.Sub(now), false // Target is in the future, return positive duration and false for not overdue
}


// CalculateWaitingDuration calculates the duration of the waiting period for a task.
// It returns the duration as time.Duration.
func CalculateWaitingDuration(task Task) time.Duration {
    if !task.StartWaitingDate.Valid || !task.EndWaitingDate.Valid {
        return 0 // No valid waiting period defined
    }

    startWaiting := task.StartWaitingDate.Time.Local() // Convert to local for calculation
    endWaiting := task.EndWaitingDate.Time.Local()     // Convert to local for calculation

    if startWaiting.After(endWaiting) {
        return 0 // Invalid waiting period (start after end)
    }

    return endWaiting.Sub(startWaiting)
}

// Helper to get working hours for a specific day of week from the database.
// Returns 0, 0, nil if no specific working hours are set for the day.
// This function should ideally be a method on TodoManager or accept a *sql.DB.
// For now, it's kept here as a standalone helper, assuming db connection is passed.
// Note: This function is currently not used directly after changes to CalculateWorkingHoursDuration
// because CalculateWorkingHoursDuration now accepts the full workingHours map.
func GetWorkingHoursForDay(db *sql.DB, dayOfWeek int) (startHour, endHour int, err error) {
    err = db.QueryRow("SELECT start_hour, end_hour FROM working_hours WHERE day_of_week = ?", dayOfWeek).Scan(&startHour, &endHour)
    if err == sql.ErrNoRows {
        return 0, 0, nil // No specific working hours set for this day
    }
    return
}

// Helper to check if a given date is a holiday.
// This function should ideally be a method on TodoManager or accept a *sql.DB.
// For now, it's kept here as a standalone helper, assuming db connection is passed.
// Note: This function is currently not used directly after changes to CalculateWorkingHoursDuration
// because CalculateWorkingHoursDuration now accepts the full holidays map.
func IsHoliday(db *sql.DB, date time.Time) (bool, error) {
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM holidays WHERE date = ?", date.Format("2006-01-02")).Scan(&count)
    if err != nil {
        return false, fmt.Errorf("error checking holiday: %w", err)
    }
    return count > 0, nil
}

// CalculateWorkingHoursDuration calculates working hours between two dates,
// considering defined working hours and skipping holidays, and subtracting breaks.
// It ensures startDate is before or equal to endDate by swapping if necessary.
// Returns the total working duration as time.Duration.
// All NullableTime inputs are assumed to be in UTC, and converted to local for calculations.
func CalculateWorkingHoursDuration(db *sql.DB, start, end NullableTime, workingHours map[time.Weekday]WorkingHours, holidays map[string]Holiday) time.Duration {
    if !start.Valid || !end.Valid {
        return 0 // Return zero duration if dates are invalid
    }

    // Convert input UTC times to local time for consistent daily calculations
    startDate := start.Time.Local()
    endDate := end.Time.Local()

    // Swap dates if start date is after end date to ensure positive duration
    if startDate.After(endDate) {
        startDate, endDate = endDate, startDate
    }

    totalWorkingDuration := time.Duration(0)

    currentDay := startDate.Truncate(24 * time.Hour)
    // Iterate through each day from startDate to endDate (inclusive for the end day if it falls within working hours)
    // Adding 24 * time.Hour to endDate ensures the end day is also considered if it has working hours.
    for currentDay.Before(endDate.Add(24 * time.Hour).Truncate(24 * time.Hour)) {
        dateKey := currentDay.Format("2006-01-02")
        if _, isHol := holidays[dateKey]; isHol {
            currentDay = currentDay.Add(24 * time.Hour)
            continue // Skip holidays
        }

        dayOfWeek := currentDay.Weekday()
        wh, hasWorkingHours := workingHours[dayOfWeek]

        // Check if working hours are defined and if there's a valid working period for the day
        if hasWorkingHours && (wh.StartHour*60+wh.StartMinute < wh.EndHour*60+wh.EndMinute) {
            // Create daily working hour times in the current day's location (Local)
            dailyWorkStart := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), wh.StartHour, wh.StartMinute, 0, 0, currentDay.Location())
            dailyWorkEnd := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), wh.EndHour, wh.EndMinute, 0, 0, currentDay.Location())

            // Handle overnight shifts (e.g., 22:00 to 06:00 next day)
            if dailyWorkEnd.Before(dailyWorkStart) {
                dailyWorkEnd = dailyWorkEnd.Add(24 * time.Hour)
            }

            // Calculate the intersection of the task's overall time range and the current day's working hours
            effectiveIntersectionStart := MaxTime(startDate, dailyWorkStart)
            effectiveIntersectionEnd := MinTime(endDate, dailyWorkEnd)

            // Add the duration of the intersection to the total working duration
            if effectiveIntersectionStart.Before(effectiveIntersectionEnd) {
                dailyWorkingTime := effectiveIntersectionEnd.Sub(effectiveIntersectionStart)
                // Subtract break duration if the working period for the day is substantial enough to include a break
                breakDuration := time.Duration(wh.BreakMinutes) * time.Minute
                if dailyWorkingTime > breakDuration { // Corrected logic: if there's enough time to work AFTER the break
                    dailyWorkingTime -= breakDuration
                } else { // If the working time is less than or equal to the break duration
                    dailyWorkingTime = 0 // No effective working time after break
                }
                totalWorkingDuration += dailyWorkingTime
            }
        }
        currentDay = currentDay.Add(24 * time.Hour)
    }

    return totalWorkingDuration
}

// MaxTime returns the later of two times.
func MaxTime(t1, t2 time.Time) time.Time {
    if t1.After(t2) {
        return t1
    }
    return t2
}

// MinTime returns the earlier of two times.
func MinTime(t1, t2 time.Time) time.Time {
    if t1.Before(t2) {
        return t1
    }
    return t2
}
