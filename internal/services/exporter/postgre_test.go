package exporter

import (
	"context"
	"os"
	"testing"

	"github.com/t1m4/db_part_dump/config"
	"github.com/t1m4/db_part_dump/internal/constants"
	"github.com/t1m4/db_part_dump/internal/repositories"
	"github.com/t1m4/db_part_dump/internal/schemas"
	"github.com/t1m4/db_part_dump/internal/testutil"

	"github.com/google/go-cmp/cmp"
)

func ReadFile(t *testing.T, filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read file err %s", err)
	}
	return string(content)
}

var expected = `-- Data for Name: alpha.users; Type: TABLE DATA;
ALTER TABLE alpha.users DISABLE TRIGGER ALL;
COPY alpha.users ("id", "username", "email", "created_at", "status") FROM stdin;
1	john_doe	john@example.com	2025-01-01T10:00:00.928501Z	active
2	jane_smith	jane@example.com	2025-01-02T10:00:00.928502Z	active
\.
ALTER TABLE alpha.users ENABLE TRIGGER ALL;


-- Data for Name: alpha.orders; Type: TABLE DATA;
ALTER TABLE alpha.orders DISABLE TRIGGER ALL;
COPY alpha.orders ("id", "user_id", "order_date", "total_amount", "status") FROM stdin;
1	1	2025-01-01T10:00:00Z	99.99	completed
3	2	2025-01-01T10:00:00Z	199.99	completed
\.
ALTER TABLE alpha.orders ENABLE TRIGGER ALL;


-- Data for Name: alpha.user_payment_methods; Type: TABLE DATA;
ALTER TABLE alpha.user_payment_methods DISABLE TRIGGER ALL;
COPY alpha.user_payment_methods ("id", "user_id", "order_id", "payment_type", "card_number", "expiry_date", "is_default", "created_at") FROM stdin;
1	1	1	credit_card	4111111111111111	2025-12-01T00:00:00Z	t	2025-01-01T10:00:00Z
2	1	\N	paypal	\N	\N	f	2025-01-02T11:00:00Z
3	2	3	credit_card	4222222222222222	2024-10-01T00:00:00Z	t	2025-01-03T12:00:00Z
\.
ALTER TABLE alpha.user_payment_methods ENABLE TRIGGER ALL;


`

func TestPostgresqlExporter(t *testing.T) {
	userTable := &schemas.Table{
		Name:    "users",
		Filters: map[string]schemas.Pks{"id": {"1": true, "2": true}},
		Fks:     map[string]*schemas.Table{},
	}
	ordersTable := &schemas.Table{
		Name:    "orders",
		Filters: map[string]schemas.Pks{"id": {"1": true, "3": true}},
		Fks:     map[string]*schemas.Table{userTable.Name: userTable},
	}
	userPaymentMethodsTable := &schemas.Table{
		Name:    "user_payment_methods",
		Filters: map[string]schemas.Pks{"id": {"1": true, "2": true, "3": true}},
		Fks: map[string]*schemas.Table{
			userTable.Name: userTable, ordersTable.Name: ordersTable,
		},
	}
	c := &config.Config{
		Settings: config.Settings{
			Output:     "test_postgresql.sql",
			SchemaName: "alpha",
			Tables: []config.Table{
				{
					Name:    "user_payment_methods",
					Filters: []config.Filter{{Name: "id", Value: "1, 2, 3"}},
				},
			},
			Direction: constants.OUTGOING,
		},
	}
	db := testutil.CreateTestDb(t)
	testutil.InsertTestData(t, db)
	repos := repositories.New(db)
	ctx := context.Background()

	exporter := PostgresqlExporter{c, repos}
	tablePks := []*schemas.Table{userTable, ordersTable, userPaymentMethodsTable}
	err := exporter.ExportToFile(ctx, tablePks)
	if err != nil {
		t.Errorf("wrong err: %v, expected %v", err, nil)
	}
	actual := ReadFile(t, c.Settings.Output)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	_ = os.Remove(c.Settings.Output)
}
