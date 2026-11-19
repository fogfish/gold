//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/gold
//

// Package gold provides a typesafe Linked Data support for Golang.
// It defines algebra for declaring identity and hierarchical relationships between data types.
package gold

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Schema returns a schema for the type T is derived from the type name
// or type registry. It is recommended to use pure types rather than containers.
// For any complex types use registry:
//
//	gold.Register[[]MyClass]("seq")
//	gold.Schema([]MyClass)()
func Schema[T any]() string {
	class := reflect.TypeOf((*T)(nil)).Elem()
	if class.Kind() == reflect.Ptr {
		class = class.Elem()
	}
	name := typeKey(class)

	schema, ok := schemaRegistry.Load(name)
	if ok {
		return schema.(string)
	}

	cat := class.Name()

	switch class.Kind() {
	case reflect.Slice, reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.UnsafePointer:
		if cat == "" {
			panic(fmt.Errorf("gold: unknown schema for %q", name))
		}
	}

	if strings.Contains(cat, "[") {
		panic(fmt.Errorf("gold: unknown schema for %q", cat))
	}

	lcat := strings.ToLower(cat)
	schemaRegistry.Store(name, lcat)
	return lcat
}

// IRIs are not hierarchical in the linked-data sense. They are flat, typed
// identifiers structured as {schema}:{id}. IRI assumes hierarchical identifiers
// but they are not hierarchical in this context {id} is a "flat" identity of the class.
// Conveniently, it is possible to use IRIs as hierarchical identifiers but it is not
// reflected in the type system and should be managed by the application itself.
//
// Relations between concepts are not implied by IRI structure itself. Instead,
// They are modeled explicitly as product of IRIs or using predicate statements:
//
//	image:Selfie schema:creator user:Alice
//
// Application models predicate statements as attributes of the IRI type
//
//	type Image struct {
//	  ID      gold.IRI[Image] `json:"id"`
//	  Creator gold.IRI[User]  `json:"creator"`
//	}
//
// Statements are not flexible enough to model convolution of statements into
// composite keys, which are essential for persisting linked-data.
//
// Generic Imply type is used to explicitly model relationships between two
// types A and B as composite structure.
//
//	user:Alice/schema:creator/image:Selfie
//
//	type Creator = gold.Imply[Image, User]
//
// Imply is a text type, it is serialised as a string "{Domain}|{Range}"
// (e.g. "user:alice|image:selfie"), including JSON and DynamoDB attributevalue
// codecs. Use a type alias to keep the methods. A defined type drops them and
// requires explicit forwarding of MarshalText and UnmarshalText.
type Imply[A, B any] struct {
	Domain IRI[A] `json:"domain"`
	Range  IRI[B] `json:"range"`
}

// Create Imply statement from two IRIs.
func ImplyFrom[A, B any](a IRI[A], b IRI[B]) Imply[A, B] {
	return Imply[A, B]{Domain: a, Range: b}
}

// Separator between Domain and Range in the text form of Imply.
const implySep = "|"

// Encodes Imply as text "{Domain}|{Range}".
//
//	empty, empty → ""
//	set,   empty → "{Domain}|"
//	set,   set   → "{Domain}|{Range}"
//
// Empty domain with non-empty range, domain with "|" and IRIs without
// the schema prefix are errors.
func (i Imply[A, B]) MarshalText() ([]byte, error) {
	if i.Domain.IsEmpty() {
		if i.Range.IsEmpty() {
			return []byte{}, nil
		}
		return nil, errors.New("gold: imply with empty domain")
	}

	if strings.Contains(string(i.Domain), implySep) {
		return nil, fmt.Errorf("gold: imply domain must not contain %q", implySep)
	}

	d, err := i.Domain.MarshalText()
	if err != nil {
		return nil, err
	}

	r, err := i.Range.MarshalText()
	if err != nil {
		return nil, err
	}

	return []byte(string(d) + implySep + string(r)), nil
}

// Decodes Imply from text "{Domain}|{Range}", splitting on the first separator.
func (i *Imply[A, B]) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		*i = Imply[A, B]{}
		return nil
	}

	seq := strings.SplitN(string(b), implySep, 2)
	if len(seq) != 2 {
		return fmt.Errorf("gold: invalid imply %q, separator %q is missing", string(b), implySep)
	}

	if len(seq[0]) == 0 {
		return errors.New("gold: imply with empty domain")
	}

	d, err := AsIRI[A](seq[0])
	if err != nil {
		return err
	}

	var r IRI[B]
	if len(seq[1]) != 0 {
		if r, err = AsIRI[B](seq[1]); err != nil {
			return err
		}
	}

	*i = Imply[A, B]{Domain: d, Range: r}
	return nil
}

// String returns text form of Imply, an invalid value yields an empty string.
func (i Imply[A, B]) String() string {
	b, err := i.MarshalText()
	if err != nil {
		return ""
	}
	return string(b)
}
