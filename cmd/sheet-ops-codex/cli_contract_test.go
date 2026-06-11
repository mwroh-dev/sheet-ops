package main

import (
	"bytes"
	"slices"
	"strings"
	"testing"
)

func TestCLICommandContractsCoverEverySurface(t *testing.T) {
	root := newRootCommand()
	contracts := sheetOpsCLIContracts(root.Name())

	wantNames := []string{
		root.Name(),
		"completion",
		"help",
		"install-skill",
		"prepare-use",
		"run-intent",
		"run-request",
		"run-validated",
	}
	slices.Sort(wantNames)
	if got := sortedContractNames(contracts); !slices.Equal(got, wantNames) {
		t.Fatalf("contract names = %v, want %v", got, wantNames)
	}

	for _, name := range wantNames {
		contract, ok := contracts[name]
		if !ok {
			t.Fatalf("missing contract for %q", name)
		}
		assertCLIContractField(t, name, "name", contract.Name)
		assertCLIContractField(t, name, "classification", contract.Classification)
		assertCLIContractField(t, name, "intended caller", contract.IntendedCaller)
		assertCLIContractField(t, name, "short description", contract.ShortDescription)
		assertCLIContractField(t, name, "usage", contract.Usage)
		assertCLIContractSlice(t, name, "options", contract.Options)
		assertCLIContractField(t, name, "output mode", contract.OutputMode)
		assertCLIContractSlice(t, name, "side effects", contract.SideEffects)
		assertCLIContractSlice(t, name, "read artifacts", contract.ReadArtifacts)
		assertCLIContractSlice(t, name, "written artifacts", contract.WrittenArtifacts)
		assertCLIContractField(t, name, "state behavior", contract.StateBehavior)
		assertCLIContractSlice(t, name, "safety notes", contract.SafetyNotes)
		assertCLIContractSlice(t, name, "related commands", contract.RelatedCommands)
	}

	assertCLIContractClassification(t, contracts, "install-skill", "public_install")
	assertCLIContractClassification(t, contracts, "prepare-use", "agent_contract")
	assertCLIContractClassification(t, contracts, "run-validated", "internal_handoff")
	assertCLIContractClassification(t, contracts, "run-intent", "maintainer_diagnostic")
	assertCLIContractClassification(t, contracts, "run-request", "compatibility_internal")
}

func TestCLIHelpSeparatesCommandSurfaces(t *testing.T) {
	root := newRootCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute(--help): %v", err)
	}

	help := stdout.String()
	for _, want := range []string{
		"Use the `sheet-ops` skill for workbook requests.",
		"Install surface:",
		"Agent-contract surface:",
		"Internal handoff surface:",
		"Maintainer diagnostic surface:",
		"install-skill",
		"preflight",
		"preview-request",
		"prepare-use",
		"run-validated",
		"run-intent",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("help output missing %q:\n%s", want, help)
		}
	}

	for _, forbidden := range []string{
		"sheet-ops-agent use",
		"second human-facing workbook entry",
	} {
		if strings.Contains(help, forbidden) {
			t.Fatalf("help output unexpectedly contains %q:\n%s", forbidden, help)
		}
	}
}

func TestCLIHelpAlignsCommandDescriptions(t *testing.T) {
	root := newRootCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute(--help): %v", err)
	}

	contracts := sheetOpsCLIAllContracts(root.Name())
	descriptionColumn := -1
	for _, line := range strings.Split(stdout.String(), "\n") {
		if !strings.HasPrefix(line, "  ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		contract, ok := contracts[fields[0]]
		if !ok || contract.Hidden || contract.Name == root.Name() {
			continue
		}

		column := strings.Index(line, contract.ShortDescription)
		if column < 0 {
			t.Fatalf("help line for %q is missing description %q:\n%s", contract.Name, contract.ShortDescription, line)
		}
		if descriptionColumn == -1 {
			descriptionColumn = column
			continue
		}
		if column != descriptionColumn {
			t.Fatalf("description column for %q = %d, want %d:\n%s", contract.Name, column, descriptionColumn, stdout.String())
		}
	}
}

func assertCLIContractClassification(t *testing.T, contracts map[string]cliCommandContract, name, want string) {
	t.Helper()

	contract, ok := contracts[name]
	if !ok {
		t.Fatalf("missing contract for %q", name)
	}
	if contract.Classification != want {
		t.Fatalf("%s classification = %q, want %q", name, contract.Classification, want)
	}
}

func assertCLIContractField(t *testing.T, commandName, fieldName, value string) {
	t.Helper()

	if strings.TrimSpace(value) == "" {
		t.Fatalf("%s %s is empty", commandName, fieldName)
	}
}

func assertCLIContractSlice(t *testing.T, commandName, fieldName string, values []string) {
	t.Helper()

	if len(values) == 0 {
		t.Fatalf("%s %s is empty", commandName, fieldName)
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			t.Fatalf("%s %s contains an empty item", commandName, fieldName)
		}
	}
}
