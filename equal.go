package xdeep

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Equal Option
type Option struct {
	// Fields than in IgnoreFields will skip compare with each other.
	// This is useful when you compare with object which come from create api results, which auto create id, create_at, updated_at.
	IgnoreFields []string
	// ignore fields with tag, eg. `json`
	IgnoreByTagName string
	// If set true, compare ignore the items order, which means `[1, 2]` equals `[2, 1]`.
	// If compared items is top level, just set empty key to true.
	IgnoreArrayOrder map[string]bool
	// deep(default), or unixNano
	TimeEqual string

	ignoredPath map[string]struct{}
}

func joinPath(ps ...string) string {
	if len(ps) > 1 && ps[0] == "" && ps[1] == "." {
		return strings.Join(ps[2:], "")
	}
	return strings.Join(ps, "")
}

// If item implements IEqual interface, Equal will compare items with item.Equal.
type IEqual interface {
	Equal(b interface{}) bool
}

// Equal Compare `expect` with `actual`. they are equal when:
// 1. Both types are same.
// 2. Length of map's key must be equal, each  pair of (key,value) must be `Equal` to the other, except the key is in the IgnoreFields.
// 3. Length of array or slice must be `Equal`. Without IgnoreArrayOrder setting, each item in the array must be `Equal` one by one.
// 4. Each exportable fields in struct must be `Equal`.
// 5. Implemented IEqual method returns true.
// 6. Otherwise, including `Time` types, the `reflect.DeepEqual(expect, actual)` return `true`.
// If return error is not nil, it contains information which field and why they are not equal.
func Equal(expect interface{}, actual interface{}, opts ...*Option) error {

	if len(opts) > 1 {
		panic("only one Option")
	}

	var opt *Option
	if len(opts) == 1 {
		opt = opts[0]
	}

	if opt == nil {
		opt = &Option{}
	}

	opt.ignoredPath = make(map[string]struct{})
	for _, v := range opt.IgnoreFields {
		opt.ignoredPath[v] = struct{}{}
	}
	return equal(expect, actual, "", opt)
}

func equal(expect interface{}, actual interface{}, path string, opt *Option) error {

	if _, ok := opt.ignoredPath[path]; ok {
		return nil
	}

	expectT := reflect.TypeOf(expect)
	actualT := reflect.TypeOf(actual)

	if expectT != actualT {
		return fmt.Errorf("%s: different Type, %v vs %v", path, expectT, actualT)
	}

	if expect == nil && actual == nil {
		return nil
	}

	if (expect != nil && actual == nil) || (expect == nil && actual != nil) {
		return fmt.Errorf("%s: different IsNil, %v vs %v", path, expect, actual)
	}

	expectV := reflect.ValueOf(expect)
	actualV := reflect.ValueOf(actual)

	switch expectV.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map:

		if expectV.IsNil() && actualV.IsNil() {
			return nil
		} else if expectV.IsNil() && !actualV.IsNil() {
			return fmt.Errorf("%s: different IsNil, %v vs %v", path, expectV.IsNil(), actualV.IsNil())
		} else if !expectV.IsNil() && actualV.IsNil() {
			return fmt.Errorf("%s: different IsNil, %v vs %v", path, expectV.IsNil(), actualV.IsNil())
		}
		if expectV.Pointer() == actualV.Pointer() {
			return nil
		}
	}

	if e, ok := expect.(IEqual); ok {
		if ok := e.Equal(actual); ok {
			return nil
		}
		return fmt.Errorf("%s: not equal, compared via implement IEqual interface", path)
	}

	if expectT.Kind() == reflect.Ptr || actualT.Kind() == reflect.Ptr {
		if expectV.Type().Kind() == reflect.Ptr {
			expectV = expectV.Elem()
		}
		if actualV.Type().Kind() == reflect.Ptr {
			actualV = actualV.Elem()
		}
		return equal(expectV.Interface(), actualV.Interface(), path, opt)
	}

	if expectT.Kind() == reflect.Map {
		return equalMap(expectV, actualV, path, opt)
	}

	if expectT.Kind() == reflect.Slice || expectT.Kind() == reflect.Array {
		return equalSlice(expectV, actualV, path, opt)
	}

	if expectT == reflect.TypeOf(time.Time{}) {
		return equalTime(expect.(time.Time), actual.(time.Time), path, opt.TimeEqual)
	}

	if expectT.Kind() == reflect.Struct {
		return equalStruct(expectV, actualV, path, opt)
	}

	if !reflect.DeepEqual(expectV.Interface(), actualV.Interface()) {
		return fmt.Errorf("%s: different value, %v vs %v", path, expectV.Interface(), actualV.Interface())
	}
	return nil
}

