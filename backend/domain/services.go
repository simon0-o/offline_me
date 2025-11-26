package domain

import "time"

// AttendanceProvider defines the interface for fetching attendance data from external systems
// This interface is defined in the domain layer, and implemented in the infrastructure layer
type AttendanceProvider interface {
	FetchCheckInTime(config *WorkConfig, date string) (*time.Time, error)
	FetchAttendanceStatus(config *WorkConfig, date string) (hasCheckedIn, hasCheckedOut bool, err error)
}
