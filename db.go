// Copyright 2017 John Scherff
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goutils

import (
	"fmt"
	"reflect"
	"strings"
)

// ObjectDbCols is a convenience method that returns relevant and valid
// database column names for use in constructing SQL statements. Only
// columns representing fields that will not cause a panic are included.
// Columns are further filtered by the tag provided. To be included, the
// tag must exist and must not have a '-' value.
func ObjectDbCols(t interface{}, tid string) (cols []string, err error) {

	v := reflect.ValueOf(t).Elem()

	if v.Type().Kind() != reflect.Struct {
		err = fmt.Errorf("kind %q is not %q", v.Type().Kind().String(), "struct")
		return cols, err
	}

	Outer: for i := 0; i < v.NumField(); i++ {

		f := v.Field(i)
		t := v.Type().Field(i)

		if !f.IsValid() || !f.CanAddr() || !f.CanInterface() {
			continue
		}

		if tag, ok := t.Tag.Lookup(tid); ok {

			tval := strings.Split(tag, `,`)

			switch {
			case tval[0] == `-`:
				continue Outer
			case tval[0] == ``:
				cols = append(cols, t.Name)
			default:
				cols = append(cols, tval[0])
			}
		}
	}

	return cols, err
}

// ObjectDbSQL generates basic SQL statements from struct fields.
func ObjectDbSQL (cmd, tbl string, cols []string) (sql string, err error) {

	cmd = strings.ToUpper(cmd)
	tbl = strings.ToLower(tbl)

	switch cmd {
	case `SELECT`:
		sql = fmt.Sprintf(`SELECT %s FROM %s`,
			strings.Join(cols, `,`), tbl,
		)
	case `INSERT`:
		sql = fmt.Sprintf(`INSERT INTO %s (%s) VALUES (?%s)`,
			tbl, strings.Join(cols, `,`),
			strings.Repeat(`,?`, len(cols)-1),
		)
	default:
		err = fmt.Errorf(`invalid SQL command %s`, cmd)
	}

	return sql, err
}

// ObjectDbVals is a convenience method that returns an entire database
// row for simpler insert statements. Only fields that will not cause a
// panic are included. Fields are further filtered by the tag provided.
// To be included, the tag must exist and must not have a '-' value. The
// table or view DDL must have the same columns with the same names in
// the same order or the query will fail. Can be combined with the method
// ObjectDbCols for a less brittle solution.
func ObjectDbVals(t interface{}, tid string) (vals []interface{}, err error) {

	v := reflect.ValueOf(t).Elem()

	if v.Type().Kind() != reflect.Struct {
		err = fmt.Errorf("kind %q is not %q", v.Type().Kind().String(), "struct")
		return vals, err
	}

	for i := 0; i < v.NumField(); i++ {

		f := v.Field(i)
		t := v.Type().Field(i)

		if !f.IsValid() || !f.CanAddr() || !f.CanInterface() {
			continue
		}

		if tag, ok := t.Tag.Lookup(tid); ok {

			if strings.Split(tag, `,`)[0] == `-` {
				continue
			}

			vals = append(vals, f.Interface())
		}
	}

	return vals, err
}

// ObjectDbValsByCol is a convenience method that returns a database row
// based on the column names provided by the caller, which are ostensibly
// taken from the database table or view. The column names are matched to
// the tag on the struct field. An error is returned if a column is not
// found or if the associated field returns false for the IsValid(),
// CanAddr(), or CanInterface() reflect methods.
func ObjectDbValsByCol(t interface{}, tid string, cols []string) (vals []interface{}, err error) {

	v := reflect.ValueOf(t).Elem()

	if v.Type().Kind() != reflect.Struct {
		err = fmt.Errorf("kind %q is not %q", v.Type().Kind().String(), "struct")
		return vals, err
	}

	Outer: for _, col := range cols {

		for i := 0; i < v.NumField(); i++ {

			f := v.Field(i)
			t := v.Type().Field(i)

			if !f.IsValid() || !f.CanAddr() || !f.CanInterface() {
				continue
			}

			if tag, ok := t.Tag.Lookup(`db`); ok {

				if col == strings.Split(tag, `,`)[0] {
					vals = append(vals, f.Interface())
					continue Outer
				}
			}

		}

		return nil, fmt.Errorf(`no match for column %q`, col)
	}

	return vals, err
}
