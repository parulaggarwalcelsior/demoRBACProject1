package schema_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratedb "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// These are independent referee tests, not the implementation's configuration
// package. They never import an implementation package or inspect its source.
var business = []string{"orders", "usercart", "user_demand", "transactions", "product", "user"}

type fixture struct {
	t   *testing.T
	db  *sql.DB
	cfg *mysql.Config
}

func targetRoot(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("REFEREE_TARGET_ROOT"); p != "" {
		p, err := filepath.Abs(p)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	p, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(p, "specs/001-complete-mysql-schema/spec.md")); err == nil {
			return p
		}
		// The authoring environment keeps the blind referee checkout beside the target.
		sibling := filepath.Join(p, "demoRBACProject1")
		if _, err := os.Stat(filepath.Join(sibling, "specs/001-complete-mysql-schema/spec.md")); err == nil {
			return sibling
		}
		parent := filepath.Dir(p)
		if parent == p {
			break
		}
		p = parent
	}
	t.Fatal("set REFEREE_TARGET_ROOT to the implementation checkout")
	return ""
}

func randomID(t *testing.T) string {
	t.Helper()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	var cfg *mysql.Config
	if dsn := os.Getenv("REFEREE_MYSQL_ADMIN_DSN"); dsn != "" {
		if os.Getenv("REFEREE_DISPOSABLE_MYSQL") != "1" {
			t.Fatal("REFEREE_MYSQL_ADMIN_DSN requires REFEREE_DISPOSABLE_MYSQL=1: this dedicated server's shopping-cart database is recreated per fixture")
		}
		var err error
		cfg, err = mysql.ParseDSN(dsn)
		if err != nil {
			t.Fatal("invalid referee admin DSN")
		}
	} else {
		if _, err := exec.LookPath("docker"); err != nil {
			t.Fatal("real MySQL 8.4 is required: supply a disposable REFEREE_MYSQL_ADMIN_DSN or Docker; tests do not skip")
		}
		name, password := "referee-schema-"+randomID(t), randomID(t)
		image := os.Getenv("REFEREE_MYSQL_IMAGE")
		if image == "" {
			image = "mysql:8.4"
		}
		cmd := exec.Command("docker", "run", "--detach", "--rm", "--name", name,
			"--publish", "127.0.0.1::3306", "--env", "MYSQL_ROOT_PASSWORD", "--env", "MYSQL_ROOT_HOST=%", image)
		cmd.Env = append(os.Environ(), "MYSQL_ROOT_PASSWORD="+password)
		if _, err := cmd.CombinedOutput(); err != nil {
			t.Fatal("could not start disposable MySQL container")
		}
		t.Cleanup(func() {
			if err := exec.Command("docker", "rm", "--force", name).Run(); err != nil {
				t.Error("could not remove owned MySQL container")
			}
		})
		out, err := exec.Command("docker", "port", name, "3306/tcp").Output()
		if err != nil {
			t.Fatal("could not find disposable MySQL port")
		}
		cfg = mysql.NewConfig()
		cfg.User, cfg.Passwd, cfg.Net, cfg.Addr = "root", password, "tcp", strings.TrimSpace(string(out))
	}
	cfg.DBName = ""
	cfg.ParseTime = false
	cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = 5*time.Second, 30*time.Second, 30*time.Second
	cfg.MultiStatements = true
	cfg.Params = map[string]string{
		"charset": "utf8mb4", "collation_connection": "'utf8mb4_0900_bin'",
		"sql_mode":           "'STRICT_ALL_TABLES,NO_ENGINE_SUBSTITUTION,NO_AUTO_VALUE_ON_ZERO'",
		"foreign_key_checks": "1", "time_zone": "'+00:00'",
	}
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal("could not configure MySQL admin connection")
	}
	admin.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = admin.Close() })
	deadline := time.Now().Add(100 * time.Second)
	for {
		if err := admin.Ping(); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("disposable MySQL did not become ready")
		}
		time.Sleep(500 * time.Millisecond)
	}
	// Namespace provisioning/cleanup is infrastructure, never feature DDL. The
	// dedicated-server opt-in above also authorizes discarding dirty fixtures.
	var locked int
	if err := admin.QueryRow("SELECT GET_LOCK('referee_001_complete_mysql_schema', 60)").Scan(&locked); err != nil || locked != 1 {
		t.Fatal("could not exclusively lock dedicated test server")
	}
	t.Cleanup(func() { _, _ = admin.Exec("SELECT RELEASE_LOCK('referee_001_complete_mysql_schema')") })
	if _, err := admin.Exec("DROP DATABASE IF EXISTS `shopping-cart`"); err != nil {
		t.Fatal("could not discard old disposable namespace")
	}
	if _, err := admin.Exec("CREATE DATABASE `shopping-cart` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_bin"); err != nil {
		t.Fatal("could not provision disposable namespace")
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP DATABASE IF EXISTS `shopping-cart`"); err != nil {
			t.Error("could not discard disposable namespace")
		}
	})
	cfg.DBName = "shopping-cart"
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal("could not open test database")
	}
	t.Cleanup(func() { _ = db.Close() })
	f := &fixture{t: t, db: db, cfg: cfg}
	var version, namespace, mode, charset, collation, zone string
	var checks int
	err = db.QueryRow("SELECT VERSION(), DATABASE(), @@sql_mode, @@character_set_connection, @@collation_connection, @@time_zone, @@foreign_key_checks").Scan(&version, &namespace, &mode, &charset, &collation, &zone, &checks)
	if err != nil {
		t.Fatal("could not inspect MySQL environment")
	}
	if !strings.HasPrefix(version, "8.4.") || namespace != "shopping-cart" || charset != "utf8mb4" || collation != "utf8mb4_0900_bin" || zone != "+00:00" || checks != 1 {
		t.Fatal("acceptance requires MySQL 8.4, shopping-cart, utf8mb4/utf8mb4_0900_bin, UTC and enabled foreign keys")
	}
	for _, wanted := range []string{"STRICT_ALL_TABLES", "NO_ENGINE_SUBSTITUTION", "NO_AUTO_VALUE_ON_ZERO"} {
		if !strings.Contains(","+mode+",", ","+wanted+",") {
			t.Fatalf("missing required SQL mode %s", wanted)
		}
	}
	return f
}

