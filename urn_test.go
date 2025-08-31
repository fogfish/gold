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
)

func TestURN(t *testing.T) {
	type N struct{}
	type T struct{}
	type U = gold.URN[N, T]

	t.Run("ToURN", func(t *testing.T) {
		testCases := []struct {
			input    []string
			expected U
		}{
			{[]string{}, U("urn:n:t")},
			{[]string{"foo"}, U("urn:n:t:foo")},
			{[]string{"foo", "bar"}, U("urn:n:t:foo:bar")},
			{[]string{"urn:n:t:existing"}, U("urn:n:t:existing")},
			{[]string{"urn:n:t:existing", "bar"}, U("urn:n:t:existing:bar")},
		}

		for _, tc := range testCases {
			if got := gold.ToURN[N, T](tc.input...); got != tc.expected {
				t.Errorf("ToURN(%v) = %q, want %q", tc.input, got, tc.expected)
			}
		}

		// Test panic case
		defer func() {
			if r := recover(); r == nil {
				t.Error("ToURN with invalid URN should panic")
			}
		}()
		gold.ToURN[N, T]("urn:other:schema:foo")
	})

	t.Run("AsURN", func(t *testing.T) {
		for in, ex := range map[string]U{
			"urn:n:t":         U("urn:n:t"),
			"urn:n:t:foo":     U("urn:n:t:foo"),
			"urn:n:t:foo:bar": U("urn:n:t:foo:bar"),
		} {
			if got, err := gold.AsURN[N, T](in); err != nil || got != ex {
				t.Errorf("AsURN(%q) = %q, %v; want %q, nil", in, got, err, ex)
			}
		}

		// Test error cases
		invalidCases := []string{
			"invalid",
			"urn:other:schema:foo",
			"urn:n:other:foo",
		}
		for _, tc := range invalidCases {
			if _, err := gold.AsURN[N, T](tc); err == nil {
				t.Errorf("AsURN(%q) should return an error", tc)
			}
		}
	})

	t.Run("ToIRI", func(t *testing.T) {
		for in, ex := range map[U]gold.IRI[T]{
			U("urn:n:t"):         gold.IRI[T](""),
			U("urn:n:t:foo"):     gold.IRI[T]("t:foo"),
			U("urn:n:t:foo:bar"): gold.IRI[T]("t:foo/bar"),
		} {
			if got := in.ToIRI(); got != ex {
				t.Errorf("(%q).ToIRI() = %q, want %q", in, got, ex)
			}
		}
	})

	t.Run("FromIRI", func(t *testing.T) {
		for in, ex := range map[gold.IRI[T]]U{
			gold.IRI[T](""):          U("urn:n:t"),
			gold.IRI[T]("t:foo"):     U("urn:n:t:foo"),
			gold.IRI[T]("t:foo/bar"): U("urn:n:t:foo:bar"),
		} {
			var urn U
			urn.FromIRI(in)
			if urn != ex {
				t.Errorf("FromIRI(%q) = %q, want %q", in, urn, ex)
			}
		}
	})

	t.Run("Norm", func(t *testing.T) {
		for in, ex := range map[U]U{
			U(""):                U("urn:n:t:"),
			U("foo"):             U("urn:n:t:foo"),
			U("urn:n:t:bar"):     U("urn:n:t:bar"),
			U("urn:n:t:foo:bar"): U("urn:n:t:foo:bar"),
		} {
			if got := in.Norm(); got != ex {
				t.Errorf("(%q).Norm() = %q, want %q", in, got, ex)
			}
		}
	})

	t.Run("Validate", func(t *testing.T) {
		validCases := []string{
			"urn:n:t",
			"urn:n:t:foo",
			"urn:n:t:foo:bar",
		}
		for _, tc := range validCases {
			urn := U(tc)
			if err := urn.Validate(); err != nil {
				t.Errorf("(%q).Validate() = %v, want nil", tc, err)
			}
		}

		invalidCases := []string{
			"invalid",
			"urn:other:schema:foo",
			"urn:n:other:foo",
		}
		for _, tc := range invalidCases {
			urn := U(tc)
			if err := urn.Validate(); err == nil {
				t.Errorf("(%q).Validate() = nil, want error", tc)
			}
		}
	})

	t.Run("IsValid", func(t *testing.T) {
		for in, ex := range map[U]bool{
			U("urn:n:t"):          true,
			U("urn:n:t:foo"):      true,
			U("urn:n:t:foo:bar"):  true,
			U("invalid"):          false,
			U("urn:other:schema"): false,
			U("urn:n:other"):      false,
		} {
			if got := in.IsValid(); got != ex {
				t.Errorf("(%q).IsValid() = %v, want %v", in, got, ex)
			}
		}
	})

	t.Run("IsEmpty", func(t *testing.T) {
		for in, ex := range map[U]bool{
			U("urn:n:t"):         true,
			U("urn:n:t:foo"):     false,
			U("urn:n:t:foo:bar"): false,
			U("invalid"):         false,
		} {
			if got := in.IsEmpty(); got != ex {
				t.Errorf("(%q).IsEmpty() = %v, want %v", in, got, ex)
			}
		}
	})

	t.Run("Reference", func(t *testing.T) {
		for in, ex := range map[U]string{
			U("urn:n:t:foo"):     "foo",
			U("urn:n:t:bar"):     "bar",
			U("urn:n:t:foo:bar"): "foo:bar",
			U("urn:n:t:complex"): "complex",
		} {
			if got := in.Reference(); got != ex {
				t.Errorf("(%q).Reference() = %q, want %q", in, got, ex)
			}
		}

		// Test panic case
		defer func() {
			if r := recover(); r == nil {
				t.Error("U(\"urn:n:t\").Reference() should panic")
			}
		}()
		U("urn:n:t").Reference()
	})

	t.Run("Split", func(t *testing.T) {
		testCases := []struct {
			input U
			base  U
			ref   U
		}{
			{U("urn:n:t"), U("urn:n:t"), U("")},
			{U("urn:n:t:foo"), U("urn:n:t"), U("urn:n:t:foo")},
			{U("urn:n:t:foo:bar"), U("urn:n:t:foo"), U("urn:n:t:bar")},
			{U("urn:n:t:foo:bar:baz"), U("urn:n:t:foo:bar"), U("urn:n:t:baz")},
		}

		for _, tc := range testCases {
			base, ref := tc.input.Split()
			if base != tc.base || ref != tc.ref {
				t.Errorf("(%q).Split() = (%q, %q), want (%q, %q)", tc.input, base, ref, tc.base, tc.ref)
			}
		}
	})

	t.Run("Join", func(t *testing.T) {
		testCases := []struct {
			urn1   U
			urn2   U
			result U
		}{
			{U("urn:n:t"), U("urn:n:t:foo"), U("urn:n:t:foo")},
			{U("urn:n:t:foo"), U("urn:n:t"), U("urn:n:t:foo")},
			{U("urn:n:t:foo"), U("urn:n:t:bar"), U("urn:n:t:foo:bar")},
			{U("urn:n:t:foo:bar"), U("urn:n:t:baz"), U("urn:n:t:foo:bar:baz")},
		}

		for _, tc := range testCases {
			if got := tc.urn1.Join(tc.urn2); got != tc.result {
				t.Errorf("(%q).Join(%q) = %q, want %q", tc.urn1, tc.urn2, got, tc.result)
			}
		}
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		for in, ex := range map[U]string{
			U("urn:n:t"):         `"urn:n:t"`,
			U("urn:n:t:foo"):     `"urn:n:t:foo"`,
			U("urn:n:t:foo:bar"): `"urn:n:t:foo:bar"`,
		} {
			if got, err := in.MarshalJSON(); err != nil || string(got) != ex {
				t.Errorf("(%q).MarshalJSON() = %q, %v; want %q, nil", in, got, err, ex)
			}
		}
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		for in, ex := range map[string]U{
			`"urn:n:t"`:         U("urn:n:t"),
			`"urn:n:t:foo"`:     U("urn:n:t:foo"),
			`"urn:n:t:foo:bar"`: U("urn:n:t:foo:bar"),
		} {
			var urn U
			if err := urn.UnmarshalJSON([]byte(in)); err != nil || urn != ex {
				t.Errorf("UnmarshalJSON(%q) = %q, %v; want %q, nil", in, urn, err, ex)
			}
		}

		// Test invalid JSON
		var urn U
		if err := urn.UnmarshalJSON([]byte(`"invalid"`)); err == nil {
			t.Error("UnmarshalJSON with invalid URN should return an error")
		}

		// Test malformed JSON
		if err := urn.UnmarshalJSON([]byte(`invalid`)); err == nil {
			t.Error("UnmarshalJSON with malformed JSON should return an error")
		}
	})
}
