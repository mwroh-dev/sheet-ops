package knowledge

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func TestTemplateResearchCorpusRecordsValidate(t *testing.T) {
	root := repoRoot(t)

	rawSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "raw_observation.schema.json"))
	patternSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "template_pattern.schema.json"))
	opportunitySchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "capability_opportunity.schema.json"))

	validateJSONL(t, rawSchema, filepath.Join(root, "knowledge", "template-research", "extracted", "templates.jsonl"))
	validateJSONLFiles(t, rawSchema, filepath.Join(root, "knowledge", "template-research", "raw", "shards", "*.jsonl"), 9)
	validateJSONFiles(t, patternSchema, filepath.Join(root, "knowledge", "template-research", "patterns", "*.json"))
	validateJSONFiles(t, opportunitySchema, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))
}

func TestTemplateResearchTripleAgentArtifactsValidate(t *testing.T) {
	root := repoRoot(t)

	triggerSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "trigger_catalog.schema.json"))
	hypothesisSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "pattern_hypotheses.schema.json"))
	judgmentSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "judgment_registry.schema.json"))
	scoringSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "round_scoring.schema.json"))

	validateJSONDocumentFromFile(t, triggerSchema, filepath.Join(root, "knowledge", "template-research", "hypotheses", "trigger_catalog.json"))
	validateJSONDocumentFromFile(t, hypothesisSchema, filepath.Join(root, "knowledge", "template-research", "hypotheses", "pattern_hypotheses.json"))
	validateJSONDocumentFromFile(t, judgmentSchema, filepath.Join(root, "knowledge", "template-research", "judgments", "pattern_registry.json"))
	validateJSONDocumentFromFile(t, scoringSchema, filepath.Join(root, "knowledge", "template-research", "judgments", "round_scoring.json"))
	validateJSONLFilesWithMinRecords(t, scoringSchema, filepath.Join(root, "knowledge", "template-research", "judgments", "scoring-history", "*.jsonl"), 1, 1)
	assertTripleAgentArtifactCounts(t, root)
}

func TestTemplateResearchClassificationCoversRoadmapCandidates(t *testing.T) {
	root := repoRoot(t)
	classificationSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "classification.schema.json"))
	sweepSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "classification_sweep.schema.json"))

	classificationPath := filepath.Join(root, "knowledge", "template-research", "classification", "roadmap-pattern-groups.json")
	validateJSONDocumentFromFile(t, classificationSchema, classificationPath)
	validateJSONDocumentFromFile(t, classificationSchema, filepath.Join(root, "knowledge", "template-research", "classification", "fixture-priority-matrix.json"))
	validateJSONDocumentFromFile(t, classificationSchema, filepath.Join(root, "knowledge", "template-research", "classification", "capability-gap-map.json"))
	validateJSONDocumentFromFile(t, classificationSchema, filepath.Join(root, "knowledge", "template-research", "classification", "verifier-strategy-map.json"))
	validateJSONDocumentFromFile(t, sweepSchema, filepath.Join(root, "knowledge", "template-research", "classification", "classification-sweep.json"))

	roadmapCandidates := loadRoadmapCandidateIDs(t, filepath.Join(root, "knowledge", "template-research", "judgments", "pattern_registry.json"))
	classified := loadPrimaryClassifiedPatternIDs(t, classificationPath)
	if len(classified) != len(roadmapCandidates) {
		t.Fatalf("primary classified pattern count=%d want %d", len(classified), len(roadmapCandidates))
	}
	for candidate := range roadmapCandidates {
		if !classified[candidate] {
			t.Fatalf("roadmap candidate %q missing from primary classification", candidate)
		}
	}
	for patternID := range classified {
		if !roadmapCandidates[patternID] {
			t.Fatalf("classification includes non-roadmap primary pattern %q", patternID)
		}
	}
	assertClassificationSweepCoverage(t, filepath.Join(root, "knowledge", "template-research", "classification", "classification-sweep.json"))
}

