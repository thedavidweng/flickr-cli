package cli

import (
	"bytes"
	"testing"
)

func TestDoctorHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"doctor", "--help"})
	rootCmd.Execute()

	if buf.Len() == 0 {
		t.Error("expected help output")
	}
}

func TestDoctorRunE(t *testing.T) {
	t.Run("valid config all checks pass", func(t *testing.T) {
		fake, cfg := setupFakeCLI(t)

		cmd, buf := cmdContext(t, cfg, true)
		err := doctorCmd.RunE(cmd, nil)
		if err != nil {
			t.Fatalf("RunE returned error: %v", err)
		}

		env := parseEnvelope(t, buf)
		if !env.OK {
			t.Fatalf("expected ok=true, got error: %v", env.Error)
		}
		if env.Meta.Command != "doctor" {
			t.Errorf("expected command=doctor, got %s", env.Meta.Command)
		}

		data, ok := env.Data.(map[string]any)
		if !ok {
			t.Fatalf("expected data to be a map, got %T", env.Data)
		}
		checksRaw, ok := data["checks"].([]any)
		if !ok {
			t.Fatalf("expected data.checks to be an array, got %T", data["checks"])
		}

		expectedNames := []string{"config", "profile", "api_key", "oauth", "api_connection"}
		if len(checksRaw) != len(expectedNames) {
			t.Fatalf("expected %d checks, got %d", len(expectedNames), len(checksRaw))
		}

		for i, raw := range checksRaw {
			c, ok := raw.(map[string]any)
			if !ok {
				t.Fatalf("check %d: expected map, got %T", i, raw)
			}
			if c["name"] != expectedNames[i] {
				t.Errorf("check %d: expected name=%s, got %v", i, expectedNames[i], c["name"])
			}
			if c["ok"] != true {
				t.Errorf("check %d (%s): expected ok=true, got %v", i, expectedNames[i], c["ok"])
			}
		}

		if fake.CountMethod("flickr.test.echo") != 1 {
			t.Errorf("expected 1 call to flickr.test.echo, got %d", fake.CountMethod("flickr.test.echo"))
		}
	})

	t.Run("unauthenticated config auth check fails", func(t *testing.T) {
		fake, _ := setupFakeCLI(t)
		cfg := setupUnauthedCLI(t, fake.Server.URL)

		cmd, buf := cmdContext(t, cfg, true)
		err := doctorCmd.RunE(cmd, nil)
		if err != nil {
			t.Fatalf("RunE returned error: %v", err)
		}

		env := parseEnvelope(t, buf)
		if !env.OK {
			t.Fatalf("expected ok=true (doctor always returns success envelope), got error: %v", env.Error)
		}

		data, ok := env.Data.(map[string]any)
		if !ok {
			t.Fatalf("expected data to be a map, got %T", env.Data)
		}
		checksRaw, ok := data["checks"].([]any)
		if !ok {
			t.Fatalf("expected data.checks to be an array, got %T", data["checks"])
		}

		// Should have 4 checks: config, profile, api_key, oauth (fails, api_connection skipped)
		if len(checksRaw) != 4 {
			t.Fatalf("expected 4 checks, got %d", len(checksRaw))
		}

		// First 3 should pass
		for i, name := range []string{"config", "profile", "api_key"} {
			c := checksRaw[i].(map[string]any)
			if c["name"] != name {
				t.Errorf("check %d: expected name=%s, got %v", i, name, c["name"])
			}
			if c["ok"] != true {
				t.Errorf("check %d (%s): expected ok=true, got %v", i, name, c["ok"])
			}
		}

		// OAuth check should fail
		oauthCheck := checksRaw[3].(map[string]any)
		if oauthCheck["name"] != "oauth" {
			t.Errorf("expected name=oauth, got %v", oauthCheck["name"])
		}
		if oauthCheck["ok"] != false {
			t.Errorf("expected ok=false for oauth, got %v", oauthCheck["ok"])
		}
		if oauthCheck["message"] == "" {
			t.Error("expected non-empty message for failed oauth check")
		}

		// API connection should not have been tested
		if fake.CountMethod("flickr.test.echo") != 0 {
			t.Errorf("expected 0 calls to flickr.test.echo, got %d", fake.CountMethod("flickr.test.echo"))
		}
	})

	t.Run("human readable output", func(t *testing.T) {
		fake, cfg := setupFakeCLI(t)

		cmd, buf := cmdContext(t, cfg, false)
		err := doctorCmd.RunE(cmd, nil)
		if err != nil {
			t.Fatalf("RunE returned error: %v", err)
		}

		output := buf.String()
		if len(output) == 0 {
			t.Fatal("expected human-readable output")
		}

		// Should contain PASS markers for each check
		for _, name := range []string{"config", "profile", "api_key", "oauth", "api_connection"} {
			expected := "[PASS] " + name
			if !bytes.Contains(buf.Bytes(), []byte(expected)) {
				t.Errorf("expected output to contain %q", expected)
			}
		}

		if !bytes.Contains(buf.Bytes(), []byte("All checks passed.")) {
			t.Error("expected output to contain 'All checks passed.'")
		}

		_ = fake // ensure fake server is used
	})

	t.Run("human readable unauthenticated", func(t *testing.T) {
		fake, _ := setupFakeCLI(t)
		cfg := setupUnauthedCLI(t, fake.Server.URL)

		cmd, buf := cmdContext(t, cfg, false)
		err := doctorCmd.RunE(cmd, nil)
		if err != nil {
			t.Fatalf("RunE returned error: %v", err)
		}

		if !bytes.Contains(buf.Bytes(), []byte("[PASS] config")) {
			t.Error("expected [PASS] config in output")
		}
		if !bytes.Contains(buf.Bytes(), []byte("[FAIL] oauth")) {
			t.Error("expected [FAIL] oauth in output")
		}
		if !bytes.Contains(buf.Bytes(), []byte("Some checks failed.")) {
			t.Error("expected 'Some checks failed.' in output")
		}
	})

	t.Run("nonexistent config file", func(t *testing.T) {
		app := &AppContext{
			ConfigFile: "/nonexistent/config.yaml",
			Profile:    "default",
		}
		checks := doctorRun(t.Context(), app)
		if len(checks) != 1 {
			t.Fatalf("expected 1 check, got %d", len(checks))
		}
		if checks[0].Name != "config" {
			t.Errorf("expected name=config, got %s", checks[0].Name)
		}
		if checks[0].OK {
			t.Error("expected ok=false for nonexistent config")
		}
	})
}
