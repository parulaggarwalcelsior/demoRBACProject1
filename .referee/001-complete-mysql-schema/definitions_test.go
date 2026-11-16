package schema_test

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Field contracts are copied only from spec.md's confirmed field register.
// The tuple is MySQL data type, nullable, default, and character length.
type field struct{ kind, nullable, def, length string }

var columns = map[string]map[string]field{
	"product": {
		"pid":       {"varchar", "NO", "<NULL>", "30"},
		"pname":     {"varchar", "YES", "<NULL>", "100"},
		"ptype":     {"varchar", "YES", "<NULL>", "20"},
		"pinfo":     {"varchar", "YES", "<NULL>", "350"},
		"pprice":    {"decimal", "NO", "0.00", "<NULL>"},
		"pquantity": {"int", "YES", "<NULL>", "<NULL>"},
		"image":     {"longblob", "YES", "<NULL>", "4294967295"},
	},
	"user": {
		"email":               {"varchar", "NO", "<NULL>", "100"},
		"name":                {"varchar", "YES", "<NULL>", "30"},
		"mobile":              {"bigint", "YES", "<NULL>", "<NULL>"},
		"address":             {"varchar", "YES", "<NULL>", "250"},
		"pincode":             {"int", "YES", "<NULL>", "<NULL>"},
		"password_hash":       {"varchar", "YES", "<NULL>", "255"},
		"must_reset_password": {"tinyint", "NO", "1", "<NULL>"},
	},
	"orders": {
		"orderid":  {"varchar", "NO", "<NULL>", "45"},
		"prodid":   {"varchar", "NO", "<NULL>", "45"},
		"quantity": {"int", "YES", "<NULL>", "<NULL>"},
		"amount":   {"decimal", "NO", "0.00", "<NULL>"},
		"shipped":  {"int", "NO", "0", "<NULL>"},
	},
	"usercart": {
		"id":       {"bigint", "NO", "<NULL>", "<NULL>"},
		"email":    {"varchar", "NO", "<NULL>", "100"},
		"pid":      {"varchar", "NO", "<NULL>", "30"},
		"quantity": {"int", "NO", "1", "<NULL>"},
		"added_at": {"datetime", "NO", "CURRENT_TIMESTAMP", "<NULL>"},
	},
	"user_demand": {
		"id":           {"bigint", "NO", "<NULL>", "<NULL>"},
		"email":        {"varchar", "NO", "<NULL>", "100"},
		"book_name":    {"varchar", "NO", "<NULL>", "100"},
		"author":       {"varchar", "YES", "<NULL>", "100"},
		"requested_at": {"datetime", "NO", "CURRENT_TIMESTAMP", "<NULL>"},
	},
	"transactions": {
		"id":             {"bigint", "NO", "<NULL>", "<NULL>"},
		"transaction_id": {"varchar", "NO", "<NULL>", "40"},
		"email":          {"varchar", "NO", "<NULL>", "100"},
		"pid":            {"varchar", "NO", "<NULL>", "30"},
		"quantity":       {"int", "NO", "<NULL>", "<NULL>"},
		"amount":         {"decimal", "NO", "0.00", "<NULL>"},
		"created_at":     {"datetime", "NO", "CURRENT_TIMESTAMP", "<NULL>"},
	},
}

func TestDeliveryArtifacts(t *testing.T) {
	root := targetRoot(t)
	paths := []string{
		"migrations/schema/000001_complete_mysql_schema.up.sql",
		"migrations/schema/000001_complete_mysql_schema.down.sql",
		"migrations/credentials/000001_replace_legacy_password.up.sql",
		"migrations/credentials/000001_replace_legacy_password.down.sql",
		"migrations/credentials-finalize/000001_require_password_hash.up.sql",
		"migrations/credentials-finalize/000001_require_password_hash.down.sql",
		"internal/config/config.go", "internal/config/config_test.go",
		"tests/schema/environment_test.go", "tests/schema/definitions_test.go",
		"tests/schema/values_test.go", "tests/schema/references_test.go",
		"tests/schema/lifecycle_test.go", "tests/schema/credentials_test.go",
		"tests/schema/testdata/legacy/000001_legacy_user.up.sql",
		"tests/schema/testdata/legacy/000001_legacy_user.down.sql",
		"tests/schema/testdata/sentinel/000001_unrelated_object.up.sql",
		"tests/schema/testdata/sentinel/000001_unrelated_object.down.sql",
		"tests/schema/testdata/sentinel/000002_external_reference.up.sql",
		"tests/schema/testdata/sentinel/000002_external_reference.down.sql",
		"go.mod", "go.sum", "compose.schema.yaml", ".golangci.yml", "Makefile",
		".github/workflows/schema.yml", "CHANGELOG.md",
		"specs/001-complete-mysql-schema/validation.md",
		"specs/001-complete-mysql-schema/quickstart.md",
	}
	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			info, err := os.Stat(filepath.Join(root, p))
			if err != nil || info.IsDir() || info.Size() == 0 {
				t.Errorf("required nonempty delivery artifact missing: %s", p)
			}
		})
	}
}

