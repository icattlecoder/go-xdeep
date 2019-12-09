# go-xdeep

## Example

### Basic type

```go
 err := xdeep.Equal("a","a")
 // err == nil
 
 err := xdeep.Equal(1,"1")
 // err != nil
```

### map type

```go
m1 := map[string]interface{}{
    "int": 1,
    "string": "1",
    "intarray": []int{1, 2, 3},
}

m2 := map[string]interface{}{
    "string": "1",
    "int": 1,
    "intarray": []int{1, 2, 3},
}

err := xdeep.Equal(m1, m2)
// err == nil

m3 := map[string]interface{}{
    "string": "2",
    "int": 1,
    "intarray": []int{1, 2, 3},
}

err := xdeep.Equal(m1, m3)
// err != nil
```

### slice type

```go
arr1 := []int{1, 2}
arr2 := []int{1, 2}

err := xdeep.Equal(arr1, arr2)
// err == nil

arr2 := []int{2, 1}
err := xdeep.Equal(arr1, arr2)
// err != nil

opt := Option{
    IgnoreArrayOrder: map[string]bool{
        "": true,
    }
}

// ignore array order
arr2 := []int{2, 1}
err := xdeep.Equal(arr1, arr2, &opt)
// err == nil
```

### struct type

```go
type Foo struct {
	Name         string
	Arr          []int
}

f1 := Foo{"name", []int{1,2}}
f2 := Foo{"name", []int{1,2}}
err := xdeep.Equal(f1, f2)
// err == nil

f3 := Foo{"bar", []int{1,2}}
err := xdeep.Equal(f1, f3)
// err != nil

// ignore field

opt := Option{
    IgnoreFields: []string{"Name"},
}
err := xdeep.Equal(f1, f3, &opt)
// err == nil

```

### IEqual interface

```go
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
bar1 := EqualBar{ "bar1" }
bar2 := EqualBar{ "bar1 copy" }
err := xdeep.Equal(bar1, bar2, &opt)
// err == nil
```

### Combine
```go
type Foo struct {
	Name         string
	Arr          []int
	M            map[string]interface{}
	InterfaceArr []interface{}
	T            time.Time
}
t1 := time.Now()
t2 := t1
f1 := Foo{
	Name: "foo",
	Arr: []int{1, 3, 9},
	M: map[string]interface{}{
		"int":1,
		"string":"2",
		"foo": EqualBar{
			Name: "bar 1",
		},
	},
	T: t1,
}

f2 := Foo{
	Name: "foo",
	Arr: []int{9,3,1},
	M: map[string]interface{}{
		"int":1,
		"string":"!2",
		"foo": EqualBar{
			Name: "bar 1 copy",
		},
	},
	T: t2,
}

opt := Option{
	IgnoreArrayOrder: map[string]bool{
		"Arr": true,
	},
	IgnoreFields: []string{"M.string"},
}

err := xdeep.Equal(f1, f2, &opt)
// err == nil
```