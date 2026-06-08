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

func TestTemplateResearchOrganismsValidateAsAdvisoryEcosystem(t *testing.T) {
	root := repoRoot(t)
	organismSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "organism_catalog.schema.json"))
	organismPath := filepath.Join(root, "knowledge", "template-research", "composition", "organisms.json")
	validateJSONDocumentFromFile(t, organismSchema, organismPath)

	roadmapCandidates := loadRoadmapCandidateIDs(t, filepath.Join(root, "knowledge", "template-research", "judgments", "pattern_registry.json"))
	molecules := loadTemplateResearchMolecules(t, filepath.Join(root, "knowledge", "template-research", "composition", "molecules.json"))
	supportedAtoms := loadSupportedCapabilityNames(t, filepath.Join(root, "contracts", "capabilities", "capability.schema.json"))
	opportunityAtoms := loadOpportunityCapabilityNames(t, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))

	raw, err := os.ReadFile(organismPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", organismPath, err)
	}
	var catalog struct {
		GeneratedFromRound string `json:"generated_from_round"`
		Authority          string `json:"authority"`
		Organisms          []struct {
			OrganismID       string   `json:"organism_id"`
			RequiredMolecule []string `json:"required_molecule_ids"`
			SupportingAtoms  []string `json:"supporting_atom_ids"`
			PlannedAtoms     []string `json:"planned_or_unsupported_atom_ids"`
			VerifierFocus    []string `json:"verifier_focus"`
			AdvisoryStatus   string   `json:"advisory_status"`
		} `json:"organisms"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", organismPath, err)
	}
	if catalog.GeneratedFromRound != "round-009" {
		t.Fatalf("organism catalog generated_from_round=%q want round-009", catalog.GeneratedFromRound)
	}
	if catalog.Authority != "advisory_only" {
		t.Fatalf("organism catalog authority=%q want advisory_only", catalog.Authority)
	}

	seenOrganisms := map[string]bool{}
	for _, organism := range catalog.Organisms {
		if organism.AdvisoryStatus != "advisory_only" {
			t.Fatalf("organism %q advisory_status=%q want advisory_only", organism.OrganismID, organism.AdvisoryStatus)
		}
		if len(organism.VerifierFocus) == 0 {
			t.Fatalf("organism %q has no verifier focus", organism.OrganismID)
		}
		if seenOrganisms[organism.OrganismID] {
			t.Fatalf("organism %q appears more than once", organism.OrganismID)
		}
		seenOrganisms[organism.OrganismID] = true
		if !roadmapCandidates[organism.OrganismID] {
			t.Fatalf("organism catalog includes non-roadmap candidate %q", organism.OrganismID)
		}
		for _, moleculeID := range organism.RequiredMolecule {
			verificationMethods, ok := molecules[moleculeID]
			if !ok {
				t.Fatalf("organism %q references unknown molecule %q", organism.OrganismID, moleculeID)
			}
			if len(verificationMethods) == 0 {
				t.Fatalf("organism %q references molecule %q without verification methods", organism.OrganismID, moleculeID)
			}
		}
		for _, atom := range organism.SupportingAtoms {
			if !supportedAtoms[atom] && !opportunityAtoms[atom] {
				t.Fatalf("organism %q references unknown supporting atom %q", organism.OrganismID, atom)
			}
		}
		for _, atom := range organism.PlannedAtoms {
			if supportedAtoms[atom] {
				t.Fatalf("organism %q planned atom %q is already supported", organism.OrganismID, atom)
			}
			if !opportunityAtoms[atom] {
				t.Fatalf("organism %q planned atom %q missing opportunity record", organism.OrganismID, atom)
			}
		}
	}
	if len(seenOrganisms) != len(roadmapCandidates) {
		t.Fatalf("organism count=%d want roadmap candidate count=%d", len(seenOrganisms), len(roadmapCandidates))
	}
	for candidate := range roadmapCandidates {
		if !seenOrganisms[candidate] {
			t.Fatalf("roadmap candidate %q missing from organism catalog", candidate)
		}
	}
}

func TestTemplateResearchExecutableOrganismCoverageLadder(t *testing.T) {
	root := repoRoot(t)
	coverageSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "executable_organism_coverage.schema.json"))
	coveragePath := filepath.Join(root, "knowledge", "template-research", "composition", "executable-organism-coverage.json")
	validateJSONDocumentFromFile(t, coverageSchema, coveragePath)

	roadmapCandidates := loadRoadmapCandidateIDs(t, filepath.Join(root, "knowledge", "template-research", "judgments", "pattern_registry.json"))
	organismPriorities := loadTemplateResearchOrganismPriorities(t, filepath.Join(root, "knowledge", "template-research", "composition", "organisms.json"))
	supportedAtoms := loadSupportedCapabilityNames(t, filepath.Join(root, "contracts", "capabilities", "capability.schema.json"))
	opportunityAtoms := loadOpportunityCapabilityNames(t, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))

	raw, err := os.ReadFile(coveragePath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", coveragePath, err)
	}
	var catalog struct {
		Authority string `json:"authority"`
		Organisms []struct {
			OrganismID         string   `json:"organism_id"`
			FixturePriority    string   `json:"fixture_priority"`
			CoverageTier       string   `json:"coverage_tier"`
			ExecutableSequence []string `json:"executable_sequence_atom_ids"`
			PlannedBlockers    []string `json:"planned_blocker_atom_ids"`
			PreviewEvidence    []string `json:"preview_evidence"`
			CoverageClaim      string   `json:"coverage_claim"`
			RemainingGaps      []string `json:"remaining_gaps"`
			NextPromotionGate  string   `json:"next_promotion_gate"`
		} `json:"organisms"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", coveragePath, err)
	}
	if catalog.Authority != "advisory_coverage_only" {
		t.Fatalf("coverage authority=%q want advisory_coverage_only", catalog.Authority)
	}
	seen := map[string]bool{}
	previewFixtureCount := 0
	verifiedClassCount := 0
	plannedBlockerCount := 0
	for _, organism := range catalog.Organisms {
		if !roadmapCandidates[organism.OrganismID] {
			t.Fatalf("coverage includes non-roadmap organism %q", organism.OrganismID)
		}
		if organismPriorities[organism.OrganismID] != organism.FixturePriority {
			t.Fatalf("organism %q fixture_priority=%q want %q", organism.OrganismID, organism.FixturePriority, organismPriorities[organism.OrganismID])
		}
		if seen[organism.OrganismID] {
			t.Fatalf("coverage includes duplicate organism %q", organism.OrganismID)
		}
		seen[organism.OrganismID] = true
		if organism.CoverageClaim == "" || organism.NextPromotionGate == "" || len(organism.RemainingGaps) == 0 {
			t.Fatalf("coverage organism %q has weak claim/gap/gate fields: %+v", organism.OrganismID, organism)
		}
		for _, atom := range organism.ExecutableSequence {
			if !supportedAtoms[atom] {
				t.Fatalf("organism %q executable sequence references unsupported atom %q", organism.OrganismID, atom)
			}
		}
		for _, atom := range organism.PlannedBlockers {
			plannedBlockerCount++
			if supportedAtoms[atom] {
				t.Fatalf("organism %q planned blocker %q is already supported", organism.OrganismID, atom)
			}
			if !opportunityAtoms[atom] {
				t.Fatalf("organism %q planned blocker %q missing opportunity record", organism.OrganismID, atom)
			}
		}
		switch organism.CoverageTier {
		case "preview_fixture", "organism_verified_class":
			previewFixtureCount++
			if organism.CoverageTier == "organism_verified_class" {
				verifiedClassCount++
			}
			if len(organism.PreviewEvidence) == 0 {
				t.Fatalf("%s organism %q has no preview evidence", organism.CoverageTier, organism.OrganismID)
			}
			if len(organism.ExecutableSequence) == 0 {
				t.Fatalf("%s organism %q has no executable sequence", organism.CoverageTier, organism.OrganismID)
			}
		case "supported_atom_plan":
			if len(organism.ExecutableSequence) == 0 {
				t.Fatalf("supported atom plan organism %q has no executable sequence", organism.OrganismID)
			}
		case "blocked_by_planned_atom":
			if len(organism.PlannedBlockers) == 0 {
				t.Fatalf("blocked organism %q has no planned blockers", organism.OrganismID)
			}
		case "advisory_only_gap":
			if len(organism.ExecutableSequence) != 0 {
				t.Fatalf("advisory-only gap organism %q should not claim executable sequence", organism.OrganismID)
			}
		default:
			t.Fatalf("organism %q has unknown coverage tier %q", organism.OrganismID, organism.CoverageTier)
		}
	}
	if len(seen) != len(roadmapCandidates) {
		t.Fatalf("coverage organism count=%d want roadmap candidate count=%d", len(seen), len(roadmapCandidates))
	}
	if previewFixtureCount != len(roadmapCandidates) {
		t.Fatalf("preview-backed organism count=%d want roadmap candidate count=%d", previewFixtureCount, len(roadmapCandidates))
	}
	if verifiedClassCount == 0 {
		t.Fatalf("verified class organism count=0 want at least one depth contract")
	}
	if plannedBlockerCount != 0 {
		t.Fatalf("planned blocker count=%d want 0 after deterministic preview coverage", plannedBlockerCount)
	}
	for candidate := range roadmapCandidates {
		if !seen[candidate] {
			t.Fatalf("roadmap candidate %q missing from executable coverage ladder", candidate)
		}
	}
}

