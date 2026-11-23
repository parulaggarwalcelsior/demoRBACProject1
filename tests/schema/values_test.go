package schema_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSchemaValues(t *testing.T) {
	if startupConfigErr != nil {
		t.Fatalf("load startup config: %v", startupConfigErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	db, err := openConfiguredDB(ctx, startupConfig)
	if err != nil {
		t.Fatalf("open schema value connection: %v", err)
	}
	defer db.Close()

	if schemaMigrationAuthored() {
		if err := migrateToVersion(ctx, startupConfig, schemaTrack, schemaTrack.Target); err != nil {
			t.Fatalf("apply schema migration: %v", err)
		}
	}

	resetSchemaValues(t, ctx, db)
	seedValueParents(t, ctx, db)

	t.Run("required and optional values", func(t *testing.T) {
		assertRequiredAndOptionalValues(t, ctx, db)
	})
	t.Run("varchar limits", func(t *testing.T) {
		assertVarcharLimits(t, ctx, db)
	})
	t.Run("binary image round trip", func(t *testing.T) {
		assertBinaryImageRoundTrip(t, ctx, db)
	})
	t.Run("money fidelity range defaults and rounding", func(t *testing.T) {
		assertMoneyValues(t, ctx, db)
	})
	t.Run("signed numeric boundaries", func(t *testing.T) {
		assertSignedNumericBoundaries(t, ctx, db)
	})
	t.Run("permitted signed domain values", func(t *testing.T) {
		assertPermittedSignedDomainValues(t, ctx, db)
	})
	t.Run("generated identities", func(t *testing.T) {
		assertGeneratedIdentities(t, ctx, db)
	})
	t.Run("quantity checks and timestamps", func(t *testing.T) {
		assertQuantityChecksAndTimestamps(t, ctx, db)
	})
	t.Run("demand duplicates and nullable author", func(t *testing.T) {
		assertDemandValues(t, ctx, db)
	})
	t.Run("hash capacity", func(t *testing.T) {
		assertHashCapacity(t, ctx, db)
	})
}

func resetSchemaValues(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM transactions",
		"DELETE FROM user_demand",
		"DELETE FROM usercart",
		"DELETE FROM orders",
		"DELETE FROM product",
		"DELETE FROM `user`",
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("reset schema values with %q: %v", stmt, err)
		}
	}
}

func seedValueParents(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	for _, pid := range []string{"seed-product", "quantity-product", "quantity-default-product"} {
		insertProduct(t, ctx, db, pid)
	}
	insertUser(t, ctx, db, "seed@example.test")
}

func assertRequiredAndOptionalValues(t *testing.T, ctx context.Context, db *sql.DB) {
	assertExecOK(t, ctx, db, "INSERT INTO product (pid) VALUES (?)", "required-pid-only")
	product := queryProductValues(t, ctx, db, "required-pid-only")
	if product.PName.Valid || product.PType.Valid || product.PInfo.Valid || product.PQuantity.Valid || product.Image.Valid {
		t.Fatalf("omitted optional product fields = %+v, want NULL optional fields", product)
	}
	if product.PPrice != "0.00" {
		t.Fatalf("omitted product price = %q, want 0.00", product.PPrice)
	}

	assertExecError(t, ctx, db, requiredDefaultErr, "INSERT INTO product (pname) VALUES (?)", "missing pid")
	assertExecError(t, ctx, db, nullErr, "INSERT INTO product (pid, pprice) VALUES (?, NULL)", "null-price")
	assertExecOK(t, ctx, db, "INSERT INTO product (pid, pname, ptype, pinfo, pquantity, image) VALUES (?, NULL, NULL, NULL, NULL, NULL)", "explicit-product-nulls")

	assertExecOK(t, ctx, db, "INSERT INTO `user` (email) VALUES (?)", "required-user@example.test")
	user := queryUserValues(t, ctx, db, "required-user@example.test")
	if user.Name.Valid || user.Mobile.Valid || user.Address.Valid || user.Pincode.Valid || user.PasswordHash.Valid {
		t.Fatalf("omitted optional user fields = %+v, want NULL optional fields", user)
	}
	if user.MustResetPassword != 1 {
		t.Fatalf("omitted must_reset_password = %d, want 1", user.MustResetPassword)
	}
	assertExecError(t, ctx, db, requiredDefaultErr, "INSERT INTO `user` (name) VALUES (?)", "missing email")
	assertExecError(t, ctx, db, nullErr, "INSERT INTO `user` (email, must_reset_password) VALUES (?, NULL)", "null-reset@example.test")

	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", "order-defaults", "seed-product")
	order := queryOrderValues(t, ctx, db, "order-defaults", "seed-product")
	if order.Quantity.Valid {
		t.Fatalf("omitted order quantity = %+v, want NULL", order.Quantity)
	}
	if order.Amount != "0.00" || order.Shipped != 0 {
		t.Fatalf("omitted order defaults amount/shipped = %q/%d, want 0.00/0", order.Amount, order.Shipped)
	}
	assertExecError(t, ctx, db, requiredDefaultErr, "INSERT INTO orders (prodid) VALUES (?)", "seed-product")
	assertExecError(t, ctx, db, nullErr, "INSERT INTO orders (orderid, prodid, amount) VALUES (?, ?, NULL)", "null-amount", "seed-product")
	assertExecError(t, ctx, db, nullErr, "INSERT INTO orders (orderid, prodid, shipped) VALUES (?, ?, NULL)", "null-shipped", "seed-product")

	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid) VALUES (?, ?)", "seed@example.test", "seed-product")
	assertIntValue(t, ctx, db, "SELECT quantity FROM usercart WHERE email = ? AND pid = ?", 1, "seed@example.test", "seed-product")
	assertExecError(t, ctx, db, requiredDefaultErr, "INSERT INTO usercart (email) VALUES (?)", "seed@example.test")
	assertExecError(t, ctx, db, nullErr, "INSERT INTO usercart (email, pid, quantity) VALUES (?, ?, NULL)", "seed@example.test", "seed-product")

	assertExecError(t, ctx, db, requiredDefaultErr, "INSERT INTO transactions (transaction_id, email, pid, amount) VALUES (?, ?, ?, ?)", "missing-quantity", "seed@example.test", "seed-product", "1.00")
	assertExecError(t, ctx, db, nullErr, "INSERT INTO transactions (transaction_id, email, pid, quantity, amount) VALUES (?, ?, ?, NULL, ?)", "null-quantity", "seed@example.test", "seed-product", "1.00")
	assertExecOK(t, ctx, db, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, ?)", "tx-default-amount", "seed@example.test", "seed-product", 1)
	assertDecimalValue(t, ctx, db, "SELECT amount FROM transactions WHERE transaction_id = ?", "0.00", "tx-default-amount")
}