func (f *fixture) safe(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	for _, secret := range []string{f.cfg.FormatDSN(), f.cfg.Passwd} {
		if secret != "" {
			s = strings.ReplaceAll(s, secret, "[redacted]")
		}
	}
	return s
}

func (f *fixture) exec(q string, args ...any) {
	f.t.Helper()
	if _, err := f.db.Exec(q, args...); err != nil {
		f.t.Fatalf("SQL fixture/assertion failed: %s", f.safe(err))
	}
}

func (f *fixture) rows(q string, args ...any) [][]string {
	f.t.Helper()
	r, err := f.db.Query(q, args...)
	if err != nil {
		f.t.Fatalf("query failed: %s", f.safe(err))
	}
	defer r.Close()
	cols, err := r.Columns()
	if err != nil {
		f.t.Fatal(err)
	}
	result := make([][]string, 0)
	for r.Next() {
		values, dest := make([]sql.NullString, len(cols)), make([]any, len(cols))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := r.Scan(dest...); err != nil {
			f.t.Fatal(err)
		}
		row := make([]string, len(values))
		for i, v := range values {
			if v.Valid {
				row[i] = v.String
			} else {
				row[i] = "<NULL>"
			}
		}
		result = append(result, row)
	}
	if err := r.Err(); err != nil {
		f.t.Fatal(err)
	}
	return result
}

func (f *fixture) wantRows(q string, want [][]string, args ...any) {
	f.t.Helper()
	if !reflect.DeepEqual(f.rows(q, args...), want) {
		f.t.Fatalf("query result differs from the specified values: %s", q)
	}
}

func (f *fixture) scalar(q, want string, args ...any) {
	f.t.Helper()
	f.wantRows(q, [][]string{{want}}, args...)
}

func (f *fixture) reset() {
	f.t.Helper()
	for _, table := range business {
		f.exec("DELETE FROM `" + table + "`")
	}
}

func (f *fixture) parents() {
	f.exec("INSERT INTO product(pid) VALUES ('p'),('p2')")
	f.exec("INSERT INTO user(email) VALUES ('u'),('u2')")
}

func (f *fixture) snapshot() map[string][][]string {
	f.t.Helper()
	out := make(map[string][][]string)
	for _, table := range business {
		r := f.rows("SELECT * FROM `" + table + "`")
		sort.Slice(r, func(i, j int) bool { return fmt.Sprint(r[i]) < fmt.Sprint(r[j]) })
		out[table] = r
	}
	return out
}

func (f *fixture) reject(codes []uint16, state, q string, args ...any) {
	f.t.Helper()
	before := f.snapshot()
	_, err := f.db.Exec(q, args...)
	var my *mysql.MySQLError
	if !errors.As(err, &my) {
		f.t.Fatalf("expected MySQL constraint error, got %s", f.safe(err))
	}
	found := false
	for _, code := range codes {
		if my.Number == code {
			found = true
		}
	}
	if !found || string(my.SQLState[:]) != state {
		f.t.Fatalf("wrong SQL error category: code=%d state=%s; wanted %v/%s", my.Number, my.SQLState, codes, state)
	}
	if !reflect.DeepEqual(before, f.snapshot()) {
		f.t.Fatal("rejected write changed business rows")
	}
}

