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

package goutil

import (
	`bytes`
	`fmt`
	`encoding/csv`
	`encoding/json`
	`os`
	`reflect`
	`strings`
)

const (
	NameIx int = 0
	ValueIx int = 1
)

// ObjectToCSV converts a single-tier struct to a string suitable for writing
// to a CSV file. For the csv package, we need to rearranage the elements from
// an ordered list of {{name, value}, {name, value}, ...} to an ordered list
// of {{name, name, ...}, {value, value, ...}}.
func ObjectToCSV (t interface{}) (b []byte, err error) {

	var ssi [][]string

	if ssi, err = ObjecToSlice(t, `csv`); err == nil {

		ss := make([][]string, 2)

		for _, si := range ssi {
			 ss[NameIx] = append(ss[NameIx], si[NameIx])
			 ss[ValueIx] = append(ss[ValueIx], si[ValueIx])
		}

		bb := new(bytes.Buffer)
		cw := csv.NewWriter(bb)
		cw.WriteAll(ss)

		b, err = bb.Bytes(), cw.Error()
	}

	return b, err
}

// ObjectToNVP converts a single-tier struct to a string containing name-
// value pairs separated by newlines.
func ObjectToNVP (t interface{}) (b []byte, err error) {

	var ssi [][]string

	if ssi, err = ObjecToSlice(t, `nvp`); err == nil {

		var s string

		for _, si := range ssi {
			s += fmt.Sprintf("%s:%s\n", si[NameIx], si[ValueIx])
		}

		b = []byte(s)
	}

	return b, err
}

// SaveObject persists an object to a JSON file.
func SaveObject(t interface{}, fn string) (err error) {

	fh, err := os.Create(fn)
	defer fh.Close()

	if err == nil {
		je := json.NewEncoder(fh)
		err = je.Encode(&t)
	}

	return err
}

// RestoreObject restores an object from a JSON file.
func RestoreObject(fn string, t interface{}) (err error) {

	fh, err := os.Open(fn)
	defer fh.Close()

	if err == nil {
		jd := json.NewDecoder(fh)
		err = jd.Decode(&t)
	}

	return err
}

// CompareObjects compares the field count, order, names, and values of two
// structs objects. If the field count or order is different, the structs are
// not comparable and the function returns an error. If the structs differ only
// in field values, the function returns a list of differences. Fields can be 
// renamed using tags (tid) and omitted with a tag value of '-'.
func CompareObjects(t1, t2 interface{}, tid string) (ss[][]string, err error) {

	if reflect.DeepEqual(t1, t2) {
		return ss, err
	}

	var (
		st1, st2 [][]string
		lt1, lt2 int
	)

	if st1, err = ObjecToSlice(t1, tid); err != nil {
		return ss, err
	}

	if st2, err = ObjecToSlice(t2, tid); err != nil {
		return ss, err
	}

	if lt1, lt2 = len(st1), len(st2); lt1 != lt2 {
		err = fmt.Errorf(`field count: %d != %d`, lt1, lt2)
		return ss, err
	}

	for i := 0; i < lt1; i++ {

		if st1[i][NameIx] != st2[i][NameIx] {
			err = fmt.Errorf(`field name %d: %q != %q`, i, st1[i][NameIx], st2[i][NameIx])
			return ss, err
		}

		if st1[i][ValueIx] != st2[i][ValueIx] {
			ss = append(ss, []string{st1[i][NameIx], st1[i][ValueIx], st2[i][ValueIx]})
		}
	}

	return ss, err
}

// ObjecToSlice converts a single-tier struct into a slice of slices in the
// form {{name, value}, {name, value}, ...} for consumption by other methods.
// The outer slice maintains the fields in the same order as the struct. The
// tag parameter is the name of the struct tag to use for special processing.
// The primary purpose of this function is to offload tag processing for other
// functions.
func ObjecToSlice(t interface{}, tid string) (ss[][]string, err error) {

	var v reflect.Value

	if v = reflect.ValueOf(t); v.Type().Kind() != reflect.Struct {
		v = reflect.ValueOf(t).Elem()
	}

	if v.Type().Kind() != reflect.Struct {
		err = fmt.Errorf(`kind %q is not %q`, v.Type().Kind().String(), `struct`)
		return ss, err
	}

	Outer: for i := 0; i < v.NumField(); i++ {

		f := v.Field(i)
		t := v.Type().Field(i)

		if !f.IsValid() || !f.CanAddr() || !f.CanInterface() {
			continue
		}

		// Process field tags. Function follows the same tag
		// rules as encoding/xml and encoding/json, but only
		// support modified field names and the '-' option.

		fname := t.Name

		if tag, ok := t.Tag.Lookup(tid); ok {

			tval := strings.Split(tag, `,`)

			switch {
			case tval[0] == `-`:
				continue Outer
			case tval[0] == ``:
				fname = t.Name
			default:
				fname = tval[0]
			}
		}

		fval := fmt.Sprintf(`%v`, f.Interface())
		ss = append(ss, []string{fname, fval})
	}

	return ss, err
}
