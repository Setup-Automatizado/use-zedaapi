package groups

import "testing"

func TestIsValidDescriptionID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "valid web client id uppercase",
			id:   "3EB08986D52542AD55611C",
			want: true,
		},
		{
			name: "valid web client id lowercase",
			id:   "3eb08986d52542ad55611c",
			want: true,
		},
		{
			name: "valid mixed case",
			id:   "3Eb08986D52542aD55611C",
			want: true,
		},
		{
			name: "valid longer id",
			id:   "3EB08986D52542AD55611C0123456789ABCDEF",
			want: true,
		},
		{
			name: "undefined string",
			id:   "undefined",
			want: false,
		},
		{
			name: "null string",
			id:   "null",
			want: false,
		},
		{
			name: "empty string",
			id:   "",
			want: false,
		},
		{
			name: "too short valid hex",
			id:   "3EB0",
			want: false,
		},
		{
			name: "exactly 15 chars hex",
			id:   "3EB08986D52542A",
			want: false,
		},
		{
			name: "exactly 16 chars hex",
			id:   "3EB08986D52542AD",
			want: true,
		},
		{
			name: "non-hex characters",
			id:   "3EB0ZZZZZZZZZZZZ",
			want: false,
		},
		{
			name: "contains spaces",
			id:   "3EB08986 D52542AD",
			want: false,
		},
		{
			name: "contains special chars",
			id:   "3EB08986-D52542AD",
			want: false,
		},
		{
			name: "numeric only valid",
			id:   "1234567890123456",
			want: true,
		},
		{
			name: "letters only valid",
			id:   "ABCDEFABCDEFABCD",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidDescriptionID(tt.id)
			if got != tt.want {
				t.Errorf("IsValidDescriptionID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}
