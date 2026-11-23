package schema_test

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"
)

var fkParentErr = expectedSQLError{number: 1451, state: "23000"}

func TestSchemaReferences(t *testing.T) {
	if startupConfigErr != nil {
		t.Fatalf("load startup config: %v", startupConfigErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db, err := openConfiguredDB(ctx, startupConfig)
	if err != nil {
		t.Fatalf("open schema reference connection: %v", err)
	}
	defer db.Close()

	if schemaMigrationAuthored() {
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("apply schema migration: %v", err)
		}
	}

	resetSchemaReferences(t, ctx, db)

	t.Run("duplicate identities and valid distinct pairs", func(t *testing.T) {
		resetSchemaReferences(t, ctx, db)
		assertDuplicateIdentitiesAndPairs(t, ctx, db)
	})
	t.Run("foreign key orphan refusals", func(t *testing.T) {
		resetSchemaReferences(t, ctx, db)
		assertFKOrphanRefusals(t, ctx, db)
	})
	t.Run("historical product protection", func(t *testing.T) {
		resetSchemaReferences(t, ctx, db)
		assertHistoricalProductProtection(t, ctx, db)
	})
	t.Run("q1 cascade and restrict actions", func(t *testing.T) {
		resetSchemaReferences(t, ctx, db)
		assertQ1Actions(t, ctx, db)
	})
	t.Run("mixed reference atomicity", func(t *testing.T) {
		resetSchemaReferences(t, ctx, db)
		assertMixedReferenceAtomicity(t, ctx, db)
	})
	t.Run("identical enforcement after reconstruction", func(t *testing.T) {
		resetSchemaReferences(t, ctx, db)
		assertReferenceEnforcementMatrix(t, ctx, db)
		if schemaMigrationAuthored() {
			if err := migrateToVersion(ctx, startupConfig, schemaTrack, 0); err != nil {
				t.Fatalf("reconstruct schema down: %v", err)
			}
			if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
				t.Fatalf("reconstruct schema up: %v", err)
			}
		}
		assertForeignKeyChecksEnabled(t, ctx, db)
		assertReferenceEnforcementMatrix(t, ctx, db)
	})
}

func resetSchemaReferences(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	resetSchemaValues(t, ctx, db)
}

func assertDuplicateIdentitiesAndPairs(t *testing.T, ctx context.Context, db *sql.DB) {
	insertProduct(t, ctx, db, "dup-product")
	before := snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, duplicateErrSpec, "INSERT INTO product (pid, pname) VALUES (?, ?)", "dup-product", "changed")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "duplicate product identity")
	assertExecOK(t, ctx, db, "INSERT INTO product (pid) VALUES (?)", "distinct-product")

	insertUser(t, ctx, db, "dup@example.test")
	before = snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, duplicateErrSpec, "INSERT INTO `user` (email, name) VALUES (?, ?)", "dup@example.test", "changed")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "duplicate user identity")
	assertExecOK(t, ctx, db, "INSERT INTO `user` (email) VALUES (?)", "distinct@example.test")

	assertExecOK(t, ctx, db, "INSERT INTO transactions (id, transaction_id, email, pid, quantity) VALUES (?, ?, ?, ?, 1)", 10, "tx-dup", "dup@example.test", "dup-product")
	before = snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, duplicateErrSpec, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", "tx-dup", "dup@example.test", "dup-product")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "duplicate transaction_id")
	assertExecError(t, ctx, db, duplicateErrSpec, "INSERT INTO transactions (id, transaction_id, email, pid, quantity) VALUES (?, ?, ?, ?, 1)", 10, "tx-other", "dup@example.test", "dup-product")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "duplicate generated transaction id")

	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "order-1", "dup-product")
	before = snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, duplicateErrSpec, "INSERT INTO orders (orderid, prodid, quantity) VALUES (?, ?, ?)", "order-1", "dup-product", 3)
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "duplicate order product pair")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "order-1", "distinct-product")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "order-2", "dup-product")

	assertExecOK(t, ctx, db, "INSERT INTO usercart (id, email, pid) VALUES (?, ?, ?)", 20, "dup@example.test", "dup-product")
	before = snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, duplicateErrSpec, "INSERT INTO usercart (email, pid, quantity) VALUES (?, ?, ?)", "dup@example.test", "dup-product", 2)
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "duplicate cart pair")
	assertExecError(t, ctx, db, duplicateErrSpec, "INSERT INTO usercart (id, email, pid) VALUES (?, ?, ?)", 20, "distinct@example.test", "distinct-product")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "duplicate generated cart id")
	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid) VALUES (?, ?)", "distinct@example.test", "dup-product")
	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid) VALUES (?, ?)", "dup@example.test", "distinct-product")

	assertExecOK(t, ctx, db, "INSERT INTO user_demand (id, email, book_name, author) VALUES (?, ?, ?, ?)", 30, "dup@example.test", "same-book", "same-author")
	before = snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, duplicateErrSpec, "INSERT INTO user_demand (id, email, book_name, author) VALUES (?, ?, ?, ?)", 30, "dup@example.test", "other-book", "other-author")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "duplicate generated demand id")
	assertExecOK(t, ctx, db, "INSERT INTO user_demand (email, book_name, author) VALUES (?, ?, ?)", "dup@example.test", "same-book", "same-author")
	assertExecOK(t, ctx, db, "INSERT INTO user_demand (email, book_name, author) VALUES (?, ?, ?)", "dup@example.test", "same-book", "same-author")
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM user_demand WHERE email = ? AND book_name = ? AND author = ?", "dup@example.test", 3, "same-book", "same-author")
}

