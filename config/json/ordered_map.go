/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
)

// orderedMap is a string-keyed map that preserves insertion order
// across JSON marshal/unmarshal cycles. flow.json sections need this
// so that round-tripping the file does not reorder entries.
//
// Receiver convention: mutating methods (Set, UnmarshalJSON) use a pointer
// receiver; all others use a value receiver. This intentionally mirrors the
// pattern used by the standard library's encoding/json marshalers (e.g.
// time.Time) — MarshalJSON must be on a value receiver because encoding/json
// only calls a pointer-receiver Marshaler when the value is addressable, which
// it is not when callers pass an orderedMap directly to json.Marshal.
type orderedMap[V any] struct {
	entries []orderedEntry[V]
}

type orderedEntry[V any] struct {
	Key   string
	Value V
}

// Set inserts or updates an entry. New keys are appended at the end;
// existing keys keep their original position.
func (m *orderedMap[V]) Set(key string, value V) {
	for i := range m.entries {
		if m.entries[i].Key == key {
			m.entries[i].Value = value
			return
		}
	}
	m.entries = append(m.entries, orderedEntry[V]{Key: key, Value: value})
}

// Get returns the value for key and whether it was found.
func (m orderedMap[V]) Get(key string) (V, bool) {
	for _, e := range m.entries {
		if e.Key == key {
			return e.Value, true
		}
	}
	var zero V
	return zero, false
}

// Len returns the number of entries.
func (m orderedMap[V]) Len() int {
	return len(m.entries)
}

// All is an iter.Seq2-shaped iterator over entries in insertion order,
// usable directly with `for k, v := range m.All`.
func (m orderedMap[V]) All(yield func(string, V) bool) {
	for _, e := range m.entries {
		if !yield(e.Key, e.Value) {
			return
		}
	}
}

func (m orderedMap[V]) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, e := range m.entries {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyBytes, err := json.Marshal(e.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')
		valueBytes, err := json.Marshal(e.Value)
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (m *orderedMap[V]) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if tok == nil {
		m.entries = nil
		return nil
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return fmt.Errorf("expected JSON object for orderedMap, got %v", tok)
	}

	m.entries = m.entries[:0]
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("expected string key, got %T", keyTok)
		}
		var value V
		if err := dec.Decode(&value); err != nil {
			return err
		}
		m.Set(key, value)
	}

	if _, err := dec.Token(); err != nil { // closing '}'
		return err
	}
	return nil
}

// JSONSchema renders orderedMap as a JSON object whose values match V's schema,
// matching invopop's default representation for map[string]V so the generated
// schema is unaffected by the switch from a plain map.
func (m orderedMap[V]) JSONSchema() *jsonschema.Schema {
	definitions := map[string]*jsonschema.Schema{}
	valueSchema := refSchemaForType(reflect.TypeFor[V](), definitions)

	schema := &jsonschema.Schema{
		Type: "object",
		PatternProperties: map[string]*jsonschema.Schema{
			".*": valueSchema,
		},
	}
	if len(definitions) > 0 {
		schema.Definitions = definitions
	}
	return schema
}

// refSchemaForType returns a schema for t. Named types become "$ref" entries
// and their reflected schema is added to definitions; slice/array types yield
// an inline array schema whose element schema is resolved recursively.
func refSchemaForType(t reflect.Type, definitions map[string]*jsonschema.Schema) *jsonschema.Schema {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if name := t.Name(); name != "" {
		if _, exists := definitions[name]; !exists {
			definitions[name] = jsonschema.Reflect(reflect.New(t).Elem().Interface())
		}
		return &jsonschema.Schema{Ref: "#/$defs/" + name}
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		return &jsonschema.Schema{
			Type:  "array",
			Items: refSchemaForType(t.Elem(), definitions),
		}
	}

	return jsonschema.Reflect(reflect.New(t).Elem().Interface())
}
