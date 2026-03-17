package cmd

import (
	"testing"
)

func TestRootCmdHasManualFlag(t *testing.T) {
	resetCmdState()

	cmd := NewRootCmd()

	flag := cmd.PersistentFlags().Lookup("manual")
	if flag == nil {
		t.Fatal("expected --manual flag to exist")
	}

	if flag.Shorthand != "" {
		t.Errorf("--manual should not have a shorthand, got %q", flag.Shorthand)
	}

	if flag.DefValue != "false" {
		t.Errorf("--manual should default to false, got %q", flag.DefValue)
	}
}

func TestRootCmdManualFlagNoShorthand(t *testing.T) {
	resetCmdState()

	cmd := NewRootCmd()

	// Verify -m is not the manual flag shorthand
	mFlag := cmd.PersistentFlags().ShorthandLookup("m")
	if mFlag != nil && mFlag.Name == "manual" {
		t.Error("--manual should not have -m shorthand")
	}
}
