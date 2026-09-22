package models

type Class struct {
	ID           int    `json:"id"`
	EventID      *int   `json:"event_id,omitempty"`
	Name         string `json:"name"`
	StudentCount int    `json:"student_count"`
	AttendCount  int    `json:"attend_count"`
}

// SpecialClassName is the non-student team that does not participate in
// student-count based rules or survey scoring.
const SpecialClassName = "専教"

// DefaultClassNames returns the classes that every sports event starts with.
func DefaultClassNames() []string {
	return []string{
		"1-1", "1-2", "1-3", "IS2", "IS3",
		"IS4", "IS5", "IT2", "IT3", "IT4",
		"IT5", "IE2", "IE3", "IE4", "IE5",
		SpecialClassName,
	}
}
