package xdeep

import (
	"strings"
	"testing"
	"time"
)

type Foo struct {
	Name         string                 `json:"name"`
	Arr          []int                  `json:"-"`
	M            map[string]interface{} `json:""`
	InterfaceArr []interface{}
	T            time.Time
}

type EqualBar struct {
	Name string
}

func (e *EqualBar) Equal(b interface{}) bool {
	if b, ok := b.(*EqualBar); ok {
		if strings.Contains(e.Name, b.Name) || strings.Contains(b.Name, e.Name) {
			return true
		}
	}
	return false
}

func TestEqual(t *testing.T) {

	arr1 := []int{2, 1}
	arr2 := []int{1, 2}
	var nilFoo *Foo
	var nil2Foo *Foo
	f1 := Foo{
		Name: "n1",
		Arr:  arr1,
	}
	f2 := Foo{
		Name: "n1",
		Arr:  arr2,
	}

	m1 := map[string]interface{}{
		"a":   "haha",
		"b":   "haha",
		"foo": f1,
	}
	m2 := map[string]interface{}{
		"a": "haha",
		"b": "haha",
		"foo": Foo{
			Name: "n1",
			Arr:  []int{1, 2},
		},
	}
	_, _ = m1, m2
	_, _ = f1, f2

	t1 := time.Now()
	t2 := t1
	t3 := t2.UTC()

	cases := []struct {
		Expect interface{}
		Actual interface{}
		Err    string
		Opt    *Option
	}{
		{
			Expect: Foo{
				Name: "foo",
				Arr:  []int{1, 3, 2},
				M: map[string]interface{}{
					"k1": 1,
					"k2": "2",
					"k3": 0.01,
					"k4": t1,
				},
				InterfaceArr: []interface{}{
					"1", "2",
				},
			},
			Actual: Foo{
				Name: "foo",
				Arr:  []int{1, 3, 2},
				M: map[string]interface{}{
					"k1": 1,
					"k2": "2",
					"k3": 0.01,
					"k4": t2,
				},
				InterfaceArr: []interface{}{
					"2", "1",
				},
			},
			Opt: &Option{
				IgnoreFields: []string{},
				IgnoreArrayOrder: map[string]bool{
					"InterfaceArr": true,
				},
			},
			Err: "",
		},
		{
			Expect: "2",
			Actual: "2",
		},
		{
			Expect: &f1,
			Actual: &f1,
		},
		{
			Expect: &EqualBar{
				Name: "bar1",
			},
			Actual: &EqualBar{
				Name: "equal bar1",
			},
		},
		{
			Expect: &EqualBar{
				Name: "bar1",
			},
			Actual: &EqualBar{
				Name: "bar2",
			},
			Err: "not equal, compared via implement IEqual interface",
		},
		{
			Expect: &Foo{},
			Actual: nilFoo,
			Err:    "different IsNil, false vs true",
		},
		{
			Expect: nilFoo,
			Actual: &Foo{},
			Err:    "different IsNil, true vs false",
		},
		{
			Expect: nilFoo,
			Actual: nil2Foo,
		},
		{
			Expect: &Foo{},
			Actual: &Foo{},
		},
		{
			Expect: &Foo{},
			Actual: &Foo{Name: "foo"},
			Opt: &Option{
				IgnoreFields: []string{"Name"},
			},
		},
		{
			Expect: time.Now(),
			Actual: time.Now(),
			Err:    ": different value",
		},
		{
			Expect: t1,
			Actual: t2,
		},
		{
			Expect: map[string]string{},
			Actual: map[string]string{
				"foo": "bar",
			},
			Err: "different map key length, 0 vs 1",
		},
		{
			Expect: map[string]interface{}{
				"foo": 1,
			},
			Actual: map[string]interface{}{
				"foo": "1",
			},
			Err: "foo: different Type, int vs string",
		},
		{
			Expect: []string{"1"},
			Actual: []string{"1", "2"},
			Err:    "different array/slice length, 1 vs 2",
		},
		{
			Expect: []string{"2", "1"},
			Actual: []string{"1", "2"},
			Err:    "different value, 2 vs 1",
		},
		{
			Expect: []string{"1", "1"},
			Actual: []string{"1", "2"},
			Opt: &Option{
				IgnoreArrayOrder: map[string]bool{
					"": true,
				},
			},
			Err:    "[1]: 2 not found in [1 1]",
		},
		{
			Expect: []string{"1", "2"},
			Actual: []string{"1", "1"},
			Opt: &Option{
				IgnoreArrayOrder: map[string]bool{
					"": true,
				},
			},
			Err:    "[1]: 2 not found in [1 1]",
		},
		{
			Expect: m1,
			Actual: m2,
			Opt: &Option{
				IgnoreArrayOrder: map[string]bool{
					"foo.Arr": true,
				},
			},
		},
		{
			Expect: Foo{
				T: time.Now(),
			},
			Actual: Foo{
				T: time.Now(),
			},
			Err: "T: different value",
		},
		{
			Expect: Foo{
				Name: "foo",
				Arr:  []int{1, 3, 9},
				M: map[string]interface{}{
					"int":    1,
					"string": "2",
					"foo": &EqualBar{
						Name: "bar 1",
					},
				},
				T: t1,
			},
			Actual: Foo{
				Name: "foo",
				Arr:  []int{9, 3, 1},
				M: map[string]interface{}{
					"int":    1,
					"string": "!2",
					"foo": &EqualBar{
						Name: "bar 1 copy",
					},
				},
				T: t1,
			},
			Opt: &Option{
				IgnoreArrayOrder: map[string]bool{
					"Arr": true,
				},
				IgnoreFields: []string{"M.string"},
			},
		},
		{
			Expect: Foo{
				Name: "f1",
			},
			Actual: Foo{
				Name: "f2",
			},
			Opt: &Option{
				IgnoreFields:    []string{"name"},
				IgnoreByTagName: "json",
			},
		},
		{
			Expect: Foo{
				Arr: []int{1},
			},
			Actual: Foo{
				Arr: []int{2},
			},
			Opt: &Option{
				IgnoreByTagName: "json",
			},
		},
		{
			Expect: t2,
			Actual: t3,
			Opt: &Option{
				TimeEqual: "unixNano",
			},
		},
		{
			Expect: t2,
			Actual: t3,
			Err:    ": different value",
		},
		{
			Expect: time.Now(),
			Actual: time.Now().Add(1),
			Opt: &Option{
				TimeEqual: "unixNano",
			},
			Err: ": different unixNano time",
		},
	}

	for k, c := range cases {
		err := Equal(cases[k].Expect, cases[k].Actual, cases[k].Opt)
		if c.Err != "" {
			if err == nil {
				t.Fatalf("expected err:%v but got nil", c.Err)
			}
			if !strings.Contains(err.Error(), c.Err) {
				t.Fatalf("expected err `%v` contains `%v`", err.Error(), c.Err)
			}
		} else {
			if err != nil {
				t.Fatalf("expected err is nil, but got %v", err)
			}
		}
	}
}