func assertFKOrphanRefusals(t *testing.T, ctx context.Context, db *sql.DB) {
	insertProduct(t, ctx, db, "fk-product")
	insertProduct(t, ctx, db, "fk-product-2")
	insertUser(t, ctx, db, "fk@example.test")
	insertUser(t, ctx, db, "fk2@example.test")

	cases := []struct {
		name      string
		insertSQL string
		insertArg []any
		seedSQL   string
		seedArg   []any
		updateSQL string
		updateArg []any
	}{
		{name: "orders.prodid", insertSQL: "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", insertArg: []any{"orphan-order", "missing-product"}, seedSQL: "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", seedArg: []any{"orphan-update-order", "fk-product"}, updateSQL: "UPDATE orders SET prodid = ? WHERE orderid = ? AND prodid = ?", updateArg: []any{"missing-product", "orphan-update-order", "fk-product"}},
		{name: "usercart.email", insertSQL: "INSERT INTO usercart (email, pid) VALUES (?, ?)", insertArg: []any{"missing@example.test", "fk-product"}, seedSQL: "INSERT INTO usercart (email, pid) VALUES (?, ?)", seedArg: []any{"fk@example.test", "fk-product"}, updateSQL: "UPDATE usercart SET email = ? WHERE email = ? AND pid = ?", updateArg: []any{"missing@example.test", "fk@example.test", "fk-product"}},
		{name: "usercart.pid", insertSQL: "INSERT INTO usercart (email, pid) VALUES (?, ?)", insertArg: []any{"fk@example.test", "missing-product"}, seedSQL: "INSERT INTO usercart (email, pid) VALUES (?, ?)", seedArg: []any{"fk2@example.test", "fk-product"}, updateSQL: "UPDATE usercart SET pid = ? WHERE email = ? AND pid = ?", updateArg: []any{"missing-product", "fk2@example.test", "fk-product"}},
		{name: "user_demand.email", insertSQL: "INSERT INTO user_demand (email, book_name) VALUES (?, ?)", insertArg: []any{"missing@example.test", "missing-user-book"}, seedSQL: "INSERT INTO user_demand (email, book_name) VALUES (?, ?)", seedArg: []any{"fk@example.test", "demand-update"}, updateSQL: "UPDATE user_demand SET email = ? WHERE email = ? AND book_name = ?", updateArg: []any{"missing@example.test", "fk@example.test", "demand-update"}},
		{name: "transactions.email", insertSQL: "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", insertArg: []any{"tx-missing-email", "missing@example.test", "fk-product"}, seedSQL: "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", seedArg: []any{"tx-update-email", "fk@example.test", "fk-product"}, updateSQL: "UPDATE transactions SET email = ? WHERE transaction_id = ?", updateArg: []any{"missing@example.test", "tx-update-email"}},
		{name: "transactions.pid", insertSQL: "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", insertArg: []any{"tx-missing-product", "fk@example.test", "missing-product"}, seedSQL: "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", seedArg: []any{"tx-update-product", "fk@example.test", "fk-product-2"}, updateSQL: "UPDATE transactions SET pid = ? WHERE transaction_id = ?", updateArg: []any{"missing-product", "tx-update-product"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := snapshotReferenceRows(t, ctx, db)
			assertExecError(t, ctx, db, fkChildErr, tc.insertSQL, tc.insertArg...)
			assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), tc.name+" orphan insert")

			assertExecOK(t, ctx, db, tc.seedSQL, tc.seedArg...)
			before = snapshotReferenceRows(t, ctx, db)
			assertExecError(t, ctx, db, fkChildErr, tc.updateSQL, tc.updateArg...)
			assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), tc.name+" orphan update")
		})
	}

	thirty := "pppppppppppppppppppppppppppppp"
	thirtyOne := thirty + "x"
	fortySix := thirtyOne + "xxxxxxxxxxxxxxx"
	insertProduct(t, ctx, db, thirty)
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "length30-parent", thirty)
	before := snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, fkChildErr, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "length31-no-parent", thirtyOne)
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "orders prodid length31 FK")
	assertExecError(t, ctx, db, tooLongErr, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "length46-child", fortySix)
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "orders prodid length46 storage")
}

