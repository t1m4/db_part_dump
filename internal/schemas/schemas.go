package schemas

type FkIds map[any]bool

type Filter struct {
	Name   string
	Values map[string]bool
}
type Pks map[string]bool

type Table struct {
	Name    string
	Filters map[string]Pks
	Fks     map[string]*Table
}
