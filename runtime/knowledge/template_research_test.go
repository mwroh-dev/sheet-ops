package knowledge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func TestTemplateResearchCorpusRecordsValidate(t *testing.T) {
	root := repoRoot(t)

	patternSchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "template_pattern.schema.json"))
	opportunitySchema := compileSchema(t, filepath.Join(root, "contracts", "template_research", "capability_opportunity.schema.json"))

	validateJSONFiles(t, patternSchema, filepath.Join(root, "knowledge", "template-research", "patterns", "*.json"))
	validateJSONFiles(t, opportunitySchema, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))
	assertTemplatePatternsReferenceKnownVocabulary(t, root)
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
	assertHypothesesReferenceTriggerCatalogAndJudgments(t, root)
	assertJudgmentsReferenceKnownCapabilities(t, root)
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
	for _, path := range []string{
		classificationPath,
		filepath.Join(root, "knowledge", "template-research", "classification", "fixture-priority-matrix.json"),
		filepath.Join(root, "knowledge", "template-research", "classification", "capability-gap-map.json"),
		filepath.Join(root, "knowledge", "template-research", "classification", "verifier-strategy-map.json"),
	} {
		classified := loadClassifiedPatternIDs(t, path)
		assertExactStringSet(t, classified, roadmapCandidates, path)
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
	moleculePath := filepath.Join(root, "knowledge", "template-research", "composition", "molecules.json")
	molecules := loadTemplateResearchMolecules(t, moleculePath)
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
	assertMoleculeOrganismLinksAreBidirectional(t, moleculePath, organismPath)
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
	runtimeDraftOrganisms := loadTemplateResearchRuntimeDraftOrganisms(t, filepath.Join(root, "knowledge", "template-research", "composition", "draft-planner-coverage.json"))
	atomBuilders := loadAtomBuilderIDs(t, filepath.Join(root, "knowledge", "atom-builders", "builders.json"))

	verifierSpecs := loadRuntimeProductizationVerifierSpecs(t, verifierPath)
	for _, organism := range runtimeDraftOrganisms {
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
	if len(plans) != len(runtimeDraftOrganisms) {
		t.Fatalf("planner plan count=%d want runtime draft organism count=%d", len(plans), len(runtimeDraftOrganisms))
	}
	for _, organism := range runtimeDraftOrganisms {
		if plans[organism].OrganismID == "" {
			t.Fatalf("planner plan missing runtime draft organism %q", organism)
		}
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
			if !atomBuilders[atom] {
				t.Fatalf("planner plan %q references atom %q without atom-builder mirror", organism, atom)
			}
			if !containsString(verifierSpecs[organism].AtomVerifierDependencies, atom) {
				t.Fatalf("planner plan %q operation atom %q missing from verifier dependencies", organism, atom)
			}
		}
		for _, verifierSpecID := range plan.RequiredVerifierSpecs {
			if !runtimeProductizationVerifierSpecIDExists(verifierSpecs, verifierSpecID) {
				t.Fatalf("planner plan %q references unknown verifier spec %q", organism, verifierSpecID)
			}
		}
	}

	harnessScenarios := loadRuntimeProductizationHarnessScenarios(t, harnessPath)
	if len(harnessScenarios) != len(runtimeDraftOrganisms) {
		t.Fatalf("template class harness scenario count=%d want runtime draft organism count=%d", len(harnessScenarios), len(runtimeDraftOrganisms))
	}
	harnessOrganisms := map[string]bool{}
	for _, scenario := range harnessScenarios {
		if plans[scenario.OrganismID].OrganismID == "" {
			t.Fatalf("harness scenario %q references organism %q without planner plan", scenario.ScenarioID, scenario.OrganismID)
		}
		if harnessOrganisms[scenario.OrganismID] {
			t.Fatalf("harness includes duplicate organism %q", scenario.OrganismID)
		}
		harnessOrganisms[scenario.OrganismID] = true
		if len(scenario.Flow) != 5 {
			t.Fatalf("harness scenario %q flow length=%d want classify/plan/execute/verify/report", scenario.ScenarioID, len(scenario.Flow))
		}
		if len(scenario.SuccessEvidence) == 0 || len(scenario.NonClaims) == 0 {
			t.Fatalf("harness scenario %q has weak evidence/non-claims: %+v", scenario.ScenarioID, scenario)
		}
	}
	for _, organism := range runtimeDraftOrganisms {
		if !harnessOrganisms[organism] {
			t.Fatalf("template class harness missing runtime draft organism %q", organism)
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
		"invoice_line_item_billing":  true,
		"expense_reimbursement":      true,
		"purchase_order_control":     true,
		"procurement_reconciliation": true,
		"monthly_budget_control":     true,
		"cash_flow_monitor":          true,
		"attendance_register":        true,
		"project_timeline_tracker":   true,
		"shift_roster_planner":       true,
		"construction_cost_tracker":  true,
		"timesheet_hours_log":        true,
		"warehouse_reorder_tracker":  true,
		"inventory_movement_log":     true,
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
	planShapePath := filepath.Join(root, "knowledge", "atom-builders", "builder-plan-shapes.json")
	validateJSONDocumentFromFile(t, planSchema, planShapePath)

	supportedAtoms := loadSupportedCapabilityNames(t, filepath.Join(root, "contracts", "capabilities", "capability.schema.json"))
	opportunityAtoms := loadOpportunityCapabilityNames(t, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))
	knownAtoms := unionStringSets(supportedAtoms, opportunityAtoms)
	planShapeAtoms := loadAtomBuilderPlanShapeAtoms(t, planShapePath)
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
	for atom := range seen {
		if !planShapeAtoms[atom] {
			t.Fatalf("atom builder %q missing from builder plan shapes", atom)
		}
	}
	for atom := range planShapeAtoms {
		if !seen[atom] {
			t.Fatalf("builder plan shape references atom %q without builder mirror", atom)
		}
		if !knownAtoms[atom] {
			t.Fatalf("builder plan shape references unknown atom %q", atom)
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

func assertMoleculeOrganismLinksAreBidirectional(t *testing.T, moleculePath, organismPath string) {
	t.Helper()
	type moleculeRecord struct {
		MoleculeID      string   `json:"molecule_id"`
		UsedByOrganisms []string `json:"used_by_organisms"`
	}
	type organismRecord struct {
		OrganismID       string   `json:"organism_id"`
		RequiredMolecule []string `json:"required_molecule_ids"`
	}

	moleculeRaw, err := os.ReadFile(moleculePath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", moleculePath, err)
	}
	var moleculeCatalog struct {
		Molecules []moleculeRecord `json:"molecules"`
	}
	if err := json.Unmarshal(moleculeRaw, &moleculeCatalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", moleculePath, err)
	}

	organismRaw, err := os.ReadFile(organismPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", organismPath, err)
	}
	var organismCatalog struct {
		Organisms []organismRecord `json:"organisms"`
	}
	if err := json.Unmarshal(organismRaw, &organismCatalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", organismPath, err)
	}

	moleculeToOrganisms := map[string]map[string]bool{}
	for _, molecule := range moleculeCatalog.Molecules {
		if moleculeToOrganisms[molecule.MoleculeID] != nil {
			t.Fatalf("duplicate molecule %q", molecule.MoleculeID)
		}
		moleculeToOrganisms[molecule.MoleculeID] = map[string]bool{}
		for _, organismID := range molecule.UsedByOrganisms {
			moleculeToOrganisms[molecule.MoleculeID][organismID] = true
		}
	}
	organismToMolecules := map[string]map[string]bool{}
	for _, organism := range organismCatalog.Organisms {
		if organismToMolecules[organism.OrganismID] != nil {
			t.Fatalf("duplicate organism %q", organism.OrganismID)
		}
		organismToMolecules[organism.OrganismID] = map[string]bool{}
		for _, moleculeID := range organism.RequiredMolecule {
			organismToMolecules[organism.OrganismID][moleculeID] = true
		}
	}

	for moleculeID, organisms := range moleculeToOrganisms {
		for organismID := range organisms {
			requiredMolecules, ok := organismToMolecules[organismID]
			if !ok {
				t.Fatalf("molecule %q used_by_organisms references unknown organism %q", moleculeID, organismID)
			}
			if !requiredMolecules[moleculeID] {
				t.Fatalf("molecule %q lists organism %q, but organism does not require that molecule", moleculeID, organismID)
			}
		}
	}
	for organismID, molecules := range organismToMolecules {
		for moleculeID := range molecules {
			usedByOrganisms, ok := moleculeToOrganisms[moleculeID]
			if !ok {
				t.Fatalf("organism %q requires unknown molecule %q", organismID, moleculeID)
			}
			if !usedByOrganisms[organismID] {
				t.Fatalf("organism %q requires molecule %q, but molecule does not list the organism", organismID, moleculeID)
			}
		}
	}
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

func loadTemplateResearchRuntimeDraftOrganisms(t *testing.T, path string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Coverage []struct {
			OrganismID  string `json:"organism_id"`
			DraftStatus string `json:"draft_status"`
		} `json:"coverage"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	var organisms []string
	for _, record := range catalog.Coverage {
		if record.DraftStatus == "runtime_draft_planner" {
			organisms = append(organisms, record.OrganismID)
		}
	}
	return organisms
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

func runtimeProductizationVerifierSpecIDExists(specs map[string]runtimeProductizationVerifierSpec, specID string) bool {
	for organismID := range specs {
		if organismID+"_verifier" == specID {
			return true
		}
	}
	return false
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

func loadAtomBuilderIDs(t *testing.T, path string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Builders []struct {
			AtomID string `json:"atom_id"`
		} `json:"builders"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	ids := map[string]bool{}
	for _, builder := range catalog.Builders {
		if ids[builder.AtomID] {
			t.Fatalf("duplicate atom builder %q", builder.AtomID)
		}
		ids[builder.AtomID] = true
	}
	return ids
}

func loadAtomBuilderPlanShapeAtoms(t *testing.T, path string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		PlanShapes []struct {
			ShapeID        string   `json:"shape_id"`
			AppliesToAtoms []string `json:"applies_to_atoms"`
		} `json:"plan_shapes"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	ids := map[string]bool{}
	for _, shape := range catalog.PlanShapes {
		for _, atom := range shape.AppliesToAtoms {
			ids[atom] = true
		}
	}
	return ids
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
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

func assertTemplatePatternsReferenceKnownVocabulary(t *testing.T, root string) {
	t.Helper()
	supportedAtoms := loadSupportedCapabilityNames(t, filepath.Join(root, "contracts", "capabilities", "capability.schema.json"))
	opportunityAtoms := loadOpportunityCapabilityNames(t, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))
	taxonomyPath := filepath.Join(root, "knowledge", "template-research", "taxonomy.md")
	visualLabels := loadTaxonomyInlineCodeSet(t, taxonomyPath, "Visual/frontend labels")
	dataLabels := loadTaxonomyInlineCodeSet(t, taxonomyPath, "Data/backend labels")
	validationLabels := loadTaxonomyInlineCodeSet(t, taxonomyPath, "Validation labels")
	workflowLabels := loadTaxonomyInlineCodeSet(t, taxonomyPath, "Workflow labels")

	paths, err := filepath.Glob(filepath.Join(root, "knowledge", "template-research", "patterns", "*.json"))
	if err != nil {
		t.Fatalf("Glob(patterns): %v", err)
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", path, err)
		}
		var pattern struct {
			PatternID                   string   `json:"pattern_id"`
			VisualPatterns              []string `json:"visual_patterns"`
			DataPatterns                []string `json:"data_patterns"`
			ValidationPatterns          []string `json:"validation_patterns"`
			WorkflowPatterns            []string `json:"workflow_patterns"`
			SourceObservations          []string `json:"source_observations"`
			ExistingCapabilities        []string `json:"existing_capabilities"`
			MissingCapabilityCandidates []string `json:"missing_capability_candidates"`
			RecurrenceStatus            string   `json:"recurrence_status"`
		}
		if err := json.Unmarshal(raw, &pattern); err != nil {
			t.Fatalf("Unmarshal(%s): %v", path, err)
		}
		if pattern.RecurrenceStatus == "recurring" && len(pattern.SourceObservations) < 2 {
			t.Fatalf("recurring pattern %q has %d source observations, want at least 2", pattern.PatternID, len(pattern.SourceObservations))
		}
		assertAllKnown(t, path, "visual_patterns", pattern.VisualPatterns, visualLabels)
		assertAllKnown(t, path, "data_patterns", pattern.DataPatterns, dataLabels)
		assertAllKnown(t, path, "validation_patterns", pattern.ValidationPatterns, validationLabels)
		assertAllKnown(t, path, "workflow_patterns", pattern.WorkflowPatterns, workflowLabels)
		for _, atom := range pattern.ExistingCapabilities {
			if !supportedAtoms[atom] {
				t.Fatalf("pattern %q existing capability %q is not supported", pattern.PatternID, atom)
			}
		}
		for _, atom := range pattern.MissingCapabilityCandidates {
			if !supportedAtoms[atom] && !opportunityAtoms[atom] {
				t.Fatalf("pattern %q missing capability candidate %q is neither supported nor tracked as opportunity", pattern.PatternID, atom)
			}
		}
	}
}

func loadTaxonomyInlineCodeSet(t *testing.T, path, marker string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	text := string(raw)
	start := strings.Index(text, marker)
	if start < 0 {
		t.Fatalf("taxonomy marker %q missing from %s", marker, path)
	}
	rest := text[start:]
	end := strings.Index(rest, "\n\n")
	if end < 0 {
		t.Fatalf("taxonomy marker %q has no paragraph boundary in %s", marker, path)
	}
	paragraph := rest[:end]
	matches := regexp.MustCompile("`([a-z0-9_]+)`").FindAllStringSubmatch(paragraph, -1)
	if len(matches) == 0 {
		t.Fatalf("taxonomy marker %q has no inline code labels", marker)
	}
	labels := map[string]bool{}
	for _, match := range matches {
		labels[match[1]] = true
	}
	return labels
}

func assertAllKnown(t *testing.T, path, field string, values []string, known map[string]bool) {
	t.Helper()
	for _, value := range values {
		if !known[value] {
			t.Fatalf("%s %s contains unknown label %q", path, field, value)
		}
	}
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

func assertHypothesesReferenceTriggerCatalogAndJudgments(t *testing.T, root string) {
	t.Helper()
	triggerPath := filepath.Join(root, "knowledge", "template-research", "hypotheses", "trigger_catalog.json")
	hypothesisPath := filepath.Join(root, "knowledge", "template-research", "hypotheses", "pattern_hypotheses.json")
	judgmentPath := filepath.Join(root, "knowledge", "template-research", "judgments", "pattern_registry.json")

	triggerIDs := loadTriggerCatalogIDs(t, triggerPath)
	judgmentDomains := loadJudgmentPatternDomains(t, judgmentPath)
	judgmentIDs := stringSetFromMapKeys(judgmentDomains)
	supportedAtoms := loadSupportedCapabilityNames(t, filepath.Join(root, "contracts", "capabilities", "capability.schema.json"))
	opportunityAtoms := loadOpportunityCapabilityNames(t, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json"))
	knownAtoms := unionStringSets(supportedAtoms, opportunityAtoms)
	assertTriggerCatalogReferencesKnownCapabilities(t, triggerPath, knownAtoms)

	raw, err := os.ReadFile(hypothesisPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", hypothesisPath, err)
	}
	var catalog struct {
		Patterns []struct {
			PatternID                  string   `json:"pattern_id"`
			Domain                     string   `json:"domain"`
			Triggers                   []string `json:"triggers"`
			LikelyExistingCapabilities []string `json:"likely_existing_capabilities"`
			LikelyMissingCapabilities  []string `json:"likely_missing_capabilities"`
		} `json:"patterns"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", hypothesisPath, err)
	}
	hypothesisIDs := map[string]bool{}
	for _, pattern := range catalog.Patterns {
		if hypothesisIDs[pattern.PatternID] {
			t.Fatalf("duplicate hypothesis pattern %q", pattern.PatternID)
		}
		hypothesisIDs[pattern.PatternID] = true
		if judgmentDomains[pattern.PatternID] != pattern.Domain {
			t.Fatalf("hypothesis %q domain=%q want judgment domain %q", pattern.PatternID, pattern.Domain, judgmentDomains[pattern.PatternID])
		}
		assertAllKnown(t, hypothesisPath, pattern.PatternID+".triggers", pattern.Triggers, triggerIDs)
		for _, atom := range pattern.LikelyExistingCapabilities {
			if !supportedAtoms[atom] {
				t.Fatalf("hypothesis %q likely existing capability %q is not supported", pattern.PatternID, atom)
			}
		}
		for _, atom := range pattern.LikelyMissingCapabilities {
			if !knownAtoms[atom] {
				t.Fatalf("hypothesis %q likely missing capability %q is neither supported nor tracked as opportunity", pattern.PatternID, atom)
			}
		}
	}
	assertExactStringSet(t, hypothesisIDs, judgmentIDs, hypothesisPath)
}

func assertTriggerCatalogReferencesKnownCapabilities(t *testing.T, path string, knownAtoms map[string]bool) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Groups map[string][]struct {
			ID                         string   `json:"id"`
			LikelySheetOpsCapabilities []string `json:"likely_sheet_ops_capabilities"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	for groupName, group := range catalog.Groups {
		for _, trigger := range group {
			for _, atom := range trigger.LikelySheetOpsCapabilities {
				if !knownAtoms[atom] {
					t.Fatalf("trigger %q in group %q references unknown capability %q", trigger.ID, groupName, atom)
				}
			}
		}
	}
}

func assertJudgmentsReferenceKnownCapabilities(t *testing.T, root string) {
	t.Helper()
	path := filepath.Join(root, "knowledge", "template-research", "judgments", "pattern_registry.json")
	knownAtoms := unionStringSets(
		loadSupportedCapabilityNames(t, filepath.Join(root, "contracts", "capabilities", "capability.schema.json")),
		loadOpportunityCapabilityNames(t, filepath.Join(root, "knowledge", "template-research", "opportunities", "*.json")),
	)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var registry struct {
		Judgments []struct {
			PatternID            string   `json:"pattern_id"`
			CapabilityCandidates []string `json:"capability_candidates"`
		} `json:"judgments"`
	}
	if err := json.Unmarshal(raw, &registry); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	for _, judgment := range registry.Judgments {
		for _, atom := range judgment.CapabilityCandidates {
			if !knownAtoms[atom] {
				t.Fatalf("judgment %q references unknown capability candidate %q", judgment.PatternID, atom)
			}
		}
	}
}

func loadTriggerCatalogIDs(t *testing.T, path string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var catalog struct {
		Groups map[string][]struct {
			ID string `json:"id"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	ids := map[string]bool{}
	for groupName, group := range catalog.Groups {
		for _, trigger := range group {
			if ids[trigger.ID] {
				t.Fatalf("duplicate trigger id %q in group %q", trigger.ID, groupName)
			}
			ids[trigger.ID] = true
		}
	}
	return ids
}

func unionStringSets(sets ...map[string]bool) map[string]bool {
	union := map[string]bool{}
	for _, set := range sets {
		for value := range set {
			union[value] = true
		}
	}
	return union
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

func loadJudgmentPatternIDs(t *testing.T, path string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var registry struct {
		Judgments []struct {
			PatternID string `json:"pattern_id"`
		} `json:"judgments"`
	}
	if err := json.Unmarshal(raw, &registry); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	ids := map[string]bool{}
	for _, judgment := range registry.Judgments {
		if ids[judgment.PatternID] {
			t.Fatalf("duplicate judgment pattern %q", judgment.PatternID)
		}
		ids[judgment.PatternID] = true
	}
	return ids
}

func loadJudgmentPatternDomains(t *testing.T, path string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	var registry struct {
		Judgments []struct {
			PatternID string `json:"pattern_id"`
			Domain    string `json:"domain"`
		} `json:"judgments"`
	}
	if err := json.Unmarshal(raw, &registry); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
	domains := map[string]string{}
	for _, judgment := range registry.Judgments {
		if domains[judgment.PatternID] != "" {
			t.Fatalf("duplicate judgment pattern %q", judgment.PatternID)
		}
		domains[judgment.PatternID] = judgment.Domain
	}
	return domains
}

func stringSetFromMapKeys[V any](values map[string]V) map[string]bool {
	set := map[string]bool{}
	for key := range values {
		set[key] = true
	}
	return set
}

func loadClassifiedPatternIDs(t *testing.T, path string) map[string]bool {
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
				t.Fatalf("pattern %q appears in multiple classification groups in %s", patternID, path)
			}
			classified[patternID] = true
		}
	}
	return classified
}

func assertExactStringSet(t *testing.T, got, want map[string]bool, label string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s set size=%d want %d", label, len(got), len(want))
	}
	for value := range want {
		if !got[value] {
			t.Fatalf("%s missing %q", label, value)
		}
	}
	for value := range got {
		if !want[value] {
			t.Fatalf("%s includes unexpected %q", label, value)
		}
	}
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
