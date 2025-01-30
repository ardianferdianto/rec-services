package request

import (
	"bytes"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type dummy struct {
	Field1 string    `json:"field_1"`
	Field2 time.Time `json:"field_2"`
}

func TestReadJSON(t *testing.T) {
	testCases := []struct {
		title   string
		body    string
		wantErr error
		want    *dummy
	}{
		{
			title: "valid request",
			body:  `{"field_1": "foo", "field_2": "2024-05-15T10:30:00+07:00"}`,
			want: &dummy{
				Field1: "foo",
				Field2: time.Date(2024, 5, 15, 10, 30, 0, 0, time.Local),
			},
		},
		{
			title:   "invalid time format",
			body:    `{"field_1": "foo", "field_2": "tomorrow"}`,
			wantErr: errors.New(`invalid request body: parsing time "tomorrow" as "2006-01-02T15:04:05Z07:00": cannot parse "tomorrow" as "2006"`),
			want: &dummy{
				Field1: "foo",
				Field2: time.Time{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			reqBody := bytes.NewReader([]byte(tc.body))
			req, _ := http.NewRequest(http.MethodPost, "/dummy", reqBody)

			var data dummy
			err := ReadJSON(req, &data)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, &data)
		})
	}
}