func assertVarcharLimits(t *testing.T, ctx context.Context, db *sql.DB) {
	fourByte := "😀"
	cases := []struct {
		name      string
		limit     int
		insertSQL string
		args      func(string) []any
		countSQL  string
	}{
		{name: "product.pid", limit: 30, insertSQL: "INSERT INTO product (pid) VALUES (?)", args: oneArg, countSQL: "SELECT COUNT(*) FROM product WHERE pid = ?"},
		{name: "product.pname", limit: 100, insertSQL: "INSERT INTO product (pid, pname) VALUES (?, ?)", args: keyedArg("v-pname"), countSQL: "SELECT COUNT(*) FROM product WHERE pname = ?"},
		{name: "product.ptype", limit: 20, insertSQL: "INSERT INTO product (pid, ptype) VALUES (?, ?)", args: keyedArg("v-ptype"), countSQL: "SELECT COUNT(*) FROM product WHERE ptype = ?"},
		{name: "product.pinfo", limit: 350, insertSQL: "INSERT INTO product (pid, pinfo) VALUES (?, ?)", args: keyedArg("v-pinfo"), countSQL: "SELECT COUNT(*) FROM product WHERE pinfo = ?"},
		{name: "user.email", limit: 100, insertSQL: "INSERT INTO `user` (email) VALUES (?)", args: oneArg, countSQL: "SELECT COUNT(*) FROM `user` WHERE email = ?"},
		{name: "user.name", limit: 30, insertSQL: "INSERT INTO `user` (email, name) VALUES (?, ?)", args: keyedArg("v-name@example.test"), countSQL: "SELECT COUNT(*) FROM `user` WHERE name = ?"},
		{name: "user.address", limit: 250, insertSQL: "INSERT INTO `user` (email, address) VALUES (?, ?)", args: keyedArg("v-address@example.test"), countSQL: "SELECT COUNT(*) FROM `user` WHERE address = ?"},
		{name: "orders.orderid", limit: 45, insertSQL: "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", args: keyedArgWithValueSecond("seed-product"), countSQL: "SELECT COUNT(*) FROM orders WHERE orderid = ?"},
		{name: "orders.prodid", limit: 45, insertSQL: "INSERT INTO orders (orderid, prodid) VALUES (?, ?)", args: orderProductArg, countSQL: "SELECT COUNT(*) FROM orders WHERE prodid = ?"},
		{name: "transactions.transaction_id", limit: 40, insertSQL: "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", args: keyedArgWithTwoFixed("seed@example.test", "seed-product"), countSQL: "SELECT COUNT(*) FROM transactions WHERE transaction_id = ?"},
		{name: "user_demand.book_name", limit: 100, insertSQL: "INSERT INTO user_demand (email, book_name) VALUES (?, ?)", args: valueSecond("seed@example.test"), countSQL: "SELECT COUNT(*) FROM user_demand WHERE book_name = ?"},
		{name: "user_demand.author", limit: 100, insertSQL: "INSERT INTO user_demand (email, book_name, author) VALUES (?, ?, ?)", args: demandAuthorArg, countSQL: "SELECT COUNT(*) FROM user_demand WHERE author = ?"},
		{name: "usercart.email", limit: 100, insertSQL: "INSERT INTO usercart (email, pid) VALUES (?, ?)", args: cartEmailArg, countSQL: "SELECT COUNT(*) FROM usercart WHERE email = ?"},
		{name: "usercart.pid", limit: 30, insertSQL: "INSERT INTO usercart (email, pid) VALUES (?, ?)", args: cartProductArg, countSQL: "SELECT COUNT(*) FROM usercart WHERE pid = ?"},
		{name: "transactions.email", limit: 100, insertSQL: "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", args: transactionEmailArg, countSQL: "SELECT COUNT(*) FROM transactions WHERE email = ?"},
		{name: "transactions.pid", limit: 30, insertSQL: "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)", args: transactionProductArg, countSQL: "SELECT COUNT(*) FROM transactions WHERE pid = ?"},
		{name: "user.password_hash", limit: 255, insertSQL: "INSERT INTO `user` (email, password_hash) VALUES (?, ?)", args: keyedArg("v-hash@example.test"), countSQL: "SELECT COUNT(*) FROM `user` WHERE password_hash = ?"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exact := strings.Repeat(fourByte, tc.limit)
			oneOver := exact + fourByte
			if tc.name == "orders.prodid" {
				assertExecError(t, ctx, db, fkChildErr, tc.insertSQL, tc.args(exact)...)
				assertExecError(t, ctx, db, tooLongErr, tc.insertSQL, tc.args(oneOver)...)
				return
			}
			seedVarcharParents(t, ctx, db, tc.name, exact)
			assertExecOK(t, ctx, db, tc.insertSQL, tc.args(exact)...)
			assertCount(t, ctx, db, tc.countSQL, exact, 1)
			before := tableCount(t, ctx, db, tableFromCountSQL(tc.countSQL))
			assertExecError(t, ctx, db, tooLongErr, tc.insertSQL, tc.args(oneOver)...)
			after := tableCount(t, ctx, db, tableFromCountSQL(tc.countSQL))
			if after != before {
				t.Fatalf("%s one-over insert changed row count from %d to %d", tc.name, before, after)
			}
			assertCount(t, ctx, db, tc.countSQL, oneOver, 0)
		})
	}
}

