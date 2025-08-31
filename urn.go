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
)

// Uniform Resource Name (URN) is a type-safe identifier for hierarchical identity,
// it is defined as
//
//	urn:{namespace}:{schema}:{id⁰}:{id¹}:...:{idⁿ}
//
// where namespace is a unique identifier for the application,
// schema is a type-safe identifier for the resource, and id is a local identifier.
//
// URNs supports convenient hierarchical identity. It is not reflected in
// the type system and should be managed by the application itself.
type URN[N, T any] string

func emptyURN[N, T any]() string {
	return fmt.Sprintf("urn:%s:%s", Schema[N](), Schema[T]())
}

// URN type constructor, creates a new URN from the string.
// It returns a valid compact URN annotated with a urn:{namespace}:{schema}.
//
// If input segment is compact URN of other kind, it panic.
func ToURN[N, T any](seq ...string) URN[N, T] {
	urn := emptyURN[N, T]()

	if len(seq) == 0 {
		return URN[N, T](urn)
	}

	if strings.HasPrefix(seq[0], urn+":") {
		return URN[N, T](strings.Join(seq, ":"))
	}

	if strings.HasPrefix(seq[0], "urn:") {
		panic(fmt.Errorf("gold: invalid URN %q for schema %q", seq[0], urn))
	}

	urn = fmt.Sprintf("%s:%s", urn, strings.Join(seq, ":"))
	return URN[N, T](urn)
}

// URN parser, converts a string into a URN type.
// Input string is expected to be a compact URN of the form urn:{namespace}:{schema}.
// It returns an error if the schema prefix is different than one associated with type T.
func AsURN[N, T any](urn string) (URN[N, T], error) {
	schema := emptyURN[N, T]()

	if !strings.HasPrefix(urn, schema+":") {
		if urn == schema {
			return URN[N, T](schema), nil
		}

		return "", fmt.Errorf("gold: invalid URN %q for schema %q", urn, schema)
	}

	return URN[N, T](urn), nil
}

// Cast URN to IRI type.
func (urn URN[N, T]) ToIRI() IRI[T] {
	prefix := emptyURN[N, T]()

	if string(urn) == prefix || len(urn) <= len(prefix)+1 {
		return IRI[T]("")
	}

	ref := string(urn[len(prefix)+1:])

	s := strings.ReplaceAll(ref, ":", "/")
	return ToIRI[T](s)
}

// Cast IRI to URN type.
func (urn *URN[N, T]) FromIRI(iri IRI[T]) {
	if iri.IsEmpty() {
		*urn = URN[N, T](emptyURN[N, T]())
		return
	}

	ref := iri.Reference()
	ref = strings.ReplaceAll(ref, "/", ":")
	*urn = ToURN[N, T](ref)
}

// Normalize URN to a URN type.
// Use it as constructor for URN types aliases.
//
//	type MyID = gold.URN[MyNamespace, MyClass]
//	var id = MyID("foo").Norm()
func (urn URN[N, T]) Norm() URN[N, T] { return ToURN[N, T](string(urn)) }

// Validate URN to be a valid compact URN type, return an error if it is not.
func (urn *URN[N, T]) Validate() (err error) {
	*urn, err = AsURN[N, T](string(*urn))
	if err != nil {
		return err
	}

	return nil
}

// Validate URN to be a valid URN type.
func (urn *URN[N, T]) IsValid() bool { return urn.Validate() == nil }

// Check URN is defined.
func (urn URN[N, T]) IsEmpty() bool { return string(urn) == emptyURN[N, T]() }

// Reference returns a local identifier of the URN, without schema prefix.
func (urn URN[N, T]) Reference() string {
	schema := emptyURN[N, T]()
	if len(urn) <= len(schema)+1 {
		panic(fmt.Errorf("gold: invalid URN %q for schema %q", urn, schema))
	}
	return string(urn[len(schema)+1:])
}

// Split splits URN into two parts: base and reference.
//
//	urn:{namespace}:{schema}:{id⁰}:{id¹}:...:{idⁿ} ⇒
//	  urn:{namespace}:{schema}:{id⁰}:{id¹}:...:{idⁿ⁻¹}
//	  urn:{namespace}:{schema}:{idⁿ}
func (urn URN[N, T]) Split() (URN[N, T], URN[N, T]) {
	if urn.IsEmpty() {
		return urn, ""
	}

	s := string(urn)
	n := strings.LastIndex(s, ":")

	// No separator exists, single segment is returned as "base"
	if n == -1 {
		return "", urn
	}

	return URN[N, T](s[:n]), ToURN[N, T](s[n+1:])
}

// Join combines two URNs into one, where the second URN is appended to the first.
func (urn URN[N, T]) Join(other URN[N, T]) URN[N, T] {
	prefix := emptyURN[N, T]()

	if string(urn) == prefix {
		return other
	}

	if string(other) == prefix {
		return urn
	}

	return URN[N, T](string(urn) + string(other[len(prefix):]))
}

// Encodes IRI into JSON as a CURIE string.
func (urn URN[N, T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(urn))
}

// Decodes IRI from JSON as a CURIE string.
func (urn *URN[N, T]) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*urn, err = AsURN[N, T](s)
	return
}
