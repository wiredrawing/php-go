package main

import (
	"fmt"
	"testing"
)

func Test_hash(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test_hash",
			args: args{
				filepath: "C:\\Users\\a-sen\\works\\php-go\\hash_dummy.txt",
			},
			want: "fd639c4f698a01e97e017cf701b32785c779872121290dd61836b302b2eb618b",
		},
	}
	fmt.Printf("%s", hash("C:\\Users\\a-sen\\works\\php-go\\hash_dummy.txt"))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hash(tt.args.filepath); got != tt.want {
				t.Errorf("hash() = %v,", string(got))
				t.Errorf("hash() = %v, want %v", got, tt.want)
			}
		})
	}
}