func TestTemplateResearchVerifiedOrganismClassesValidateDepthContracts(t *testing.T) {
	root := repoRoot(t)
	classSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "verified_organism_classes.schema.json"))
	classPath := filepath.Join(root, "knowledge", "template-research", "composition", "verified-organism-classes.json")
	validateJSONDocumentFromFile(t, classSchema, classPath)

	coverageClasses := loadTemplateResearchCoverageTiers(t, filepath.Join(root, "knowledge", "template-research", "composition", "executable-organism-coverage.json"))
	supportedAtoms := loadSupportedCapabilityNames(t, filepath.Join(root, "contracts", "capabilities", "capability.schema.json"))
	opportunityAtoms := loadOpportunityCapabilityNames(t, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))

	raw, err := os.ReadFile(classPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", classPath, err)
	}
	var catalog struct {
		Authority string `json:"authority"`
		Classes   []struct {
			OrganismID         string   `json:"organism_id"`
			DepthFocus         string   `json:"depth_focus"`
			RuntimeAtomIDs     []string `json:"runtime_atom_ids"`
			OpportunityAtomIDs []string `json:"opportunity_atom_ids"`
			PreviewEvidence    []string `json:"preview_evidence"`
			AcceptanceCriteria []string `json:"acceptance_criteria"`
			ExplicitNonClaims  []string `json:"explicit_non_claims"`
			PromotionGate      string   `json:"promotion_gate"`
		} `json:"classes"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", classPath, err)
	}
	if catalog.Authority != "advisory_verified_class" {
		t.Fatalf("verified class authority=%q want advisory_verified_class", catalog.Authority)
	}
	wantOrganisms := map[string]bool{
		"student_gradebook":          true,
		"training_completion_matrix": true,
		"service_ticket_queue":       true,
		"sales_pipeline_tracker":     true,
		"maintenance_issue_log":      true,
		"compliance_action_register": true,
		"safety_compliance_register": true,
		"loan_repayment_calculator":  true,
	}
	seen := map[string]bool{}
	for _, class := range catalog.Classes {
		if !wantOrganisms[class.OrganismID] {
			t.Fatalf("unexpected verified organism class %q", class.OrganismID)
		}
		if seen[class.OrganismID] {
			t.Fatalf("duplicate verified organism class %q", class.OrganismID)
		}
		seen[class.OrganismID] = true
		if coverageClasses[class.OrganismID] != "organism_verified_class" {
			t.Fatalf("coverage tier for %q=%q want organism_verified_class", class.OrganismID, coverageClasses[class.OrganismID])
		}
		if class.DepthFocus == "" || len(class.PreviewEvidence) == 0 || len(class.AcceptanceCriteria) < 3 || len(class.ExplicitNonClaims) == 0 || class.PromotionGate == "" {
			t.Fatalf("verified organism class %q has weak depth contract: %+v", class.OrganismID, class)
		}
		for _, atom := range class.RuntimeAtomIDs {
			if !supportedAtoms[atom] {
				t.Fatalf("verified organism class %q references unsupported runtime atom %q", class.OrganismID, atom)
			}
		}
		for _, atom := range class.OpportunityAtomIDs {
			if supportedAtoms[atom] {
				t.Fatalf("verified organism class %q opportunity atom %q is already supported", class.OrganismID, atom)
			}
			if !opportunityAtoms[atom] {
				t.Fatalf("verified organism class %q opportunity atom %q missing opportunity record", class.OrganismID, atom)
			}
		}
	}
	if len(seen) != len(wantOrganisms) {
		t.Fatalf("verified organism class count=%d want %d", len(seen), len(wantOrganisms))
	}
	for organism := range wantOrganisms {
		if !seen[organism] {
			t.Fatalf("verified organism class missing %q", organism)
		}
	}
}

func TestTemplateResearchRuntimeProductizationArtifactsValidate(t *testing.T) {
	root := repoRoot(t)
	productizationDir := filepath.Join(root, "knowledge", "template-research", "runtime")

	verifierSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "organism_verifier_specs.schema.json"))
	plannerSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "operation_planner.schema.json"))
	harnessSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "template_class_harness.schema.json"))
	promotionSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "advanced_atom_promotion.schema.json"))

	verifierPath := filepath.Join(productizationDir, "organism-verifier-specs.json")
	plannerPath := filepath.Join(productizationDir, "operation-planner.json")
	harnessPath := filepath.Join(productizationDir, "template-class-harness.json")
	promotionPath := filepath.Join(productizationDir, "advanced-atom-promotion.json")

	validateJSONDocumentFromFile(t, verifierSchema, verifierPath)
	validateJSONDocumentFromFile(t, plannerSchema, plannerPath)
	validateJSONDocumentFromFile(t, harnessSchema, harnessPath)
	validateJSONDocumentFromFile(t, promotionSchema, promotionPath)

	supportedAtoms := loadSupportedCapabilityNames(t, filepath.Join(root, "contracts", "capabilities", "capability.schema.json"))
	opportunityAtoms := loadOpportunityCapabilityNames(t, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))
	coverageTiers := loadTemplateResearchCoverageTiers(t, filepath.Join(root, "knowledge", "template-research", "composition", "executable-organism-coverage.json"))

	verifierSpecs := loadRuntimeProductizationVerifierSpecs(t, verifierPath)
	for _, organism := range []string{
		"invoice_line_item_billing",
		"monthly_budget_control",
		"inventory_movement_log",
		"student_gradebook",
		"loan_repayment_calculator",
	} {
		spec, ok := verifierSpecs[organism]
		if !ok {
			t.Fatalf("organism verifier spec missing %q", organism)
		}
		if len(spec.AcceptanceChecks) < 3 || len(spec.AtomVerifierDependencies) == 0 || len(spec.NonClaims) == 0 {
			t.Fatalf("organism verifier spec %q is too weak: %+v", organism, spec)
		}
		for _, atom := range spec.AtomVerifierDependencies {
			if !supportedAtoms[atom] {
				t.Fatalf("organism verifier spec %q references unsupported atom %q", organism, atom)
			}
		}
	}

	plans := loadRuntimeProductizationPlannerPlans(t, plannerPath)
	if len(plans) < 5 {
		t.Fatalf("planner plan count=%d want at least 5", len(plans))
	}
	for organism, plan := range plans {
		if coverageTiers[organism] == "" {
			t.Fatalf("planner references organism %q missing from coverage ladder", organism)
		}
		if len(plan.OperationSequence) == 0 || len(plan.RequiredVerifierSpecs) == 0 || plan.FallbackPolicy == "" {
			t.Fatalf("planner plan %q is incomplete: %+v", organism, plan)
		}
		for _, atom := range plan.OperationSequence {
			if !supportedAtoms[atom] {
				t.Fatalf("planner plan %q references unsupported atom %q", organism, atom)
			}
		}
	}

	harnessScenarios := loadRuntimeProductizationHarnessScenarios(t, harnessPath)
	if len(harnessScenarios) < 5 {
		t.Fatalf("template class harness scenario count=%d want at least 5", len(harnessScenarios))
	}
	for _, scenario := range harnessScenarios {
		if plans[scenario.OrganismID].OrganismID == "" {
			t.Fatalf("harness scenario %q references organism %q without planner plan", scenario.ScenarioID, scenario.OrganismID)
		}
		if len(scenario.Flow) != 5 {
			t.Fatalf("harness scenario %q flow length=%d want classify/plan/execute/verify/report", scenario.ScenarioID, len(scenario.Flow))
		}
		if len(scenario.SuccessEvidence) == 0 || len(scenario.NonClaims) == 0 {
			t.Fatalf("harness scenario %q has weak evidence/non-claims: %+v", scenario.ScenarioID, scenario)
		}
	}

	promotions := loadRuntimeProductizationPromotionDecisions(t, promotionPath)
	for _, atom := range []string{"create_pivot_summary", "matrix_growth", "timeline_grid_projection", "calculation_schedule_verifier", "printable_render_qa"} {
		decision, ok := promotions[atom]
		if !ok {
			t.Fatalf("advanced atom promotion decision missing %q", atom)
		}
		if decision.Status == "promote_now" {
			t.Fatalf("advanced atom %q must not promote without runtime verifier evidence", atom)
		}
		if supportedAtoms[atom] {
			t.Fatalf("advanced atom %q unexpectedly present in supported capability enum", atom)
		}
		if atom == "create_pivot_summary" && !opportunityAtoms[atom] {
			t.Fatalf("advanced atom %q should remain tracked as an opportunity", atom)
		}
		if len(decision.RequiredEvidence) < 2 || decision.Reason == "" {
			t.Fatalf("advanced atom promotion decision %q is weak: %+v", atom, decision)
		}
	}
}

func TestTemplateResearchDraftPlannerCoverageStaysExplicit(t *testing.T) {
	root := repoRoot(t)
	coverageSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "draft_planner_coverage.schema.json"))
	coveragePath := filepath.Join(root, "knowledge", "template-research", "composition", "draft-planner-coverage.json")
	validateJSONDocumentFromFile(t, coverageSchema, coveragePath)

	roadmapCandidates := loadRoadmapCandidateIDs(t, filepath.Join(root, "knowledge", "template-research", "judgments", "pattern_registry.json"))
	raw, err := os.ReadFile(coveragePath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", coveragePath, err)
	}
	var catalog struct {
		Authority          string `json:"authority"`
		GeneratedFromRound string `json:"generated_from_round"`
		Coverage           []struct {
			OrganismID        string   `json:"organism_id"`
			DraftStatus       string   `json:"draft_status"`
			RuntimeEntrypoint string   `json:"runtime_entrypoint"`
			EvidenceTests     []string `json:"evidence_tests"`
			CurrentLimit      string   `json:"current_limit"`
			NextGate          string   `json:"next_gate"`
		} `json:"coverage"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", coveragePath, err)
	}
	if catalog.Authority != "runtime_draft_planner_coverage" {
		t.Fatalf("draft planner coverage authority=%q want runtime_draft_planner_coverage", catalog.Authority)
	}
	if catalog.GeneratedFromRound != "round-009" {
		t.Fatalf("draft planner generated_from_round=%q want round-009", catalog.GeneratedFromRound)
	}

	wantRuntimeDraft := map[string]bool{
		"invoice_line_item_billing": true,
		"monthly_budget_control":    true,
		"cash_flow_monitor":         true,
		"timesheet_hours_log":       true,
		"warehouse_reorder_tracker": true,
		"inventory_movement_log":    true,
		"student_gradebook":         true,
		"loan_repayment_calculator": true,
	}
	seen := map[string]bool{}
	runtimeDraftCount := 0
	explicitOnlyCount := 0
	for _, record := range catalog.Coverage {
		if !roadmapCandidates[record.OrganismID] {
			t.Fatalf("draft planner coverage includes non-roadmap organism %q", record.OrganismID)
		}
		if seen[record.OrganismID] {
			t.Fatalf("draft planner coverage includes duplicate organism %q", record.OrganismID)
		}
		seen[record.OrganismID] = true
		if record.CurrentLimit == "" || record.NextGate == "" {
			t.Fatalf("draft planner coverage %q has weak limit/gate fields: %+v", record.OrganismID, record)
		}
		switch record.DraftStatus {
		case "runtime_draft_planner":
			runtimeDraftCount++
			if !wantRuntimeDraft[record.OrganismID] {
				t.Fatalf("organism %q unexpectedly claims runtime draft planner", record.OrganismID)
			}
			if record.RuntimeEntrypoint != "requestcompiler.DraftOrganismExecutionRequest" {
				t.Fatalf("runtime draft %q entrypoint=%q", record.OrganismID, record.RuntimeEntrypoint)
			}
			if len(record.EvidenceTests) < 2 {
				t.Fatalf("runtime draft %q has weak evidence tests: %+v", record.OrganismID, record.EvidenceTests)
			}
		case "explicit_request_only":
			explicitOnlyCount++
			if wantRuntimeDraft[record.OrganismID] {
				t.Fatalf("organism %q should be runtime_draft_planner", record.OrganismID)
			}
			if record.RuntimeEntrypoint != "" || len(record.EvidenceTests) != 0 {
				t.Fatalf("explicit-only organism %q must not claim runtime entrypoint/tests: %+v", record.OrganismID, record)
			}
		default:
			t.Fatalf("organism %q has unknown draft status %q", record.OrganismID, record.DraftStatus)
		}
	}
	if len(seen) != len(roadmapCandidates) {
		t.Fatalf("draft planner coverage organism count=%d want roadmap candidate count=%d", len(seen), len(roadmapCandidates))
	}
	if runtimeDraftCount != len(wantRuntimeDraft) {
		t.Fatalf("runtime draft count=%d want %d", runtimeDraftCount, len(wantRuntimeDraft))
	}
	if explicitOnlyCount != len(roadmapCandidates)-len(wantRuntimeDraft) {
		t.Fatalf("explicit-only count=%d want %d", explicitOnlyCount, len(roadmapCandidates)-len(wantRuntimeDraft))
	}
	for candidate := range roadmapCandidates {
		if !seen[candidate] {
			t.Fatalf("roadmap candidate %q missing from draft planner coverage", candidate)
		}
	}
}

