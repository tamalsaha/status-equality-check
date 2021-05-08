package main

import "testing"

func TestStatusEqual(t *testing.T) {
	type args struct {
		old interface{}
		new interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusEqual(tt.args.old, tt.args.new); got != tt.want {
				t.Errorf("StatusEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