func assertBinaryImageRoundTrip(t *testing.T, ctx context.Context, db *sql.DB) {
	want := []byte{0x00, 0x01, 0xfe, 0xff, 'i', 'm', 'g'}
	assertExecOK(t, ctx, db, "INSERT INTO product (pid, image) VALUES (?, ?)", "image-product", want)
	var got []byte
	if err := db.QueryRowContext(ctx, "SELECT image FROM product WHERE pid = ?", "image-product").Scan(&got); err != nil {
		t.Fatalf("read image product: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("image bytes = %v, want %v", got, want)
	}
}

func assertMoneyValues(t *testing.T, ctx context.Context, db *sql.DB) {
	cases := []struct {
		name      string
		insertSQL string
		selectSQL string
		args      func(string) []any
	}{
		{name: "product.pprice", insertSQL: "INSERT INTO product (pid, pprice) VALUES (?, ?)", selectSQL: "SELECT pprice FROM product WHERE pid = ?", args: keyedArg("money-product")},
		{name: "orders.amount", insertSQL: "INSERT INTO orders (orderid, prodid, amount) VALUES (?, ?, ?)", selectSQL: "SELECT amount FROM orders WHERE orderid = ?", args: orderAmountArg},
		{name: "transactions.amount", insertSQL: "INSERT INTO transactions (transaction_id, email, pid, quantity, amount) VALUES (?, ?, ?, 1, ?)", selectSQL: "SELECT amount FROM transactions WHERE transaction_id = ?", args: transactionAmountArg},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, value := range []string{"0.00", "12.34", "99999999.99", "-99999999.99"} {
				assertExecOK(t, ctx, db, tc.insertSQL, tc.args(value)...)
				assertDecimalValue(t, ctx, db, tc.selectSQL, value, valueKey(tc.args(value)))
			}
			assertExecError(t, ctx, db, outOfRangeErr, tc.insertSQL, tc.args("100000000.00")...)
			assertExecError(t, ctx, db, outOfRangeErr, tc.insertSQL, tc.args("-100000000.00")...)
			assertExecOK(t, ctx, db, tc.insertSQL, tc.args("12.345")...)
			assertDecimalValue(t, ctx, db, tc.selectSQL, "12.35", valueKey(tc.args("12.345")))
			assertExecOK(t, ctx, db, tc.insertSQL, tc.args("-12.345")...)
			assertDecimalValue(t, ctx, db, tc.selectSQL, "-12.35", valueKey(tc.args("-12.345")))
		})
	}
}

