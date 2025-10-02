package repositories_test

import (
	"testing"
	"time"

	"github.com/t1m4/db_part_dump/internal/repositories"

	"github.com/google/go-cmp/cmp"
)

func TestAnyToPsqlString(t *testing.T) {
	type TestData struct {
		name     string
		value    any
		expected string
	}
	tests := []TestData{
		{name: "test bool", value: true, expected: "t"},
		{name: "test bool", value: false, expected: "f"},
		{name: "test nil", value: nil, expected: "\\N"},
		{name: "test slice of byte", value: []byte("hello"), expected: "hello"},
		{name: "test time", value: time.Date(2008, 6, 8, 12, 50, 31, 42, time.UTC), expected: "2008-06-08T12:50:31.000000042Z"},
	}
	for _, test := range tests {
		actual := repositories.AnyToPsqlString(test.value)
		if diff := cmp.Diff(test.expected, actual); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	}
}
