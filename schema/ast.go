package schema

type DDL interface {
	Statement() string
}

type CreateTable struct {
	statement string
	table     Table
}

type CreateIndex struct {
	statement string
	tableName string
	index     Index
}

type AddIndex struct {
	statement string
	tableName string
	index     Index
}

type AddPrimaryKey struct {
	statement string
	tableName string
	index     Index
}

type AddForeignKey struct {
	statement  string
	tableName  string
	foreignKey ForeignKey
}

type AddPolicy struct {
	statement string
	tableName string
	policy    Policy
}

type Table struct {
	name        string
	columns     []Column
	indexes     []Index
	foreignKeys []ForeignKey
	policies    []Policy
	// XXX: have options and alter on its change?
}

type Column struct {
	name           string
	position       int
	typeName       string
	unsigned       bool
	notNull        *bool
	autoIncrement  bool
	array          bool
	defaultDef     *DefaultDefinition
	length         *Value
	scale          *Value
	check          *CheckDefinition
	checkNoInherit bool
	charset        string
	collate        string
	timezone       bool // for Postgres `with time zone`
	keyOption      ColumnKeyOption
	onUpdate       *Value
	enumValues     []string
	references     string
	identity       string
	sequence       *Sequence
	// TODO: keyopt
	// XXX: zerofill?
}

type Index struct {
	name      string
	indexType string // Parsed only in "create table" but not parsed in "add index". Only used inside `generateDDLsForCreateTable`.
	columns   []IndexColumn
	primary   bool
	unique    bool
	where     string // for Postgres `Partial Indexes`
	clustered bool   // for MSSQL
	options   []IndexOption
}

type IndexColumn struct {
	column string
	length *int
}

type IndexOption struct {
	optionName string
	value      *Value
}

type ForeignKey struct {
	constraintName   string
	indexName        string
	indexColumns     []string
	referenceName    string
	referenceColumns []string
	onDelete         string
	onUpdate         string
}

type Policy struct {
	name          string
	referenceName string
	permissive    string
	scope         string
	roles         []string
	using         string
	withCheck     string
}

type View struct {
	statement  string
	name       string
	definition string
}

type Value struct {
	valueType ValueType
	raw       []byte

	// ValueType-specific. Should be union?
	strVal   string  // ValueTypeStr
	intVal   int     // ValueTypeInt
	floatVal float64 // ValueTypeFloat
	bitVal   bool    // ValueTypeBit
}

type ValueType int

const (
	ValueTypeStr = ValueType(iota)
	ValueTypeInt
	ValueTypeFloat
	ValueTypeHexNum
	ValueTypeHex
	ValueTypeValArg
	ValueTypeBit
	ValueTypeBool
)

type ColumnKeyOption int

const (
	ColumnKeyNone = ColumnKeyOption(iota)
	ColumnKeyPrimary
	ColumnKeySpatialKey
	ColumnKeyUnique
	ColumnKeyUniqueKey
	ColumnKey
)

type Sequence struct {
	Name        string
	IfNotExists bool
	Type        string
	IncrementBy *int
	MinValue    *int
	NoMinValue  bool
	MaxValue    *int
	NoMaxValue  bool
	StartWith   *int
	Cache       *int
	Cycle       bool
	NoCycle     bool
	OwnedBy     string
}

type DefaultDefinition struct {
	value          *Value
	constraintName string // only for MSSQL
}

type CheckDefinition struct {
	definition     string
	constraintName string
}

func (c *CreateTable) Statement() string {
	return c.statement
}

func (c *CreateIndex) Statement() string {
	return c.statement
}

func (a *AddIndex) Statement() string {
	return a.statement
}

func (a *AddPrimaryKey) Statement() string {
	return a.statement
}

func (a *AddForeignKey) Statement() string {
	return a.statement
}

func (a *AddPolicy) Statement() string {
	return a.statement
}

func (v *View) Statement() string {
	return v.statement
}

func (t *Table) PrimaryKey() *Index {
	for _, index := range t.indexes {
		if index.primary {
			return &index
		}
	}

	primaryColumns := []IndexColumn{}
	for _, column := range t.columns {
		if column.keyOption == ColumnKeyPrimary {
			primaryColumns = append(primaryColumns, IndexColumn{
				column: column.name,
			})
		}
	}

	if len(primaryColumns) == 0 {
		return nil
	}

	return &Index{
		name:      "PRIMARY",
		indexType: "primary key",
		columns:   primaryColumns,
		primary:   true,
		unique:    true,
		clustered: true,
	}
}

func (keyOption ColumnKeyOption) isUnique() bool {
	return keyOption == ColumnKeyUnique || keyOption == ColumnKeyUniqueKey
}