func (f *fixture) migration(s source.Driver, history string, up bool) error {
	db, err := sql.Open("mysql", f.cfg.FormatDSN())
	if err != nil {
		return err
	}
	driver, err := migratedb.WithInstance(db, &migratedb.Config{MigrationsTable: history})
	if err != nil {
		_ = db.Close()
		return err
	}
	m, err := migrate.NewWithInstance("referee", s, "mysql", driver)
	if err != nil {
		_ = driver.Close()
		return err
	}
	defer m.Close()
	if up {
		return m.Up()
	}
	return m.Steps(-1)
}

func (f *fixture) track(track string, up bool) error {
	f.t.Helper()
	p := filepath.Join(targetRoot(f.t), "migrations", track)
	u := (&url.URL{Scheme: "file", Path: filepath.ToSlash(p)}).String()
	s, err := (&file.File{}).Open(u)
	if err != nil {
		return err
	}
	return f.migration(s, historyName(track), up)
}

func historyName(track string) string {
	return "referee_" + strings.ReplaceAll(track, "-", "_") + "_history"
}

func (f *fixture) apply(track string, up bool) {
	f.t.Helper()
	if err := f.track(track, up); err != nil {
		f.t.Fatalf("%s migration failed: %s", track, f.safe(err))
	}
}

func (f *fixture) history(track, version, dirty string) {
	f.t.Helper()
	f.wantRows("SELECT CAST(version AS CHAR), CAST(dirty AS CHAR) FROM `"+historyName(track)+"`", [][]string{{version, dirty}})
}

func (f *fixture) fixtureMigration(name, upSQL, downSQL string, up bool) error {
	f.t.Helper()
	s, err := iofs.New(fstest.MapFS{
		"000001_fixture.up.sql":   &fstest.MapFile{Data: []byte(upSQL)},
		"000001_fixture.down.sql": &fstest.MapFile{Data: []byte(downSQL)},
	}, ".")
	if err != nil {
		return err
	}
	return f.migration(s, "referee_fixture_"+name, up)
}

func (f *fixture) fixtureUp(name, upSQL, downSQL string) {
	f.t.Helper()
	if err := f.fixtureMigration(name, upSQL, downSQL, true); err != nil {
		f.t.Fatalf("versioned referee fixture failed: %s", f.safe(err))
	}
}

const sentinelUp = "CREATE TABLE referee_unrelated (id INT PRIMARY KEY, payload VARBINARY(40) NOT NULL) ENGINE=InnoDB"
const sentinelDown = "DROP TABLE referee_unrelated"

func (f *fixture) sentinel() {
	f.fixtureUp("sentinel", sentinelUp, sentinelDown)
	f.exec("INSERT INTO referee_unrelated VALUES (7, X'00616263FF')")
}

func (f *fixture) sentinelState() [][]string { return f.rows("SHOW CREATE TABLE referee_unrelated") }

func (f *fixture) assertSentinel(def [][]string) {
	f.t.Helper()
	if !reflect.DeepEqual(def, f.sentinelState()) {
		f.t.Fatal("unrelated table definition changed")
	}
	f.wantRows("SELECT id, HEX(payload) FROM referee_unrelated", [][]string{{"7", "00616263FF"}})
}

func schemaFixture(t *testing.T) *fixture {
	t.Helper()
	f := newFixture(t)
	f.apply("schema", true)
	f.history("schema", "1", "0")
	return f
}

func runCommand(t *testing.T, f *fixture, timeout time.Duration, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	tool := args[0]
	// A runner can invoke this standalone module with an absolute Go path even
	// when Go is missing from the inherited PATH. Use that same toolchain below.
	if tool == "go" {
		tool = filepath.Join(runtime.GOROOT(), "bin", "go")
	}
	cmd := exec.CommandContext(ctx, tool, args[1:]...)
	cmd.Dir = targetRoot(t)
	toolPath := filepath.Join(runtime.GOROOT(), "bin") + string(os.PathListSeparator) + os.Getenv("PATH")
	cmd.Env = append(os.Environ(), "PATH="+toolPath, "SCHEMA_TEST_DSN="+f.cfg.FormatDSN(), "SCHEMA_TEST_MYSQL_PASSWORD="+f.cfg.Passwd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Implementation output may contain synthetic hashes or secrets. Retain only
		// the command identity/status; reproduce locally to inspect its own diagnostics.
		t.Fatalf("required implementation gate %q failed (output withheld to protect runtime credentials): %v", args, err)
	}
	return out
}
