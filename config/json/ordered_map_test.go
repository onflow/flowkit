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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_OrderedMap_SetAppendsNewKeys(t *testing.T) {
	var m orderedMap[int]
	m.Set("zeta", 1)
	m.Set("alpha", 2)
	m.Set("mu", 3)

	assert.Equal(t, 3, m.Len())
	assert.Equal(t, []string{"zeta", "alpha", "mu"}, collectKeys(m))
}

func Test_OrderedMap_SetUpdatesPreservesPosition(t *testing.T) {
	var m orderedMap[int]
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	m.Set("b", 99)

	assert.Equal(t, []string{"a", "b", "c"}, collectKeys(m))
	v, ok := m.Get("b")
	assert.True(t, ok)
	assert.Equal(t, 99, v)
}

func Test_OrderedMap_GetMissing(t *testing.T) {
	var m orderedMap[string]
	m.Set("x", "X")

	v, ok := m.Get("y")
	assert.False(t, ok)
	assert.Equal(t, "", v)
}

func Test_OrderedMap_MarshalPreservesInsertionOrder(t *testing.T) {
	var m orderedMap[int]
	m.Set("zeta", 1)
	m.Set("alpha", 2)
	m.Set("mu", 3)

	b, err := json.Marshal(m)
	require.NoError(t, err)
	assert.Equal(t, `{"zeta":1,"alpha":2,"mu":3}`, string(b))
}

func Test_OrderedMap_MarshalEmpty(t *testing.T) {
	var m orderedMap[int]

	b, err := json.Marshal(m)
	require.NoError(t, err)
	assert.Equal(t, `{}`, string(b))
}

func Test_OrderedMap_UnmarshalPreservesSourceOrder(t *testing.T) {
	input := []byte(`{"zeta":1,"alpha":2,"mu":3}`)

	var m orderedMap[int]
	err := json.Unmarshal(input, &m)
	require.NoError(t, err)

	assert.Equal(t, []string{"zeta", "alpha", "mu"}, collectKeys(m))
}

func Test_OrderedMap_UnmarshalReusesExistingBacking(t *testing.T) {
	var m orderedMap[int]
	m.Set("preexisting", 7)

	err := json.Unmarshal([]byte(`{"b":2,"a":1}`), &m)
	require.NoError(t, err)

	assert.Equal(t, []string{"b", "a"}, collectKeys(m))
	_, ok := m.Get("preexisting")
	assert.False(t, ok)
}

func Test_OrderedMap_UnmarshalNull(t *testing.T) {
	var m orderedMap[int]
	m.Set("x", 1)

	err := json.Unmarshal([]byte(`null`), &m)
	require.NoError(t, err)
	assert.Equal(t, 0, m.Len())
}

func Test_OrderedMap_UnmarshalRejectsNonObject(t *testing.T) {
	var m orderedMap[int]
	err := json.Unmarshal([]byte(`[1, 2, 3]`), &m)
	assert.Error(t, err)
}

func Test_OrderedMap_RoundTripPreservesOrder(t *testing.T) {
	input := []byte(`{"first":"a","second":"b","third":"c","fourth":"d"}`)

	var m orderedMap[string]
	require.NoError(t, json.Unmarshal(input, &m))

	out, err := json.Marshal(m)
	require.NoError(t, err)
	assert.Equal(t, string(input), string(out))
}

func Test_OrderedMap_StructValue(t *testing.T) {
	type point struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	input := []byte(`{"second":{"x":2,"y":20},"first":{"x":1,"y":10}}`)

	var m orderedMap[point]
	require.NoError(t, json.Unmarshal(input, &m))

	assert.Equal(t, []string{"second", "first"}, collectKeys(m))

	first, ok := m.Get("first")
	assert.True(t, ok)
	assert.Equal(t, point{X: 1, Y: 10}, first)

	out, err := json.Marshal(m)
	require.NoError(t, err)
	assert.Equal(t, string(input), string(out))
}

func Test_OrderedMap_RangeBreak(t *testing.T) {
	var m orderedMap[int]
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	var visited []string
	for k := range m.All {
		visited = append(visited, k)
		if k == "b" {
			break
		}
	}

	assert.Equal(t, []string{"a", "b"}, visited)
}

func collectKeys[V any](m orderedMap[V]) []string {
	keys := make([]string, 0, m.Len())
	for k := range m.All {
		keys = append(keys, k)
	}
	return keys
}