func assertSignedNumericBoundaries(t *testing.T, ctx context.Context, db *sql.DB) {
	intCases := []struct {
		name      string
		insertSQL string
		args      func(string) []any
		selectSQL string
	}{
		{name: "product.pquantity", insertSQL: "INSERT INTO product (pid, pquantity) VALUES (?, ?)", args: keyedArg("int-product"), selectSQL: "SELECT pquantity FROM product WHERE pid = ?"},
		{name: "orders.quantity", insertSQL: "INSERT INTO orders (orderid, prodid, quantity) VALUES (?, ?, ?)", args: orderAmountArg, selectSQL: "SELECT quantity FROM orders WHERE orderid = ?"},
		{name: "orders.shipped", insertSQL: "INSERT INTO orders (orderid, prodid, shipped) VALUES (?, ?, ?)", args: orderAmountArg, selectSQL: "SELECT shipped FROM orders WHERE orderid = ?"},
		{name: "user.pincode", insertSQL: "INSERT INTO `user` (email, pincode) VALUES (?, ?)", args: keyedArg("int-user@example.test"), selectSQL: "SELECT pincode FROM `user` WHERE email = ?"},
	}
	for _, tc := range intCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, value := range []string{"-2147483648", "2147483647"} {
				assertExecOK(t, ctx, db, tc.insertSQL, tc.args(value)...)
				assertIntStringValue(t, ctx, db, tc.selectSQL, value, valueKey(tc.args(value)))
			}
			assertExecError(t, ctx, db, outOfRangeErr, tc.insertSQL, tc.args("-2147483649")...)
			assertExecError(t, ctx, db, outOfRangeErr, tc.insertSQL, tc.args("2147483648")...)
		})
	}

	for _, value := range []string{"-9223372036854775808", "9223372036854775807"} {
		assertExecOK(t, ctx, db, "INSERT INTO `user` (email, mobile) VALUES (?, ?)", "bigint-"+value+"@example.test", value)
		assertIntStringValue(t, ctx, db, "SELECT mobile FROM `user` WHERE email = ?", value, "bigint-"+value+"@example.test")
	}
	assertExecError(t, ctx, db, outOfRangeErr, "INSERT INTO `user` (email, mobile) VALUES (?, ?)", "bigint-under@example.test", "-9223372036854775809")
	assertExecError(t, ctx, db, outOfRangeErr, "INSERT INTO `user` (email, mobile) VALUES (?, ?)", "bigint-over@example.test", "9223372036854775808")
}

func assertPermittedSignedDomainValues(t *testing.T, ctx context.Context, db *sql.DB) {
	assertExecOK(t, ctx, db, "INSERT INTO product (pid, pquantity, pprice) VALUES (?, ?, ?)", "signed-product", -7, "-1.23")
	assertExecOK(t, ctx, db, "INSERT INTO orders (orderid, prodid, quantity, amount, shipped) VALUES (?, ?, ?, ?, ?)", "signed-order", "seed-product", -3, "-4.56", 2)
	product := queryProductValues(t, ctx, db, "signed-product")
	order := queryOrderValues(t, ctx, db, "signed-order", "seed-product")
	if product.PQuantity.Int64 != -7 || product.PPrice != "-1.23" {
		t.Fatalf("signed product values = %+v, want stock -7 and price -1.23", product)
	}
	if order.Quantity.Int64 != -3 || order.Amount != "-4.56" || order.Shipped != 2 {
		t.Fatalf("signed order values = %+v, want quantity -3 amount -4.56 shipped 2", order)
	}
}

func assertGeneratedIdentities(t *testing.T, ctx context.Context, db *sql.DB) {
	for _, table := range []string{"usercart", "user_demand", "transactions"} {
		t.Run(table, func(t *testing.T) {
			first := insertGeneratedRow(t, ctx, db, table, nil)
			second := insertGeneratedRow(t, ctx, db, table, nil)
			if first == second {
				t.Fatalf("%s generated duplicate ids %d and %d", table, first, second)
			}
			nullID := insertGeneratedRow(t, ctx, db, table, sql.NullInt64{})
			if nullID == 0 {
				t.Fatalf("%s NULL generated id = 0, want native generated nonzero id", table)
			}
			insertGeneratedRow(t, ctx, db, table, int64(0))
			insertGeneratedRow(t, ctx, db, table, int64(-9223372036854775808))
			insertGeneratedRow(t, ctx, db, table, int64(9223372036854775807))
			duplicateErr := insertGeneratedRowErr(ctx, db, table, int64(0))
			assertExpectedSQLError(t, duplicateErr, duplicateErrSpec)
			overflowErr := insertGeneratedRowErr(ctx, db, table, "9223372036854775808")
			assertExpectedSQLError(t, overflowErr, outOfRangeErr)
		})
	}
}

