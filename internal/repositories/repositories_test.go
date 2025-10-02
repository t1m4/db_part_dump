package repositories_test

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"testing"

	"github.com/t1m4/db_part_dump/config"
	"github.com/t1m4/db_part_dump/internal/constants"
	"github.com/t1m4/db_part_dump/internal/db"
	"github.com/t1m4/db_part_dump/internal/repositories"
	"github.com/t1m4/db_part_dump/internal/schemas"
	"github.com/t1m4/db_part_dump/internal/testutil"

	"github.com/google/go-cmp/cmp"
)

func TestGetPKColumnName(t *testing.T) {
	type TestData struct {
		name       string
		schemaName string
		tableName  string
		expected   string
		err        error
	}
	db := testutil.CreateTestDb(t)
	testutil.InsertTestData(t, db)
	repos := repositories.New(db)
	ctx := context.Background()
	tests := []TestData{
		{"test users table", "alpha", "users", "id", nil},
		{"test table_one table", "alpha", "table_one", "id", nil},
		{"test wrong table", "alpha", "test", "", sql.ErrNoRows},
	}
	for _, test := range tests {
		pkColumnName, err := repos.GetPKColumnName(ctx, test.schemaName, test.tableName)
		if diff := cmp.Diff(test.expected, pkColumnName); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
		if err != test.err {
			t.Errorf("wrong err: %v, expected %v", err, test.err)
		}
	}
}

func TestGetPks(t *testing.T) {
	type TestData struct {
		name         string
		schemaName   string
		table        config.Table
		pkColumnName string
		expected     []map[string]any
		err          error
	}
	db := testutil.CreateTestDb(t)
	testutil.InsertTestData(t, db)
	repos := repositories.New(db)
	ctx := context.Background()
	tests := []TestData{
		{
			"test int pk",
			"alpha",
			config.Table{Name: "users", Filters: []config.Filter{{Name: "id", Value: "1, 2, 3"}}},
			"id",
			[]map[string]any{
				{"id": int64(1)},
				{"id": int64(2)},
				{"id": int64(3)},
			},
			nil,
		},
		{
			"test guid pk",
			"alpha",
			config.Table{Name: "table_one", Filters: []config.Filter{{Name: "id", Value: "'11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222'"}}},
			"id",
			[]map[string]any{
				{"id": []byte("11111111-1111-1111-1111-111111111111")},
			},
			nil,
		},
		{
			"test empty",
			"alpha",
			config.Table{Name: "users", Filters: []config.Filter{{Name: "id", Value: "-1, -2, -3"}}},
			"id",
			[]map[string]any{},
			nil,
		},
	}
	for _, test := range tests {
		actual, err := repos.GetPkIdRows(ctx, test.schemaName, test.table, test.pkColumnName)
		if diff := cmp.Diff(test.expected, actual); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
		if test.err != err {
			t.Errorf("wrong err: %v, expected %v", err, test.expected)
		}
	}
}

func TestGetFks(t *testing.T) {
	type TestData struct {
		name              string
		direction         string
		schemaName        string
		tableName         string
		isIncludeIncoming bool
		expected          []db.Fk
		err               error
	}
	testDb := testutil.CreateTestDb(t)
	testutil.InsertTestData(t, testDb)
	repos := repositories.New(testDb)
	ctx := context.Background()
	tests := []TestData{
		{
			name:              "test outgoing",
			direction:         constants.OUTGOING,
			schemaName:        "alpha",
			tableName:         "orders",
			isIncludeIncoming: false,
			expected: []db.Fk{
				{
					ColumnName:         "user_id",
					ForeignTableSchema: "alpha",
					ForeignTableName:   "users",
					ForeignColumnName:  "id",
					Direction:          constants.OUTGOING,
				},
			},
			err: nil,
		},
		{
			name:              "test outgoing with isIncludeIncoming",
			direction:         constants.OUTGOING,
			schemaName:        "alpha",
			tableName:         "orders",
			isIncludeIncoming: true,
			expected: []db.Fk{
				{
					ColumnName:         "user_id",
					ForeignTableSchema: "alpha",
					ForeignTableName:   "users",
					ForeignColumnName:  "id",
					Direction:          constants.OUTGOING,
				},
				{
					ColumnName:         "id",
					ForeignTableSchema: "alpha",
					ForeignTableName:   "order_items",
					ForeignColumnName:  "order_id", Direction: constants.INCOMING},
				{
					ColumnName:         "id",
					ForeignTableSchema: "alpha",
					ForeignTableName:   "user_payment_methods",
					ForeignColumnName:  "order_id",
					Direction:          constants.INCOMING,
				},
				{
					ColumnName:         "id",
					ForeignTableSchema: "alpha",
					ForeignTableName:   "order_coupons",
					ForeignColumnName:  "order_id",
					Direction:          constants.INCOMING,
				},
			},
			err: nil,
		},
	}
	for _, test := range tests {
		actual, err := repos.GetFKs(ctx, test.direction, test.schemaName, test.tableName, test.isIncludeIncoming)
		if diff := cmp.Diff(test.expected, actual); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
		if test.err != err {
			t.Errorf("wrong err: %v, expected %v", err, test.expected)
		}
	}
}

