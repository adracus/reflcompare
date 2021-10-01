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

package reflcompare_test

import (
	"fmt"
	"reflect"

	. "github.com/adracus/reflcompare"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func intPtr(i int) *int {
	return &i
}

type Struct struct {
	A int
	B *int
	C []int
	D map[int]int
	E error
	F *error
	G func()
}

var _ = Describe("Reflcompare", func() {
	Context("Comparisons", func() {
		var (
			c        Comparisons
			intArray [2]int
			intSlice []int
			s        Struct
			_        = s
			nilErr   error
			err      error
			errPtr   *error
			m        map[int]int
		)
		setup := func() {
			intArray = [2]int{1, 2}
			intSlice = intArray[:]
			err = fmt.Errorf("some err")
			errPtr = &err
			m = map[int]int{1: 1}
		}
		setup()
		BeforeEach(setup)

		DescribeTable("DeepCompare regular",
			func(c Comparisons, v1, v2 interface{}, expect int) {
				Expect(c.DeepCompare(v1, v2)).To(Equal(expect))
				Expect(c.DeepCompare(v2, v1)).To(Equal(-expect))
			},
			Entry("nil - not nil", c, (*int)(nil), intPtr(1), -1),
			Entry("not nil - nil", c, intPtr(1), (*int)(nil), 1),
			Entry("custom struct{A: 0, B: *1} == struct{A: 0, B: *2}", NewComparisonsOrDie(func(a, b Struct) int { return a.A - b.A }), Struct{B: intPtr(1)}, Struct{B: intPtr(2)}, 0),
			Entry("array1 == array2", c, [2]int{1, 2}, [2]int{1, 2}, 0),
			Entry("array1[1] < array2[1]", c, [2]int{1, 1}, [2]int{1, 2}, -1),
			Entry("array1[1] > array2[1]", c, [2]int{1, 2}, [2]int{1, 1}, 1),
			Entry("slice1(nil) == slice2(empty)", c, ([]int)(nil), []int{}, 0),
			Entry("slice1(empty) == slice2(nil)", c, []int{}, ([]int)(nil), 0),
			Entry("len(slice1) > len(slice2)", c, []int{1, 2}, []int{1}, 1),
			Entry("slice1(arrayx) == slice2(arrayx)", c, intSlice, intSlice, 0),
			Entry("slice1 == slice2", c, []int{1, 2}, []int{1, 2}, 0),
			Entry("slice1[1] < slice2[1]", c, []int{1, 1}, []int{1, 2}, -1),
			Entry("slice1[1] > slice2[1]", c, []int{1, 2}, []int{1, 1}, 1),
			Entry("iface1(nil) == iface2(nil) via ptr", c, &nilErr, &nilErr, 0),
			Entry("iface1(v) == iface2(v) via ptr", c, errPtr, errPtr, 0),
			Entry("*int == *int", c, intPtr(1), intPtr(1), 0),
			Entry("*int < *int", c, intPtr(1), intPtr(2), -1),
			Entry("struct == struct", c, Struct{}, Struct{}, 0),
			Entry("struct{A: 1} < struct{A: 2}", c, Struct{A: 1}, Struct{A: 2}, -1),
			Entry("map1(nil) == map2(nil)", c, (map[int]int)(nil), (map[int]int)(nil), 0),
			Entry("map1(empty) == map2(nil)", c, map[int]int{}, (map[int]int)(nil), 0),
			Entry("map1(nil) == map2(empty)", c, (map[int]int)(nil), map[int]int{}, 0),
			Entry("len(map1) < len(map2)", c, map[int]int{1: 1}, map[int]int{1: 1, 2: 2}, -1),
			Entry("map1 === map1", c, m, m, 0),
			Entry("map{1: 1} < map{1: 2}", c, map[int]int{1: 1}, map[int]int{1: 2}, -1),
			Entry("map1 == map2", c, map[int]int{1: 1}, map[int]int{1: 1}, 0),
			Entry("f1(nil) == f2(nil)", c, (func())(nil), (func())(nil), 0),
			Entry("f1(nil) < f2", c, (func())(nil), func() {}, -1),
			Entry("false < true", c, false, true, -1),
			Entry("false == false", c, false, false, 0),
			Entry("true == true", c, true, true, 0),
			Entry("uint1 < uint2", c, uint(1), uint(2), -1),
			Entry("uint1 == uint2", c, uint(1), uint(1), 0),
			Entry("uintptr1 < uintptr2", c, uintptr(1), uintptr(2), -1),
			Entry("uintptr1 == uintptr2", c, uintptr(1), uintptr(1), 0),
			Entry("uint81 < uint82", c, uint8(1), uint8(2), -1),
			Entry("uint81 == uint82", c, uint8(1), uint8(1), 0),
			Entry("uint161 < uint162", c, uint16(1), uint16(2), -1),
			Entry("uint161 == uint162", c, uint16(1), uint16(1), 0),
			Entry("uint321 < uint322", c, uint32(1), uint32(2), -1),
			Entry("uint321 == uint322", c, uint32(1), uint32(1), 0),
			Entry("uint641 < uint642", c, uint64(1), uint64(2), -1),
			Entry("uint641 == uint642", c, uint64(1), uint64(1), 0),
			Entry("int1 < int2", c, 1, 2, -1),
			Entry("int1 == int2", c, 1, 1, 0),
			Entry("int81 < int82", c, int8(1), int8(2), -1),
			Entry("int81 == int82", c, int8(1), int8(1), 0),
			Entry("int161 < int162", c, int16(1), int16(2), -1),
			Entry("int161 == int162", c, int16(1), int16(1), 0),
			Entry("int321 < int322", c, int32(1), int32(2), -1),
			Entry("int321 == int322", c, int32(1), int32(1), 0),
			Entry("int641 < int642", c, int64(1), int64(2), -1),
			Entry("int641 == int642", c, int64(1), int64(1), 0),
			Entry("float321 < float322", c, float32(1), float32(2), -1),
			Entry("float321 == float322", c, float32(1), float32(1), 0),
			Entry("float641 < float642", c, float64(1), float64(2), -1),
			Entry("float641 == float642", c, float64(1), float64(1), 0),
			Entry("string1 < string2", c, "a", "b", -1),
			Entry("string1 == string2", c, "a", "a", 0),
			Entry("fallback interface == interface", c, complex(1, 1), complex(1, 1), 0),
		)

		DescribeTable("DeepCompare panic",
			func(c Comparisons, v1, v2 interface{}) {
				Expect(func() {
					c.DeepCompare(v1, v2)
				}).To(Panic())
			},
			Entry("different types", c, 1, "foo"),
			Entry("two non-nil functions", c, func() {}, func() {}),
		)
	})

	Describe("AddFunc", func() {
		It("should add the function", func() {
			c := make(Comparisons)
			Expect(c.AddFunc(func(a, b int) int {
				return a*a - b*b
			})).To(Succeed())
			Expect(c[reflect.TypeOf(1)].Interface().(func(a, b int) int)(2, 1)).To(Equal(3))
		})

		It("should error if the given argument is no function", func() {
			c := make(Comparisons)
			Expect(c.AddFunc(1)).To(HaveOccurred())
		})
	})

	Describe("AddFuncs", func() {
		It("should add the function", func() {
			c := make(Comparisons)
			Expect(c.AddFuncs(func(a, b int) int {
				return a*a - b*b
			})).To(Succeed())
			Expect(c[reflect.TypeOf(1)].Interface().(func(a, b int) int)(2, 1)).To(Equal(3))
		})

		It("should error if the given argument is no function", func() {
			c := make(Comparisons)
			Expect(c.AddFuncs(1)).To(HaveOccurred())
		})
	})

	Describe("NewComparisons", func() {
		It("should create a new conversion with the given functions", func() {
			c, err := NewComparisons(func(a, b int) int {
				return a*a - b*b
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(c[reflect.TypeOf(1)].Interface().(func(a, b int) int)(2, 1)).To(Equal(3))
		})

		It("should error if the given argument is no function", func() {
			_, err := NewComparisons(1)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("NewComparisonsOrDie", func() {
		It("should create a new conversion with the given functions", func() {
			c := NewComparisonsOrDie(func(a, b int) int {
				return a*a - b*b
			})
			Expect(c[reflect.TypeOf(1)].Interface().(func(a, b int) int)(2, 1)).To(Equal(3))
		})

		It("should panic if the given argument is no function", func() {
			Expect(func() {
				NewComparisonsOrDie(1)
			}).To(Panic())
		})
	})
})
