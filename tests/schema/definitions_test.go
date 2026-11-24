package schema_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"demoRBACProject1/internal/config"
)

func TestSchemaDefinitions(t *testing.T) {
	if startupConfigErr != nil {
		t.Fatalf("load startup config: %v", startupConfigErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, err := openConfiguredDB(ctx, startupConfig)
	if err != nil {
		t.Fatalf("open schema definition connection: %v", err)
	}
	defer db.Close()

	if schemaMigrationAuthored() {
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("apply schema migration: %v", err)
		}
	}

	snapshot, err := inspectSchemaDefinition(ctx, db)
	if err != nil {
		t.Fatalf("inspect schema definition: %v", err)
	}

	assertExactBusinessInventory(t, snapshot)
	assertBusinessTablesEmpty(t, snapshot)
	assertColumns(t, snapshot)
	assertPrimaryKeys(t, snapshot)
	assertIndexes(t, snapshot)
	assertForeignKeys(t, snapshot)
	assertChecks(t, snapshot)
	assertCreateTableFragments(t, snapshot)
}

type schemaDefinitionSnapshot struct {
	Tables      map[string]tableDefinition
	Columns     map[string]map[string]columnDefinition
	PrimaryKeys map[string][]string
	Indexes     map[string]map[string]indexDefinition
	ForeignKeys map[string]foreignKeyDefinition
	Checks      map[string]checkDefinition
	CreateSQL   map[string]string
	Rows        map[string]int64
}

type tableDefinition struct {
	Engine    string
	Collation string
}

type columnDefinition struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
	Extra    string
}

type indexDefinition struct {
	Name    string
	Unique  bool
	Columns []string
}

type foreignKeyDefinition struct {
	Name        string
	Table       string
	Columns     []string
	ParentTable string
	ParentCols  []string
	OnUpdate    string
	OnDelete    string
}

type checkDefinition struct {
	Name       string
	Table      string
	Expression string
	Enforced   bool
}

