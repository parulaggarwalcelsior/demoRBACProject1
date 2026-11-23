package schema_test

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type record map[string]any

func base(table string) record {
	switch table {
	case "product":
		return record{"pid": "new-product"}
	case "user":
		return record{"email": "new-user"}
	case "orders":
		return record{"orderid": "order", "prodid": "p"}
	case "usercart":
		return record{"email": "u", "pid": "p"}
	case "user_demand":
		return record{"email": "u", "book_name": "request"}
	case "transactions":
		return record{"transaction_id": "transaction", "email": "u", "pid": "p", "quantity": 1}
	}
	panic("unknown table")
}

func insertSQL(table string, r record) (string, []any) {
	var keys []string
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	marks, values, quoted := make([]string, len(keys)), make([]any, len(keys)), make([]string, len(keys))
	for i, key := range keys {
		marks[i], values[i], quoted[i] = "?", r[key], "`"+key+"`"
	}
	return "INSERT INTO `" + table + "` (" + strings.Join(quoted, ",") + ") VALUES (" + strings.Join(marks, ",") + ")", values
}

func (f *fixture) insert(table string, r record) {
	f.t.Helper()
	q, args := insertSQL(table, r)
	result, err := f.db.Exec(q, args...)
	if err != nil {
		f.t.Fatalf("valid %s write rejected: %s", table, f.safe(err))
	}
	if table == "usercart" || table == "user_demand" || table == "transactions" {
		if v, exists := r["id"]; !exists || v == nil {
			id, err := result.LastInsertId()
			if err != nil {
				f.t.Fatal(err)
			}
			r["id"] = id
		}
	}
}

func (f *fixture) value(table string, r record, column, want string) {
	f.t.Helper()
	key := "id"
	switch table {
	case "product":
		key = "pid"
	case "user":
		key = "email"
	case "orders":
		key = "orderid"
	}
	f.scalar("SELECT `"+column+"` FROM `"+table+"` WHERE `"+key+"`=?", want, r[key])
}

func (f *fixture) badInsert(table string, r record, codes []uint16, state string) {
	f.t.Helper()
	q, args := insertSQL(table, r)
	f.reject(codes, state, q, args...)
}

