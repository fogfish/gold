//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/gold
//

package gold_test

import (
	"encoding"
	"encoding/json"
	"testing"

	"github.com/fogfish/gold"
	"github.com/fogfish/gold/internal/testa"
	"github.com/fogfish/gold/internal/testb"
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

func TestRegistryTypeIdentity(t *testing.T) {
	it.Then(t).Should(
		it.Equal(gold.Register[testa.Thread]("alpha"), "alpha"),
		it.Equal(gold.Register[testb.Thread]("beta"), "beta"),
		it.Equal(gold.Schema[testa.Thread](), "alpha"),
		it.Equal(gold.Schema[testb.Thread](), "beta"),
		it.Equal(gold.Schema[*testa.Thread](), "alpha"),
	)

	it.Then(t).Should(
		it.Fail(func() { gold.Register[testa.Thread]("other") }).Contain("duplicate"),
	)

	s := gold.Schemas()
	it.Then(t).Should(
		it.Equal(s["github.com/fogfish/gold/internal/testa.Thread"], "alpha"),
		it.Equal(s["github.com/fogfish/gold/internal/testb.Thread"], "beta"),
	)
}

func TestImply(t *testing.T) {
	type A struct{}
	type B struct{}
	type I = gold.Imply[A, B]

	var (
		_ encoding.TextMarshaler   = I{}
		_ encoding.TextUnmarshaler = (*I)(nil)
	)

	a := gold.ToIRI[A]("a")
	b := gold.ToIRI[B]("b")

	t.Run("MarshalText", func(t *testing.T) {
		for _, tc := range []struct {
			in  I
			out string
		}{
			{I{}, ""},
			{I{Domain: a}, "a:a|"},
			{I{Domain: a, Range: b}, "a:a|b:b"},
			{I{Domain: a, Range: gold.Any[B]()}, "a:a|b:"},
			{I{Domain: gold.Any[A](), Range: b}, "a:|b:b"},
		} {
			got, err := tc.in.MarshalText()
			it.Then(t).Should(
				it.Nil(err),
				it.Equal(string(got), tc.out),
				it.Equal(tc.in.String(), tc.out),
			)
		}
	})

	t.Run("MarshalTextErrors", func(t *testing.T) {
		for _, in := range map[string]I{
			"empty domain":       {Range: b},
			"domain with |":      {Domain: "a:x|y"},
			"domain wrong":       {Domain: "x"},
			"range not normed":   {Domain: a, Range: "b"},
			"range wrong schema": {Domain: a, Range: "a:b"},
		} {
			got, err := in.MarshalText()
			it.Then(t).Should(
				it.True(err != nil),
				it.Equal(len(got), 0),
				it.Equal(in.String(), ""),
			)
		}
		_, err := I{Range: b}.MarshalText()
		it.Then(t).Should(it.True(err.Error() == "gold: imply with empty domain"))
		_, err = I{Domain: "a:x|y"}.MarshalText()
		it.Then(t).Should(it.True(err.Error() == `gold: imply domain must not contain "|"`))
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		for in, ex := range map[string]I{
			"":        {},
			"a:a|":    {Domain: a},
			"a:a|b:b": {Domain: a, Range: b},
			"a:|b:":   {Domain: gold.Any[A](), Range: gold.Any[B]()},
			// range with | survives, split on the first separator is used
			"a:a|b:x|y": {Domain: a, Range: "b:x|y"},
		} {
			out := I{Domain: "a:old"}
			err := out.UnmarshalText([]byte(in))
			it.Then(t).Should(
				it.Nil(err),
				it.Equal(out, ex),
			)
		}
	})

	t.Run("UnmarshalTextErrors", func(t *testing.T) {
		for _, in := range []string{"a:a", "|b:b", "|", "x:a|b:b", "a:a|x:b", "a:a|b"} {
			var out I
			it.Then(t).Should(
				it.True(out.UnmarshalText([]byte(in)) != nil),
				it.Equal(out, I{}),
			)
		}
	})

	t.Run("RoundTrip", func(t *testing.T) {
		for _, in := range []I{{}, {Domain: a}, {Domain: a, Range: b}, {Domain: a, Range: "b:x|y"}} {
			txt, err := in.MarshalText()
			var out I
			it.Then(t).Should(
				it.Nil(err),
				it.Nil(out.UnmarshalText(txt)),
				it.Equal(out, in),
			)
		}
	})

	t.Run("JSON", func(t *testing.T) {
		raw, err := json.Marshal(I{Domain: a, Range: b})
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(string(raw), `"a:a|b:b"`),
		)

		var out I
		it.Then(t).Should(
			it.Nil(json.Unmarshal(raw, &out)),
			it.Equal(out, I{Domain: a, Range: b}),
		)
		it.Then(t).Should(it.True(json.Unmarshal([]byte(`"x:a|b:b"`), &out) != nil))
	})

	t.Run("Alias", func(t *testing.T) {
		type Account struct{}
		type Story struct{}
		type ThreadID = gold.Imply[Account, Story]
		type Doc struct {
			K ThreadID `json:"k"`
		}
		raw, err := json.Marshal(Doc{K: gold.ImplyFrom(gold.ToIRI[Account]("x"), gold.ToIRI[Story]("y"))})
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(string(raw), `{"k":"account:x|story:y"}`),
		)
	})
}
