package db

import (
	"strings"
	"testing"
)

func env(vars map[string]string) func(string) string {
	return func(name string) string { return vars[name] }
}

func TestModeFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		vars map[string]string
		want Mode
	}{
		{"nothing set", nil, ModeOff},
		{"url", map[string]string{"DATABASE_URL": "postgres://localhost/app"}, ModePassword},
		{"password", map[string]string{"DATABASE_PASSWORD": "secret"}, ModePassword},
		{"iam", map[string]string{"DATABASE_HOST": "10.0.0.5", "DATABASE_NAME": "app", "DATABASE_USER": "app-sa"}, ModeIAM},
		{"iam missing a var", map[string]string{"DATABASE_HOST": "10.0.0.5", "DATABASE_NAME": "app"}, ModeOff},
		{"password wins over iam", map[string]string{
			"DATABASE_PASSWORD": "secret",
			"DATABASE_HOST":     "10.0.0.5", "DATABASE_NAME": "app", "DATABASE_USER": "app-sa",
		}, ModePassword},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ModeFrom(env(tt.vars)); got != tt.want {
				t.Errorf("ModeFrom() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfigPasswordParts(t *testing.T) {
	t.Parallel()

	cfg, err := config(env(map[string]string{
		"DATABASE_HOST":     "db.example.internal",
		"DATABASE_NAME":     "app",
		"DATABASE_USER":     "app",
		"DATABASE_PASSWORD": "p@ss:w/rd",
	}))
	if err != nil {
		t.Fatal(err)
	}
	conn := cfg.ConnConfig
	if conn.Host != "db.example.internal" || conn.Database != "app" || conn.User != "app" {
		t.Errorf("got host %q, database %q, user %q", conn.Host, conn.Database, conn.User)
	}
	if conn.Password != "p@ss:w/rd" {
		t.Errorf("password did not survive escaping: got %q", conn.Password)
	}
}

func TestConfigIAM(t *testing.T) {
	t.Parallel()

	cfg, err := config(env(map[string]string{
		"DATABASE_HOST": "10.0.0.5",
		"DATABASE_NAME": "app",
		"DATABASE_USER": "app-sa@example-project.iam",
	}))
	if err != nil {
		t.Fatal(err)
	}
	conn := cfg.ConnConfig
	if conn.Host != "10.0.0.5" || conn.Port != 5432 || conn.Database != "app" {
		t.Errorf("got host %q, port %d, database %q", conn.Host, conn.Port, conn.Database)
	}
	if conn.User != "app-sa@example-project.iam" {
		t.Errorf("got user %q", conn.User)
	}
	// IAM mode always encrypts, and never checks the per-instance certificate.
	if conn.TLSConfig == nil || !conn.TLSConfig.InsecureSkipVerify {
		t.Errorf("want TLS without certificate checks, got %+v", conn.TLSConfig)
	}
	if len(conn.Fallbacks) != 0 {
		t.Errorf("want no plaintext fallback, got %d", len(conn.Fallbacks))
	}
}

func TestConfigOff(t *testing.T) {
	t.Parallel()

	_, err := config(env(map[string]string{"DATABASE_HOST": "10.0.0.5"}))
	if err == nil {
		t.Fatal("want an error when nothing is configured")
	}
	if !strings.Contains(err.Error(), "DATABASE_NAME, DATABASE_USER") {
		t.Errorf("error should name the missing variables: %v", err)
	}
}

func TestConfigPasswordWithoutHost(t *testing.T) {
	t.Parallel()

	// No DATABASE_HOST: pgx falls back to its own default rather than failing.
	cfg, err := config(env(map[string]string{"DATABASE_NAME": "app", "DATABASE_USER": "app", "DATABASE_PASSWORD": "pw"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConnConfig.Host == "" {
		t.Error("want pgx's default host, got none")
	}
}
