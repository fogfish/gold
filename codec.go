//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/gold
//

package gold

import (
	"fmt"
	"strings"

	"github.com/fogfish/golem/optics"
)

type Encoder[T any] interface {
	Encode(*T) (string, error)
}

type Decoder[T any] interface {
	Decode(string, *T) error
}

type Codec[T any] interface {
	Encoder[T]
	Decoder[T]
}

//------------------------------------------------------------------------------

type n3[T, A, B any] struct {
	schema string
	shape  optics.Lens2[T, IRI[A], IRI[B]]
}

func N3[T, A, B any]() Codec[T] {
	return n3[T, A, B]{
		schema: Schema[T](),
		shape:  optics.ForShape2[T, IRI[A], IRI[B]](),
	}
}

func (c n3[T, A, B]) Encode(obj *T) (string, error) {
	a, b := c.shape.Get(obj)
	return fmt.Sprintf("%s %s %s.", a.String(), c.schema, b.String()), nil
}

func (c n3[T, A, B]) Decode(s string, obj *T) error {
	if !strings.HasSuffix(s, ".") {
		return fmt.Errorf("gold: invalid n3 format for %q", s)
	}

	seq := strings.SplitN(s[:len(s)-1], " ", 3)
	if len(seq) != 3 || seq[1] != c.schema {
		return fmt.Errorf("gold: invalid schema for %q", s)
	}

	a, err := AsIRI[A](seq[0])
	if err != nil {
		return fmt.Errorf("gold: invalid IRI for %q: %w", seq[0], err)
	}

	b, err := AsIRI[B](seq[2])
	if err != nil {
		return fmt.Errorf("gold: invalid IRI for %q: %w", seq[2], err)
	}

	c.shape.Put(obj, a, b)
	return nil
}
