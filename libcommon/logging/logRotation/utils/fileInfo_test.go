package utils

import "testing"

func TestFileInfo_GenerateFileSize(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "test-20B",
			args: struct{ line string }{line: "20B"},
			want: 20,
		},
		{
			name: "test-20K",
			args: struct{ line string }{line: "20K"},
			want: 20480,
		},
		{
			name: "test-20M",
			args: struct{ line string }{line: "20M"},
			want: 20971520,
		},
		{
			name: "test-20G",
			args: struct{ line string }{line: "20G"},
			want: 21474836480,
		},
		{
			name: "test-20KB",
			args: struct{ line string }{line: "20KB"},
			want: 20480,
		},
		{
			name: "test-20MB",
			args: struct{ line string }{line: "20MB"},
			want: 20971520,
		},
		{
			name: "test-20GB",
			args: struct{ line string }{line: "20GB"},
			want: 21474836480,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileSize{}
			if got := f.GenerateFileSize(tt.args.line); got != tt.want {
				t.Errorf("GenerateFileSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
