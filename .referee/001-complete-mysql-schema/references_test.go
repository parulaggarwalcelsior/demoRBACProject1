package schema_test

import (
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/golang-migrate/migrate/v4"
)

func referenceMatrix(t *testing.T, f *fixture) {
	run := func(name string, fn func()) {
		t.Run(name, func(t *testing.T) { f.t = t; f.reset(); f.parents(); fn() })
	}
	for _, table := range []string{"product", "user", "transactions"} {
		run("duplicate-identity/"+table, func() {
			r := base(table)
			f.insert(table, r)
			duplicate := base(table)
			if table == "transactions" {
				delete(duplicate, "id")
			}
			f.badInsert(table, duplicate, []uint16{1062}, "23000")
		})
	}
	for _, table := range []string{"orders", "usercart"} {
		run("composite-identity/"+table, func() {
			r := base(table)
			f.insert(table, r)
			f.badInsert(table, base(table), []uint16{1062}, "23000")
			r = base(table)
			if table == "orders" {
				r["prodid"] = "p2"
			} else {
				r["pid"] = "p2"
			}
			f.insert(table, r)
			r = base(table)
			if table == "orders" {
				r["orderid"] = "another-order"
			} else {
				r["email"] = "u2"
			}
			f.insert(table, r)
			f.scalar("SELECT COUNT(*) FROM `"+table+"`", "3")
		})
	}
	run("demand-has-no-unsupplied-unique", func() {
		f.insert("user_demand", base("user_demand"))
		f.insert("user_demand", base("user_demand"))
		f.scalar("SELECT COUNT(*) FROM user_demand", "2")
	})
	for _, link := range []struct{ table, column string }{
		{"orders", "prodid"}, {"usercart", "email"}, {"usercart", "pid"},
		{"user_demand", "email"}, {"transactions", "email"}, {"transactions", "pid"},
	} {
		run("orphan-insert/"+link.table+"/"+link.column, func() {
			r := base(link.table)
			r[link.column] = "missing-parent"
			f.badInsert(link.table, r, []uint16{1452}, "23000")
		})
		run("orphan-update/"+link.table+"/"+link.column, func() {
			f.insert(link.table, base(link.table))
			f.reject([]uint16{1452}, "23000", "UPDATE `"+link.table+"` SET `"+link.column+"`='missing-parent'")
		})
	}
	for _, op := range []string{"delete", "update"} {
		run("historical-product/"+op, func() {
			query := "DELETE FROM product WHERE pid='p'"
			if op == "update" {
				query = "UPDATE product SET pid='renamed' WHERE pid='p'"
			}
			f.exec("INSERT INTO orders(orderid,prodid) VALUES ('a','p'),('b','p'),('c','p')")
			f.reject([]uint16{1451}, "23000", query)
			f.exec("UPDATE product SET pname='metadata is editable',pquantity=-2 WHERE pid='p'")
			f.scalar("SELECT COUNT(*) FROM orders WHERE prodid='p'", "3")
			f.exec("DELETE FROM orders WHERE orderid IN ('a','b')")
			f.reject([]uint16{1451}, "23000", query)
			f.exec("DELETE FROM orders WHERE orderid='c'")
			f.exec(query)
			f.scalar("SELECT COUNT(*) FROM product WHERE pid='p'", "0")
			if op == "update" {
				f.scalar("SELECT COUNT(*) FROM product WHERE pid='renamed'", "1")
			}
		})
	}
	for _, table := range []string{"orders", "usercart", "transactions"} {
		run("product-delete-restrict/"+table, func() {
			f.insert(table, base(table))
			f.reject([]uint16{1451}, "23000", "DELETE FROM product WHERE pid='p'")
		})
	}
	run("user-delete-cascades-cart-demand", func() {
		f.insert("usercart", base("usercart"))
		f.insert("user_demand", base("user_demand"))
		other := base("user_demand")
		other["email"] = "u2"
		f.insert("user_demand", other)
		f.exec("DELETE FROM user WHERE email='u'")
		f.scalar("SELECT COUNT(*) FROM usercart", "0")
		f.scalar("SELECT COUNT(*) FROM user_demand WHERE email='u'", "0")
		f.scalar("SELECT COUNT(*) FROM user_demand WHERE email='u2'", "1")
		f.scalar("SELECT COUNT(*) FROM product", "2")
	})
	run("transaction-user-delete-restrict", func() {
		f.insert("transactions", base("transactions"))
		f.reject([]uint16{1451}, "23000", "DELETE FROM user WHERE email='u'")
	})
	run("account-update-cascades-every-q1-email", func() {
		for _, table := range []string{"usercart", "user_demand", "transactions"} {
			f.insert(table, base(table))
		}
		f.exec("UPDATE user SET email='renamed-user' WHERE email='u'")
		for _, table := range []string{"usercart", "user_demand", "transactions"} {
			f.scalar("SELECT email FROM `"+table+"`", "renamed-user")
		}
	})
	run("product-update-cascades-q1-pids", func() {
		f.insert("usercart", base("usercart"))
		f.insert("transactions", base("transactions"))
		f.exec("UPDATE product SET pid='renamed-product' WHERE pid='p'")
		f.scalar("SELECT pid FROM usercart", "renamed-product")
		f.scalar("SELECT pid FROM transactions", "renamed-product")
	})
	run("mixed-product-update-is-atomic", func() {
		for _, table := range []string{"orders", "usercart", "transactions"} {
			f.insert(table, base(table))
		}
		f.reject([]uint16{1451}, "23000", "UPDATE product SET pid='renamed-product' WHERE pid='p'")
		f.reject([]uint16{1451}, "23000", "DELETE FROM product WHERE pid='p'")
	})
	run("transaction-restrict-keeps-user-cart-demand", func() {
		for _, table := range []string{"usercart", "user_demand", "transactions"} {
			f.insert(table, base(table))
		}
		f.reject([]uint16{1451}, "23000", "DELETE FROM user WHERE email='u'")
	})
}