func assertQuantityChecksAndTimestamps(t *testing.T, ctx context.Context, db *sql.DB) {
	before := time.Now().UTC().Add(-2 * time.Second)
	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid) VALUES (?, ?)", "seed@example.test", "quantity-default-product")
	after := time.Now().UTC().Add(2 * time.Second)
	assertTimestampBetween(t, ctx, db, "SELECT added_at FROM usercart WHERE email = ? AND pid = ?", before, after, "seed@example.test", "quantity-default-product")

	assertExecOK(t, ctx, db, "INSERT INTO usercart (email, pid, quantity, added_at) VALUES (?, ?, ?, ?)", "seed@example.test", "quantity-product", 2147483647, "2026-11-17 12:34:56")
	assertTimestampExact(t, ctx, db, "SELECT added_at FROM usercart WHERE email = ? AND pid = ?", "2026-11-17 12:34:56", "seed@example.test", "quantity-product")
	for _, value := range []any{0, -1} {
		assertExecError(t, ctx, db, checkErr, "INSERT INTO usercart (email, pid, quantity) VALUES (?, ?, ?)", "seed@example.test", "seed-product", value)
	}
	assertExecError(t, ctx, db, outOfRangeErr, "INSERT INTO usercart (email, pid, quantity) VALUES (?, ?, ?)", "seed@example.test", "seed-product", "2147483648")
	assertExecError(t, ctx, db, nullErr, "INSERT INTO usercart (email, pid, added_at) VALUES (?, ?, NULL)", "seed@example.test", "seed-product")

	before = time.Now().UTC().Add(-2 * time.Second)
	assertExecOK(t, ctx, db, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, ?)", "tx-time-default", "seed@example.test", "seed-product", 1)
	after = time.Now().UTC().Add(2 * time.Second)
	assertTimestampBetween(t, ctx, db, "SELECT created_at FROM transactions WHERE transaction_id = ?", before, after, "tx-time-default")
	assertExecOK(t, ctx, db, "INSERT INTO transactions (transaction_id, email, pid, quantity, created_at) VALUES (?, ?, ?, ?, ?)", "tx-time-explicit", "seed@example.test", "seed-product", 2147483647, "2026-11-17 23:45:01")
	assertTimestampExact(t, ctx, db, "SELECT created_at FROM transactions WHERE transaction_id = ?", "2026-11-17 23:45:01", "tx-time-explicit")
	for _, value := range []any{0, -1} {
		assertExecError(t, ctx, db, checkErr, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, ?)", "tx-check-"+fmt.Sprint(value), "seed@example.test", "seed-product", value)
	}
	assertExecError(t, ctx, db, outOfRangeErr, "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, ?)", "tx-check-over", "seed@example.test", "seed-product", "2147483648")
	assertExecError(t, ctx, db, nullErr, "INSERT INTO transactions (transaction_id, email, pid, quantity, created_at) VALUES (?, ?, ?, ?, NULL)", "tx-null-time", "seed@example.test", "seed-product", 1)
}

func assertDemandValues(t *testing.T, ctx context.Context, db *sql.DB) {
	assertExecOK(t, ctx, db, "INSERT INTO user_demand (email, book_name) VALUES (?, ?)", "seed@example.test", "same-book")
	assertExecOK(t, ctx, db, "INSERT INTO user_demand (email, book_name, author) VALUES (?, ?, NULL)", "seed@example.test", "same-book")
	assertExecOK(t, ctx, db, "INSERT INTO user_demand (email, book_name, author) VALUES (?, ?, ?)", "seed@example.test", "same-book", "same-author")
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM user_demand WHERE email = ? AND book_name = ?", "seed@example.test", 3, "same-book")
	var nullAuthors int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_demand WHERE book_name = ? AND author IS NULL", "same-book").Scan(&nullAuthors); err != nil {
		t.Fatalf("count demand null authors: %v", err)
	}
	if nullAuthors != 2 {
		t.Fatalf("null author demand rows = %d, want 2", nullAuthors)
	}
}

