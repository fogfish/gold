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

func Register[T any](schema string) string {
	class := reflect.TypeOf((*T)(nil)).Elem()
	if class.Kind() == reflect.Ptr {
		class = class.Elem()
	}
	name := class.String()

	if s, loaded := schemaRegistry.LoadOrStore(name, schema); loaded {
		panic(fmt.Sprintf("gold: registering duplicate types for %q: %s != %s", name, s, schema))
	}

	return schema
}