func definitions(f *fixture) map[string][][]string {
	result := make(map[string][][]string)
	for _, table := range business {
		rows := f.rows("SHOW CREATE TABLE `" + table + "`")
		// AUTO_INCREMENT's current counter is state, not a schema difference.
		rows[0][1] = regexp.MustCompile(` AUTO_INCREMENT=\d+`).ReplaceAllString(rows[0][1], "")
		result[table] = rows
	}
	return result
}

func TestSchemaLifecycleAndReferenceReplay(t *testing.T) {
	f := newFixture(t)
	f.sentinel()
	sentinel := f.sentinelState()
	f.apply("schema", true)
	f.history("schema", "1", "0")
	initial := definitions(f)
	if err := f.track("schema", true); !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("repeated up must report no change: %s", f.safe(err))
	}
	f.history("schema", "1", "0")
	t.Run("initial-reference-matrix", func(t *testing.T) { referenceMatrix(t, f) })
	f.t = t
	// Deliberately populated teardown exercises all dependency paths.
	f.reset()
	f.parents()
	for _, table := range []string{"orders", "usercart", "user_demand", "transactions"} {
		f.insert(table, base(table))
	}
	f.apply("schema", false)
	for _, table := range business {
		f.scalar("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?", "0", table)
	}
	f.assertSentinel(sentinel)
	f.scalar("SELECT DATABASE()", "shopping-cart")
	f.apply("schema", true)
	if !reflect.DeepEqual(initial, definitions(f)) {
		t.Fatal("recreated definitions differ from initial definitions")
	}
	for _, table := range business {
		f.scalar("SELECT COUNT(*) FROM `"+table+"`", "0")
	}
	f.assertSentinel(sentinel)
	f.history("schema", "1", "0")
	// Force newly acquired sessions as well as checking the current connection.
	f.db.SetMaxIdleConns(0)
	f.scalar("SELECT @@foreign_key_checks", "1")
	t.Run("recreated-reference-matrix", func(t *testing.T) { referenceMatrix(t, f) })
	f.t = t
	f.assertSentinel(sentinel)
}

func TestTeardownRefusesExternalDependency(t *testing.T) {
	f := schemaFixture(t)
	f.parents()
	f.sentinel()
	sentinel := f.sentinelState()
	f.fixtureUp("external", `CREATE TABLE referee_external (
id INT PRIMARY KEY,pid VARCHAR(30) NOT NULL,
CONSTRAINT referee_external_product FOREIGN KEY(pid) REFERENCES product(pid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin`, "DROP TABLE referee_external")
	f.exec("INSERT INTO referee_external VALUES (1,'p')")
	external := f.rows("SHOW CREATE TABLE referee_external")
	if err := f.track("schema", false); err == nil {
		t.Fatal("teardown must refuse an external product reference")
	}
	// MySQL DDL can partially commit. Do not assert that removed feature children
	// return, force history, retry a dirty track, or broaden its cleanup inventory.
	f.scalar("SELECT dirty FROM `"+historyName("schema")+"`", "1")
	f.scalar("SELECT pid FROM referee_external WHERE id=1", "p")
	f.scalar("SELECT COUNT(*) FROM product WHERE pid='p'", "1")
	if !reflect.DeepEqual(external, f.rows("SHOW CREATE TABLE referee_external")) {
		t.Fatal("external dependency definition changed")
	}
	f.assertSentinel(sentinel)
	f.scalar("SELECT @@foreign_key_checks", "1")
}
