// Copyright 2021 Axel Christ
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package reflcompare provides utilities for comparing values.
package reflcompare

import (
	"fmt"
	"reflect"
	"strings"
)

// Comparisons enables comparing arbitrary values of the same type.
//
// It allows overriding comparison functions by supplying a custom function via
// Comparison.AddFunc or Comparison.AddFuncs.
type Comparisons map[reflect.Type]reflect.Value

// AddFuncs adds the given functions as a comparison functions.
// The functions have to have a signature of func(A, A) int where A can be any type.
// If any function does not match that signature, an error is returned.
func (c Comparisons) AddFuncs(funcs ...interface{}) error {
	for _, f := range funcs {
		if err := c.AddFunc(f); err != nil {
			return err
		}
	}
	return nil
}

// AddFunc adds the given function as a comparison function.
// The function has to have a signature of func(A, A) int where A can be any type.
// If the function does not match that signature, an error is returned.
func (c Comparisons) AddFunc(compFunc interface{}) error {
	fv := reflect.ValueOf(compFunc)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		return fmt.Errorf("expected func, got: %v", ft)
	}
	if ft.NumIn() != 2 {
		return fmt.Errorf("expected two 'in' params, got: %v", ft)
	}
	if ft.NumOut() != 1 {
		return fmt.Errorf("expected one 'out' param, got: %v", ft)
	}
	if ft.In(0) != ft.In(1) {
		return fmt.Errorf("expected arg 1 and 2 to have same type, but got %v", ft)
	}
	var forReturnType int
	intType := reflect.TypeOf(forReturnType)
	if ft.Out(0) != intType {
		return fmt.Errorf("expected bool return, got: %v", ft)
	}
	c[ft.In(0)] = fv
	return nil
}

// Below here is forked from go's reflect/deepequal.go

// During deepValueEqual, must keep track of checks that are
// in progress.  The comparison algorithm assumes that all
// checks in progress are true when it reencounters them.
// Visited comparisons are stored in a map indexed by visit.
type visit struct {
	a1  uintptr
	a2  uintptr
	typ reflect.Type
}

// unexportedTypePanic is thrown when you use this DeepEqual on something that has an
// unexported type. It indicates a programmer error, so should not occur at runtime,
// which is why it's not public and thus impossible to catch.
type unexportedTypePanic []reflect.Type

func (u unexportedTypePanic) Error() string { return u.String() }
func (u unexportedTypePanic) String() string {
	strs := make([]string, len(u))
	for i, t := range u {
		strs[i] = fmt.Sprintf("%v", t)
	}
	return "an unexported field was encountered, nested like this: " + strings.Join(strs, " -> ")
}

func makeUsefulPanic(v reflect.Value) {
	if x := recover(); x != nil {
		if u, ok := x.(unexportedTypePanic); ok {
			u = append(unexportedTypePanic{v.Type()}, u...)
			x = u
		}
		panic(x)
	}
}

func compareBool(b1, b2 bool) int {
	if b1 {
		if !b2 {
			return 1
		}
		return 0
	}
	if b2 {
		return -1
	}
	return 0
}

