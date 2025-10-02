package dump

import (
	"context"
	"testing"

	"github.com/t1m4/db_part_dump/config"
	"github.com/t1m4/db_part_dump/internal/constants"
	"github.com/t1m4/db_part_dump/internal/repositories"
	"github.com/t1m4/db_part_dump/internal/schemas"
	"github.com/t1m4/db_part_dump/internal/testutil"

	"github.com/google/go-cmp/cmp"
)

var c = &config.Config{
	Settings: config.Settings{
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

var userTable = &schemas.Table{
	Name:    "users",
	Filters: map[string]schemas.Pks{"id": {"1": true, "2": true}},
	Fks:     map[string]*schemas.Table{},
}
var ordersTable = &schemas.Table{
	Name:    "orders",
	Filters: map[string]schemas.Pks{"id": {"1": true, "3": true}},
	Fks:     map[string]*schemas.Table{userTable.Name: userTable},
}
var userPaymentMethodsTable = &schemas.Table{
	Name:    "user_payment_methods",
	Filters: map[string]schemas.Pks{"id": {"1": true, "2": true, "3": true}},
	Fks: map[string]*schemas.Table{
		userTable.Name: userTable, ordersTable.Name: ordersTable,
	},
}

func TestCollectTableFkIdsWithOutgoing(t *testing.T) {
	db := testutil.CreateTestDb(t)
	testutil.InsertTestData(t, db)
	repos := repositories.New(db)
	ctx := context.Background()

	expected := tablePksByTableT{
		userTable.Name:               userTable,
		ordersTable.Name:             ordersTable,
		userPaymentMethodsTable.Name: userPaymentMethodsTable,
	}
	dumpService := New(c, repos)
	actual, err := dumpService.collectTableFkIds(ctx)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	if err != nil {
		t.Errorf("wrong err: %v, expected %v", err, nil)
	}
}

func TestCollectTableFkIdsWithIncludeIncomingTables(t *testing.T) {
	db := testutil.CreateTestDb(t)
	testutil.InsertTestData(t, db)
	repos := repositories.New(db)
	ctx := context.Background()

	ordersTable := &schemas.Table{
		Name: ordersTable.Name,
		Filters: map[string]schemas.Pks{
			"id":      {"1": true, "3": true},
			"user_id": {"1": true, "2": true},
		},
		Fks: ordersTable.Fks,
	}
	userPaymentMethodsTable := &schemas.Table{
		Name: "user_payment_methods",
		Filters: map[string]schemas.Pks{
			"id":      {"1": true, "2": true, "3": true},
			"user_id": {"1": true, "2": true},
		},
		Fks: map[string]*schemas.Table{
			userTable.Name: userTable, ordersTable.Name: ordersTable,
		},
	}
	userAddressesTable := &schemas.Table{
		Name:    "user_addresses",
		Filters: map[string]schemas.Pks{"user_id": {"1": true, "2": true}},
		Fks:     map[string]*schemas.Table{userTable.Name: userTable},
	}
	userPreferencesTable := &schemas.Table{
		Name:    "user_preferences",
		Filters: map[string]schemas.Pks{"user_id": {"1": true, "2": true}},
		Fks:     map[string]*schemas.Table{userTable.Name: userTable},
	}

	testC := *c
	testC.Settings.IncludeIncomingTables = []string{userTable.Name}
	expected := tablePksByTableT{
		userTable.Name:               userTable,
		ordersTable.Name:             ordersTable,
		userPaymentMethodsTable.Name: userPaymentMethodsTable,
		userAddressesTable.Name:      userAddressesTable,
		userPreferencesTable.Name:    userPreferencesTable,
	}
	dumpService := New(&testC, repos)
	actual, err := dumpService.collectTableFkIds(ctx)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	if err != nil {
		t.Errorf("wrong err: %v, expected %v", err, nil)
	}
}

func TestCollectTableFkIdsCycles(t *testing.T) {
	db := testutil.CreateTestDb(t)
	testutil.InsertTestData(t, db)
	repos := repositories.New(db)
	ctx := context.Background()
	c := &config.Config{
		Settings: config.Settings{
			SchemaName: "alpha",
			Tables: []config.Table{
				{
					Name:    "table_one",
					Filters: []config.Filter{{Name: "id", Value: "'11111111-1111-1111-1111-111111111111'"}},
				},
			},
			Direction: constants.OUTGOING,
		},
	}

	tableOne := &schemas.Table{
		Name:    "table_one",
		Filters: map[string]schemas.Pks{"id": {"'11111111-1111-1111-1111-111111111111'": true}},
		Fks:     make(map[string]*schemas.Table),
	}
	tableTwo := &schemas.Table{
		Name:    "table_two",
		Filters: map[string]schemas.Pks{"id": {"'22222222-2222-2222-2222-222222222222'": true}},
		Fks:     make(map[string]*schemas.Table),
	}
	tableThree := &schemas.Table{
		Name:    "table_three",
		Filters: map[string]schemas.Pks{"id": {"'33333333-3333-3333-3333-333333333333'": true}},
		Fks:     make(map[string]*schemas.Table),
	}
	tableOne.Fks[tableTwo.Name] = tableTwo
	tableTwo.Fks[tableThree.Name] = tableThree
	tableThree.Fks[tableOne.Name] = tableOne
	expected := tablePksByTableT{
		tableOne.Name:   tableOne,
		tableTwo.Name:   tableTwo,
		tableThree.Name: tableThree,
	}
	dumpService := New(c, repos)
	actual, err := dumpService.collectTableFkIds(ctx)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	if err != nil {
		t.Errorf("wrong err: %v, expected %v", err, nil)
	}
}
