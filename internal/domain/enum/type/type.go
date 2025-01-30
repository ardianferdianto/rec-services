package enum_type

import "encoding/json"

type Type int

const (
	DEBIT Type = iota + 1
	CREDIT
)

func (t Type) String() string {
	return [...]string{"",
		"DEBIT",
		"CREDIT",
	}[t]
}

func FromString(status string) Type {
	return map[string]Type{
		"DEBIT":  DEBIT,
		"CREDIT": CREDIT,
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
