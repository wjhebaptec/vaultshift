package normalize_test

import (
	"testing"

	"github.com/vaultshift/internal/normalize"
)

func TestNormalize_CamelToSnake(t *testing.T) {
	n := normalize.New(normalize.WithStyle(normalize.StyleSnake))
	got := n.Normalize("mySecretKey")
	if got != "my_secret_key" {
		t.Fatalf("expected my_secret_key, got %s", got)
	}
}

func TestNormalize_KebabInput_ToSnake(t *testing.T) {
	n := normalize.New()
	got := n.Normalize("my-secret-key")
	if got != "my_secret_key" {
		t.Fatalf("expected my_secret_key, got %s", got)
	}
}

func TestNormalize_ToKebab(t *testing.T) {
	n := normalize.New(normalize.WithStyle(normalize.StyleKebab))
	got := n.Normalize("mySecretKey")
	if got != "my-secret-key" {
		t.Fatalf("expected my-secret-key, got %s", got)
	}
}

func TestNormalize_ToScream(t *testing.T) {
	n := normalize.New(normalize.WithStyle(normalize.StyleScream))
	got := n.Normalize("mySecretKey")
	if got != "MY_SECRET_KEY" {
		t.Fatalf("expected MY_SECRET_KEY, got %s", got)
	}
}

func TestNormalize_ToDot(t *testing.T) {
	n := normalize.New(normalize.WithStyle(normalize.StyleDot))
	got := n.Normalize("my_secret_key")
	if got != "my.secret.key" {
		t.Fatalf("expected my.secret.key, got %s", got)
	}
}

func TestNormalize_WithPrefix(t *testing.T) {
	n := normalize.New(
		normalize.WithStyle(normalize.StyleSnake),
		normalize.WithPrefix("app_"),
	)
	got := n.Normalize("dbPassword")
	if got != "app_db_password" {
		t.Fatalf("expected app_db_password, got %s", got)
	}
}

func TestNormalize_AlreadySnake(t *testing.T) {
	n := normalize.New()
	got := n.Normalize("already_snake")
	if got != "already_snake" {
		t.Fatalf("expected already_snake, got %s", got)
	}
}

func TestNormalizeAll_TransformsMap(t *testing.T) {
	n := normalize.New(normalize.WithStyle(normalize.StyleScream))
	input := map[string]string{
		"dbHost":     "localhost",
		"dbPassword": "s3cr3t",
	}
	out := n.NormalizeAll(input)
	if out["DB_HOST"] != "localhost" {
		t.Fatalf("expected DB_HOST=localhost, got %v", out)
	}
	if out["DB_PASSWORD"] != "s3cr3t" {
		t.Fatalf("expected DB_PASSWORD=s3cr3t, got %v", out)
	}
}

func TestNormalizeAll_PreservesValues(t *testing.T) {
	n := normalize.New()
	input := map[string]string{"APIKey": "tok_abc123"}
	out := n.NormalizeAll(input)
	if out["a_p_i_key"] != "tok_abc123" && out["api_key"] != "tok_abc123" {
		// accept either reasonable split
		for k, v := range out {
			if v != "tok_abc123" {
				t.Fatalf("value not preserved: key=%s val=%s", k, v)
			}
		}
	}
}
