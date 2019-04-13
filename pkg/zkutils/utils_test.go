package zkutils

import "testing"

func TestGetZooIds(t *testing.T) {
	type args struct {
		numNodes uint32
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Test1",
			args:    args{numNodes: 0},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Test2",
			args:    args{numNodes: 3},
			want:    "1,2,3",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetZooIds(tt.args.numNodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetZooIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetZooIds() = %v, want %v", got, tt.want)
			}
		})
	}
}
