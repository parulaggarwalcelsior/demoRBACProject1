package schema_test

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
)

// The credential-only entry fixture deliberately retains legacy email(60).
// Its upgrade must not pretend to be a complete target-schema reconstruction.
const legacyUp = `CREATE TABLE user (
email VARCHAR(60) NOT NULL PRIMARY KEY,
name VARCHAR(30) NULL DEFAULT NULL,
mobile BIGINT NULL DEFAULT NULL,
address VARCHAR(250) NULL DEFAULT NULL,
pincode INT NULL DEFAULT NULL,
password VARCHAR(20) NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin`
const legacyDown = "DROP TABLE user"

func legacyFixture(t *testing.T) *fixture {
	t.Helper()
	f := newFixture(t)
	f.fixtureUp("legacy", legacyUp, legacyDown)
	// No plaintext fixture value is stored, selected, printed, or converted.
	f.exec("INSERT INTO user(email,name,mobile,address,pincode,password) VALUES (?,NULL,NULL,NULL,NULL,NULL)", strings.Repeat("e", 60))
	f.exec("INSERT INTO user(email,name,mobile,address,pincode,password) VALUES ('synthetic','Reader',-9223372036854775808,'Address',-2147483648,NULL)")
	f.exec("INSERT INTO user(email,name,mobile,address,pincode,password) VALUES ('','',9223372036854775807,'',2147483647,NULL)")
	return f
}

func profiles(f *fixture) [][]string {
	return f.rows("SELECT email,name,mobile,address,pincode FROM user ORDER BY email")
}

func accounts(f *fixture) [][]string {
	return f.rows("SELECT email,name,mobile,address,pincode,password_hash,must_reset_password FROM user ORDER BY email")
}

func otherUserMetadata(f *fixture) [][]string {
	return f.rows(`SELECT column_name,column_type,is_nullable,column_default,extra,collation_name
FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='user'
AND column_name NOT IN ('password','password_hash','must_reset_password') ORDER BY ordinal_position`)
}

func (f *fixture) hashColumn(nullable string) {
	f.t.Helper()
	f.wantRows("SELECT data_type,character_maximum_length,is_nullable,column_default,collation_name FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='user' AND column_name='password_hash'", [][]string{{"varchar", "255", nullable, "<NULL>", "utf8mb4_0900_bin"}})
}

func stableCredentialDefinitions(f *fixture) map[string][][]string {
	// Ignore only the changing hash column. Every other column, constraint,
	// index, table option and column position must survive both transitions.
	result := make(map[string][][]string)
	hashLine := regexp.MustCompile("(?m)^  `password_hash`[^\\n]*")
	counter := regexp.MustCompile(` AUTO_INCREMENT=\d+`)
	for _, r := range f.rows(`SELECT table_name FROM information_schema.tables
WHERE table_schema=DATABASE() AND table_name IN ('product','user','orders','usercart','user_demand','transactions')`) {
		def := f.rows("SHOW CREATE TABLE `" + r[0] + "`")
		def[0][1] = hashLine.ReplaceAllString(counter.ReplaceAllString(def[0][1], ""), "  HASH_COLUMN")
		result[r[0]] = def
	}
	return result
}

func (f *fixture) rejectAccountWrite(code uint16, state, q string, args ...any) {
	f.t.Helper()
	before := accounts(f)
	_, err := f.db.Exec(q, args...)
	var my *mysql.MySQLError
	if !errors.As(err, &my) || my.Number != code || string(my.SQLState[:]) != state {
		f.t.Fatalf("wrong account-write refusal category: %s", f.safe(err))
	}
	if !reflect.DeepEqual(before, accounts(f)) {
		f.t.Fatal("rejected account write changed accounts")
	}
}