func assertHashCapacity(t *testing.T, ctx context.Context, db *sql.DB) {
	syntheticEncoded := "$argon2id$v=19$m=65536,t=3,p=4$" + strings.Repeat("s", 43) + "$" + strings.Repeat("h", 43)
	if len(syntheticEncoded) != 117 {
		t.Fatalf("synthetic encoded hash length = %d, want 117", len(syntheticEncoded))
	}
	boundary255 := strings.Repeat("h", 255)
	boundary256 := boundary255 + "h"
	assertExecOK(t, ctx, db, "INSERT INTO `user` (email, password_hash) VALUES (?, ?)", "hash-encoded@example.test", syntheticEncoded)
	assertStringValue(t, ctx, db, "SELECT password_hash FROM `user` WHERE email = ?", syntheticEncoded, "hash-encoded@example.test")
	assertExecOK(t, ctx, db, "INSERT INTO `user` (email, password_hash) VALUES (?, ?)", "hash-255@example.test", boundary255)
	assertStringValue(t, ctx, db, "SELECT password_hash FROM `user` WHERE email = ?", boundary255, "hash-255@example.test")
	assertExecError(t, ctx, db, tooLongErr, "INSERT INTO `user` (email, password_hash) VALUES (?, ?)", "hash-256@example.test", boundary256)
	assertCount(t, ctx, db, "SELECT COUNT(*) FROM `user` WHERE email = ?", "hash-256@example.test", 0)
}

type expectedSQLError struct {
	number uint16
	state  string
}

var (
	requiredDefaultErr = expectedSQLError{number: 1364, state: "HY000"}
	nullErr            = expectedSQLError{number: 1048, state: "23000"}
	tooLongErr         = expectedSQLError{number: 1406, state: "22001"}
	outOfRangeErr      = expectedSQLError{number: 1264, state: "22003"}
	duplicateErrSpec   = expectedSQLError{number: 1062, state: "23000"}
	fkChildErr         = expectedSQLError{number: 1452, state: "23000"}
	checkErr           = expectedSQLError{number: 3819, state: "HY000"}
)

func assertExecOK(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("exec %q with %v: %v", query, args, err)
	}
}

func assertExecError(t *testing.T, ctx context.Context, db *sql.DB, want expectedSQLError, query string, args ...any) {
	t.Helper()
	_, err := db.ExecContext(ctx, query, args...)
	assertExpectedSQLError(t, err, want)
}

func assertExpectedSQLError(t *testing.T, err error, want expectedSQLError) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected MySQL error %d/%s, got nil", want.number, want.state)
	}
	assertMySQLError(t, err, want.number, want.state)
}

func insertProduct(t *testing.T, ctx context.Context, db *sql.DB, pid string) {
	t.Helper()
	assertExecOK(t, ctx, db, "INSERT INTO product (pid) VALUES (?)", pid)
}

func insertUser(t *testing.T, ctx context.Context, db *sql.DB, email string) {
	t.Helper()
	assertExecOK(t, ctx, db, "INSERT INTO `user` (email) VALUES (?)", email)
}

func insertProductIfMissing(t *testing.T, ctx context.Context, db *sql.DB, pid string) {
	t.Helper()
	assertExecOK(t, ctx, db, "INSERT INTO product (pid) VALUES (?) ON DUPLICATE KEY UPDATE pid = pid", pid)
}

func insertUserIfMissing(t *testing.T, ctx context.Context, db *sql.DB, email string) {
	t.Helper()
	assertExecOK(t, ctx, db, "INSERT INTO `user` (email) VALUES (?) ON DUPLICATE KEY UPDATE email = email", email)
}

func oneArg(value string) []any {
	return []any{value}
}

func keyedArg(key string) func(string) []any {
	return func(value string) []any {
		return []any{key + "-" + shortValueKey(value), value}
	}
}

func keyedArgWithValueSecond(fixed string) func(string) []any {
	return func(value string) []any {
		return []any{value, fixed}
	}
}

func keyedArgWithTwoFixed(first, second string) func(string) []any {
	return func(value string) []any {
		return []any{value, first, second}
	}
}

func valueSecond(first string) func(string) []any {
	return func(value string) []any {
		return []any{first, value}
	}
}

func orderProductArg(value string) []any {
	return []any{"order-product-" + shortValueKey(value), value}
}

func orderAmountArg(value string) []any {
	return []any{"order-" + shortValueKey(value), "seed-product", value}
}

func transactionAmountArg(value string) []any {
	return []any{"tx-" + shortValueKey(value), "seed@example.test", "seed-product", value}
}

func demandAuthorArg(value string) []any {
	return []any{"seed@example.test", "book-" + shortValueKey(value), value}
}

func cartEmailArg(value string) []any {
	return []any{value, "seed-product"}
}

func cartProductArg(value string) []any {
	return []any{"seed@example.test", value}
}

func transactionEmailArg(value string) []any {
	return []any{"tx-email-" + shortValueKey(value), value, "seed-product"}
}