func TestSchemaValues(t *testing.T) {
	f := schemaFixture(t)
	run := func(name string, fn func()) {
		t.Run(name, func(t *testing.T) { f.t = t; f.reset(); f.parents(); fn() })
	}
	for table, expected := range map[string]map[string]string{
		"product":      {"pname": "<NULL>", "ptype": "<NULL>", "pinfo": "<NULL>", "pquantity": "<NULL>", "image": "<NULL>", "pprice": "0.00"},
		"user":         {"name": "<NULL>", "mobile": "<NULL>", "address": "<NULL>", "pincode": "<NULL>", "password_hash": "<NULL>", "must_reset_password": "1"},
		"orders":       {"quantity": "<NULL>", "amount": "0.00", "shipped": "0"},
		"usercart":     {"quantity": "1"},
		"user_demand":  {"author": "<NULL>"},
		"transactions": {"amount": "0.00"},
	} {
		run(table+"/omitted-defaults", func() {
			r := base(table)
			f.insert(table, r)
			for column, want := range expected {
				f.value(table, r, column, want)
			}
		})
	}
	for table, fields := range map[string][]string{
		"product": {"pname", "ptype", "pinfo", "pquantity", "image"},
		"user":    {"name", "mobile", "address", "pincode", "password_hash"},
		"orders":  {"quantity"}, "user_demand": {"author"},
	} {
		for _, column := range fields {
			run(table+"/explicit-null/"+column, func() {
				r := base(table)
				r[column] = nil
				f.insert(table, r)
				f.value(table, r, column, "<NULL>")
			})
		}
	}
	for table, fields := range map[string][]string{
		"product": {"pid", "pprice"}, "user": {"email", "must_reset_password"},
		"orders":       {"orderid", "prodid", "amount", "shipped"},
		"usercart":     {"email", "pid", "quantity", "added_at"},
		"user_demand":  {"email", "book_name", "requested_at"},
		"transactions": {"transaction_id", "email", "pid", "quantity", "amount", "created_at"},
	} {
		for _, column := range fields {
			run(table+"/reject-null/"+column, func() {
				r := base(table)
				r[column] = nil
				f.badInsert(table, r, []uint16{1048}, "23000")
			})
		}
	}
	for table, fields := range map[string][]string{
		"product": {"pid"}, "user": {"email"}, "orders": {"orderid", "prodid"},
		"usercart": {"email", "pid"}, "user_demand": {"email", "book_name"},
		"transactions": {"transaction_id", "email", "pid", "quantity"},
	} {
		for _, column := range fields {
			run(table+"/reject-omission/"+column, func() {
				r := base(table)
				delete(r, column)
				f.badInsert(table, r, []uint16{1364}, "HY000")
			})
		}
	}
	// All VARCHAR limits, in characters rather than bytes. Referenced values get
	// real exact-limit parents; overlength cases must fail for length, not FK.
	for table, fields := range columns {
		for column, def := range fields {
			if def.kind != "varchar" {
				continue
			}
			if table == "orders" && column == "prodid" {
				continue
			} // Separate FK/length case below.
			limit, _ := strconv.Atoi(def.length)
			for _, alphabet := range []string{"x", "📚"} {
				run(table+"/length/"+column+"/"+alphabet, func() {
					value := strings.Repeat(alphabet, limit)
					if column == "password_hash" {
						value = syntheticHash(limit)
					}
					r := base(table)
					r[column] = value
					if column == "email" && table != "user" {
						f.exec("INSERT INTO user(email) VALUES (?)", value)
					}
					if column == "pid" && table != "product" {
						f.exec("INSERT INTO product(pid) VALUES (?)", value)
					}
					f.insert(table, r)
					f.value(table, r, column, value)
					f.reset()
					f.parents()
					r = base(table)
					r[column] = value + alphabet
					f.badInsert(table, r, []uint16{1406}, "22001")
				})
			}
		}
	}
	run("orders/product-reference-length", func() {
		parent := strings.Repeat("📚", 30)
		f.exec("INSERT INTO product(pid) VALUES (?)", parent)
		r := base("orders")
		r["prodid"] = parent
		f.insert("orders", r)
		f.value("orders", r, "prodid", parent)
		for _, n := range []int{31, 45} {
			r = base("orders")
			r["prodid"] = strings.Repeat("x", n)
			f.badInsert("orders", r, []uint16{1452}, "23000")
		}
		r = base("orders")
		r["prodid"] = strings.Repeat("x", 46)
		f.badInsert("orders", r, []uint16{1406}, "22001")
	})
	for table, column := range map[string]string{"product": "pprice", "orders": "amount", "transactions": "amount"} {
		for _, value := range []string{"0.00", "12.34", "99999999.99", "-99999999.99"} {
			run(table+"/money/"+value, func() { r := base(table); r[column] = value; f.insert(table, r); f.value(table, r, column, value) })
		}
		for _, value := range []string{"100000000.00", "-100000000.00"} {
			run(table+"/money-overflow/"+value, func() { r := base(table); r[column] = value; f.badInsert(table, r, []uint16{1264}, "22003") })
		}
		for _, value := range []struct{ input, want string }{{"12.345", "12.35"}, {"-12.345", "-12.35"}} {
			run(table+"/environment-rounding/"+value.input, func() {
				r := base(table)
				r[column] = value.input
				f.insert(table, r)
				// Characterize warnings immediately, before a SELECT replaces them.
				warnings := f.rows("SHOW WARNINGS")
				for _, w := range warnings {
					if w[1] != "1265" {
						f.t.Errorf("unexpected rounding warning code %s", w[1])
					}
				}
				f.value(table, r, column, value.want)
			})
		}
	}
	for _, tc := range []struct{ table, column, min, max, below, above string }{
		{"product", "pquantity", "-2147483648", "2147483647", "-2147483649", "2147483648"},
		{"orders", "quantity", "-2147483648", "2147483647", "-2147483649", "2147483648"},
		{"orders", "shipped", "-2147483648", "2147483647", "-2147483649", "2147483648"},
		{"user", "pincode", "-2147483648", "2147483647", "-2147483649", "2147483648"},
		{"user", "mobile", "-9223372036854775808", "9223372036854775807", "-9223372036854775809", "9223372036854775808"},
	} {
		for _, value := range []string{tc.min, tc.max, "0", "-1", "2"} {
			run(tc.table+"/signed/"+tc.column+"/"+value, func() {
				r := base(tc.table)
				r[tc.column] = value
				f.insert(tc.table, r)
				f.value(tc.table, r, tc.column, value)
			})
		}
		for _, value := range []string{tc.below, tc.above} {
			run(tc.table+"/overflow/"+tc.column+"/"+value, func() { r := base(tc.table); r[tc.column] = value; f.badInsert(tc.table, r, []uint16{1264}, "22003") })
		}
	}
	for _, table := range []string{"usercart", "transactions"} {
		for _, value := range []string{"1", "2147483647"} {
			run(table+"/quantity/"+value, func() {
				r := base(table)
				r["quantity"] = value
				f.insert(table, r)
				f.value(table, r, "quantity", value)
			})
		}
		for _, value := range []string{"0", "-1", "-2147483648"} {
			run(table+"/quantity-check/"+value, func() { r := base(table); r["quantity"] = value; f.badInsert(table, r, []uint16{3819}, "HY000") })
		}
		run(table+"/quantity-overflow", func() { r := base(table); r["quantity"] = "2147483648"; f.badInsert(table, r, []uint16{1264}, "22003") })
		run(table+"/quantity-update-check", func() {
			r := base(table)
			f.insert(table, r)
			f.reject([]uint16{3819}, "HY000", "UPDATE `"+table+"` SET quantity=0 WHERE id=?", r["id"])
		})
	}
	for table, stamp := range map[string]string{"usercart": "added_at", "user_demand": "requested_at", "transactions": "created_at"} {
		run(table+"/timestamp-default", func() {
			r := base(table)
			before := f.rows("SELECT DATE_FORMAT(NOW(),'%Y-%m-%d %H:%i:%s')")[0][0]
			f.insert(table, r)
			after := f.rows("SELECT DATE_FORMAT(NOW(),'%Y-%m-%d %H:%i:%s')")[0][0]
			got := f.rows("SELECT DATE_FORMAT(`"+stamp+"`,'%Y-%m-%d %H:%i:%s') FROM `"+table+"` WHERE id=?", r["id"])[0][0]
			if got < before || got > after {
				f.t.Fatal("timestamp default outside database-clock bounds")
			}
		})
		run(table+"/timestamp-explicit", func() {
			r := base(table)
			r[stamp] = "2001-02-03 04:05:06"
			f.insert(table, r)
			f.value(table, r, stamp, "2001-02-03 04:05:06")
			field, value := "quantity", any(2)
			if table == "user_demand" {
				field, value = "author", "different"
			}
			f.exec("UPDATE `"+table+"` SET `"+field+"`=? WHERE id=?", value, r["id"])
			f.value(table, r, stamp, "2001-02-03 04:05:06")
		})
	}
	run("product/binary-image", func() {
		r := base("product")
		r["image"] = []byte{0, 255, 0, 1, 128, 39, 92}
		f.insert("product", r)
		f.scalar("SELECT HEX(image) FROM product WHERE pid=?", "00FF000180275C", r["pid"])
	})
	run("no-invented-identity-format", func() {
		f.exec("INSERT INTO product(pid) VALUES ('')")
		f.exec("INSERT INTO user(email) VALUES ('')")
		f.exec("INSERT INTO orders(orderid,prodid,quantity,shipped) VALUES ('','',-1,2)")
		f.exec("INSERT INTO usercart(email,pid) VALUES ('','')")
		f.exec("INSERT INTO user_demand(email,book_name,author) VALUES ('','','')")
		f.exec("INSERT INTO transactions(transaction_id,email,pid,quantity) VALUES ('','','',1)")
	})
	run("environment-binary-collation", func() {
		f.exec("INSERT INTO product(pid) VALUES ('Key'),('key'),('kéy'),('key ')")
		f.scalar("SELECT COUNT(*) FROM product WHERE pid IN ('Key','key','kéy','key ')", "4")
	})
}