func TestAtomBuilderLayerValidatesRuntimeMirrorAndPlanningContracts(t *testing.T) {
	root := repoRoot(t)
	domainSchema := compileSchema(t, filepath.Join(root, "contracts", "workbook_app", "domain_schema.schema.json"))
	builderSchema := compileSchema(t, filepath.Join(root, "contracts", "atom_builders", "atom_builder.schema.json"))
	planSchema := compileSchema(t, filepath.Join(root, "contracts", "atom_builders", "builder_plan.schema.json"))

	validateJSONDocumentFromFile(t, domainSchema, filepath.Join(root, "knowledge", "atom-builders", "workbook-app-primitives.json"))
	builderPath := filepath.Join(root, "knowledge", "atom-builders", "builders.json")
	validateJSONDocumentFromFile(t, builderSchema, builderPath)
	validateJSONDocumentFromFile(t, planSchema, filepath.Join(root, "knowledge", "atom-builders", "builder-plan-shapes.json"))

	supportedAtoms := loadSupportedCapabilityNames(t, filepath.Join(root, "contracts", "capabilities", "capability.schema.json"))
	opportunityAtoms := loadOpportunityCapabilityNames(t, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))
	raw, err := os.ReadFile(builderPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", builderPath, err)
	}
	var catalog struct {
		Authority string `json:"authority"`
		Builders  []struct {
			AtomID            string   `json:"atom_id"`
			Status            string   `json:"status"`
			RuntimePaths      []string `json:"runtime_paths"`
			CapabilityRecord  string   `json:"capability_record,omitempty"`
			OpportunityRecord string   `json:"opportunity_record,omitempty"`
		} `json:"builders"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", builderPath, err)
	}
	if catalog.Authority != "advisory_mirror" {
		t.Fatalf("atom builder catalog authority=%q want advisory_mirror", catalog.Authority)
	}
	seen := map[string]bool{}
	for _, builder := range catalog.Builders {
		if seen[builder.AtomID] {
			t.Fatalf("atom builder %q appears more than once", builder.AtomID)
		}
		seen[builder.AtomID] = true
		switch builder.Status {
		case "supported", "runtime_primitive":
			if !supportedAtoms[builder.AtomID] {
				t.Fatalf("%s builder %q missing from supported capability enum", builder.Status, builder.AtomID)
			}
			if builder.CapabilityRecord == "" {
				t.Fatalf("%s builder %q missing capability record", builder.Status, builder.AtomID)
			}
			requireRepoRelativePathsExist(t, root, append([]string{builder.CapabilityRecord}, builder.RuntimePaths...))
		case "planned":
			if supportedAtoms[builder.AtomID] {
				t.Fatalf("planned builder %q must not be in supported capability enum", builder.AtomID)
			}
			if !opportunityAtoms[builder.AtomID] {
				t.Fatalf("planned builder %q missing opportunity record", builder.AtomID)
			}
			if builder.OpportunityRecord == "" {
				t.Fatalf("planned builder %q missing opportunity record path", builder.AtomID)
			}
			requireRepoRelativePathsExist(t, root, []string{builder.OpportunityRecord})
			if len(builder.RuntimePaths) != 0 {
				t.Fatalf("planned builder %q should not claim runtime paths", builder.AtomID)
			}
		default:
			t.Fatalf("builder %q has unknown status %q", builder.AtomID, builder.Status)
		}
	}
	for atom := range supportedAtoms {
		if !seen[atom] {
			t.Fatalf("supported capability %q missing atom builder mirror", atom)
		}
	}
}

func TestTemplateResearchKeepsUnsupportedCapabilitiesOutOfSupportedRegistry(t *testing.T) {
	root := repoRoot(t)
	capabilitySchemaPath := filepath.Join(root, "contracts", "capabilities", "capability.schema.json")
	existing := loadSupportedCapabilityNames(t, capabilitySchemaPath)

	plannedNames := []string{
		"create_pivot_summary",
	}
	for _, planned := range plannedNames {
		if existing[planned] {
			t.Fatalf("planned capability %q must stay out of supported capability enum", planned)
		}
	}

	found := loadOpportunityCapabilityNames(t, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))
	for _, planned := range plannedNames {
		if !found[planned] {
			t.Fatalf("planned capability %q missing opportunity record", planned)
		}
	}
}

func loadTemplateResearchMolecules(t *testing.T, path string) map[string][]string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Molecules []struct {
			MoleculeID          string   `json:"molecule_id"`
			VerificationMethods []string `json:"verification_methods"`
		} `json:"molecules"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	molecules := map[string][]string{}
	for _, molecule := range catalog.Molecules {
		molecules[molecule.MoleculeID] = append([]string(nil), molecule.VerificationMethods...)
	}
	return molecules
}