func transactionProductArg(value string) []any {
	return []any{"tx-product-" + shortValueKey(value), "seed@example.test", value}
}

func valueKey(args []any) any {
	if len(args) == 0 {
		return nil
	}
	return args[0]
}

func shortValueKey(value string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%03d", len([]rune(value))))
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r + ('a' - 'A'))
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		default:
			builder.WriteByte('x')
		}
		if builder.Len() >= 24 {
			break
		}
	}
	key := builder.String()
	if key == "" {
		key = "empty"
	}
	return key
}

func seedVarcharParents(t *testing.T, ctx context.Context, db *sql.DB, name string, value string) {
	t.Helper()
	switch name {
	case "usercart.email", "transactions.email":
		insertUserIfMissing(t, ctx, db, value)
	case "usercart.pid", "transactions.pid":
		insertProductIfMissing(t, ctx, db, value)
	}
}

func tableFromCountSQL(query string) string {
	switch {
	case strings.Contains(query, "product"):
		return "product"
	case strings.Contains(query, "`user`"):
		return "user"
	case strings.Contains(query, "orders"):
		return "orders"
	case strings.Contains(query, "usercart"):
		return "usercart"
	case strings.Contains(query, "user_demand"):
		return "user_demand"
	case strings.Contains(query, "transactions"):
		return "transactions"
	default:
		return ""
	}
}

func insertGeneratedRow(t *testing.T, ctx context.Context, db *sql.DB, table string, id any) int64 {
	t.Helper()
	result, err := execGeneratedRow(ctx, db, table, id)
	if err != nil {
		t.Fatalf("insert generated row into %s with id %v: %v", table, id, err)
	}
	if explicit, ok := id.(int64); ok {
		return explicit
	}
	got, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id for %s: %v", table, err)
	}
	return got
}

func insertGeneratedRowErr(ctx context.Context, db *sql.DB, table string, id any) error {
	_, err := execGeneratedRow(ctx, db, table, id)
	return err
}

var generatedSequence int

func execGeneratedRow(ctx context.Context, db *sql.DB, table string, id any) (sql.Result, error) {
	var query string
	var args []any
	token := fmt.Sprint(id)
	if id == nil {
		generatedSequence++
		token = fmt.Sprintf("auto-%d", generatedSequence)
	}
	token = shortValueKey(token)
	switch table {
	case "usercart":
		pid := "generated-" + token
		if _, err := db.ExecContext(ctx, "INSERT INTO product (pid) VALUES (?) ON DUPLICATE KEY UPDATE pid = pid", pid); err != nil {
			return nil, err
		}
		query = "INSERT INTO usercart (id, email, pid, quantity) VALUES (?, ?, ?, 1)"
		args = []any{id, "seed@example.test", pid}
	case "user_demand":
		query = "INSERT INTO user_demand (id, email, book_name) VALUES (?, ?, ?)"
		args = []any{id, "seed@example.test", "generated-" + token}
	case "transactions":
		query = "INSERT INTO transactions (id, transaction_id, email, pid, quantity) VALUES (?, ?, ?, ?, 1)"
		args = []any{id, "generated-" + token, "seed@example.test", "seed-product"}
	default:
		return nil, fmt.Errorf("unsupported generated table %s", table)
	}
	if id == nil {
		switch table {
		case "usercart":
			query = "INSERT INTO usercart (email, pid, quantity) VALUES (?, ?, 1)"
		case "user_demand":
			query = "INSERT INTO user_demand (email, book_name) VALUES (?, ?)"
		case "transactions":
			query = "INSERT INTO transactions (transaction_id, email, pid, quantity) VALUES (?, ?, ?, 1)"
		}
		args = args[1:]
	}
	return db.ExecContext(ctx, query, args...)
}

type productValues struct {
	PName     sql.NullString
	PType     sql.NullString
	PInfo     sql.NullString
	PPrice    string
	PQuantity sql.NullInt64
	Image     sql.NullString
}

func queryProductValues(t *testing.T, ctx context.Context, db *sql.DB, pid string) productValues {
	t.Helper()
	var values productValues
	var image []byte
	err := db.QueryRowContext(ctx, "SELECT pname, ptype, pinfo, CAST(pprice AS CHAR), pquantity, image FROM product WHERE pid = ?", pid).
		Scan(&values.PName, &values.PType, &values.PInfo, &values.PPrice, &values.PQuantity, &image)
	if err != nil {
		t.Fatalf("query product %q: %v", pid, err)
	}
	values.Image.Valid = image != nil
	return values
}

type userValues struct {
	Name              sql.NullString
	Mobile            sql.NullInt64
	Address           sql.NullString
	Pincode           sql.NullInt64
	PasswordHash      sql.NullString
	MustResetPassword int64
}

