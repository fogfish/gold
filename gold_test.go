//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/gold
//

package gold_test

import (
	"encoding/json"
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

func TestToIRI(t *testing.T) {
	type Person struct{}

	it.Then(t).Should(
		it.Equal(gold.ToIRI[Person]("123"), "123"),
	)
}

func TestAsIRI(t *testing.T) {
	type Person struct{}

	t.Run("Valid IRI", func(t *testing.T) {
		id, err := gold.AsIRI[Person]("person:123")
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(id, gold.IRI[Person]("123")),
			it.Equal(id, "123"),
		)
	})

	t.Run("Empty IRI", func(t *testing.T) {
		id, err := gold.AsIRI[Person]("person:")
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(id, gold.IRI[Person]("")),
			it.Equal(id, ""),
		)
	})

	t.Run("Invalid IRI", func(t *testing.T) {
		it.Then(t).Should(
			it.Error(gold.AsIRI[Person]("strings:123")).Contain("invalid IRI"),
			it.Error(gold.AsIRI[Person]("person")).Contain("invalid IRI"),
			it.Error(gold.AsIRI[Person]("")).Contain("invalid IRI"),
		)
	})
}

func TestJsonify(t *testing.T) {
	type Person struct {
		ID   gold.IRI[Person] `json:"id"`
		Name string           `json:"name"`
	}

	t.Run("Encode", func(t *testing.T) {
		p := Person{
			ID:   gold.IRI[Person]("123"),
			Name: "John Doe",
		}

		b, err := json.Marshal(p)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(string(b), `{"id":"person:123","name":"John Doe"}`),
		)
	})

	t.Run("Decode", func(t *testing.T) {
		var p Person

		err := json.Unmarshal([]byte(`{"id":"person:123","name":"John Doe"}`), &p)

		it.Then(t).Should(
			it.Nil(err),
			it.Equal(p.ID, gold.IRI[Person]("123")),
			it.Equal(p.Name, "John Doe"),
		)
	})
}
