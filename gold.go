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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Compact IRI (CURIE) is universal identifier for a resource, defined as
//
//	{schema}:{id}
//
// where `schema“ is a namespace prefix and `id` is a local identifier. The schema
// is always derived from the type of the resource T, making it safe at compiletime.
//
// It is recommended to derive schema from the class itself:
//
//	type MyClass struct {
//	  ID gold.IRI[MyClass] `json:"id"`
//	}
//
// Locally, IRI is ALWAY kept as {id} and type annotation, formatting into CURIE
// excuted at serialization.
type IRI[T any] string

// Schema returns a CURIE schema for the type T is derived from the type name
// or type registry. It is recommended to use pure types rather than containers.
// For any complex types use registry:
//
//	gold.Register[[]MyClass]("seq")
//	gold.Schema([]MyClass)()
func Schema[T any]() string {
	class := reflect.TypeOf(new(T)).Elem()
	if class.Kind() == reflect.Ptr {
		class = class.Elem()
	}
	name := class.String()

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

	return strings.ToLower(cat)
}

// Convert a local string identifier to compact IRI.
// It returns id string annotated with a type.
func ToIRI[T any](id string) IRI[T] {
	return IRI[T](id)
}

// Convert IRI string into a compact IRI type. It fails schema prefix is
// different than one assotiated with type T.
func AsIRI[T any](iri string) (IRI[T], error) {
	schema := Schema[T]()
	if !strings.HasPrefix(iri, schema+":") {
		return "", fmt.Errorf("gold: invalid IRI %q for schema %q", iri, schema)
	}

	if len(iri) < len(schema)+1 {
		return "", fmt.Errorf("gold: invalid IRI %q for schema %q", iri, schema)
	}

	return IRI[T](iri[len(schema)+1:]), nil
}

// Converts IRI to fully qualified CURIE as a string type.
func (iri IRI[T]) String() string {
	schema := Schema[T]()
	return fmt.Sprintf("%s:%s", schema, string(iri))
}

// Encodes IRI into JSON as a CURIE string.
func (iri IRI[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(iri.String())
}

// Decodes IRI from JSON as a CURIE string.
func (iri *IRI[T]) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*iri, err = AsIRI[T](s)
	return
}

// Expands IRI into absolute URI using the registered prefix for the type T.
func URI[T any](iri IRI[T]) string {
	class := reflect.TypeOf(new(T)).Elem()
	if class.Kind() == reflect.Ptr {
		class = class.Elem()
	}
	name := class.String()

	prefix, ok := uriRegistry.Load(name)
	if !ok {
		return iri.String()
	}

	return prefix.(string) + ":" + string(iri)
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
//	type Creator gold.Imply[Image, User]
type Imply[A, B any] struct {
	Domain IRI[A] `json:"domain"`
	Range  IRI[B] `json:"range"`
}
