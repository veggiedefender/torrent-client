package console_print

import (
	"reflect"
	"testing"
)

func TestConsolePrint_pushSlice(t *testing.T) {
	type fields struct {
		top       []string
		bottom    []string
		center    []string
		log       []string
		maxCenter int
		maxLog    int
	}
	type args struct {
		a   []string
		max int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "1",
			args: args{
				a:   []string{"1", "2", "3", "4"},
				max: 3,
			},
			want: []string{"2", "3", "4"},
		},
		{
			name: "2",
			args: args{
				a:   []string{"1", "2", "3"},
				max: 3,
			},
			want: []string{"1", "2", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ConsolePrint{
				top:       tt.fields.top,
				bottom:    tt.fields.bottom,
				center:    tt.fields.center,
				log:       tt.fields.log,
				maxCenter: tt.fields.maxCenter,
				maxLog:    tt.fields.maxLog,
			}
			if got := p.pushSlice(tt.args.a, tt.args.max); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pushSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