func assertHistoricalProductProtection(t *testing.T, ctx context.Context, db *sql.DB) {
	insertProduct(t, ctx, db, "history-one")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "history-order-1", "history-one")
	assertProductParentMutationRefused(t, ctx, db, "history-one", "one order reference")

	insertProduct(t, ctx, db, "history-many")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "history-order-a", "history-many")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "history-order-b", "history-many")
	assertProductParentMutationRefused(t, ctx, db, "history-many", "several order references")
	assertExecOK(t, ctx, db, "DELETE FROM orders WHERE orderid = ? AND prodid = ?", "history-order-a", "history-many")
	assertProductParentMutationRefused(t, ctx, db, "history-many", "last remaining order reference")

	insertProduct(t, ctx, db, "history-delete-free")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "history-delete-last", "history-delete-free")
	assertExecOK(t, ctx, db, "DELETE FROM orders WHERE orderid = ? AND prodid = ?", "history-delete-last", "history-delete-free")
	assertExecOK(t, ctx, db, "DELETE FROM product WHERE pid = ?", "history-delete-free")
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM product WHERE pid = ?", "history-delete-free", 0)

	insertProduct(t, ctx, db, "history-update-free")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "history-update-last", "history-update-free")
	assertExecOK(t, ctx, db, "DELETE FROM orders WHERE orderid = ? AND prodid = ?", "history-update-last", "history-update-free")
	assertExecOK(t, ctx, db, "UPDATE product SET pid = ? WHERE pid = ?", "history-update-new", "history-update-free")
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM product WHERE pid = ?", "history-update-new", 1)

	insertProduct(t, ctx, db, "history-metadata")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "history-metadata-order", "history-metadata")
	assertExecOK(t, ctx, db, "UPDATE product SET pname = ?, pquantity = ? WHERE pid = ?", "permitted metadata", -4, "history-metadata")
	product := queryProductValues(t, ctx, db, "history-metadata")
	if !product.PName.Valid || product.PName.String != "permitted metadata" || !product.PQuantity.Valid || product.PQuantity.Int64 != -4 {
		t.Fatalf("metadata update product = %+v, want changed non-identity fields", product)
	}
}

