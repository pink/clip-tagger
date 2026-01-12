// flags/flags_test.go
package flags

import (
	"flag"
	"os"
	"testing"
)

// resetFlags resets the flag package for testing
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestParse_ValidDirectory(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "/tmp"}

	config, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Directory != "/tmp" {
		t.Errorf("expected directory '/tmp', got '%s'", config.Directory)
	}
}

func TestParse_MissingDirectory(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd"}

	_, err := Parse()
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func TestParse_SortByName(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "--sort-by=name", "/tmp"}

	config, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.SortBy != "name" {
		t.Errorf("expected sort-by 'name', got '%s'", config.SortBy)
	}
}

func TestParse_SortByModified(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "--sort-by=modified", "/tmp"}

	config, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.SortBy != "modified" {
		t.Errorf("expected sort-by 'modified', got '%s'", config.SortBy)
	}
}

func TestParse_SortByCreated(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "--sort-by=created", "/tmp"}

	config, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.SortBy != "created" {
		t.Errorf("expected sort-by 'created', got '%s'", config.SortBy)
	}
}

func TestParse_InvalidSortBy(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "--sort-by=invalid", "/tmp"}

	_, err := Parse()
	if err == nil {
		t.Fatal("expected error for invalid sort-by value")
	}
}

func TestParse_ResetFlag(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "--reset", "/tmp"}

	config, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !config.Reset {
		t.Error("expected reset flag to be true")
	}
}

func TestParse_CleanMissingFlag(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "--clean-missing", "/tmp"}

	config, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !config.CleanMissing {
		t.Error("expected clean-missing flag to be true")
	}
}

func TestParse_PreviewFlag(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "--preview", "/tmp"}

	config, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !config.Preview {
		t.Error("expected preview flag to be true")
	}
}

func TestParse_MultipleFlags(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "--sort-by=name", "--clean-missing", "--preview", "/tmp"}

	config, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.SortBy != "name" {
		t.Errorf("expected sort-by 'name', got '%s'", config.SortBy)
	}
	if !config.CleanMissing {
		t.Error("expected clean-missing flag to be true")
	}
	if !config.Preview {
		t.Error("expected preview flag to be true")
	}
}

func TestParse_DefaultValues(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "/tmp"}

	config, err := Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.SortBy != "" {
		t.Errorf("expected empty sort-by, got '%s'", config.SortBy)
	}
	if config.Reset {
		t.Error("expected reset flag to be false")
	}
	if config.CleanMissing {
		t.Error("expected clean-missing flag to be false")
	}
	if config.Preview {
		t.Error("expected preview flag to be false")
	}
}