func queryUserValues(t *testing.T, ctx context.Context, db *sql.DB, email string) userValues {
	t.Helper()
	var values userValues
	if err := db.QueryRowContext(ctx, "SELECT name, mobile, address, pincode, password_hash, must_reset_password FROM `user` WHERE email = ?", email).
		Scan(&values.Name, &values.Mobile, &values.Address, &values.Pincode, &values.PasswordHash, &values.MustResetPassword); err != nil {
		t.Fatalf("query user %q: %v", email, err)
	}
	return values
}

type orderValues struct {
	Quantity sql.NullInt64
	Amount   string
	Shipped  int64
}

func queryOrderValues(t *testing.T, ctx context.Context, db *sql.DB, orderID, productID string) orderValues {
	t.Helper()
	var values orderValues
	if err := db.QueryRowContext(ctx, "SELECT quantity, CAST(amount AS CHAR), shipped FROM orders WHERE orderid = ? AND prodid = ?", orderID, productID).
		Scan(&values.Quantity, &values.Amount, &values.Shipped); err != nil {
		t.Fatalf("query order %q/%q: %v", orderID, productID, err)
	}
	return values
}

func assertCount(t *testing.T, ctx context.Context, db *sql.DB, query string, first any, want int64, rest ...any) {
	t.Helper()
	var got int64
	args := append([]any{first}, rest...)
	if err := db.QueryRowContext(ctx, query, args...).Scan(&got); err != nil {
		t.Fatalf("count query %q with %v: %v", query, args, err)
	}
	if got != want {
		t.Fatalf("count query %q with %v = %d, want %d", query, args, got, want)
	}
}

func tableCount(t *testing.T, ctx context.Context, db *sql.DB, table string) int64 {
	t.Helper()
	var got int64
	if table == "" {
		t.Fatalf("missing table for count")
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM `"+escapeIdentifier(table)+"`").Scan(&got); err != nil {
		t.Fatalf("count table %s: %v", table, err)
	}
	return got
}

func assertDecimalValue(t *testing.T, ctx context.Context, db *sql.DB, query string, want string, args ...any) {
	t.Helper()
	var got string
	if err := db.QueryRowContext(ctx, query, args...).Scan(&got); err != nil {
		t.Fatalf("decimal query %q with %v: %v", query, args, err)
	}
	if got != want {
		t.Fatalf("decimal query %q with %v = %q, want %q", query, args, got, want)
	}
}

func assertIntValue(t *testing.T, ctx context.Context, db *sql.DB, query string, want int64, args ...any) {
	t.Helper()
	var got int64
	if err := db.QueryRowContext(ctx, query, args...).Scan(&got); err != nil {
		t.Fatalf("int query %q with %v: %v", query, args, err)
	}
	if got != want {
		t.Fatalf("int query %q with %v = %d, want %d", query, args, got, want)
	}
}

func assertIntStringValue(t *testing.T, ctx context.Context, db *sql.DB, query string, want string, args ...any) {
	t.Helper()
	var got string
	if err := db.QueryRowContext(ctx, query, args...).Scan(&got); err != nil {
		t.Fatalf("int-string query %q with %v: %v", query, args, err)
	}
	if got != want {
		t.Fatalf("int-string query %q with %v = %q, want %q", query, args, got, want)
	}
}

func assertStringValue(t *testing.T, ctx context.Context, db *sql.DB, query string, want string, args ...any) {
	t.Helper()
	var got string
	if err := db.QueryRowContext(ctx, query, args...).Scan(&got); err != nil {
		t.Fatalf("string query %q with %v: %v", query, args, err)
	}
	if got != want {
		t.Fatalf("string query %q with %v = %q, want %q", query, args, got, want)
	}
}

func assertTimestampBetween(t *testing.T, ctx context.Context, db *sql.DB, query string, before time.Time, after time.Time, args ...any) {
	t.Helper()
	var got time.Time
	if err := db.QueryRowContext(ctx, query, args...).Scan(&got); err != nil {
		t.Fatalf("timestamp query %q with %v: %v", query, args, err)
	}
	if got.Before(before) || got.After(after) {
		t.Fatalf("timestamp query %q with %v = %s, want between %s and %s", query, args, got, before, after)
	}
}

func assertTimestampExact(t *testing.T, ctx context.Context, db *sql.DB, query string, want string, args ...any) {
	t.Helper()
	var got time.Time
	if err := db.QueryRowContext(ctx, query, args...).Scan(&got); err != nil {
		t.Fatalf("timestamp query %q with %v: %v", query, args, err)
	}
	if got.Format("2006-01-02 15:04:05") != want {
		t.Fatalf("timestamp query %q with %v = %s, want %s", query, args, got.Format("2006-01-02 15:04:05"), want)
	}
}

