package schema_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMigrationPairsAreTracked(t *testing.T) {
	// FR-015/T039 explicitly leave this gate unmet when no Git checkout exists.
	root := targetRoot(t)
	for _, pair := range []struct{ track, name string }{
		{"schema", "complete_mysql_schema"},
		{"credentials", "replace_legacy_password"},
		{"credentials-finalize", "require_password_hash"},
	} {
		for _, direction := range []string{"up", "down"} {
			path := filepath.Join("migrations", pair.track, "000001_"+pair.name+"."+direction+".sql")
			cmd := exec.Command("git", "ls-files", "--error-unmatch", "--", path)
			cmd.Dir = root
			if err := cmd.Run(); err != nil {
				t.Errorf("required versioned migration is not tracked in the implementation Git checkout: %s", path)
			}
		}
	}
}

func TestReviewableValidationEvidence(t *testing.T) {
	// This checks the required handoff inventory, not the truth or chronology of
	// historical claims. The independent executable tests remain the SQL oracle.
	p := filepath.Join(targetRoot(t), "specs/001-complete-mysql-schema/validation.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal("missing reviewable validation.md")
	}
	doc := string(b)
	tokens := []string{
		"TestConfig", "TestEnvironmentPreflight", "TestEnvironment",
		"TestSchemaDefinitions", "TestSchemaValues", "TestSchemaReferences",
		"TestSchemaLifecycle", "TestCredentialUpgrade", "TestCredentialFinalization",
		"SC-001", "SC-002", "SC-003", "SC-004", "SC-005",
		"US1", "US2", "US3", "D1", "D2", "D3", "D4", "D5",
		"STRICT_ALL_TABLES", "NO_ENGINE_SUBSTITUTION", "NO_AUTO_VALUE_ON_ZERO",
		"utf8mb4_0900_bin", "credentials-finalize",
	}
	for i := 1; i <= 16; i++ {
		tokens = append(tokens, fmt.Sprintf("FR-%03d", i))
	}
	for i := 1; i <= 40; i++ {
		tokens = append(tokens, fmt.Sprintf("T%03d", i))
	}
	for _, token := range tokens {
		if !strings.Contains(doc, token) {
			t.Errorf("validation handoff omits required case/source/task/environment accounting: %s", token)
		}
	}
	for _, token := range []string{"red", "green", "dirty", "quiesc", "digest", "exit"} {
		if !strings.Contains(strings.ToLower(doc), token) {
			t.Errorf("validation handoff omits %s evidence", token)
		}
	}
}

func TestImplementationRequiredGatesAndRefusals(t *testing.T) {
	f := newFixture(t)
	// The plan specifies no exported API for the test-only entry preflights.
	// Invoke their named acceptance suites through Go's public test interface.
	// This also checks they are real, executed, non-skipping tests, rather than
	// counting a successful `go test` with no matching tests as proof of refusal.
	out := runCommand(t, f, 15*time.Minute, "go", "test", "-race", "-count=1", "-json", "./internal/config", "./tests/schema")
	type event struct{ Action, Package, Test string }
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	for scanner.Scan() {
		var e event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}
		if e.Action == "skip" {
			t.Error("required implementation acceptance suite skipped a test")
		}
		if e.Action == "pass" && e.Test != "" {
			seen[e.Test] = true
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal("could not parse implementation test events")
	}
	for _, name := range []string{"TestConfig", "TestEnvironmentPreflight", "TestEnvironment", "TestSchemaDefinitions", "TestSchemaValues", "TestSchemaReferences", "TestSchemaLifecycle", "TestCredentialUpgrade", "TestCredentialFinalization"} {
		if !seen[name] {
			t.Errorf("required implementation test did not execute and pass: %s", name)
		}
	}
	// T003/T036/T038 require the actual formatting, imports, vet, lint and race
	// gate. Run the delivered entry point; a missing tool is an unmet gate.
	runCommand(t, f, 15*time.Minute, "make", "schema-check")
}
