//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/gold
//

package gold_test

import (
	"errors"
	"testing"

	"github.com/fogfish/gold"
	"github.com/fogfish/it/v2"
)

// Test types used across codec tests.
type (
	codecA struct{}
	codecB struct{}

	codecPair struct {
		A gold.IRI[codecA]
		B gold.IRI[codecB]
	}
)

// ---------------------------------------------------------------------------
// EncodeJSON / DecodeJSON
// ---------------------------------------------------------------------------

func TestEncodeJSON(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		b, err := gold.EncodeJSON("hello", nil)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(string(b), `"hello"`),
		)
	})

	t.Run("PropagatesError", func(t *testing.T) {
		it.Then(t).Should(
			it.Error(gold.EncodeJSON("", errors.New("test"))),
		)
	})
}

func TestDecodeJSON(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s, err := gold.DecodeJSON([]byte(`"world"`))
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(s, "world"),
		)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		it.Then(t).Should(
			it.Error(gold.DecodeJSON([]byte(`not-json`))),
		)
	})
}

// ---------------------------------------------------------------------------
// N3
// ---------------------------------------------------------------------------

func TestN3(t *testing.T) {
	codec := gold.N3[codecPair, codecA, codecB]()

	t.Run("EncodeEmpty", func(t *testing.T) {
		obj := codecPair{}
		s, err := codec.Encode(&obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(s, ""),
		)
	})

	t.Run("EncodeMissingB", func(t *testing.T) {
		obj := codecPair{A: gold.ToIRI[codecA]("x")}
		s, err := codec.Encode(&obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(s, ""),
		)
	})

	t.Run("EncodeBothFields", func(t *testing.T) {
		obj := codecPair{
			A: gold.ToIRI[codecA]("x"),
			B: gold.ToIRI[codecB]("y"),
		}
		s, err := codec.Encode(&obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(s, "codeca:x codecpair codecb:y."),
		)
	})

	t.Run("DecodeEmpty", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("", &obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(obj.A, gold.IRI[codecA]("")),
			it.Equal(obj.B, gold.IRI[codecB]("")),
		)
	})

	t.Run("DecodeValid", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("codeca:x codecpair codecb:y.", &obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(obj.A, gold.ToIRI[codecA]("x")),
			it.Equal(obj.B, gold.ToIRI[codecB]("y")),
		)
	})

	t.Run("DecodeMissingDot", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("codeca:x codecpair codecb:y", &obj)
		it.Then(t).Should(
			it.True(err != nil),
		)
	})

	t.Run("DecodeWrongSchema", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("codeca:x wrongschema codecb:y.", &obj)
		it.Then(t).Should(
			it.True(err != nil),
		)
	})

	t.Run("DecodeInvalidIRI_A", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("badiri codecpair codecb:y.", &obj)
		it.Then(t).Should(
			it.True(err != nil),
		)
	})

	t.Run("DecodeInvalidIRI_B", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("codeca:x codecpair badiri.", &obj)
		it.Then(t).Should(
			it.True(err != nil),
		)
	})

	t.Run("RoundTrip", func(t *testing.T) {
		orig := codecPair{
			A: gold.ToIRI[codecA]("alice"),
			B: gold.ToIRI[codecB]("selfie"),
		}
		s, err := codec.Encode(&orig)
		it.Then(t).Should(it.Nil(err))

		var decoded codecPair
		err = codec.Decode(s, &decoded)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(decoded.A, orig.A),
			it.Equal(decoded.B, orig.B),
		)
	})
}

// ---------------------------------------------------------------------------
// HashKey
// ---------------------------------------------------------------------------

