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
	"reflect"
	"sync"
)

var schemaRegistry sync.Map

// typeKey identifies the type by its full identity (import path and name),
// so distinct types sharing a package name and a type name do not collide.
func typeKey(t reflect.Type) string {
	if t.Name() != "" && t.PkgPath() != "" {
		return t.PkgPath() + "." + t.Name()
	}
	return t.String() // unnamed / builtin / container types
}

// Schemas returns a copy of every schema derived or registered so far,
// keyed by the full type identity. Use it to pin persisted schemas in a golden test.
func Schemas() map[string]string {
	out := map[string]string{}
	schemaRegistry.Range(func(k, v any) bool {
		out[k.(string)] = v.(string)
		return true
	})
	return out
}

// Register binds a schema to the type T.
func Register[T any](schema string) string {
	class := reflect.TypeOf((*T)(nil)).Elem()
	if class.Kind() == reflect.Ptr {
		class = class.Elem()
	}
	name := typeKey(class)

	if s, loaded := schemaRegistry.LoadOrStore(name, schema); loaded {
		panic(fmt.Sprintf("gold: registering duplicate types for %q: %s != %s", name, s, schema))
	}

	return schema
}