func equalTime(expect, actual time.Time, path string, equal string) error {
	if equal == "unixNano" {
		t1 := expect.UnixNano()
		t2 := actual.UnixNano()
		if t1 != t2 {
			return fmt.Errorf("%s: different unixNano time, %v vs %v", path, t1, t2)
		}
		return nil
	}
	if !reflect.DeepEqual(expect, actual) {
		return fmt.Errorf("%s: different value, %v vs %v", path, expect, actual)
	}
	return nil
}

func equalMap(expectV, actualV reflect.Value, path string, opt *Option) error {

	expectKeys := expectV.MapKeys()
	actualKeys := actualV.MapKeys()
	if len(expectKeys) != len(actualKeys) {
		return fmt.Errorf("%s: different map key length, %v vs %v", path, len(expectKeys), len(actualKeys))
	}

	for _, key := range expectKeys {
		v1 := expectV.MapIndex(key)
		v2 := actualV.MapIndex(key)
		ipath := joinPath(path, ".", key.String())
		if !v1.IsValid() || !v2.IsValid() {
			return fmt.Errorf("%s: different map key isValid, %v vs %v", ipath, v1.IsValid(), v2.IsValid())
		}

		if err := equal(expectV.MapIndex(key).Interface(), actualV.MapIndex(key).Interface(), ipath, opt); err != nil {
			return err
		}
	}
	return nil
}

func equalStruct(expectV, actualV reflect.Value, path string, opt *Option) error {

	fieldNum := expectV.Type().NumField()
	t := expectV.Type()

	for i := 0; i < fieldNum; i++ {

		ipath := joinPath(path, ".", t.Field(i).Name)
		if opt.IgnoreByTagName != "" {
			tag, ok := t.Field(i).Tag.Lookup(opt.IgnoreByTagName)
			if ok {
				tag = strings.Split(tag, ",")[0]
				if tag == "-" {
					continue
				}
				if tag == "" {
					ipath = joinPath(path, ".", t.Field(i).Name)
				} else {
					ipath = joinPath(path, ".", tag)
				}
			}
		}

		if expectV.Field(i).CanInterface() {
			if err := equal(expectV.Field(i).Interface(), actualV.Field(i).Interface(), ipath, opt); err != nil {
				return err
			}
		}
	}
	return nil
}

func equalSlice(expectV, actualV reflect.Value, path string, opt *Option) error {

	size := expectV.Len()
	if size != actualV.Len() {
		return fmt.Errorf("%s: different array/slice length, %v vs %v", path, size, actualV.Len())
	}
	// `*` means ignore all path
	if !opt.IgnoreArrayOrder[path] && !opt.IgnoreArrayOrder["*"] {
		for i := 0; i < size; i++ {
			ipath := joinPath(path, "[", strconv.Itoa(i), "]")
			if err := equal(expectV.Index(i).Interface(), actualV.Index(i).Interface(), ipath, opt); err != nil {
				return err
			}
		}
	}

	compare := func(a, b reflect.Value) error {
		size := a.Len()
	L1:
		for i := 0; i < size; i++ {
			for j := 0; j < size; j++ {
				ipath := joinPath(path, "[", strconv.Itoa(i), "]")
				if err := equal(a.Index(i).Interface(), b.Index(j).Interface(), ipath, opt); err == nil {
					continue L1
				}
			}
			return fmt.Errorf("%s[%d]: %v not found in %v", path, i, a.Index(i).Interface(), b.Interface())
		}
		return nil
	}

	if err := compare(expectV, actualV); err != nil {
		return err
	}
	return compare(actualV, expectV)
}
