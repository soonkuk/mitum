package condition

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestCondition(t *testing.T) {
	cases := []struct {
		name     string
		query    string
		expected string
		err      string
	}{
		{
			name:     "simple: int",
			query:    "a=2",
			expected: "(a = [2])",
		},
		{
			name:     "simple: string val",
			query:    `a="showme"`,
			expected: "(a = [showme])",
		},
		{
			name:     "simple: negative int",
			query:    "a=-2",
			expected: "(a = [-2])",
		},
		{
			name:     "simple: float",
			query:    "a=3.141592",
			expected: "(a = [3.141592])",
		},
		{
			name:     "simple: dot connected field",
			query:    "a.b.c.d.e = 2",
			expected: "(a.b.c.d.e = [2])",
		},
		{
			name:     "simple: a>=2",
			query:    "a>=2",
			expected: "(a >= [2])",
		},
		{
			name:     "simple: equal a = 3",
			query:    "a=3",
			expected: "(a = [3])",
		},
		{
			name:     "simple: lessthan a < 3",
			query:    "a<3",
			expected: "(a < [3])",
		},
		{
			name:     "simple: greaterthan a > 3",
			query:    "a>3",
			expected: "(a > [3])",
		},
		{
			name:     "simple: lessequal a <= 3",
			query:    "a<=3",
			expected: "(a <= [3])",
		},
		{
			name:     "simple: greaterequal a >= 3",
			query:    "a>=3",
			expected: "(a >= [3])",
		},
		{
			name:     "simple: notequal a != 3",
			query:    "a!=3",
			expected: "(a != [3])",
		},
		{
			name:     "simple: in a in 3",
			query:    "a in (3, 4, 5, 6)",
			expected: "(a in [3,4,5,6])",
		},
		{
			name:     "simple: notin a not in 3",
			query:    "a not in (3, 4, 5, 6)",
			expected: "(a not in [3,4,5,6])",
		},
		{
			name:     "simple: in a in 3",
			query:    "a in (3, 4, 5, 6)",
			expected: "(a in [3,4,5,6])",
		},
		{
			name:     "simple: notin a not in 3",
			query:    "a not in (3, 4, 5, 6)",
			expected: "(a not in [3,4,5,6])",
		},
		{
			name:     "simple: regexp a regexp 3",
			query:    `a regexp "foo.*"`,
			expected: "(a regexp [foo.*])",
		},
		{
			name:     "simple: notregexp a not regexp 3",
			query:    `a not regexp "foo.*"`,
			expected: "(a not regexp [foo.*])",
		},
		{
			name:  "simple: bad regexp expression",
			query: `a not regexp "foo(.*"`,
			err:   "error parsing regexp",
		},
		{
			name:     "joint: and with 2 comparison",
			query:    `a = 1 and b = 2`,
			expected: "(and:(a = [1]), (b = [2]))",
		},
		{
			name:     "joint: and with 3 comparison",
			query:    `a = 1 and b = 2 and c = 3`,
			expected: "(and:(a = [1]), (b = [2]), (c = [3]))",
		},
		{
			name:     "joint: or with 2 comparison",
			query:    `a = 1 or b = 2`,
			expected: "(or:(a = [1]), (b = [2]))",
		},
		{
			name:     "joint: or with 3 comparison",
			query:    `a = 1 or b = 2 or c = 3`,
			expected: "(or:(a = [1]), (b = [2]), (c = [3]))",
		},
		{
			name:     "joint: and first, complex with 3 comparison",
			query:    `a = 1 and b = 2 or c = 3`,
			expected: "(or:(and:(a = [1]), (b = [2])), (c = [3]))",
		},
		{
			name:     "joint: or first, complex with 3 comparison",
			query:    `(a = 1 or b = 2) and c = 3`,
			expected: "(and:(or:(a = [1]), (b = [2])), (c = [3]))",
		},
		{
			name:     "joint: complex #0",
			query:    `(a > 1 or b < 2) and (c >= 3 and d <= 4) or (e != 5 and f not in (6, 7))`,
			expected: "(or:(and:(or:(a > [1]), (b < [2])), (and:(c >= [3]), (d <= [4]))), (and:(e != [5]), (f not in [6,7])))",
		},
		{
			name:     "joint: complex #1",
			query:    `(a.x.y.z > 1 or b < 2) and (c.o.p.q.r >= 3 and d.s.t.u <= 4) or (e.v.w != 5 and f.m.n not in (6, 7))`,
			expected: "(or:(and:(or:(a.x.y.z > [1]), (b < [2])), (and:(c.o.p.q.r >= [3]), (d.s.t.u <= [4]))), (and:(e.v.w != [5]), (f.m.n not in [6,7])))",
		},
		{
			name:     "null: #0",
			query:    `a = null`,
			expected: "(a = [])",
		},
	}

	cp := NewConditionParser()
	for i, c := range cases {
		i := i
		c := c
		t.Run(
			c.name,
			func(*testing.T) {
				result, err := cp.Parse(c.query)
				if len(c.err) > 0 {
					errString := ""
					if err != nil {
						errString = err.Error()
					}

					assert.Contains(t, errString, c.err, "%d: %v; %v != %v", i, c.name, c.expected, result)
					return
				} else if err != nil {
					assert.NoError(t, err)
					return
				}

				assert.Equal(t, c.expected, result.String(), "%d: %v; %v != %v", i, c.name, c.expected, result)
			},
		)
	}
}

type testConditionJoint struct {
	suite.Suite
}

func (t *testConditionJoint) TestPrepend() {
	query := `(a > 1 or b < 2) and (c >= 3 and d <= 4) or (e != 5 and f not in (6, 7))`
	cp := NewConditionParser()
	cd, err := cp.Parse(query)
	t.NoError(err)

	cmp := NewComparison("node", "equal", []interface{}{"n0"}, reflect.String)

	{
		newcd := cd.(JointConditions).Prepend("and", cmp)
		t.Equal("(and:(node equal [n0]), (or:(and:(or:(a > [1]), (b < [2])), (and:(c >= [3]), (d <= [4]))), (and:(e != [5]), (f not in [6,7]))))", newcd.String())
	}

	{
		newcd := cd.(JointConditions).Prepend("or", cmp)
		t.Equal("(or:(node equal [n0]), (and:(or:(a > [1]), (b < [2])), (and:(c >= [3]), (d <= [4]))), (and:(e != [5]), (f not in [6,7])))", newcd.String())
	}

}

func TestConditionJoint(t *testing.T) {
	suite.Run(t, new(testConditionJoint))
}
