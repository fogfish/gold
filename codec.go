//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/gold
//

package gold

import (
	"encoding/json"
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

func EncodeJSON(s string, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	return json.Marshal(s)
}

func DecodeJSON(b []byte) (string, error) {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return "", err
	}
	return s, nil
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
	if len(a) == 0 || len(b) == 0 {
		return "", nil
	}

	return fmt.Sprintf("%s %s %s.", a, c.schema, b), nil
}

func (c n3[T, A, B]) Decode(s string, obj *T) error {
	if len(s) == 0 {
		c.shape.Put(obj, ToIRI[A](""), ToIRI[B](""))
		return nil
	}

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

//------------------------------------------------------------------------------

type hashkey[T, A, B any] struct {
	schema string
	shape  optics.Lens2[T, IRI[A], IRI[B]]
}

func HashKey[T, A, B any]() Codec[T] {
	return hashkey[T, A, B]{
		schema: Schema[T](),
		shape:  optics.ForShape2[T, IRI[A], IRI[B]](),
	}
}

func (c hashkey[T, A, B]) Encode(obj *T) (string, error) {
	a, b := c.shape.Get(obj)
	if len(a) == 0 || len(b) == 0 {
		return "", nil
	}
	return fmt.Sprintf("%s|%s", a, b), nil
}

func (c hashkey[T, A, B]) Decode(s string, obj *T) error {
	if len(s) == 0 {
		c.shape.Put(obj, ToIRI[A](""), ToIRI[B](""))
		return nil
	}

	seq := strings.SplitN(s, "|", 2)
	if len(seq) != 2 {
		return fmt.Errorf("gold: invalid schema for %q", s)
	}

	a, err := AsIRI[A](seq[0])
	if err != nil {
		return fmt.Errorf("gold: invalid IRI for %q: %w", seq[0], err)
	}

	b, err := AsIRI[B](seq[1])
	if err != nil {
		return fmt.Errorf("gold: invalid IRI for %q: %w", seq[1], err)
	}

	c.shape.Put(obj, a, b)
	return nil
}