func TestHashKey(t *testing.T) {
	codec := gold.HashKey[codecPair, codecA, codecB]()

	t.Run("EncodeEmpty", func(t *testing.T) {
		obj := codecPair{}
		s, err := codec.Encode(&obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(s, ""),
		)
	})

	t.Run("EncodeBothFields", func(t *testing.T) {
		obj := codecPair{
			A: gold.ToIRI[codecA]("x"),
			B: gold.ToIRI[codecB]("y"),
		}
		s, err := codec.Encode(&obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(s, "codeca:x|codecb:y"),
		)
	})

	t.Run("DecodeEmpty", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("", &obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(obj.A, gold.IRI[codecA]("")),
			it.Equal(obj.B, gold.IRI[codecB]("")),
		)
	})

	t.Run("DecodeValid", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("codeca:x|codecb:y", &obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(obj.A, gold.ToIRI[codecA]("x")),
			it.Equal(obj.B, gold.ToIRI[codecB]("y")),
		)
	})

	t.Run("DecodeMissingSeparator", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("codeca:x", &obj)
		it.Then(t).Should(
			it.True(err != nil),
		)
	})

	t.Run("DecodeInvalidIRI_A", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("badiri|codecb:y", &obj)
		it.Then(t).Should(
			it.True(err != nil),
		)
	})

	t.Run("DecodeInvalidIRI_B", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("codeca:x|badiri", &obj)
		it.Then(t).Should(
			it.True(err != nil),
		)
	})

	t.Run("RoundTrip", func(t *testing.T) {
		orig := codecPair{
			A: gold.ToIRI[codecA]("alice"),
			B: gold.ToIRI[codecB]("selfie"),
		}
		s, err := codec.Encode(&orig)
		it.Then(t).Should(it.Nil(err))

		var decoded codecPair
		err = codec.Decode(s, &decoded)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(decoded.A, orig.A),
			it.Equal(decoded.B, orig.B),
		)
	})
}

// ---------------------------------------------------------------------------
// SortKey
// ---------------------------------------------------------------------------

func TestSortKey(t *testing.T) {
	codec := gold.SortKey[codecPair, codecA, codecB]()

	t.Run("EncodeBothEmpty", func(t *testing.T) {
		obj := codecPair{}
		s, err := codec.Encode(&obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(s, ""),
		)
	})

	t.Run("EncodeOnlyA", func(t *testing.T) {
		obj := codecPair{A: gold.ToIRI[codecA]("x")}
		s, err := codec.Encode(&obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(s, "codeca:x"),
		)
	})

	t.Run("EncodeBothFields", func(t *testing.T) {
		obj := codecPair{
			A: gold.ToIRI[codecA]("x"),
			B: gold.ToIRI[codecB]("y"),
		}
		s, err := codec.Encode(&obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(s, "codeca:x|codecb:y"),
		)
	})

	t.Run("DecodeEmpty", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("", &obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(obj.A, gold.IRI[codecA]("")),
			it.Equal(obj.B, gold.IRI[codecB]("")),
		)
	})

	t.Run("DecodeOnlyA", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("codeca:x", &obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(obj.A, gold.ToIRI[codecA]("x")),
			it.Equal(obj.B, gold.IRI[codecB]("")),
		)
	})

	t.Run("DecodeBothFields", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("codeca:x|codecb:y", &obj)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(obj.A, gold.ToIRI[codecA]("x")),
			it.Equal(obj.B, gold.ToIRI[codecB]("y")),
		)
	})

	t.Run("DecodeInvalidIRI_A", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("badiri", &obj)
		it.Then(t).Should(
			it.True(err != nil),
		)
	})

	t.Run("DecodeInvalidIRI_B", func(t *testing.T) {
		obj := codecPair{}
		err := codec.Decode("codeca:x|badiri", &obj)
		it.Then(t).Should(
			it.True(err != nil),
		)
	})

	t.Run("RoundTripOnlyA", func(t *testing.T) {
		orig := codecPair{A: gold.ToIRI[codecA]("alice")}
		s, err := codec.Encode(&orig)
		it.Then(t).Should(it.Nil(err))

		var decoded codecPair
		err = codec.Decode(s, &decoded)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(decoded.A, orig.A),
			it.Equal(decoded.B, orig.B),
		)
	})

	t.Run("RoundTripBoth", func(t *testing.T) {
		orig := codecPair{
			A: gold.ToIRI[codecA]("alice"),
			B: gold.ToIRI[codecB]("selfie"),
		}
		s, err := codec.Encode(&orig)
		it.Then(t).Should(it.Nil(err))

		var decoded codecPair
		err = codec.Decode(s, &decoded)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(decoded.A, orig.A),
			it.Equal(decoded.B, orig.B),
		)
	})
}