// deep compare values using reflected types. The map argument tracks
// comparisons that have already been seen, which allows short circuiting on
// recursive types.
func (c Comparisons) deepValueCompare(v1, v2 reflect.Value, visited map[visit]int, depth int) (res int) {
	defer makeUsefulPanic(v1)

	if !v1.IsValid() || !v2.IsValid() {
		return compareBool(v1.IsValid(), v2.IsValid())
	}
	if v1.Type() != v2.Type() {
		panic(fmt.Sprintf("cannot compare different types: %s - %s", v1.Type(), v2.Type()))
	}
	if fv, ok := c[v1.Type()]; ok {
		return int(fv.Call([]reflect.Value{v1, v2})[0].Int())
	}

	hard := func(k reflect.Kind) bool {
		switch k {
		case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
			return true
		}
		return false
	}

	if v1.CanAddr() && v2.CanAddr() && hard(v1.Kind()) {
		addr1 := v1.UnsafeAddr()
		addr2 := v2.UnsafeAddr()
		var swapped bool
		if addr1 > addr2 {
			// Canonicalize order to reduce number of entries in visited.
			addr1, addr2 = addr2, addr1
			swapped = true
		}

		// Short circuit if references are identical ...
		if addr1 == addr2 {
			return 0
		}

		// ... or already seen
		typ := v1.Type()
		v := visit{addr1, addr2, typ}
		if res, ok := visited[v]; ok {
			return res
		}

		defer func() {
			// Remember for later.
			cache := res
			if swapped {
				cache = -cache
			}
			visited[v] = cache
		}()
	}

	switch v1.Kind() {
	case reflect.Array:
		// We don't need to check length here because length is part of
		// an array's type, which has already been filtered for.
		for i := 0; i < v1.Len(); i++ {
			if res := c.deepValueCompare(v1.Index(i), v2.Index(i), visited, depth+1); res != 0 {
				return res
			}
		}
		return 0
	case reflect.Slice:
		if (v1.IsNil() || v1.Len() == 0) != (v2.IsNil() || v2.Len() == 0) {
			return 0
		}
		if res := v1.Len() - v2.Len(); res != 0 {
			return res
		}
		if v1.Pointer() == v2.Pointer() {
			return 0
		}
		for i := 0; i < v1.Len(); i++ {
			if res := c.deepValueCompare(v1.Index(i), v2.Index(i), visited, depth+1); res != 0 {
				return res
			}
		}
		return 0
	case reflect.Interface:
		if res := compareBool(!v1.IsNil(), !v2.IsNil()); res != 0 {
			return res
		}
		return c.deepValueCompare(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Ptr:
		return c.deepValueCompare(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if res := c.deepValueCompare(v1.Field(i), v2.Field(i), visited, depth+1); res != 0 {
				return res
			}
		}
		return 0
	case reflect.Map:
		if (v1.IsNil() || v1.Len() == 0) != (v2.IsNil() || v2.Len() == 0) {
			return 0
		}
		if res := v1.Len() - v2.Len(); res != 0 {
			return res
		}
		if v1.Pointer() == v2.Pointer() {
			return 0
		}
		for _, k := range v1.MapKeys() {
			if res := c.deepValueCompare(v1.MapIndex(k), v2.MapIndex(k), visited, depth+1); res != 0 {
				return res
			}
		}
		return 0
	case reflect.Func:
		if !v1.IsNil() && !v2.IsNil() {
			panic("cannot compare two non-nil functions")
		}
		return compareBool(!v1.IsNil(), !v2.IsNil())

	case reflect.Bool:
		return compareBool(v1.Bool(), v2.Bool())

	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return compareUInt64(v1.Uint(), v2.Uint())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return compareInt64(v1.Int(), v2.Int())

	case reflect.Float32, reflect.Float64:
		return compareFloat64(v1.Float(), v2.Float())

	case reflect.String:
		return strings.Compare(v1.String(), v2.String())

	default:
		// Normal equality suffices
		if !v1.CanInterface() || !v2.CanInterface() {
			panic(unexportedTypePanic{})
		}
		return compareInterface(v1.Interface(), v2.Interface())
	}
}

// compareInt64 compares two int64 values. We compare 'manually' to avoid any overflow.
func compareInt64(i1, i2 int64) int {
	if i1 < i2 {
		return -1
	}
	if i1 > i2 {
		return 1
	}
	return 0
}

func compareFloat64(f1, f2 float64) int {
	if f1 < f2 {
		return -1
	}
	if f1 > f2 {
		return 1
	}
	return 0
}

func compareUInt64(u1, u2 uint64) int {
	if u1 < u2 {
		return -1
	}
	if u1 > u2 {
		return 1
	}
	return 0
}

func compareInterface(v1, v2 interface{}) int {
	// utmost fallback: regular equality
	if v1 == v2 {
		return 0
	}
	panic(fmt.Sprintf("cannot compare values of type %T", v1))
}

// DeepCompare compares two values, traversing through them if they
// are complex data types.
//
// It will use c's comparison functions if it finds types that match.
//
// An empty slice *is* equal to a nil slice for our purposes; same for maps.
//
// Unexported field members cannot be compared and will cause an informative panic; you must add an Equality
// function for these types.
func (c Comparisons) DeepCompare(a1, a2 interface{}) int {
	if res := compareBool(a1 == nil, a2 == nil); res != 0 {
		return res
	}
	v1 := reflect.ValueOf(a1)
	v2 := reflect.ValueOf(a2)
	if v1.Type() != v2.Type() {
		panic(fmt.Sprintf("cannot compare different types: %T - %T", a1, a2))
	}
	return c.deepValueCompare(v1, v2, make(map[visit]int), 0)
}

// NewComparisons creates new Comparisons with the given functions added.
// If any of the given functions is *not* a comparison function, it errors.
func NewComparisons(funcs ...interface{}) (Comparisons, error) {
	c := make(Comparisons)
	if err := c.AddFuncs(funcs...); err != nil {
		return nil, err
	}
	return c, nil
}

// NewComparisonsOrDie creates new Comparisons with the given functions added.
// If any of the given functions is *not* a comparison function, it panics.
func NewComparisonsOrDie(funcs ...interface{}) Comparisons {
	c, err := NewComparisons(funcs...)
	if err != nil {
		panic(err)
	}
	return c
}
