package condition

import (
	"sort"
	"testing"
)

func TestExtractTopLevelVars(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       []string
		wantErr    bool
	}{
		// ── Simple identifiers ───────────────────────────────────────
		{
			name:       "single simple variable",
			expression: `age >= 18`,
			want:       []string{"age"},
		},
		{
			name:       "multiple simple variables",
			expression: `age >= 18 && name == "Alice"`,
			want:       []string{"age", "name"},
		},
		// ── Select expressions ───────────────────────────────────────
		{
			name:       "single-level select returns root only",
			expression: `user.age >= 18`,
			want:       []string{"user"},
		},
		{
			name:       "multi-level select returns root only",
			expression: `user.address.city == "Beijing"`,
			want:       []string{"user"},
		},
		{
			name:       "multiple distinct roots",
			expression: `user.age >= 18 && order.amount > 100`,
			want:       []string{"order", "user"},
		},
		{
			name:       "same root deduplicated",
			expression: `user.age >= 18 && user.name == "Alice"`,
			want:       []string{"user"},
		},
		// ── Function calls ───────────────────────────────────────────
		{
			name:       "global function",
			expression: `size(items) > 0`,
			want:       []string{"items"},
		},
		{
			name:       "member function",
			expression: `items.size() > 0`,
			want:       []string{"items"},
		},
		{
			name:       "member function with variable argument",
			expression: `name.startsWith(prefix)`,
			want:       []string{"name", "prefix"},
		},
		// ── List and map literals ────────────────────────────────────
		{
			name:       "variables inside list literal",
			expression: `[a, b, c].size() == 3`,
			want:       []string{"a", "b", "c"},
		},
		{
			name:       "variables inside map literal",
			expression: `{"key": value}.key == expected`,
			want:       []string{"expected", "value"},
		},
		// ── Comprehensions (macro expansion) ─────────────────────────
		{
			name:       "exists: iter variable excluded",
			expression: `items.exists(x, x > threshold)`,
			want:       []string{"items", "threshold"},
		},
		{
			name:       "exists inner exists: iter variable excluded",
			expression: `items.exists(x, items2.exists(y, x > y))`,
			want:       []string{"items", "items2"},
		},
		{
			name:       "all: iter variable excluded",
			expression: `scores.all(s, s >= passing)`,
			want:       []string{"passing", "scores"},
		},
		{
			name:       "filter: iter variable excluded, outer variable included",
			expression: `items.filter(x, x.price < budget).size() > 0`,
			want:       []string{"budget", "items"},
		},
		{
			name:       "map macro: iter variable excluded",
			expression: `nums.map(n, n * factor)`,
			want:       []string{"factor", "nums"},
		},
		{
			name:       "nested comprehensions each bind their own iter variable",
			expression: `outer.exists(a, inner.exists(b, a + b > limit))`,
			want:       []string{"inner", "limit", "outer"},
		},
		// ── Pure literals (no variables) ─────────────────────────────
		{
			name:       "pure numeric literal",
			expression: `1 + 2 == 3`,
			want:       []string{},
		},
		{
			name:       "boolean literals",
			expression: `true || false`,
			want:       []string{},
		},
		// ── Error handling ───────────────────────────────────────────
		{
			name:       "syntax error returns error",
			expression: `age >=`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTopLevelVars(tt.expression)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ExtractTopLevelVars(%q) error = %v, wantErr %v", tt.expression, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			sort.Strings(got)
			sort.Strings(tt.want)
			if !equalSlices(got, tt.want) {
				t.Errorf("ExtractTopLevelVars(%q)\n  got  %v\n  want %v", tt.expression, got, tt.want)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