func inspectSchemaDefinition(ctx context.Context, db *sql.DB) (schemaDefinitionSnapshot, error) {
	snapshot := schemaDefinitionSnapshot{
		Tables:      map[string]tableDefinition{},
		Columns:     map[string]map[string]columnDefinition{},
		PrimaryKeys: map[string][]string{},
		Indexes:     map[string]map[string]indexDefinition{},
		ForeignKeys: map[string]foreignKeyDefinition{},
		Checks:      map[string]checkDefinition{},
		CreateSQL:   map[string]string{},
		Rows:        map[string]int64{},
	}

	tableRows, err := db.QueryContext(ctx, `
SELECT TABLE_NAME, ENGINE, TABLE_COLLATION
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_TYPE = 'BASE TABLE'
ORDER BY TABLE_NAME`)
	if err != nil {
		return snapshot, err
	}
	defer tableRows.Close()
	for tableRows.Next() {
		var name, engine, collation string
		if err := tableRows.Scan(&name, &engine, &collation); err != nil {
			return snapshot, err
		}
		snapshot.Tables[name] = tableDefinition{Engine: engine, Collation: collation}
	}
	if err := tableRows.Err(); err != nil {
		return snapshot, err
	}

	columnRows, err := db.QueryContext(ctx, `
SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
ORDER BY TABLE_NAME, ORDINAL_POSITION`)
	if err != nil {
		return snapshot, err
	}
	defer columnRows.Close()
	for columnRows.Next() {
		var table, name, typ, nullable, extra string
		var def sql.NullString
		if err := columnRows.Scan(&table, &name, &typ, &nullable, &def, &extra); err != nil {
			return snapshot, err
		}
		if snapshot.Columns[table] == nil {
			snapshot.Columns[table] = map[string]columnDefinition{}
		}
		snapshot.Columns[table][name] = columnDefinition{
			Name:     name,
			Type:     normalizeColumnType(typ),
			Nullable: nullable == "YES",
			Default:  normalizeDefault(def, nullable == "YES"),
			Extra:    strings.ToLower(extra),
		}
	}
	if err := columnRows.Err(); err != nil {
		return snapshot, err
	}

	keyRows, err := db.QueryContext(ctx, `
SELECT TABLE_NAME, INDEX_NAME, NON_UNIQUE, COLUMN_NAME, SEQ_IN_INDEX
FROM INFORMATION_SCHEMA.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
ORDER BY TABLE_NAME, INDEX_NAME, SEQ_IN_INDEX`)
	if err != nil {
		return snapshot, err
	}
	defer keyRows.Close()
	for keyRows.Next() {
		var table, name, column string
		var nonUnique, seq int
		if err := keyRows.Scan(&table, &name, &nonUnique, &column, &seq); err != nil {
			return snapshot, err
		}
		if snapshot.Indexes[table] == nil {
			snapshot.Indexes[table] = map[string]indexDefinition{}
		}
		index := snapshot.Indexes[table][name]
		index.Name = name
		index.Unique = nonUnique == 0
		index.Columns = append(index.Columns, column)
		snapshot.Indexes[table][name] = index
		if name == "PRIMARY" {
			snapshot.PrimaryKeys[table] = append(snapshot.PrimaryKeys[table], column)
		}
		_ = seq
	}
	if err := keyRows.Err(); err != nil {
		return snapshot, err
	}

	fkRows, err := db.QueryContext(ctx, `
SELECT
	k.CONSTRAINT_NAME, k.TABLE_NAME, k.COLUMN_NAME, k.REFERENCED_TABLE_NAME,
	k.REFERENCED_COLUMN_NAME, r.UPDATE_RULE, r.DELETE_RULE, k.POSITION_IN_UNIQUE_CONSTRAINT
FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE k
JOIN INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS r
  ON r.CONSTRAINT_SCHEMA = k.CONSTRAINT_SCHEMA
 AND r.CONSTRAINT_NAME = k.CONSTRAINT_NAME
WHERE k.CONSTRAINT_SCHEMA = DATABASE()
  AND k.REFERENCED_TABLE_NAME IS NOT NULL
ORDER BY k.CONSTRAINT_NAME, k.POSITION_IN_UNIQUE_CONSTRAINT`)
	if err != nil {
		return snapshot, err
	}
	defer fkRows.Close()
	for fkRows.Next() {
		var name, table, column, parent, parentColumn, updateRule, deleteRule string
		var position int
		if err := fkRows.Scan(&name, &table, &column, &parent, &parentColumn, &updateRule, &deleteRule, &position); err != nil {
			return snapshot, err
		}
		fk := snapshot.ForeignKeys[name]
		fk.Name = name
		fk.Table = table
		fk.Columns = append(fk.Columns, column)
		fk.ParentTable = parent
		fk.ParentCols = append(fk.ParentCols, parentColumn)
		fk.OnUpdate = normalizeRule(updateRule)
		fk.OnDelete = normalizeRule(deleteRule)
		snapshot.ForeignKeys[name] = fk
		_ = position
	}
	if err := fkRows.Err(); err != nil {
		return snapshot, err
	}

	checkRows, err := db.QueryContext(ctx, `
SELECT tc.CONSTRAINT_NAME, tc.TABLE_NAME, cc.CHECK_CLAUSE, tc.ENFORCED
FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
JOIN INFORMATION_SCHEMA.CHECK_CONSTRAINTS cc
  ON cc.CONSTRAINT_SCHEMA = tc.CONSTRAINT_SCHEMA
 AND cc.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
WHERE tc.CONSTRAINT_SCHEMA = DATABASE()
  AND tc.CONSTRAINT_TYPE = 'CHECK'
ORDER BY tc.CONSTRAINT_NAME`)
	if err != nil {
		return snapshot, err
	}
	defer checkRows.Close()
	for checkRows.Next() {
		var name, table, clause, enforced string
		if err := checkRows.Scan(&name, &table, &clause, &enforced); err != nil {
			return snapshot, err
		}
		snapshot.Checks[name] = checkDefinition{
			Name:       name,
			Table:      table,
			Expression: normalizeExpression(clause),
			Enforced:   enforced == "YES",
		}
	}
	if err := checkRows.Err(); err != nil {
		return snapshot, err
	}

	for table := range snapshot.Tables {
		var ignored, createSQL string
		if err := db.QueryRowContext(ctx, "SHOW CREATE TABLE `"+escapeIdentifier(table)+"`").Scan(&ignored, &createSQL); err != nil {
			return snapshot, fmt.Errorf("show create table %s: %w", table, err)
		}
		snapshot.CreateSQL[table] = strings.ToLower(createSQL)
		if isBusinessTable(table) {
			var count int64
			if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM `"+escapeIdentifier(table)+"`").Scan(&count); err != nil {
				return snapshot, err
			}
			snapshot.Rows[table] = count
		}
	}

	return snapshot, nil
}

