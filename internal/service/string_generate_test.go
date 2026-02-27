package service

import "testing"

func TestGenerateRandomString(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateRandomString(tt.args.length); got != tt.want {
				t.Errorf("GenerateRandomString() = %v, want %v", got, tt.want)
			}
		})
	}
}
