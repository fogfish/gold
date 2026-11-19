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
)

func TestIRI(t *testing.T) {
	type T struct{}
	type I = gold.IRI[T]

	t.Run("ToIRI", func(t *testing.T) {
		for in, ex := range map[string]I{
			"":      I(""),
			"foo":   I("t:foo"),
			"t:bar": I("t:bar"),
			"x:foo": I("t:x:foo"),
		} {
			if got := gold.ToIRI[T](in); got != ex {
				t.Errorf("ToIRI(%q) = %q, want %q", in, got, ex)
			}
		}
	})

	t.Run("AsIRI", func(t *testing.T) {
		for in, ex := range map[string]I{
			"t:foo":   I("t:foo"),
			"t:bar":   I("t:bar"),
			"t:x:foo": I("t:x:foo"),
		} {
			if got, err := gold.AsIRI[T](in); err != nil || got != ex {
				t.Errorf("AsIRI(%q) = %q, %v; want %q, nil", in, got, err, ex)
			}
		}

		if _, err := gold.AsIRI[T]("invalid"); err == nil {
			t.Error("AsIRI(invalid) should return an error")
		}
	})

	t.Run("Norm", func(t *testing.T) {
		for in, ex := range map[I]I{
			I(""):        I(""),
			I("foo"):     I("t:foo"),
			I("t:bar"):   I("t:bar"),
			I("x:foo"):   I("t:x:foo"),
			I("t:x:foo"): I("t:x:foo"),
		} {
			if got := in.Norm(); got != ex {
				t.Errorf("(%q).Norm() = %q, want %q", in, got, ex)
			}
		}
	})

	t.Run("Validate", func(t *testing.T) {
		validCases := []string{
			"t:foo",
			"t:bar",
			"t:x:foo",
		}
		for _, tc := range validCases {
			iri := I(tc)
			if err := iri.Validate(); err != nil {
				t.Errorf("(%q).Validate() = %v, want nil", tc, err)
			}
		}

		invalidCases := []string{
			"invalid",
			"x:foo",
			"",
		}
		for _, tc := range invalidCases {
			iri := I(tc)
			if err := iri.Validate(); err == nil {
				t.Errorf("(%q).Validate() = nil, want error", tc)
			}
		}
	})

	t.Run("IsValid", func(t *testing.T) {
		for in, ex := range map[I]bool{
			I("t:foo"):   true,
			I("t:bar"):   true,
			I("t:x:foo"): true,
			I("invalid"): false,
			I("x:foo"):   false,
			I(""):        false,
		} {
			if got := in.IsValid(); got != ex {
				t.Errorf("(%q).IsValid() = %v, want %v", in, got, ex)
			}
		}
	})

	t.Run("IsEmpty", func(t *testing.T) {
		for in, ex := range map[I]bool{
			I(""):        true,
			I("t:foo"):   false,
			I("t:bar"):   false,
			I("invalid"): false,
		} {
			if got := in.IsEmpty(); got != ex {
				t.Errorf("(%q).IsEmpty() = %v, want %v", in, got, ex)
			}
		}
	})

	t.Run("Reference", func(t *testing.T) {
		for in, ex := range map[I]string{
			I("t:foo"):     "foo",
			I("t:bar"):     "bar",
			I("t:x:foo"):   "x:foo",
			I("t:complex"): "complex",
		} {
			if got := in.Reference(); got != ex {
				t.Errorf("(%q).Reference() = %q, want %q", in, got, ex)
			}
		}

		// Test panic case
		defer func() {
			if r := recover(); r == nil {
				t.Error("I(\"\").Reference() should panic")
			}
		}()
		I("").Reference()
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		for in, ex := range map[I]string{
			I(""):        `""`,
			I("t:foo"):   `"t:foo"`,
			I("t:bar"):   `"t:bar"`,
			I("t:x:foo"): `"t:x:foo"`,
		} {
			if got, err := in.MarshalJSON(); err != nil || string(got) != ex {
				t.Errorf("(%q).MarshalJSON() = %q, %v; want %q, nil", in, got, err, ex)
			}
		}
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		for in, ex := range map[string]I{
			`"t:foo"`:   I("t:foo"),
			`"t:bar"`:   I("t:bar"),
			`"t:x:foo"`: I("t:x:foo"),
			`""`:        I(""),
		} {
			var iri I
			if err := iri.UnmarshalJSON([]byte(in)); err != nil || iri != ex {
				t.Errorf("UnmarshalJSON(%q) = %q, %v; want %q, nil", in, iri, err, ex)
			}
		}

		// Test invalid JSON
		var iri I
		if err := iri.UnmarshalJSON([]byte(`"invalid"`)); err == nil {
			t.Error("UnmarshalJSON with invalid IRI should return an error")
		}

		// Test malformed JSON
		if err := iri.UnmarshalJSON([]byte(`invalid`)); err == nil {
			t.Error("UnmarshalJSON with malformed JSON should return an error")
		}
	})

}