func TestTemplateResearchMoleculesValidateAndReferenceRoadmapOrganisms(t *testing.T) {
	root := repoRoot(t)
	moleculeSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "molecule_catalog.schema.json"))
	moleculePath := filepath.Join(root, "knowledge", "template-research", "composition", "molecules.json")
	validateJSONDocumentFromFile(t, moleculeSchema, moleculePath)

	roadmapCandidates := loadRoadmapCandidateIDs(t, filepath.Join(root, "knowledge", "template-research", "judgments", "pattern_registry.json"))
	raw, err := os.ReadFile(moleculePath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", moleculePath, err)
	}
	var catalog struct {
		Molecules []struct {
			MoleculeID      string   `json:"molecule_id"`
			UsedByOrganisms []string `json:"used_by_organisms"`
			RequiredAtoms   []string `json:"required_atoms"`
			Verification    []string `json:"verification_methods"`
		} `json:"molecules"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", moleculePath, err)
	}
	if len(catalog.Molecules) < 16 {
		t.Fatalf("molecule count=%d want at least 16", len(catalog.Molecules))
	}
	coveredOrganisms := map[string]bool{}
	for _, molecule := range catalog.Molecules {
		if molecule.MoleculeID == "" || len(molecule.RequiredAtoms) == 0 || len(molecule.Verification) == 0 {
			t.Fatalf("incomplete molecule: %+v", molecule)
		}
		for _, organism := range molecule.UsedByOrganisms {
			if !roadmapCandidates[organism] {
				t.Fatalf("molecule %q references non-roadmap organism %q", molecule.MoleculeID, organism)
			}
			coveredOrganisms[organism] = true
		}
	}
	for organism := range roadmapCandidates {
		if !coveredOrganisms[organism] {
			t.Fatalf("roadmap organism %q is not covered by any molecule", organism)
		}
	}
}

func TestTemplateResearchKeepsUnsupportedCapabilitiesOutOfSupportedRegistry(t *testing.T) {
	root := repoRoot(t)
	capabilitySchemaPath := filepath.Join(root, "contracts", "capabilities", "capability.schema.json")
	raw, err := os.ReadFile(capabilitySchemaPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", capabilitySchemaPath, err)
	}

	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("Unmarshal capability schema: %v", err)
	}

	properties := document["properties"].(map[string]any)
	name := properties["name"].(map[string]any)
	enum := name["enum"].([]any)
	existing := map[string]bool{}
	for _, value := range enum {
		existing[value.(string)] = true
	}

	plannedNames := []string{
		"append_structured_rows",
		"extend_table_formulas",
		"copy_period_sheet",
		"normalize_headers",
		"create_pivot_summary",
		"add_data_validation",
		"protect_formula_cells",
		"reconcile_tables",
		"generate_printable_form",
		"roll_forward_period",
	}
	for _, planned := range plannedNames {
		if existing[planned] {
			t.Fatalf("planned capability %q must stay out of supported capability enum", planned)
		}
	}

	opportunityPaths, err := filepath.Glob(filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))
	if err != nil {
		t.Fatalf("Glob(opportunities): %v", err)
	}
	found := map[string]bool{}
	for _, path := range opportunityPaths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", path, err)
		}
		var record struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &record); err != nil {
			t.Fatalf("Unmarshal(%s): %v", path, err)
		}
		found[record.Name] = true
	}
	for _, planned := range plannedNames {
		if !found[planned] {
			t.Fatalf("planned capability %q missing opportunity record", planned)
		}
	}
}

func validateJSONL(t *testing.T, schema *jsonschema.Schema, path string) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open(%s): %v", path, err)
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		count++
		validateJSONDocument(t, schema, []byte(line), path)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Scan(%s): %v", path, err)
	}
	if count < 20 {
		t.Fatalf("%s record count=%d want at least 20 seed observations", path, count)
	}
}

func validateJSONLFiles(t *testing.T, schema *jsonschema.Schema, glob string, minFiles int) {
	t.Helper()
	validateJSONLFilesWithMinRecords(t, schema, glob, minFiles, 20)
}

func validateJSONLFilesWithMinRecords(t *testing.T, schema *jsonschema.Schema, glob string, minFiles, minRecords int) {
	t.Helper()
	paths, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("Glob(%s): %v", glob, err)
	}
	if len(paths) < minFiles {
		t.Fatalf("Glob(%s) found %d files, want at least %d", glob, len(paths), minFiles)
	}
	for _, path := range paths {
		pathMinRecords := minRecords
		if strings.Contains(filepath.Base(path), "round-009") {
			pathMinRecords = 15
		}
		validateJSONLWithMinRecords(t, schema, path, pathMinRecords)
	}
}

func validateJSONLWithMinRecords(t *testing.T, schema *jsonschema.Schema, path string, minRecords int) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open(%s): %v", path, err)
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		count++
		validateJSONDocument(t, schema, []byte(line), path)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Scan(%s): %v", path, err)
	}
	if count < minRecords {
		t.Fatalf("%s record count=%d want at least %d", path, count, minRecords)
	}
}

func validateJSONFiles(t *testing.T, schema *jsonschema.Schema, glob string) {
	t.Helper()
	paths, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("Glob(%s): %v", glob, err)
	}
	if len(paths) == 0 {
		t.Fatalf("Glob(%s) found no records", glob)
	}
	if strings.Contains(glob, string(filepath.Join("patterns", "*.json"))) && len(paths) < 8 {
		t.Fatalf("Glob(%s) found %d pattern records, want at least 8", glob, len(paths))
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", path, err)
		}
		validateJSONDocument(t, schema, raw, path)
	}
}

func validateJSONDocumentFromFile(t *testing.T, schema *jsonschema.Schema, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	validateJSONDocument(t, schema, raw, path)
}

func validateJSONDocument(t *testing.T, schema *jsonschema.Schema, raw []byte, label string) {
	t.Helper()
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("Unmarshal(%s): %v", label, err)
	}
	if err := schema.Validate(document); err != nil {
		t.Fatalf("Validate(%s): %v", label, err)
	}
}

func assertTripleAgentArtifactCounts(t *testing.T, root string) {
	t.Helper()
	triggerPath := filepath.Join(root, "knowledge", "template-research", "hypotheses", "trigger_catalog.json")
	triggerRaw, err := os.ReadFile(triggerPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", triggerPath, err)
	}
	var triggerCatalog struct {
		Groups map[string][]any `json:"groups"`
	}
	if err := json.Unmarshal(triggerRaw, &triggerCatalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", triggerPath, err)
	}
	triggerCount := 0
	for _, group := range triggerCatalog.Groups {
		triggerCount += len(group)
	}
	if triggerCount < 50 {
		t.Fatalf("trigger count=%d want at least 50", triggerCount)
	}

	judgmentPath := filepath.Join(root, "knowledge", "template-research", "judgments", "pattern_registry.json")
	judgmentRaw, err := os.ReadFile(judgmentPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", judgmentPath, err)
	}
	var registry struct {
		RoundID   string `json:"round_id"`
		Judgments []struct {
			Status string `json:"status"`
		} `json:"judgments"`
	}
	if err := json.Unmarshal(judgmentRaw, &registry); err != nil {
		t.Fatalf("Unmarshal(%s): %v", judgmentPath, err)
	}
	if registry.RoundID != "round-009" {
		t.Fatalf("judgment registry round_id=%q want round-009", registry.RoundID)
	}
	if len(registry.Judgments) < 20 {
		t.Fatalf("judgment count=%d want at least 20 after round-002", len(registry.Judgments))
	}
	statuses := map[string]int{}
	for _, judgment := range registry.Judgments {
		statuses[judgment.Status]++
	}
	for _, status := range []string{"recurring", "roadmap_candidate"} {
		if statuses[status] == 0 {
			t.Fatalf("judgment status %q missing from registry counts=%v", status, statuses)
		}
	}

	scoringPath := filepath.Join(root, "knowledge", "template-research", "judgments", "round_scoring.json")
	scoringRaw, err := os.ReadFile(scoringPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", scoringPath, err)
	}
	var scoring struct {
		RoundID  string `json:"round_id"`
		Findings []struct {
			Severity string `json:"severity"`
			Status   string `json:"status"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(scoringRaw, &scoring); err != nil {
		t.Fatalf("Unmarshal(%s): %v", scoringPath, err)
	}
	if scoring.RoundID != "round-009" {
		t.Fatalf("scoring round_id=%q want round-009", scoring.RoundID)
	}
	for _, finding := range scoring.Findings {
		switch finding.Severity {
		case "critical", "high":
			if finding.Status != "resolved" {
				t.Fatalf("%s finding has status=%q want resolved", finding.Severity, finding.Status)
			}
		case "medium":
			if finding.Status != "backlogged" && finding.Status != "resolved" {
				t.Fatalf("medium finding has status=%q want backlogged or resolved", finding.Status)
			}
		}
	}
}