func assertQ1Actions(t *testing.T, ctx context.Context, db *sql.DB) {
	insertProduct(t, ctx, db, "q1-product")
	insertUser(t, ctx, db, "q1@example.test")
	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid) VALUES (?, ?)", "q1@example.test", "q1-product")
	assertExecOK(t, ctx, db, "INSERT INTO user_demand (email, book_name) VALUES (?, ?)", "q1@example.test", "q1-book")
	assertExecOK(t, ctx, db, "DELETE FROM `user` WHERE email = ?", "q1@example.test")
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM usercart WHERE email = ?", "q1@example.test", 0)
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM user_demand WHERE email = ?", "q1@example.test", 0)

	insertProduct(t, ctx, db, "q1-user-restrict-product")
	insertUser(t, ctx, db, "q1-restrict@example.test")
	assertExecOK(t, ctx, db, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", "q1-user-restrict-tx", "q1-restrict@example.test", "q1-user-restrict-product")
	before := snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, fkParentErr, "DELETE FROM `user` WHERE email = ?", "q1-restrict@example.test")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "transaction restricts user delete")

	insertProduct(t, ctx, db, "q1-product-restrict")
	insertUser(t, ctx, db, "q1-product-restrict@example.test")
	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid) VALUES (?, ?)", "q1-product-restrict@example.test", "q1-product-restrict")
	before = snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, fkParentErr, "DELETE FROM product WHERE pid = ?", "q1-product-restrict")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "cart restricts product delete")

	insertProduct(t, ctx, db, "q1-tx-product-restrict")
	insertUser(t, ctx, db, "q1-tx-product-restrict@example.test")
	assertExecOK(t, ctx, db, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", "q1-product-restrict-tx", "q1-tx-product-restrict@example.test", "q1-tx-product-restrict")
	before = snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, fkParentErr, "DELETE FROM product WHERE pid = ?", "q1-tx-product-restrict")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "transaction restricts product delete")

	insertProduct(t, ctx, db, "q1-cascade-product")
	insertUser(t, ctx, db, "q1-cascade@example.test")
	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid) VALUES (?, ?)", "q1-cascade@example.test", "q1-cascade-product")
	assertExecOK(t, ctx, db, "INSERT INTO user_demand (email, book_name) VALUES (?, ?)", "q1-cascade@example.test", "q1-cascade-book")
	assertExecOK(t, ctx, db, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", "q1-cascade-tx", "q1-cascade@example.test", "q1-cascade-product")
	assertExecOK(t, ctx, db, "UPDATE `user` SET email = ? WHERE email = ?", "q1-cascade-new@example.test", "q1-cascade@example.test")
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM usercart WHERE email = ?", "q1-cascade-new@example.test", 1)
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM user_demand WHERE email = ?", "q1-cascade-new@example.test", 1)
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM transactions WHERE email = ?", "q1-cascade-new@example.test", 1)

	assertExecOK(t, ctx, db, "UPDATE product SET pid = ? WHERE pid = ?", "q1-cascade-product-new", "q1-cascade-product")
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM usercart WHERE pid = ?", "q1-cascade-product-new", 1)
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM transactions WHERE pid = ?", "q1-cascade-product-new", 1)
}

func assertMixedReferenceAtomicity(t *testing.T, ctx context.Context, db *sql.DB) {
	insertProduct(t, ctx, db, "mixed-product")
	insertUser(t, ctx, db, "mixed@example.test")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "mixed-order", "mixed-product")
	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid) VALUES (?, ?)", "mixed@example.test", "mixed-product")
	assertExecOK(t, ctx, db, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", "mixed-tx", "mixed@example.test", "mixed-product")
	before := snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, fkParentErr, "UPDATE product SET pid = ? WHERE pid = ?", "mixed-product-new", "mixed-product")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "order NO ACTION blocks mixed product rewrite")

	insertProduct(t, ctx, db, "mixed-user-product")
	insertUser(t, ctx, db, "mixed-user@example.test")
	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid) VALUES (?, ?)", "mixed-user@example.test", "mixed-user-product")
	assertExecOK(t, ctx, db, "INSERT INTO user_demand (email, book_name) VALUES (?, ?)", "mixed-user@example.test", "mixed-user-book")
	assertExecOK(t, ctx, db, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", "mixed-user-tx", "mixed-user@example.test", "mixed-user-product")
	before = snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, fkParentErr, "DELETE FROM `user` WHERE email = ?", "mixed-user@example.test")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "transaction restrict leaves user cascades untouched")
}