func TestSchemaDefinitions(t *testing.T) {
	f := schemaFixture(t)
	wantedTables := []string{"orders", "product", "transactions", "user", "user_demand", "usercart", historyName("schema")}
	sort.Strings(wantedTables)
	var actual []string
	for _, row := range f.rows("SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() ORDER BY table_name") {
		actual = append(actual, row[0])
	}
	if !reflect.DeepEqual(actual, wantedTables) {
		t.Fatal("schema must create exactly six business tables plus its migration history")
	}
	for table, expected := range columns {
		t.Run(table, func(t *testing.T) {
			f.t = t
			f.wantRows("SELECT engine, table_collation FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?", [][]string{{"InnoDB", "utf8mb4_0900_bin"}}, table)
			f.scalar("SELECT COUNT(*) FROM `"+table+"`", "0")
			rows := f.rows("SELECT column_name, data_type, is_nullable, column_default, character_maximum_length, column_type, extra, numeric_precision, numeric_scale, collation_name FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=?", table)
			if len(rows) != len(expected) {
				t.Fatalf("%s: got %d columns, want %d", table, len(rows), len(expected))
			}
			for _, r := range rows {
				want, ok := expected[r[0]]
				if !ok {
					t.Fatalf("unexpected column %s.%s", table, r[0])
				}
				def := r[3]
				if strings.EqualFold(strings.TrimSuffix(def, "()"), "CURRENT_TIMESTAMP") {
					def = "CURRENT_TIMESTAMP"
				}
				if (field{r[1], r[2], def, r[4]}) != want {
					t.Errorf("incorrect type/null/default/length for %s.%s", table, r[0])
				}
				if strings.Contains(strings.ToLower(r[5]), "unsigned") {
					t.Errorf("unsourced unsigned numeric: %s.%s", table, r[0])
				}
				if r[1] == "decimal" && (r[7] != "10" || r[8] != "2") {
					t.Errorf("%s.%s must be DECIMAL(10,2)", table, r[0])
				}
				if r[1] == "varchar" && r[9] != "utf8mb4_0900_bin" {
					t.Errorf("wrong character collation for %s.%s", table, r[0])
				}
				generated := r[0] == "id" && (table == "usercart" || table == "user_demand" || table == "transactions")
				if strings.Contains(r[6], "auto_increment") != generated {
					t.Errorf("wrong AUTO_INCREMENT contract for %s.%s", table, r[0])
				}
				if strings.Contains(strings.ToLower(r[6]), "on update") {
					t.Errorf("timestamp must not change on unrelated updates: %s.%s", table, r[0])
				}
			}
		})
	}
	f.t = t
	primary := map[string]string{"product": "pid", "user": "email", "orders": "orderid,prodid", "usercart": "id", "user_demand": "id", "transactions": "id"}
	for table, key := range primary {
		f.scalar("SELECT GROUP_CONCAT(column_name ORDER BY seq_in_index) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=? AND index_name='PRIMARY'", key, table)
	}
	wantUnique := [][]string{{"transactions", "transaction_id"}, {"usercart", "email,pid"}}
	f.wantRows("SELECT table_name, GROUP_CONCAT(column_name ORDER BY seq_in_index) AS cols FROM information_schema.statistics WHERE table_schema=DATABASE() AND index_name <> 'PRIMARY' AND non_unique=0 GROUP BY table_name,index_name ORDER BY table_name,cols", wantUnique)
	f.wantRows(`SELECT k.table_name,k.column_name,k.referenced_table_name,k.referenced_column_name,r.update_rule,r.delete_rule
FROM information_schema.key_column_usage k JOIN information_schema.referential_constraints r
ON r.constraint_schema=k.constraint_schema AND r.table_name=k.table_name AND r.constraint_name=k.constraint_name
WHERE k.constraint_schema=DATABASE() AND k.referenced_table_name IS NOT NULL ORDER BY k.table_name,k.column_name`, [][]string{
		{"orders", "prodid", "product", "pid", "NO ACTION", "NO ACTION"},
		{"transactions", "email", "user", "email", "CASCADE", "RESTRICT"},
		{"transactions", "pid", "product", "pid", "CASCADE", "RESTRICT"},
		{"user_demand", "email", "user", "email", "CASCADE", "CASCADE"},
		{"usercart", "email", "user", "email", "CASCADE", "CASCADE"},
		{"usercart", "pid", "product", "pid", "CASCADE", "RESTRICT"},
	})
	f.scalar("SELECT constraint_name FROM information_schema.referential_constraints WHERE constraint_schema=DATABASE() AND table_name='orders'", "productid")
	f.wantRows("SELECT table_name,enforced FROM information_schema.table_constraints WHERE constraint_schema=DATABASE() AND constraint_type='CHECK' ORDER BY table_name", [][]string{{"transactions", "YES"}, {"usercart", "YES"}})
	f.scalar("SELECT COUNT(*) FROM information_schema.triggers WHERE trigger_schema=DATABASE()", "0")
	f.scalar("SELECT COUNT(*) FROM information_schema.routines WHERE routine_schema=DATABASE()", "0")
	// Bootstrap never selects credential finalization, including on an empty user table.
	f.scalar("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name IN (?,?)", "0", historyName("credentials"), historyName("credentials-finalize"))
}
