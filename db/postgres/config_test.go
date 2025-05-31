package postgres

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestGetConnectionConfigFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T)
		want    *ConnectionConfig
		wantErr error
	}{
		{
			name: "set application name",
			prepare: func(t *testing.T) {
				t.Setenv("POSTGRES_HOST", "globalhost")
				t.Setenv("POSTGRES_MAXOPENCONNS", "777")
				t.Setenv("POSTGRES_APPLICATIONNAME", "searcher")
			},
			want: &ConnectionConfig{
				Host:            "globalhost",
				Port:            "5432",
				User:            "default",
				Password:        "default",
				DB:              "default",
				MaxOpenConns:    777,
				ConnMaxLifetime: -1,
				ApplicationName: "searcher",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(t)
			got, err := GetConnectionConfigFromEnv()
			testutils.Equal(t, err, tt.wantErr)
			testutils.Equal(t, tt.want, got)
		})
	}
}
