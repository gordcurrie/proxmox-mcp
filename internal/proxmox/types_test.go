package proxmox

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *APIError
		want string
	}{
		{
			name: "500 internal server error",
			err:  &APIError{StatusCode: 500, Body: `{"message":"internal error"}`},
			want: `proxmox API error 500: {"message":"internal error"}`,
		},
		{
			name: "403 forbidden",
			err:  &APIError{StatusCode: 403, Body: "permission denied"},
			want: "proxmox API error 403: permission denied",
		},
		{
			name: "422 unprocessable",
			err:  &APIError{StatusCode: 422, Body: "invalid parameter"},
			want: "proxmox API error 422: invalid parameter",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestItoa(t *testing.T) {
	t.Parallel()

	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "10"},
		{42, "42"},
		{100, "100"},
		{403, "403"},
		{500, "500"},
		{1234567890, "1234567890"},
		{-1, "-1"},
		{-42, "-42"},
		{-500, "-500"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := itoa(tc.n); got != tc.want {
				t.Errorf("itoa(%d) = %q, want %q", tc.n, got, tc.want)
			}
		})
	}
}
