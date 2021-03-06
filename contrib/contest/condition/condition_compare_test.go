package condition

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConditionCompare(t *testing.T) {
	cases := []struct {
		name     string
		op       string
		a        interface{}
		b        interface{}
		hint     reflect.Kind
		expected bool
	}{
		{
			name:     "equal: int",
			op:       "equal",
			a:        int(1),
			b:        int(1),
			hint:     reflect.Int,
			expected: true,
		},
		{
			name:     "equal: int, not equal",
			op:       "equal",
			a:        int(1),
			b:        int(2),
			hint:     reflect.Int64,
			expected: false,
		},
		{
			name:     "equal: int, different base",
			op:       "equal",
			a:        int(3),
			b:        int64(3),
			hint:     reflect.Int64,
			expected: true,
		},
		{
			name:     "equal: int, different type, float64 == int64",
			op:       "equal",
			a:        float64(1),
			b:        int64(1),
			hint:     reflect.Int64,
			expected: true,
		},
		{
			name:     "equal: float, different type, int64 == float64",
			op:       "equal",
			a:        int64(1),
			b:        float64(1),
			hint:     reflect.Float64,
			expected: true,
		},
		{
			name:     "equal: int, different type, int64 == string",
			op:       "equal",
			a:        "1",
			b:        int64(1),
			hint:     reflect.Int64,
			expected: true,
		},
		{
			name:     "equal: int, different type, slice == int64",
			op:       "equal",
			a:        []int{1, 2},
			b:        int64(1),
			hint:     reflect.Int64,
			expected: false,
		},
		{
			name:     "greater than: int",
			op:       "greater_than",
			a:        int(2),
			b:        int(1),
			hint:     reflect.Int,
			expected: true,
		},
		{
			name:     "greater than: int, but equal",
			op:       "greater_than",
			a:        int(2),
			b:        int(2),
			hint:     reflect.Int,
			expected: false,
		},
		{
			name:     "equal or greater than: int #0",
			op:       "equal_or_greater_than",
			a:        int(2),
			b:        int(2),
			hint:     reflect.Int,
			expected: true,
		},
		{
			name:     "equal or greater than: int #1",
			op:       "equal_or_greater_than",
			a:        int(3),
			b:        int(2),
			hint:     reflect.Int,
			expected: true,
		},
		{
			name:     "lesser than: int",
			op:       "lesser_than",
			a:        int(1),
			b:        int(2),
			hint:     reflect.Int,
			expected: true,
		},
		{
			name:     "lesser than: int, but equal",
			op:       "lesser_than",
			a:        int(2),
			b:        int(2),
			hint:     reflect.Int,
			expected: false,
		},
		{
			name:     "equal or lesser than: int #0",
			op:       "equal_or_lesser_than",
			a:        int(2),
			b:        int(2),
			hint:     reflect.Int,
			expected: true,
		},
		{
			name:     "equal or lesser than: int #1",
			op:       "equal_or_lesser_than",
			a:        int(2),
			b:        int(3),
			hint:     reflect.Int,
			expected: true,
		},
		{
			name:     "equal: string #0",
			op:       "equal",
			a:        "showme",
			b:        "showme",
			hint:     reflect.String,
			expected: true,
		},
		{
			name:     "equal: string #1",
			op:       "equal",
			a:        int64(1),
			b:        "1",
			hint:     reflect.String,
			expected: true,
		},
		{
			name:     "greater than: string",
			op:       "greater_than",
			a:        int64(2),
			b:        "1",
			hint:     reflect.String,
			expected: true,
		},
		{
			name:     "lesser than: string",
			op:       "lesser_than",
			a:        int64(2),
			b:        "3",
			hint:     reflect.String,
			expected: true,
		},
		{
			name:     "in: string",
			op:       "in",
			a:        "a",
			b:        []string{"a", "b"},
			hint:     reflect.String,
			expected: true,
		},
		{
			name:     "in: string, but not in",
			op:       "in",
			a:        "c",
			b:        []string{"a", "b"},
			hint:     reflect.String,
			expected: false,
		},
		{
			name:     "not in: string",
			op:       "not in",
			a:        "c",
			b:        []string{"a", "b"},
			hint:     reflect.String,
			expected: true,
		},
		{
			name:     "not in: string, but in",
			op:       "not in",
			a:        "a",
			b:        []string{"a", "b"},
			hint:     reflect.String,
			expected: false,
		},
		{
			name:     "in: int",
			op:       "in",
			a:        33,
			b:        []int{33, 34},
			hint:     reflect.Int,
			expected: true,
		},
		{
			name:     "in: int, but not in",
			op:       "in",
			a:        35,
			b:        []int{33, 34},
			hint:     reflect.Int,
			expected: false,
		},
		{
			name:     "not in: int",
			op:       "not in",
			a:        35,
			b:        []int{33, 34},
			hint:     reflect.Int,
			expected: true,
		},
		{
			name:     "not in: int, but in",
			op:       "not in",
			a:        33,
			b:        []int{33, 34},
			hint:     reflect.Int,
			expected: false,
		},
		{
			name:     "equal: int",
			op:       "equal",
			a:        int(1),
			b:        int(1),
			hint:     reflect.Int,
			expected: true,
		},
		{
			name:     "regexp: matched",
			op:       "regexp",
			a:        "show me",
			b:        "^show",
			hint:     reflect.String,
			expected: true,
		},
		{
			name:     "not regexp: matched",
			op:       "rnot egexp",
			a:        "show me",
			b:        "^show",
			hint:     reflect.String,
			expected: false,
		},
		{
			name:     "not regexp: not matched",
			op:       "not regexp",
			a:        "show me",
			b:        "^show0",
			hint:     reflect.String,
			expected: true,
		},
		{
			name:     "regexp: not matched",
			op:       "regexp",
			a:        "show me",
			b:        "^show0",
			hint:     reflect.String,
			expected: false,
		},
		{
			name:     "regexp: int matched",
			op:       "regexp",
			a:        int(12345),
			b:        "^12",
			hint:     reflect.String,
			expected: true,
		},
		{
			name:     "regexp: int not matched",
			op:       "regexp",
			a:        int(12345),
			b:        "678",
			hint:     reflect.String,
			expected: false,
		},
	}

	for i, c := range cases {
		i := i
		c := c
		t.Run(
			c.name,
			func(*testing.T) {
				result := compare(c.op, c.a, c.b, c.hint)
				assert.Equal(t, c.expected, result, "%d: %v; %v %v %v", i, c.name, c.op, c.expected, result)
			},
		)
	}
}