func syntheticHash(n int) string {
	prefix := "$argon2id$v=19$m=65536,t=3,p=1$c3ludGhldGlj$"
	return prefix + strings.Repeat("A", n-len(prefix))
}

func TestGeneratedSignedBIGINTIdentities(t *testing.T) {
	// Explicit maximum ids may exhaust AUTO_INCREMENT even after DML rollback.
	// Each table's generator is used first, then endpoints/overflow run last.
	f := schemaFixture(t)
	f.parents()
	for _, table := range []string{"usercart", "user_demand", "transactions"} {
		t.Run(table, func(t *testing.T) {
			f.t = t
			makeRow := func(label string) record {
				r := base(table)
				if table == "usercart" {
					r["pid"] = label
					f.exec("INSERT INTO product(pid) VALUES (?)", label)
				}
				if table == "transactions" {
					r["transaction_id"] = label
				}
				return r
			}
			a, b := makeRow("generated-a"), makeRow("generated-b")
			b["id"] = nil
			f.insert(table, a)
			f.insert(table, b)
			if fmt.Sprint(a["id"]) == fmt.Sprint(b["id"]) || fmt.Sprint(a["id"]) == "0" || fmt.Sprint(b["id"]) == "0" {
				t.Fatal("omitted and NULL ids must generate distinct identities")
			}
			for i, id := range []string{"0", "-9223372036854775808", "9223372036854775807"} {
				r := makeRow(fmt.Sprintf("boundary-%d", i))
				r["id"] = id
				f.insert(table, r)
				f.value(table, r, "id", id)
				duplicate := makeRow(fmt.Sprintf("duplicate-%d", i))
				duplicate["id"] = id
				f.badInsert(table, duplicate, []uint16{1062}, "23000")
			}
			for i, id := range []string{"-9223372036854775809", "9223372036854775808"} {
				r := makeRow(fmt.Sprintf("overflow-%d", i))
				r["id"] = id
				f.badInsert(table, r, []uint16{1264}, "22003")
			}
		})
	}
}
