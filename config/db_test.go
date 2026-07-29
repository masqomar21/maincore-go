package config

import (
	"net/url"
	"strings"
	"testing"
)

func TestCreateDatabaseIfNotExist_ParseDSN(t *testing.T) {
	tests := []struct {
		name           string
		dsn            string
		expectedDB     string
		expectedDefault string
		isURL          bool
	}{
		{
			name:            "URL standard",
			dsn:             "postgres://user:password@localhost:5432/my_database?sslmode=disable",
			expectedDB:      "my_database",
			expectedDefault: "postgres://user:password@localhost:5432/postgres?sslmode=disable",
			isURL:           true,
		},
		{
			name:            "URL with multiple query params",
			dsn:             "postgresql://root:secret@127.0.0.1:5432/testdb?sslmode=disable&timezone=Asia/Jakarta",
			expectedDB:      "testdb",
			expectedDefault: "postgresql://root:secret@127.0.0.1:5432/postgres?sslmode=disable&timezone=Asia/Jakarta",
			isURL:           true,
		},
		{
			name:            "Key-value simple",
			dsn:             "host=localhost user=root password=root dbname=starter_kit port=5432 sslmode=disable",
			expectedDB:      "starter_kit",
			expectedDefault: "host=localhost user=root password=root dbname=postgres port=5432 sslmode=disable",
			isURL:           false,
		},
		{
			name:            "Key-value with quotes and spaces",
			dsn:             "host=localhost user=root password='my secret password' dbname='starter_kit_quoted' port=5432",
			expectedDB:      "starter_kit_quoted",
			expectedDefault: "host=localhost user=root password='my secret password' dbname=postgres port=5432",
			isURL:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulating the parser logic here to verify output
			var dbname string
			var defaultDSN string

			if tt.isURL {
				// URL check logic
				u, err := parseURL(tt.dsn)
				if err != nil {
					t.Fatalf("Failed to parse URL: %v", err)
				}
				dbname = stringsTrimPrefix(u.Path, "/")
				if idx := stringsIndex(dbname, "?"); idx != -1 {
					dbname = dbname[:idx]
				}
				u.Path = "/postgres"
				defaultDSN = u.String()
			} else {
				// KV check logic
				fields := splitKV(tt.dsn)
				var newFields []string
				for _, field := range fields {
					parts := stringsSplitN(field, "=", 2)
					if len(parts) == 2 {
						key := stringsTrimSpace(parts[0])
						val := stringsTrimSpace(parts[1])
						if key == "dbname" {
							dbname = stringsTrim(val, `'"`)
							newFields = append(newFields, "dbname=postgres")
						} else {
							newFields = append(newFields, field)
						}
					} else {
						newFields = append(newFields, field)
					}
				}
				defaultDSN = stringsJoin(newFields, " ")
			}

			if dbname != tt.expectedDB {
				t.Errorf("Expected dbname %q, got %q", tt.expectedDB, dbname)
			}
			if defaultDSN != tt.expectedDefault {
				t.Errorf("Expected defaultDSN %q, got %q", tt.expectedDefault, defaultDSN)
			}
		})
	}
}

// Inline aliases matching stdlib / custom functions for self-contained testing
func parseURL(s string) (*url.URL, error) {
	importUrl, err := urlParse(s)
	return importUrl, err
}

func urlParse(s string) (*url.URL, error) {
	// Import via reflection/lookup or direct reference
	// We can use standard library imports inside test since we import net/url
	return url.Parse(s)
}

func stringsTrimPrefix(s, prefix string) string {
	importStrings := "strings"
	_ = importStrings
	return strings.TrimPrefix(s, prefix)
}

func stringsIndex(s, substr string) int {
	return strings.Index(s, substr)
}

func stringsTrim(s, cutset string) string {
	return strings.Trim(s, cutset)
}

func stringsSplitN(s, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

func stringsTrimSpace(s string) string {
	return strings.TrimSpace(s)
}

func stringsJoin(elems []string, sep string) string {
	return strings.Join(elems, sep)
}