func TestCredentialReplacementPreservesAccountsAndRefusesDown(t *testing.T) {
	f := legacyFixture(t)
	f.sentinel()
	sentinel := f.sentinelState()
	before, metadata := profiles(f), otherUserMetadata(f)
	f.apply("credentials", true)
	f.history("credentials", "1", "0")
	if !reflect.DeepEqual(before, profiles(f)) || !reflect.DeepEqual(metadata, otherUserMetadata(f)) {
		t.Fatal("credential upgrade changed account/profile records or legacy profile definitions")
	}
	f.scalar("SELECT COUNT(*) FROM user WHERE password_hash IS NULL AND must_reset_password=1", "3")
	f.scalar("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='user' AND column_name='password'", "0")
	f.wantRows("SELECT column_name,data_type,character_maximum_length,is_nullable,column_default FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='user' AND column_name IN ('password_hash','must_reset_password') ORDER BY column_name", [][]string{
		{"must_reset_password", "tinyint", "<NULL>", "NO", "1"},
		{"password_hash", "varchar", "255", "YES", "<NULL>"},
	})
	if err := f.track("credentials", true); !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("repeated credential upgrade should be no-change: %s", f.safe(err))
	}
	f.history("credentials", "1", "0")
	f.assertSentinel(sentinel)
	hash := syntheticHash(255)
	f.exec("UPDATE user SET password_hash=?,must_reset_password=0 WHERE email='synthetic'", hash)
	f.scalar("SELECT password_hash FROM user WHERE email='synthetic'", hash)
	f.rejectAccountWrite(1406, "22001", "UPDATE user SET password_hash=? WHERE email='synthetic'", syntheticHash(256))
	f.rejectAccountWrite(1048, "23000", "UPDATE user SET must_reset_password=NULL WHERE email='synthetic'")
	protected := accounts(f)
	defs := f.rows("SHOW CREATE TABLE user")
	err := f.track("credentials", false)
	if err == nil || !strings.Contains(err.Error(), "45000") {
		t.Fatal("credential down must deliberately signal refusal before restoring plaintext storage")
	}
	if !reflect.DeepEqual(protected, accounts(f)) || !reflect.DeepEqual(defs, f.rows("SHOW CREATE TABLE user")) {
		t.Fatal("refused credential downgrade changed protected data/schema")
	}
	f.scalar("SELECT dirty FROM `"+historyName("credentials")+"`", "1")
	f.assertSentinel(sentinel)
	// Discard this dirty fixture; never force or misreport it as a rollback.
}

func TestFinalizationSQLGuardRejectsEveryPendingState(t *testing.T) {
	for _, tc := range []struct {
		name string
		hash any
		flag int
	}{
		{"null-and-reset", nil, 1}, {"null-and-no-reset", nil, 0},
		{"hash-and-reset", syntheticHash(90), 1},
		{"hash-and-other-positive-flag", syntheticHash(90), 2},
		{"hash-and-negative-flag", syntheticHash(90), -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := schemaFixture(t)
			f.sentinel()
			sentinel := f.sentinelState()
			f.exec("INSERT INTO user(email,password_hash,must_reset_password) VALUES ('ready',?,0),('pending',?,?)", syntheticHash(100), tc.hash, tc.flag)
			before, defs := f.snapshot(), definitions(f)
			// Deliberately bypass the implementation's Go preflight. This proves the
			// separately required SQL guard itself refuses through golang-migrate.
			// No concurrent account writer exists in this disposable fixture.
			err := f.track("credentials-finalize", true)
			if err == nil || !strings.Contains(err.Error(), "45000") {
				t.Fatal("finalization SQL must SIGNAL SQLSTATE 45000 for any pending account")
			}
			if !reflect.DeepEqual(before, f.snapshot()) || !reflect.DeepEqual(defs, definitions(f)) {
				t.Fatal("failed finalization changed business data or definitions")
			}
			f.history("credentials-finalize", "1", "1")
			f.history("schema", "1", "0")
			f.scalar("SELECT is_nullable FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='user' AND column_name='password_hash'", "YES")
			for _, r := range f.rows("SELECT routine_name FROM information_schema.routines WHERE routine_schema=DATABASE()") {
				if r[0] != "schema_guard_finalize_credentials" {
					t.Error("failed finalization left an unowned routine")
				}
			}
			f.assertSentinel(sentinel)
			// A SIGNAL may leave a guard routine and dirty migration history. The
			// whole explicitly disposable fixture is discarded by cleanup.
		})
	}
}

