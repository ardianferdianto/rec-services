package enum_status

import "encoding/json"

type Type int

const (
	INITIATED Type = iota + 1
	PENDING
	IN_PROGRESS
	COMPLETED
	FAILED
)

func (t Type) String() string {
	return [...]string{"",
		"INITIATED",
		"PENDING",
		"IN_PROGRESS",
		"COMPLETED",
		"FAILED",
	}[t]
}

func FromString(status string) Type {
	return map[string]Type{
		"INITIATED":   INITIATED,
		"PENDING":     PENDING,
		"IN_PROGRESS": IN_PROGRESS,
		"COMPLETED":   COMPLETED,
		"FAILED":      FAILED,
	}[status]
}

func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *Type) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*t = FromString(s)
	return nil
}