func loadTemplateResearchOrganismPriorities(t *testing.T, path string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Organisms []struct {
			OrganismID      string `json:"organism_id"`
			FixturePriority string `json:"fixture_priority"`
		} `json:"organisms"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	priorities := map[string]string{}
	for _, organism := range catalog.Organisms {
		priorities[organism.OrganismID] = organism.FixturePriority
	}
	return priorities
}

func loadTemplateResearchCoverageTiers(t *testing.T, path string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Organisms []struct {
			OrganismID   string `json:"organism_id"`
			CoverageTier string `json:"coverage_tier"`
		} `json:"organisms"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	tiers := map[string]string{}
	for _, organism := range catalog.Organisms {
		tiers[organism.OrganismID] = organism.CoverageTier
	}
	return tiers
}

type runtimeProductizationVerifierSpec struct {
	AcceptanceChecks         []string `json:"acceptance_checks"`
	AtomVerifierDependencies []string `json:"atom_verifier_dependencies"`
	NonClaims                []string `json:"non_claims"`
}

func loadRuntimeProductizationVerifierSpecs(t *testing.T, path string) map[string]runtimeProductizationVerifierSpec {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Specs []struct {
			OrganismID               string   `json:"organism_id"`
			AcceptanceChecks         []string `json:"acceptance_checks"`
			AtomVerifierDependencies []string `json:"atom_verifier_dependencies"`
			NonClaims                []string `json:"non_claims"`
		} `json:"specs"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	specs := map[string]runtimeProductizationVerifierSpec{}
	for _, spec := range catalog.Specs {
		specs[spec.OrganismID] = runtimeProductizationVerifierSpec{
			AcceptanceChecks:         spec.AcceptanceChecks,
			AtomVerifierDependencies: spec.AtomVerifierDependencies,
			NonClaims:                spec.NonClaims,
		}
	}
	return specs
}

type runtimeProductizationPlannerPlan struct {
	OrganismID            string   `json:"organism_id"`
	OperationSequence     []string `json:"operation_sequence_atom_ids"`
	RequiredVerifierSpecs []string `json:"required_verifier_spec_ids"`
	FallbackPolicy        string   `json:"fallback_policy"`
}

func loadRuntimeProductizationPlannerPlans(t *testing.T, path string) map[string]runtimeProductizationPlannerPlan {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Plans []runtimeProductizationPlannerPlan `json:"plans"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	plans := map[string]runtimeProductizationPlannerPlan{}
	for _, plan := range catalog.Plans {
		plans[plan.OrganismID] = plan
	}
	return plans
}

