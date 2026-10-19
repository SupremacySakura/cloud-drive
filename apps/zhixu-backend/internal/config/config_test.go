package config

import (
	"encoding/base64"
	"net"
	"os"
	"path/filepath"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestModelEncryptionKeyConfiguration(t *testing.T) {
	for _, key := range []string{"", base64.StdEncoding.EncodeToString(make([]byte, 32))} {
		cfg := Config{JWTSecret: "test", DBUser: "root", DBHost: "localhost", DBPort: "3306", DBName: "zhixu", ModelConfigEncryptionKey: key}
		require.NoError(t, cfg.validate())
	}
	cfg := Config{ModelConfigEncryptionKey: "invalid-secret"}
	err := cfg.validate()
	require.Error(t, err)
	require.NotContains(t, err.Error(), "invalid-secret")
	v := viper.New()
	bindEnvVariables(v)
	t.Setenv("MODEL_CONFIG_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	var loaded Config
	require.NoError(t, v.Unmarshal(&loaded))
	require.NotEmpty(t, loaded.ModelConfigEncryptionKey)
}

func TestLoadEnvFileUsesExplicitDevelopmentFile(t *testing.T) {
	dir := t.TempDir()
	development := filepath.Join(dir, ".env.development")
	require.NoError(t, os.WriteFile(development, []byte("DB_HOST=127.0.0.1\nDB_NAME=zhixu\nPORT=9100\n"), 0600))
	t.Setenv("ZHIXU_ENV_FILE", development)

	v := viper.New()
	require.NoError(t, loadEnvFile(v))
	require.Equal(t, "127.0.0.1", v.GetString("DB_HOST"))
	require.Equal(t, "zhixu", v.GetString("DB_NAME"))
	require.Equal(t, "9100", v.GetString("PORT"))
}

func TestExplicitMissingEnvFileDoesNotFallBackToDockerEnv(t *testing.T) {
	t.Setenv("ZHIXU_ENV_FILE", filepath.Join(t.TempDir(), ".env.development"))
	require.Error(t, loadEnvFile(viper.New()))
}

func TestDatabaseDefaultUsesZhixu(t *testing.T) {
	v := viper.New()
	setDefaults(v)
	require.Equal(t, "zhixu", v.GetString("DB_NAME"))
}

func TestGetDSNFormatsDatabaseAddress(t *testing.T) {
	for _, scenario := range []struct {
		host, port, address string
	}{
		{"127.0.0.1", "3306", "127.0.0.1:3306"},
		{"localhost", "3306", "localhost:3306"},
		{"mysql", "3306", "mysql:3306"},
		{"::1", "3306", "[::1]:3306"},
		{"2001:db8::1", "3307", "[2001:db8::1]:3307"},
	} {
		t.Run(scenario.host, func(t *testing.T) {
			cfg := Config{
				DBUser: "testuser", DBPassword: "testpassword",
				DBHost: scenario.host, DBPort: scenario.port, DBName: "zhixu",
			}
			dsn := cfg.GetDSN()
			require.Equal(t, "testuser:testpassword@tcp("+scenario.address+")/zhixu?charset=utf8mb4&parseTime=True&loc=Local", dsn)

			parsed, err := mysqlDriver.ParseDSN(dsn)
			require.NoError(t, err)
			require.Equal(t, scenario.address, parsed.Addr)
			host, port, err := net.SplitHostPort(parsed.Addr)
			require.NoError(t, err)
			require.Equal(t, scenario.host, host)
			require.Equal(t, scenario.port, port)
		})
	}
}
