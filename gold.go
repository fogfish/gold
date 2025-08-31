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
//	type Creator gold.Imply[Image, User]
type Imply[A, B any] struct {
	Domain IRI[A] `json:"domain"`
	Range  IRI[B] `json:"range"`
}

// Create Imply statement from two IRIs.
func ImplyFrom[A, B any](a IRI[A], b IRI[B]) Imply[A, B] {
	return Imply[A, B]{Domain: a, Range: b}
}
