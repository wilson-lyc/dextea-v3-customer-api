package order

import "testing"

func TestStripAPIPrefix(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"order create", "/api/v1/orders", "/orders"},
		{"order detail with id", "/api/v1/orders/123", "/orders/123"},
		{"order status", "/api/v1/orders/123/status", "/orders/123/status"},
		{"pre-build", "/api/v1/orders/pre-build", "/orders/pre-build"},
		{"exact prefix", "/api/v1", "/"},
		{"similar prefix not stripped", "/api/v10/orders", "/api/v10/orders"},
		{"non prefix", "/other/orders", "/other/orders"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := stripAPIPrefix(c.in); got != c.want {
				t.Fatalf("stripAPIPrefix(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestBindCustomerOwnership(t *testing.T) {
	const customerID int64 = 42

	cases := []struct {
		name string
		body string
		want string
	}{
		{"override numeric customer_id", `{"customer_id": 1, "items": []}`, `{"customer_id":42,"items":[]}`},
		{"override string customerId", `{"customerId": "1"}`, `{"customerId":"42"}`},
		{"override user_id", `{"user_id": 1}`, `{"user_id":42}`},
		{"same value unchanged", `{"customer_id": 42}`, `{"customer_id": 42}`},
		{"no customer field unchanged", `{"a": 1}`, `{"a": 1}`},
		{"non json passthrough", `not-json`, `not-json`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := string(bindCustomerOwnership([]byte(c.body), customerID)); got != c.want {
				t.Fatalf("bindCustomerOwnership(%q) = %q, want %q", c.body, got, c.want)
			}
		})
	}
}
