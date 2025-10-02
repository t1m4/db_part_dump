package dump

import (
	"testing"

	"github.com/t1m4/db_part_dump/config"
	"github.com/t1m4/db_part_dump/internal/schemas"

	"github.com/google/go-cmp/cmp"
)

func TestDfsSort(t *testing.T) {
	type TestData struct {
		name            string
		tablePksByTable tablePksByTableT
		configTables    []config.Table
		expected        []*schemas.Table
	}
	userTable := &schemas.Table{Name: "users"}
	ordersTable := &schemas.Table{Name: "orders", Fks: map[string]*schemas.Table{userTable.Name: userTable}}
	userPaymentMethodsTable := &schemas.Table{
		Name: "user_payment_methods",
		Fks: map[string]*schemas.Table{
			userTable.Name: userTable, ordersTable.Name: ordersTable,
		},
	}
	tests := []TestData{
		{
			name: "test dfs",
			tablePksByTable: tablePksByTableT{
				userTable.Name:               userTable,
				ordersTable.Name:             ordersTable,
				userPaymentMethodsTable.Name: userPaymentMethodsTable,
			},
			configTables: []config.Table{{Name: userPaymentMethodsTable.Name}},
			expected:     []*schemas.Table{userTable, ordersTable, userPaymentMethodsTable},
		},
	}
	for _, test := range tests {
		actual := dfsSort(test.tablePksByTable, test.configTables)
		if diff := cmp.Diff(test.expected, actual); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	}
}
