<!--
Sync Impact Report:
- Version: 1.0.0 → 1.1.1 (PATCH)
- Amendment: Fixed security and usability issues in code review requirement
- Changes from v1.1.0:
  - Fixed HIGH severity: Added security guidance for data privacy
  - Fixed HIGH severity: Clarified command syntax (Claude Code vs CLI)
  - Fixed MEDIUM severity: Changed from "unstaged" to "staged" changes
  - Added: Security safeguards (redact secrets, exclude sensitive files)
  - Added: Exemptions for trivial changes and hotfixes
  - Added: Evidence format requirement (review-report.md)
  - Added: API key handling guidance (referenced in rationale)
- Principles modified:
  - Principle II (Quality Gates): Added security requirements and exemptions
- Pre-Commit Requirements updated:
  - Step 1: Enhanced with security, privacy, and exemptions guidance
- Templates requiring updates:
  ✅ plan-template.md - No changes needed
  ✅ spec-template.md - No changes needed
  ✅ tasks-template.md - No changes needed
  ✅ CLAUDE.md - Updated with security guidance
- Follow-up: None (all security issues addressed)
-->


# Dungo Constitution

## Core Principles

### I. Test-First Development (NON-NEGOTIABLE)

All code MUST be written following strict Test-Driven Development (TDD):

- Tests written FIRST → User approves tests → Tests FAIL → Implementation begins
- Red-Green-Refactor cycle is mandatory, no exceptions
- Unit tests for all packages (graph, synthesis, embedding, carving, content, validation)
- Property-based tests for constraints (connectivity, key-before-lock, degree bounds)
- Golden tests for determinism (fixed seeds produce identical SVG/JSON snapshots)
- Integration tests via simulated agent (verifies boss and key findability)
- Test artifacts stored in `testdata/` with versioned schema

**Rationale**: The dungeon generator is a constraint-driven system with complex invariants. Without TDD, subtle bugs in graph synthesis or spatial embedding will compound through the pipeline. Property-based testing is essential for validating constraints across random seeds.

### II. Quality Gates

Code MUST pass quality gates before commit:

- **Code review MUST run**: `mcp-pr` review of staged changes with OpenAI provider
  - **Security requirement**: Only review non-sensitive code (exclude secrets, credentials, PII)
  - **Data privacy**: Redact sensitive information before sending to external LLM
  - **Action**: Address all HIGH severity findings before committing
  - **Tech debt**: Document MEDIUM severity as tech debt if not immediately fixable
  - **Exemptions**: Trivial changes (<10 lines), automated commits, emergency hotfixes may skip
  - **Evidence**: Save review report (`--output=review-report.md`) for PR compliance
- **Linting MUST pass**: `golangci-lint run` with zero errors (configuration in `.golangci.yml`)
- **All tests MUST pass**: `go test ./...` with 100% success rate
- **No exceptions**: Commits with failing lint, tests, or unaddressed HIGH severity code review findings are PROHIBITED

Configured linters (non-negotiable): gofmt, govet, staticcheck, errcheck, gosimple, ineffassign, unused, typecheck, gocyclo (max 15), misspell

**Rationale**: Quality gates prevent technical debt accumulation. The codebase must maintain high standards from day one. Automated checks and AI-powered code review catch issues before they propagate. mcp-pr provides additional analysis for security, best practices, and architectural consistency. Security safeguards ensure sensitive data never leaves the organization.

### III. Deterministic Design

The generator MUST produce identical output given identical inputs:

- Single master seed determines all randomness
- **Stage-specific sub-seeds via**: `seed_stage = H(master_seed, stage_name, config_hash)` where:
  - `H` = SHA-256 hash function (crypto/sha256)
  - Output: First 8 bytes of hash as big-endian uint64
  - `master_seed`: 8-byte big-endian encoding of the master seed
  - `stage_name`: UTF-8 bytes of stage identifier ("synthesis", "embedding", etc.)
  - `config_hash`: SHA-256 hash of serialized configuration
- All random decisions use stage-local RNG with `math/rand.NewSource(int64(seed_stage))`
- Never use global `rand` functions or `crypto/rand`
- Pure functions throughout pipeline (Graph → Embed → Carve → Content → Validate)
- Reproducibility enables caching, stepwise testing, and debugging

**Rationale**: Determinism is a core requirement per specification §10. Non-deterministic generation breaks golden tests, makes debugging impossible, and violates the fundamental contract with users who expect consistent results from seeds. Precise specification of hash function and seed derivation ensures cross-platform reproducibility.

### IV. Performance Consciousness

Implementation MUST meet performance targets:

- 60-room dungeon: < 50ms for graph + embedding stage
- Full generation: < 200ms including carve and content
- Memory: < 50MB per generation
- **Runtime generation: Single-threaded** (v1 - ensures determinism and predictable memory)
- O(N log N) typical complexity for N rooms

Performance regressions caught via benchmarks (`go test -bench ./...`)

**Rationale**: Performance targets defined in specification §17 are non-negotiable. The generator must be production-ready for game engines. Performance debt is difficult to fix later; design with targets from the start. Single-threaded runtime ensures deterministic behavior and predictable memory usage; parallelism is reserved for development tooling (tests, CI, concurrent agents).

