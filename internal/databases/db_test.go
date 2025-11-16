package databases

import (
	"testing"

	"github.com/merkulovlad/avito-internship-test/internal/config"
)

func TestNewDB_PingFails(t *testing.T) {
	fakeCfg := &config.DatabaseConfig{
		Host:              "wrong",
		Port:              59999,
		User:              "wrong",
		Password:          "wrong",
		Name:              "wrong",
		SSLMode:           "disable",
		MaxConnections:    1,
		ConnectionTimeout: 10,
		ConnMaxLifetime:   10,
	}

	db, err := NewDB(fakeCfg)
	if err == nil {
		if db != nil {
			_ = db.Close()
		}

		t.Fatalf("expected error, got nil")
	}
}
