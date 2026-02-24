package proxmox

import (
	"encoding/json"
	"log/slog"
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

func TestSensitiveString_Redaction(t *testing.T) {
	t.Parallel()

	const secret = "s3cr3t-p@ssw0rd"
	s := SensitiveString(secret)

	t.Run("String() redacts", func(t *testing.T) {
		if got := s.String(); got != "[REDACTED]" {
			t.Errorf("String() = %q, want \"[REDACTED]\"", got)
		}
	})

	t.Run("LogValue() redacts", func(t *testing.T) {
		got := s.LogValue()
		want := slog.StringValue("[REDACTED]")
		if !got.Equal(want) {
			t.Errorf("LogValue() = %v, want %v", got, want)
		}
	})
}

func TestSensitiveString_JSON(t *testing.T) {
	t.Parallel()

	const secret = "s3cr3t-p@ssw0rd"
	s := SensitiveString(secret)

	t.Run("MarshalJSON emits real value", func(t *testing.T) {
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}
		// json.Marshal of a plain string would give `"s3cr3t-p@ssw0rd"` (quoted)
		want := `"` + secret + `"`
		if string(data) != want {
			t.Errorf("MarshalJSON = %s, want %s", data, want)
		}
	})

	t.Run("UnmarshalJSON round-trips", func(t *testing.T) {
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("MarshalJSON: %v", err)
		}
		var got SensitiveString
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("UnmarshalJSON: %v", err)
		}
		if string(got) != secret {
			t.Errorf("round-trip got %q, want %q", string(got), secret)
		}
	})

	t.Run("UnmarshalJSON rejects non-string", func(t *testing.T) {
		var got SensitiveString
		if err := json.Unmarshal([]byte(`123`), &got); err == nil {
			t.Error("expected error unmarshalling number, got nil")
		}
	})
}
