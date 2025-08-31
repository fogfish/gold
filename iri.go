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

// Compact IRI (CURIE) is universal identifier for a resource, defined as
//
//	{schema}:{id}
//
// where schema is a namespace prefix and id is a local identifier. The schema
// is always derived from the type of the resource T, making it safe at compiletime.
//
// It is recommended to derive schema from the class itself:
//
//	type MyClass struct {
//	  ID gold.IRI[MyClass] `json:"id"`
//	}
//
// IRIs are not hierarchical in the linked-data sense. They are flat, typed
// identifiers structured as {schema}:{id}.
type IRI[T any] string

// IRI type constructor, creates a new IRI from the string.
// It returns a valid compact IRI annotated with a schema.
//
// If input string is compact IRI of other kind, it is still annotated with a schema.
func ToIRI[T any](id string) IRI[T] {
	if len(id) == 0 {
		return IRI[T]("")
	}

	schema := Schema[T]()
	prefix := fmt.Sprintf("%s:", schema)

	if strings.HasPrefix(id, prefix) {
		return IRI[T](id)
	}

	return IRI[T](prefix + id)
}

// IRI parser, converts a string into a compact IRI type.
// Input string is expected to be a compact IRI of the form {schema}:{id}.
// It returns an error if the schema prefix is different than one associated with type T.
func AsIRI[T any](iri string) (IRI[T], error) {
	schema := Schema[T]()
	prefix := fmt.Sprintf("%s:", schema)

	if !strings.HasPrefix(iri, prefix) {
		return "", fmt.Errorf("gold: invalid IRI %q for schema %q", iri, schema)
	}

	return IRI[T](iri), nil
}

// Normalize IRI to a compact IRI type.
// Use it as constructor for IRI types aliases.
//
//	type MyID = gold.IRI[MyClass]
//	var id = MyID("foo").Norm()
func (iri IRI[T]) Norm() IRI[T] { return ToIRI[T](string(iri)) }

// Validate IRI to be a valid compact IRI type, return an error if it is not.
func (iri *IRI[T]) Validate() (err error) {
	*iri, err = AsIRI[T](string(*iri))
	if err != nil {
		return err
	}
	return nil
}

// Validate IRI to be a valid compact IRI type.
func (iri *IRI[T]) IsValid() bool { return iri.Validate() == nil }

// Check IRI is defined.
func (iri IRI[T]) IsEmpty() bool { return len(iri) == 0 }

// Reference returns a local identifier of the IRI, without schema prefix.
func (iri IRI[T]) Reference() string {
	schema := Schema[T]()
	if len(iri) <= len(schema)+1 {
		panic(fmt.Errorf("gold: invalid IRI %q for schema %q", iri, schema))
	}

	return string(iri[len(schema)+1:])
}

// Encodes IRI into JSON as a CURIE string.
func (iri IRI[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(iri))
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