func TestCredentialFinalizationAndSafeDown(t *testing.T) {
	for _, entry := range []string{"empty-bootstrap", "ready-bootstrap", "ready-legacy-upgrade"} {
		t.Run(entry, func(t *testing.T) {
			var f *fixture
			if entry == "ready-legacy-upgrade" {
				f = legacyFixture(t)
				f.apply("credentials", true)
			} else {
				f = schemaFixture(t)
			}
			f.sentinel()
			sentinel := f.sentinelState()
			if entry == "ready-bootstrap" {
				f.parents()
				for _, table := range []string{"orders", "usercart", "user_demand", "transactions"} {
					f.insert(table, base(table))
				}
			}
			f.exec("UPDATE user SET password_hash=?,must_reset_password=0", syntheticHash(255))
			before, metadata := accounts(f), otherUserMetadata(f)
			stable := stableCredentialDefinitions(f)
			var allBefore map[string][][]string
			if entry != "ready-legacy-upgrade" {
				allBefore = f.snapshot()
			}
			// This test owns every writer. Guard and ALTER run inside this quiesced
			// boundary; a migration advisory lock alone is not writer isolation.
			f.apply("credentials-finalize", true)
			f.history("credentials-finalize", "1", "0")
			if !reflect.DeepEqual(before, accounts(f)) || !reflect.DeepEqual(metadata, otherUserMetadata(f)) {
				t.Fatal("finalization changed account values or unrelated profile definitions")
			}
			if !reflect.DeepEqual(stable, stableCredentialDefinitions(f)) {
				t.Fatal("finalization changed a definition beyond hash nullability")
			}
			if allBefore != nil && !reflect.DeepEqual(allBefore, f.snapshot()) {
				t.Fatal("finalization changed other business records")
			}
			f.hashColumn("NO")
			f.scalar("SELECT COUNT(*) FROM information_schema.routines WHERE routine_schema=DATABASE()", "0")
			f.rejectAccountWrite(1048, "23000", "INSERT INTO user(email,password_hash,must_reset_password) VALUES ('null-hash',NULL,0)")
			f.rejectAccountWrite(1364, "HY000", "INSERT INTO user(email,must_reset_password) VALUES ('omitted-hash',0)")
			f.rejectAccountWrite(1406, "22001", "INSERT INTO user(email,password_hash,must_reset_password) VALUES ('long-hash',?,0)", syntheticHash(256))
			// Readiness is a transition gate, not a permanent business CHECK.
			f.exec("INSERT INTO user(email,password_hash,must_reset_password) VALUES ('future-reset',?,1)", syntheticHash(120))
			f.rejectAccountWrite(1048, "23000", "UPDATE user SET password_hash=NULL WHERE email='future-reset'")
			beforeDown := accounts(f)
			f.apply("credentials-finalize", false)
			if !reflect.DeepEqual(beforeDown, accounts(f)) {
				t.Fatal("finalization down changed hashes/reset flags/profile values")
			}
			if !reflect.DeepEqual(stable, stableCredentialDefinitions(f)) {
				t.Fatal("finalization down changed an unrelated definition")
			}
			f.hashColumn("YES")
			f.scalar("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='user' AND column_name='password'", "0")
			f.exec("INSERT INTO user(email) VALUES ('reset-after-down')")
			f.wantRows("SELECT password_hash,must_reset_password FROM user WHERE email='reset-after-down'", [][]string{{"<NULL>", "1"}})
			f.assertSentinel(sentinel)
			if entry == "ready-legacy-upgrade" {
				f.history("credentials", "1", "0")
			} else {
				f.history("schema", "1", "0")
			}
		})
	}
}

func TestFinalizedSchemaRecreationUsesSeparateHistory(t *testing.T) {
	f := schemaFixture(t)
	f.sentinel()
	sentinel := f.sentinelState()
	f.exec("INSERT INTO user(email,password_hash,must_reset_password) VALUES ('ready',?,0)", syntheticHash(100))
	f.apply("credentials-finalize", true)
	f.history("schema", "1", "0")
	f.history("credentials-finalize", "1", "0")
	f.apply("credentials-finalize", false)
	f.scalar("SELECT COUNT(*) FROM `"+historyName("credentials-finalize")+"`", "0")
	f.apply("schema", false)
	f.apply("schema", true)
	f.scalar("SELECT COUNT(*) FROM user", "0")
	f.scalar("SELECT is_nullable FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='user' AND column_name='password_hash'", "YES")
	f.scalar("SELECT COUNT(*) FROM `"+historyName("credentials-finalize")+"`", "0")
	f.exec("INSERT INTO user(email) VALUES ('pending-after-reconstruction')")
	f.wantRows("SELECT password_hash,must_reset_password FROM user", [][]string{{"<NULL>", "1"}})
	f.exec("UPDATE user SET password_hash=?,must_reset_password=0", syntheticHash(100))
	f.apply("credentials-finalize", true)
	f.history("credentials-finalize", "1", "0")
	f.assertSentinel(sentinel)
}
