package contracts

import (
	"testing"
)

func TestDataTypeString(t *testing.T) {
	cases := []struct {
		dt   DataType
		want string
	}{
		{TypeCredentials, "Credentials"},
		{TypeNotes, "Notes"},
		{TypeFiles, "Files"},
		{TypeCards, "Cards"},
		{DataType(999), "Unknown"},
	}
	for _, c := range cases {
		if got := c.dt.String(); got != c.want {
			t.Errorf("%v: want %s, got %s", c.dt, c.want, got)
		}
	}
}