func assertReferenceEnforcementMatrix(t *testing.T, ctx context.Context, db *sql.DB) {
	resetSchemaReferences(t, ctx, db)
	insertProduct(t, ctx, db, "matrix-product")
	insertUser(t, ctx, db, "matrix@example.test")
	before := snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, fkChildErr, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "matrix-missing-order", "matrix-missing-product")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "matrix orphan order insert")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "matrix-order", "matrix-product")
	before = snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, fkParentErr, "UPDATE product SET pid = ? WHERE pid = ?", "matrix-product-new", "matrix-product")
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), "matrix protected product update")
	assertExecOK(t, ctx, db, "UPDATE product SET pname = ? WHERE pid = ?", "matrix metadata", "matrix-product")
	product := queryProductValues(t, ctx, db, "matrix-product")
	if !product.PName.Valid || product.PName.String != "matrix metadata" {
		t.Fatalf("matrix metadata product = %+v, want updated name", product)
	}
}

func assertProductParentMutationRefused(t *testing.T, ctx context.Context, db *sql.DB, pid string, label string) {
	t.Helper()
	before := snapshotReferenceRows(t, ctx, db)
	assertExecError(t, ctx, db, fkParentErr, "DELETE FROM product WHERE pid = ?", pid)
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), label+" delete")
	assertExecError(t, ctx, db, fkParentErr, "UPDATE product SET pid = ? WHERE pid = ?", pid+"-new", pid)
	assertReferenceRowsEqual(t, before, snapshotReferenceRows(t, ctx, db), label+" identity update")
}

func assertForeignKeyChecksEnabled(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	var got string
	if err := db.QueryRowContext(ctx, "SELECT @@session.foreign_key_checks").Scan(&got); err != nil {
		t.Fatalf("read foreign_key_checks: %v", err)
	}
	if got != "1" {
		t.Fatalf("foreign_key_checks = %q, want 1", got)
	}
}

type referenceRows map[string]string

func snapshotReferenceRows(t *testing.T, ctx context.Context, db *sql.DB) referenceRows {
	t.Helper()
	queries := map[string]string{
		"product":      "SELECT COALESCE(GROUP_CONCAT(CAST(JSON_ARRAY(pid, pname, ptype, pinfo, CAST(pprice AS CHAR), pquantity) AS CHAR) ORDER BY pid SEPARATOR '\\n'), '') FROM product",
		"user":         "SELECT COALESCE(GROUP_CONCAT(CAST(JSON_ARRAY(email, name, mobile, address, pincode, password_hash, must_reset_password) AS CHAR) ORDER BY email SEPARATOR '\\n'), '') FROM `user`",
		"orders":       "SELECT COALESCE(GROUP_CONCAT(CAST(JSON_ARRAY(orderid, prodid, quantity, CAST(amount AS CHAR), shipped) AS CHAR) ORDER BY orderid, prodid SEPARATOR '\\n'), '') FROM orders",
		"usercart":     "SELECT COALESCE(GROUP_CONCAT(CAST(JSON_ARRAY(id, email, pid, quantity, added_at) AS CHAR) ORDER BY id SEPARATOR '\\n'), '') FROM usercart",
		"user_demand":  "SELECT COALESCE(GROUP_CONCAT(CAST(JSON_ARRAY(id, email, book_name, author, requested_at) AS CHAR) ORDER BY id SEPARATOR '\\n'), '') FROM user_demand",
		"transactions": "SELECT COALESCE(GROUP_CONCAT(CAST(JSON_ARRAY(id, transaction_id, email, pid, quantity, CAST(amount AS CHAR), created_at) AS CHAR) ORDER BY id SEPARATOR '\\n'), '') FROM transactions",
	}
	snapshot := referenceRows{}
	for table, query := range queries {
		var value string
		if err := db.QueryRowContext(ctx, query).Scan(&value); err != nil {
			t.Fatalf("snapshot %s rows: %v", table, err)
		}
		snapshot[table] = value
	}
	return snapshot
}

func assertReferenceRowsEqual(t *testing.T, want, got referenceRows, label string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s changed rows\nwant: %#v\n got: %#v", label, want, got)
	}
}