### V. Concurrent Execution

Development workflow MUST leverage parallel execution (**development time only**):

- Use concurrent agents whenever possible (multiple Task tool calls in single message)
- Parallel test execution across packages when independent (`go test -parallel`)
- Parallel implementation of independent user stories
- Parallel linting and testing during CI/CD
- Task planning explicitly marks parallelizable work with [P] prefix

**Scope**: This principle applies to **development tooling and CI/CD only**. Runtime dungeon generation remains single-threaded (Principle IV) for determinism and predictable resource usage.

**Rationale**: Maximizes development velocity and mirrors how the pipeline stages can be independently developed and tested. Reduces latency in AI-assisted development by batching tool calls. Clear separation between development parallelism (encouraged) and runtime behavior (single-threaded).

## Quality Standards

### Code Organization

- Package-per-stage architecture: `pkg/dungeon`, `pkg/graph`, `pkg/synthesis`, `pkg/embedding`, `pkg/carving`, `pkg/content`, `pkg/validation`, `pkg/rng`
- Use `pkg/` for library packages (not `internal/` initially)
- No shared `util` packages; utilities live in respective domain packages
- Each package independently testable with minimal dependencies

### Error Handling

- Return errors, never panic (except unrecoverable programmer errors)
- Wrap errors with context: `fmt.Errorf("stage failed: %w", err)`
- Validate inputs at pipeline stage boundaries
- Use context.Context for cancellation and timeouts

### Constraint Validation

Hard constraints MUST pass before proceeding to next stage:

- Single connected component (or `allowDisconnected=true` for teleport motifs)
- Degree bounds per room and globally
- Key-before-lock reachability for all lock gates
- At least one viable Start→Boss path within `minLen..maxLen`
- No spatial overlaps; corridor feasibility

Soft constraints optimized but not required:

- Pacing adherence to target curve
- Thematic clustering (biome continuity)
- Cycle density within acceptable band
- Secret density and reward pacing

## Development Workflow

### Pre-Commit Requirements

Before committing, developer MUST:

1. **Run code review**: Use mcp-pr to review staged changes before commit
   - **Security**: Review only non-sensitive code (redact secrets, credentials, PII before review)
   - **Command** (in Claude Code): `mcp__mcp-pr__review_staged` with OpenAI provider
   - **CLI alternative**: `mcp-pr review --staged --provider=openai --output=review-report.md`
   - **Data privacy**: Never review files containing secrets, API keys, credentials, or PII
   - **Evidence**: Save review report for PR inclusion
   - **Action**: Address all HIGH severity findings before committing
   - **Tech debt**: Document MEDIUM severity findings if not immediately fixable
   - **Exemptions**: Automated commits, hotfixes, and trivial changes (<10 lines) may skip review
2. Run `golangci-lint run` → MUST pass with zero errors
3. Run `go test ./...` → MUST pass 100% tests
4. Run `gofmt -l .` → MUST return empty (no unformatted files)
5. Verify no debug code, TODOs, or commented-out blocks

### Implementation Process

1. Review specification in `specs/dungo_specification.md`
2. Use `/speckit.specify` to create feature spec with user stories
3. Use `/speckit.plan` to generate implementation plan
4. Use `/speckit.tasks` to break down into task list
5. Write tests FIRST (TDD red phase)
6. Verify tests FAIL appropriately
7. Implement feature (TDD green phase)
8. Refactor for clarity and performance (TDD refactor phase)
9. Run quality gates (lint + tests)
10. Commit with descriptive message

### Testing Discipline

For each feature:

- **Unit tests**: Test individual functions and types
- **Integration tests**: Test pipeline stage interactions
- **Property tests**: Test constraints across random seeds (use `gopter` or similar)
- **Golden tests**: Snapshot SVG/JSON for fixed seeds; diff on changes
- **Fuzz tests**: Push boundaries until violations surface

Test coverage tracked; aim for >80% overall, >90% for core packages.

## Governance

This constitution supersedes all other development practices. All code reviews, pull requests, and commits MUST verify compliance with these principles.

### Amendment Process

1. Propose amendment with rationale and impact analysis
2. Discuss with team/stakeholders
3. Update constitution with version bump (semantic versioning)
4. Update dependent templates and documentation
5. Commit with sync impact report

### Versioning

Constitution version follows MAJOR.MINOR.PATCH:

- **MAJOR**: Backward incompatible principle removals or redefinitions
- **MINOR**: New principle or materially expanded guidance
- **PATCH**: Clarifications, wording, typo fixes

### Compliance Review

All pull requests MUST include:

- Evidence of mcp-pr code review with HIGH severity findings addressed
- Evidence of passing lint (`golangci-lint run` output)
- Evidence of passing tests (`go test ./...` output)
- Confirmation of TDD workflow (tests written first)
- Performance benchmark results if touching hot paths

Use `CLAUDE.md` for runtime development guidance specific to this repository.

**Version**: 1.1.1 | **Ratified**: 2025-11-04 | **Last Amended**: 2025-11-04