type runtimeProductizationHarnessScenario struct {
	ScenarioID      string   `json:"scenario_id"`
	OrganismID      string   `json:"organism_id"`
	Flow            []string `json:"flow"`
	SuccessEvidence []string `json:"success_evidence"`
	NonClaims       []string `json:"non_claims"`
}

func loadRuntimeProductizationHarnessScenarios(t *testing.T, path string) []runtimeProductizationHarnessScenario {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Scenarios []runtimeProductizationHarnessScenario `json:"scenarios"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	return catalog.Scenarios
}

type runtimeProductizationPromotionDecision struct {
	Status           string   `json:"status"`
	Reason           string   `json:"reason"`
	RequiredEvidence []string `json:"required_evidence"`
}

func loadRuntimeProductizationPromotionDecisions(t *testing.T, path string) map[string]runtimeProductizationPromotionDecision {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Decisions []struct {
			AtomID           string   `json:"atom_id"`
			Status           string   `json:"status"`
			Reason           string   `json:"reason"`
			RequiredEvidence []string `json:"required_evidence"`
		} `json:"decisions"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	decisions := map[string]runtimeProductizationPromotionDecision{}
	for _, decision := range catalog.Decisions {
		decisions[decision.AtomID] = runtimeProductizationPromotionDecision{
			Status:           decision.Status,
			Reason:           decision.Reason,
			RequiredEvidence: decision.RequiredEvidence,
		}
	}
	return decisions
}

func loadSupportedCapabilityNames(t *testing.T, schemaPath string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", schemaPath, err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("Unmarshal(%s): %v", schemaPath, err)
	}
	properties := document["properties"].(map[string]any)
	name := properties["name"].(map[string]any)
	enum := name["enum"].([]any)
	names := map[string]bool{}
	for _, value := range enum {
		names[value.(string)] = true
	}
	return names
}

func loadOpportunityCapabilityNames(t *testing.T, glob string) map[string]bool {
	t.Helper()
	opportunityPaths, err := filepath.Glob(glob)
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
	return found
}

func requireRepoRelativePathsExist(t *testing.T, root string, paths []string) {
	t.Helper()
	for _, rel := range paths {
		if filepath.IsAbs(rel) {
			t.Fatalf("path %q must be repository-relative", rel)
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("required path %q: %v", rel, err)
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