func assertExactBusinessInventory(t *testing.T, snapshot schemaDefinitionSnapshot) {
	t.Helper()
	wantBusiness := []string{"orders", "product", "transactions", "user", "user_demand", "usercart"}
	if got := presentBusinessTables(snapshot); !reflect.DeepEqual(got, wantBusiness) {
		t.Fatalf("business tables = %v, want exactly %v", got, wantBusiness)
	}

	for _, obsolete := range []string{"books", "order", "order_header", "order_item", "users"} {
		if _, ok := snapshot.Tables[obsolete]; ok {
			t.Fatalf("obsolete or extra business object %q is present", obsolete)
		}
	}
}

func assertBusinessTablesEmpty(t *testing.T, snapshot schemaDefinitionSnapshot) {
	t.Helper()
	for _, table := range []string{"product", "user", "orders", "usercart", "user_demand", "transactions"} {
		if got := snapshot.Rows[table]; got != 0 {
			t.Fatalf("%s row count = %d, want 0", table, got)
		}
	}
}

func assertColumns(t *testing.T, snapshot schemaDefinitionSnapshot) {
	t.Helper()
	expected := map[string][]columnDefinition{
		"product": {
			{Name: "pid", Type: "varchar(30)", Nullable: false, Default: ""},
			{Name: "pname", Type: "varchar(100)", Nullable: true, Default: "null"},
			{Name: "ptype", Type: "varchar(20)", Nullable: true, Default: "null"},
			{Name: "pinfo", Type: "varchar(350)", Nullable: true, Default: "null"},
			{Name: "pprice", Type: "decimal(10,2)", Nullable: false, Default: "0.00"},
			{Name: "pquantity", Type: "int", Nullable: true, Default: "null"},
			{Name: "image", Type: "longblob", Nullable: true, Default: "null"},
		},
		"user": {
			{Name: "email", Type: "varchar(100)", Nullable: false, Default: ""},
			{Name: "name", Type: "varchar(30)", Nullable: true, Default: "null"},
			{Name: "mobile", Type: "bigint", Nullable: true, Default: "null"},
			{Name: "address", Type: "varchar(250)", Nullable: true, Default: "null"},
			{Name: "pincode", Type: "int", Nullable: true, Default: "null"},
			{Name: "password_hash", Type: "varchar(255)", Nullable: true, Default: "null"},
			{Name: "must_reset_password", Type: "tinyint(1)", Nullable: false, Default: "1"},
		},
		"usercart": {
			{Name: "id", Type: "bigint", Nullable: false, Default: "", Extra: "auto_increment"},
			{Name: "email", Type: "varchar(100)", Nullable: false, Default: ""},
			{Name: "pid", Type: "varchar(30)", Nullable: false, Default: ""},
			{Name: "quantity", Type: "int", Nullable: false, Default: "1"},
			{Name: "added_at", Type: "datetime", Nullable: false, Default: "current_timestamp"},
		},
		"orders": {
			{Name: "orderid", Type: "varchar(45)", Nullable: false, Default: ""},
			{Name: "prodid", Type: "varchar(45)", Nullable: false, Default: ""},
			{Name: "quantity", Type: "int", Nullable: true, Default: "null"},
			{Name: "amount", Type: "decimal(10,2)", Nullable: false, Default: "0.00"},
			{Name: "shipped", Type: "int", Nullable: false, Default: "0"},
		},
		"transactions": {
			{Name: "id", Type: "bigint", Nullable: false, Default: "", Extra: "auto_increment"},
			{Name: "transaction_id", Type: "varchar(40)", Nullable: false, Default: ""},
			{Name: "email", Type: "varchar(100)", Nullable: false, Default: ""},
			{Name: "pid", Type: "varchar(30)", Nullable: false, Default: ""},
			{Name: "quantity", Type: "int", Nullable: false, Default: ""},
			{Name: "amount", Type: "decimal(10,2)", Nullable: false, Default: "0.00"},
			{Name: "created_at", Type: "datetime", Nullable: false, Default: "current_timestamp"},
		},
		"user_demand": {
			{Name: "id", Type: "bigint", Nullable: false, Default: "", Extra: "auto_increment"},
			{Name: "email", Type: "varchar(100)", Nullable: false, Default: ""},
			{Name: "book_name", Type: "varchar(100)", Nullable: false, Default: ""},
			{Name: "author", Type: "varchar(100)", Nullable: true, Default: "null"},
			{Name: "requested_at", Type: "datetime", Nullable: false, Default: "current_timestamp"},
		},
	}

	for table, wantColumns := range expected {
		gotColumns := snapshot.Columns[table]
		if len(gotColumns) != len(wantColumns) {
			t.Fatalf("%s column count = %d, want %d: %+v", table, len(gotColumns), len(wantColumns), gotColumns)
		}
		for _, want := range wantColumns {
			got, ok := gotColumns[want.Name]
			if !ok {
				t.Fatalf("%s.%s missing", table, want.Name)
			}
			if got.Type != want.Type || got.Nullable != want.Nullable || got.Default != want.Default {
				t.Fatalf("%s.%s = type %q nullable %t default %q, want type %q nullable %t default %q",
					table, want.Name, got.Type, got.Nullable, got.Default, want.Type, want.Nullable, want.Default)
			}
			if want.Extra != "" && !strings.Contains(got.Extra, want.Extra) {
				t.Fatalf("%s.%s extra = %q, want to contain %q", table, want.Name, got.Extra, want.Extra)
			}
		}
	}
}