func TestIRIText(t *testing.T) {
	type T struct{}
	type X struct{}
	type I = gold.IRI[T]

	var (
		_ encoding.TextMarshaler   = I("")
		_ encoding.TextUnmarshaler = (*I)(nil)
	)

	t.Run("MarshalText", func(t *testing.T) {
		for in, ex := range map[I]string{
			I(""):      "",
			I("t:foo"): "t:foo",
			I("t:"):    "t:",
		} {
			if got, err := in.MarshalText(); err != nil || string(got) != ex {
				t.Errorf("(%q).MarshalText() = %q, %v; want %q, nil", in, got, err, ex)
			}
		}
	})

	t.Run("MarshalTextRejectsInvalid", func(t *testing.T) {
		for _, in := range []I{"alice", "x:foo", "tt:foo"} {
			if _, err := in.MarshalText(); err == nil {
				t.Errorf("(%q).MarshalText() should fail", in)
			}
		}
		if _, err := gold.IRI[X]("t:foo").MarshalText(); err == nil {
			t.Error("wrong schema should be rejected by MarshalText")
		}
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		for in, ex := range map[string]I{
			"":      I(""),
			"t:foo": I("t:foo"),
			"t:":    I("t:"),
		} {
			iri := I("t:old")
			if err := iri.UnmarshalText([]byte(in)); err != nil || iri != ex {
				t.Errorf("UnmarshalText(%q) = %q, %v; want %q, nil", in, iri, err, ex)
			}
		}
	})

	t.Run("UnmarshalTextRejectsInvalid", func(t *testing.T) {
		for _, in := range []string{"alice", "x:foo", "blog:x"} {
			var iri I
			if err := iri.UnmarshalText([]byte(in)); err == nil {
				t.Errorf("UnmarshalText(%q) should fail", in)
			}
		}
	})

	t.Run("RoundTrip", func(t *testing.T) {
		in := gold.ToIRI[T]("bar")
		b, err := in.MarshalText()
		var out I
		if err != nil || out.UnmarshalText(b) != nil || in != out {
			t.Errorf("round trip failed: %q -> %q -> %q (%v)", in, b, out, err)
		}
	})

	t.Run("JSONEmpty", func(t *testing.T) {
		var out struct {
			ID I `json:"id"`
		}
		if err := json.Unmarshal([]byte(`{"id":""}`), &out); err != nil || out.ID != "" {
			t.Errorf("empty JSON IRI should decode: %v", err)
		}
	})
}

func TestAny(t *testing.T) {
	type User struct{}
	type Blog struct{}

	t.Run("Value", func(t *testing.T) {
		if got := gold.Any[User](); got != gold.IRI[User]("user:") {
			t.Errorf("Any[User]() = %q", got)
		}
	})

	t.Run("Validation", func(t *testing.T) {
		any := gold.Any[User]()
		if _, err := gold.AsIRI[User](string(any)); err != nil {
			t.Error(err)
		}
		if !any.IsValid() || any.IsEmpty() {
			t.Error("Any must be valid and not empty")
		}
		if b, err := any.MarshalText(); err != nil || string(b) != "user:" {
			t.Errorf("MarshalText = %q, %v", b, err)
		}
		var out gold.IRI[User]
		if err := out.UnmarshalText([]byte("user:")); err != nil || out != any {
			t.Errorf("UnmarshalText = %q, %v", out, err)
		}
	})

	t.Run("Reference", func(t *testing.T) {
		if got := gold.Any[User]().Reference(); got != "" {
			t.Errorf("Reference() = %q, want empty", got)
		}
		for _, in := range []gold.IRI[User]{"", "user", "x:"} {
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("(%q).Reference() should panic", in)
					}
				}()
				in.Reference()
			}()
		}
	})

	t.Run("ImplyRange", func(t *testing.T) {
		i := gold.ImplyFrom(gold.ToIRI[Blog]("1"), gold.Any[User]())
		b, err := i.MarshalText()
		if err != nil || string(b) != "blog:1|user:" {
			t.Errorf("MarshalText = %q, %v", b, err)
		}
	})
}
