package main

import "testing"

func Test_parseAaguidAsUuid(t *testing.T) {
	type args struct {
		aaguid []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1Password",
			args: args{
				aaguid: []byte{186, 218, 85, 102, 167, 170, 64, 31, 189, 150, 69, 97, 154, 85, 18, 13},
			},
			want: "bada5566-a7aa-401f-bd96-45619a55120d",
		},
		{
			name: "Chrome on Mac",
			args: args{
				aaguid: []byte{173, 206, 0, 2, 53, 188, 198, 10, 100, 139, 11, 37, 241, 240, 85, 3},
			},
			want: "adce0002-35bc-c60a-648b-0b25f1f05503",
		},
		{
			name: "iCloud Keychain",
			args: args{
				aaguid: []byte{251, 252, 48, 7, 21, 78, 78, 204, 140, 11, 110, 2, 5, 87, 215, 189},
			},
			want: "fbfc3007-154e-4ecc-8c0b-6e020557d7bd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseAaguidAsUuid(tt.args.aaguid); got != tt.want {
				t.Errorf("parseAaguidAsUuid() = %v, want %v", got, tt.want)
			}
		})
	}
}