func assertPrimaryKeys(t *testing.T, snapshot schemaDefinitionSnapshot) {
	t.Helper()
	expected := map[string][]string{
		"product":      {"pid"},
		"user":         {"email"},
		"usercart":     {"id"},
		"orders":       {"orderid", "prodid"},
		"transactions": {"id"},
		"user_demand":  {"id"},
	}
	for table, want := range expected {
		if got := snapshot.PrimaryKeys[table]; !reflect.DeepEqual(got, want) {
			t.Fatalf("%s primary key = %v, want %v", table, got, want)
		}
	}
}

func assertIndexes(t *testing.T, snapshot schemaDefinitionSnapshot) {
	t.Helper()
	expected := map[string]indexDefinition{
		"usercart.uq_usercart_email_pid":                 {Unique: true, Columns: []string{"email", "pid"}},
		"transactions.uq_transactions_transaction_id":    {Unique: true, Columns: []string{"transaction_id"}},
		"orders.productid":                              {Unique: false, Columns: []string{"prodid"}},
		"usercart.fk_usercart_pid":                      {Unique: false, Columns: []string{"pid"}},
		"user_demand.fk_user_demand_email":              {Unique: false, Columns: []string{"email"}},
		"transactions.fk_transactions_email":            {Unique: false, Columns: []string{"email"}},
		"transactions.fk_transactions_pid":              {Unique: false, Columns: []string{"pid"}},
	}
	for key, want := range expected {
		table, name, ok := strings.Cut(key, ".")
		if !ok {
			t.Fatalf("bad expected index key %q", key)
		}
		got, ok := snapshot.Indexes[table][name]
		if !ok {
			t.Fatalf("index %s.%s missing; got indexes %+v", table, name, snapshot.Indexes[table])
		}
		if got.Unique != want.Unique || !reflect.DeepEqual(got.Columns, want.Columns) {
			t.Fatalf("index %s.%s = unique %t columns %v, want unique %t columns %v",
				table, name, got.Unique, got.Columns, want.Unique, want.Columns)
		}
	}
}

func assertForeignKeys(t *testing.T, snapshot schemaDefinitionSnapshot) {
	t.Helper()
	expected := map[string]foreignKeyDefinition{
		"productid":              {Table: "orders", Columns: []string{"prodid"}, ParentTable: "product", ParentCols: []string{"pid"}, OnUpdate: "NO ACTION", OnDelete: "NO ACTION"},
		"fk_usercart_email":      {Table: "usercart", Columns: []string{"email"}, ParentTable: "user", ParentCols: []string{"email"}, OnUpdate: "CASCADE", OnDelete: "CASCADE"},
		"fk_usercart_pid":        {Table: "usercart", Columns: []string{"pid"}, ParentTable: "product", ParentCols: []string{"pid"}, OnUpdate: "CASCADE", OnDelete: "RESTRICT"},
		"fk_user_demand_email":   {Table: "user_demand", Columns: []string{"email"}, ParentTable: "user", ParentCols: []string{"email"}, OnUpdate: "CASCADE", OnDelete: "CASCADE"},
		"fk_transactions_email":  {Table: "transactions", Columns: []string{"email"}, ParentTable: "user", ParentCols: []string{"email"}, OnUpdate: "CASCADE", OnDelete: "RESTRICT"},
		"fk_transactions_pid":    {Table: "transactions", Columns: []string{"pid"}, ParentTable: "product", ParentCols: []string{"pid"}, OnUpdate: "CASCADE", OnDelete: "RESTRICT"},
	}
	if len(snapshot.ForeignKeys) != len(expected) {
		t.Fatalf("foreign key count = %d, want %d: %+v", len(snapshot.ForeignKeys), len(expected), snapshot.ForeignKeys)
	}
	for name, want := range expected {
		got, ok := snapshot.ForeignKeys[name]
		if !ok {
			t.Fatalf("foreign key %s missing", name)
		}
		if got.Table != want.Table || !reflect.DeepEqual(got.Columns, want.Columns) ||
			got.ParentTable != want.ParentTable || !reflect.DeepEqual(got.ParentCols, want.ParentCols) ||
			got.OnUpdate != want.OnUpdate || got.OnDelete != want.OnDelete {
			t.Fatalf("foreign key %s = %+v, want %+v", name, got, want)
		}
	}
}