func TestGetFkIdRows(t *testing.T) {
	type TestData struct {
		name       string
		schemaName string
		table      config.Table
		fks        []db.Fk
		expected   []map[string]any
		err        error
	}
	testDb := testutil.CreateTestDb(t)
	testutil.InsertTestData(t, testDb)
	repos := repositories.New(testDb)
	ctx := context.Background()
	tests := []TestData{
		{
			name:       "test get user_payment_methods",
			schemaName: "alpha",
			table:      config.Table{Name: "user_payment_methods", Filters: []config.Filter{{Name: "id", Value: "1, 2, 3"}}},
			fks:        []db.Fk{{ColumnName: "user_id"}, {ColumnName: "order_id"}},
			expected: []map[string]any{
				{"user_id": int64(1), "order_id": int64(1)},
				{"user_id": int64(1), "order_id": nil},
				{"user_id": int64(2), "order_id": int64(3)},
			},
		},
	}
	for _, test := range tests {
		actual, err := repos.GetFkIdRows(ctx, test.schemaName, test.table, test.fks)
		if diff := cmp.Diff(test.expected, actual); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
		if test.err != err {
			t.Errorf("wrong err: %v, expected %v", err, test.expected)
		}
	}
}

func TestGetRows(t *testing.T) {
	type TestData struct {
		name       string
		schemaName string
		table      *schemas.Table
		buf        *bytes.Buffer
		expected   string
		err        error
	}
	testDb := testutil.CreateTestDb(t)
	testutil.InsertTestData(t, testDb)
	repos := repositories.New(testDb)
	ctx := context.Background()
	tests := []TestData{
		{
			name:       "test get users",
			schemaName: "alpha",
			table: &schemas.Table{
				Name: "user_payment_methods",
				Filters: map[string]schemas.Pks{
					"id":      {"1": true, "2": true, "3": true},
					"user_id": {"1": true, "2": true},
				},
			},
			buf: &bytes.Buffer{},
			expected: `-- Data for Name: alpha.user_payment_methods; Type: TABLE DATA;
ALTER TABLE alpha.user_payment_methods DISABLE TRIGGER ALL;
COPY alpha.user_payment_methods ("id", "user_id", "order_id", "payment_type", "card_number", "expiry_date", "is_default", "created_at") FROM stdin;
1	1	1	credit_card	4111111111111111	2025-12-01T00:00:00Z	t	2025-01-01T10:00:00Z
2	1	\N	paypal	\N	\N	f	2025-01-02T11:00:00Z
3	2	3	credit_card	4222222222222222	2024-10-01T00:00:00Z	t	2025-01-03T12:00:00Z
8	1	\N	bank_transfer	\N	\N	f	2025-01-08T17:00:00Z
9	2	\N	credit_card	4555555555555555	2024-12-01T00:00:00Z	f	2025-01-09T18:00:00Z
\.
ALTER TABLE alpha.user_payment_methods ENABLE TRIGGER ALL;


`,
		},
	}
	for _, test := range tests {
		w := bufio.NewWriter(test.buf)
		err := repos.GetRows(ctx, test.schemaName, test.table, w)
		actual := test.buf.String()
		if diff := cmp.Diff(test.expected, actual); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
		if test.err != err {
			t.Errorf("wrong err: %v, expected %v", err, test.expected)
		}
	}
}