func loadRoadmapCandidateIDs(t *testing.T, path string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var registry struct {
		Judgments []struct {
			PatternID string `json:"pattern_id"`
			Status    string `json:"status"`
		} `json:"judgments"`
	}
	if err := json.Unmarshal(raw, &registry); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	candidates := map[string]bool{}
	for _, judgment := range registry.Judgments {
		if judgment.Status == "roadmap_candidate" {
			candidates[judgment.PatternID] = true
		}
	}
	if len(candidates) == 0 {
		t.Fatalf("%s has no roadmap candidates", path)
	}
	return candidates
}

func loadPrimaryClassifiedPatternIDs(t *testing.T, path string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var classification struct {
		Groups []struct {
			Patterns []string `json:"patterns"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(raw, &classification); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	classified := map[string]bool{}
	for _, group := range classification.Groups {
		for _, patternID := range group.Patterns {
			if classified[patternID] {
				t.Fatalf("pattern %q appears in multiple primary classification groups", patternID)
			}
			classified[patternID] = true
		}
	}
	return classified
}

func assertClassificationSweepCoverage(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var sweep struct {
		Items []struct {
			PatternID        string   `json:"pattern_id"`
			CurrentGroup     string   `json:"current_group"`
			AlternativeGroup []string `json:"alternative_groups"`
			Decision         string   `json:"decision"`
			Rationale        string   `json:"rationale"`
		} `json:"items"`
	}
	if err := json.Unmarshal(raw, &sweep); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	if len(sweep.Items) < 5 {
		t.Fatalf("classification sweep item count=%d want at least 5", len(sweep.Items))
	}
	for _, item := range sweep.Items {
		if item.PatternID == "" || item.CurrentGroup == "" || len(item.AlternativeGroup) == 0 || item.Decision == "" || item.Rationale == "" {
			t.Fatalf("incomplete classification sweep item: %+v", item)
		}
	}
}

func compileSchema(t *testing.T, path string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(path)
	if err != nil {
		t.Fatalf("Compile(%s): %v", path, err)
	}
	return schema
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
