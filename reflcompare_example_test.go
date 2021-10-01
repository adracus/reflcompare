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
	"github.com/adracus/reflcompare"
	"strings"
)

func ExampleComparisons_DeepCompare() {
	type Pet string
	const (
		Cat Pet = "Cat"
		Dog Pet = "Dog"
		Bat Pet = "Bat"
	)
	petAdorability := map[Pet]int{
		// Everybody has to like cats :D
		Cat: 10,
		Dog: 9,
		// Who wants to have a bat in their house (except Batman)?
		Bat: 0,
	}

	// Lexicographically, dogs are above cats
	fmt.Println("cat vs dog:", strings.Compare(string(Cat), string(Dog)))

	comparePets := func(p1, p2 Pet) int {
		return petAdorability[p1] - petAdorability[p2]
	}
	c := reflcompare.NewComparisonsOrDie(comparePets)

	fmt.Println("cat vs dog (adorability considered):", c.DeepCompare(Cat, Dog))
	// Output:
	// cat vs dog: -1
	// cat vs dog (adorability considered): 1
}
