//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/gold
//

package gold_test

import (
	"testing"

	"github.com/fogfish/gold"
	"github.com/fogfish/it/v2"
)

func TestSchema(t *testing.T) {
	type Person struct{}
	type Persons []Person
	type T[A any] struct{}

	t.Run("Pure Types", func(t *testing.T) {
		it.Then(t).Should(
			it.Equal(gold.Schema[Person](), "person"),
			it.Equal(gold.Schema[Persons](), "persons"),
		)
	})

	t.Run("Unknown Container Types", func(t *testing.T) {
		for _, f := range []func(){
			func() { gold.Schema[[]Person]() },
			func() { gold.Schema[map[string]Person]() },
			func() { gold.Schema[chan Person]() },
			func() { gold.Schema[func() Person]() },
			func() { gold.Schema[T[Person]]() },
		} {
			it.Then(t).Should(
				it.Fail(f).Contain("unknown schema for"),
			)
		}
	})

	t.Run("Register Schema", func(t *testing.T) {
		gold.Register[T[string]]("tstring")

		it.Then(t).Should(
			it.Equal(gold.Schema[T[string]](), "tstring"),
		)
	})
}
