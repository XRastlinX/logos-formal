package main

import "testing"

func TestParseArgumentsAllowsTargetBeforeOptions(t *testing.T) {
	options, err := parseArguments([]string{
		"runtime/cyonic-service",
		"--manifest",
		"cyonic.validation.json",
		"--format",
		"json",
	})
	if err != nil {
		t.Fatal(err)
	}
	if options.target != "runtime/cyonic-service" ||
		options.manifestPath != "cyonic.validation.json" ||
		options.format != "json" {
		t.Fatalf("unexpected options: %#v", options)
	}
}

func TestParseArgumentsRejectsUnknownOption(t *testing.T) {
	if _, err := parseArguments([]string{"--execute", "whoami"}); err == nil {
		t.Fatal("expected unknown option to be rejected")
	}
}