func assertChecks(t *testing.T, snapshot schemaDefinitionSnapshot) {
	t.Helper()
	expected := map[string]checkDefinition{
		"ck_usercart_quantity":     {Table: "usercart", Expression: "`quantity`>=1", Enforced: true},
		"ck_transactions_quantity": {Table: "transactions", Expression: "`quantity`>=1", Enforced: true},
	}
	for name, want := range expected {
		got, ok := snapshot.Checks[name]
		if !ok {
			t.Fatalf("check %s missing; got %+v", name, snapshot.Checks)
		}
		if got.Table != want.Table || !got.Enforced || got.Expression != want.Expression {
			t.Fatalf("check %s = %+v, want %+v", name, got, want)
		}
	}
}

func assertCreateTableFragments(t *testing.T, snapshot schemaDefinitionSnapshot) {
	t.Helper()
	for _, table := range []string{"product", "user", "orders", "usercart", "user_demand", "transactions"} {
		definition, ok := snapshot.Tables[table]
		if !ok {
			t.Fatalf("table %s missing", table)
		}
		if definition.Engine != "InnoDB" {
			t.Fatalf("%s engine = %q, want InnoDB", table, definition.Engine)
		}
		if definition.Collation != config.RequiredCollation {
			t.Fatalf("%s collation = %q, want %q", table, definition.Collation, config.RequiredCollation)
		}
		createSQL := snapshot.CreateSQL[table]
		for _, fragment := range []string{"engine=innodb", "default charset=utf8mb4", "collate=utf8mb4_0900_bin"} {
			if !strings.Contains(createSQL, fragment) {
				t.Fatalf("SHOW CREATE TABLE %s missing %q in:\n%s", table, fragment, createSQL)
			}
		}
	}
}

func presentBusinessTables(snapshot schemaDefinitionSnapshot) []string {
	var tables []string
	for table := range snapshot.Tables {
		if isBusinessTable(table) {
			tables = append(tables, table)
		}
	}
	sort.Strings(tables)
	return tables
}

func isBusinessTable(table string) bool {
	switch table {
	case "product", "user", "orders", "usercart", "user_demand", "transactions":
		return true
	default:
		return false
	}
}

func schemaMigrationAuthored() bool {
	_, err := os.Stat(schemaTrack.SourcePath + "/000001_complete_mysql_schema.up.sql")
	if err == nil {
		return true
	}
	return !errors.Is(err, os.ErrNotExist)
}

func normalizeColumnType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " int(", " int(")
	value = strings.TrimSuffix(value, " signed")
	if value == "tinyint(1)" {
		return value
	}
	return value
}

func normalizeDefault(value sql.NullString, nullable bool) string {
	if !value.Valid {
		if nullable {
			return "null"
		}
		return ""
	}
	normalized := strings.ToLower(strings.TrimSpace(value.String))
	normalized = strings.Trim(normalized, "()")
	normalized = strings.Trim(normalized, "'")
	if normalized == "current_timestamp()" {
		return "current_timestamp"
	}
	if normalized == "b'1'" || normalized == "true" {
		return "1"
	}
	if normalized == "b'0'" || normalized == "false" {
		return "0"
	}
	return normalized
}

func normalizeRule(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "NO ACTION" || value == "RESTRICT" || value == "CASCADE" {
		return value
	}
	return value
}

func normalizeExpression(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "(", "")
	value = strings.ReplaceAll(value, ")", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func escapeIdentifier(value string) string {
	return strings.ReplaceAll(value, "`", "``")
}
