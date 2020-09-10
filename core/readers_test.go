package core

import (
	"reflect"
	"testing"

	"github.com/binacsgo/log"
)

func Test_reader_buildNewData(t *testing.T) {
	type fields struct {
		manager *ManagerService
		logger  log.Logger
		ipcFile string
	}
	type args struct {
		oridata string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   updata
	}{
		// Add test cases.
		{
			name: "test1",
			fields: fields{
				logger: log.NewNopLogger(),
			},
			args: args{
				oridata: "test_oridata_eg{instance=\"test_instance\"}",
			},
			want: updata{
				data:     "test_oridata_eg{instance=\"test_instance\"}",
				instance: "test_instance",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reader{
				manager: tt.fields.manager,
				logger:  tt.fields.logger,
				ipcFile: tt.fields.ipcFile,
			}
			if got := r.buildNewData(tt.args.oridata); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reader.buildNewData() = %v, want %v", got, tt.want)
			}
		})
	}
}
