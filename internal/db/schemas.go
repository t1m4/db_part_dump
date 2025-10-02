package db

type Fk struct {
	ColumnName         string
	ForeignTableSchema string
	ForeignTableName   string
	ForeignColumnName  string
	Direction          string
}
