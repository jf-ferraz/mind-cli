# Convergence Analysis: Phase 2 TUI Dashboard Implementation Strategy

**Topic**: Phase 2 TUI Dashboard -- implementation strategy for the mind-cli interactive terminal interface. Addresses: (a) priority ordering of SHOULD fixes vs new TUI features, (b) Bubble Tea architecture (single model vs per-tab models), (c) service integration patterns (how tabs access existing services), (d) testing strategy for TUI components, (e) scope management given Phase 2 is significantly UI-focused.
**Date**: 2026-03-11
**Personas**: Architect, Pragmatist, Critic, Researcher
**Variant**: Deep (4 personas, 2 rounds)
**Effective Persona Count**: 3.5 (see Diversity Audit)

---

## Phase 2: Opening Positions

### Position A -- The Architect

**Core Recommendation**: Fix all 7 SHOULD items in a focused pre-Phase-2 sprint, then implement the TUI using a delegated per-tab model architecture with strict service interface boundaries. The TUI is a new presentation layer that must integrate cleanly with existing services; building it atop known deviations compounds architectural risk.

**Detailed Position**:

1. **SHOULD fixes first, no exceptions.** The 7 outstanding SHOULD items represent architectural deviations that will complicate TUI integration:
   - **S-1 (mutually exclusive flags)**: Trivial fix, but the pattern of unchecked flag conflicts could propagate to TUI command dispatch.
   - **S-2 (missing documents check in ReconcileSuite)**: The Checks tab (Tab 4) will display ReconcileSuite results. Missing this check means the TUI will show incomplete validation data from day one.
   - **S-3 (transitive reason strings)**: The Status tab displays staleness reasons. Generic "may be outdated" strings undermine the TUI's value proposition of surfacing *why* documents are stale.
   - **`--project` rename**: The TUI will use the same project root resolution. Fixing the flag name now prevents a breaking change later.
   - **`docs search` abstraction**: The Docs tab (Tab 2) will need DocRepo for search. If search bypasses the abstraction, the TUI either duplicates the bypass or cannot offer search.
   - **GoDoc gaps**: Adding a new presentation layer means more consumers of these exported methods. Documentation must be in place.
   - **Wiring centralization**: Phase 1.5 partially centralized wiring via PersistentPreRunE. The TUI needs services injected at construction time (BP-05 Section 3 specifies this). Completing centralization in main.go enables clean injection for both CLI and TUI entry points.

2. **Per-tab model architecture.** BP-05 Section 3 specifies this explicitly: `App` delegates to tab-specific `tea.Model` implementations. This is not a design question -- it is a specification compliance question. The architecture:
   - `App` (top-level model) manages active tab index, global keys, and service references.
   - Each tab (`StatusView`, `DocsView`, `IterationsView`, `ChecksView`, `QualityView`) is an independent `tea.Model` with its own state.
   - Communication via Bubble Tea messages only -- no direct parent-child state mutation.
   - Services injected into `App` at construction, passed to tab models during initialization.

3. **Service interface boundaries.** The TUI must access services through the same interfaces as the CLI:
   - `ProjectService.Health()` for StatusView
   - `ValidationService.RunAll()` for ChecksView
   - `DocRepo.Read()` + Glamour for preview
   - `IterationRepo.List()` for IterationsView
   - `QualityRepo.ReadLog()` for QualityView
   - `ReconciliationService.ReadStaleness()` for staleness panels

   No new service methods should be needed. The existing service layer was designed with multiple consumers in mind (CLI, TUI, MCP).

4. **Implementation sequence.** Follow BP-08 Phase 2 structure, bottom-up:
   - Step 1: `tui/styles.go` + `tui/keys.go` (theme and keybindings)
   - Step 2: `tui/app.go` (shell: chrome, tab bar, delegation)
   - Step 3: `tui/status.go` (Tab 1 -- surfaces the most data, validates service integration)
   - Step 4: `tui/docs.go` (Tab 2 -- list + filter + preview)
   - Step 5: `tui/iterations.go` (Tab 3 -- table + filter + expand)
   - Step 6: `tui/checks.go` (Tab 4 -- accordion + lazy loading)
   - Step 7: `tui/quality.go` (Tab 5 -- chart + detail)
   - Step 8: Help overlay, responsive design, monochrome fallback
   - Step 9: `cmd/tui_cmd.go` wiring

**Evidence Registry**:

| ID | Claim | Source | Confidence |
|----|-------|--------|------------|
| A1 | Per-tab delegation is Bubble Tea best practice | Charm docs: "Composing models" tutorial, lip.sh/blog/tea-models | Medium [Single-source] |
| A2 | BP-05 specifies per-tab model delegation | BP-05 Section 3, Component Hierarchy | Very High |
| A3 | Service interfaces designed for multiple consumers | BP-01 Section 2.2, BP-08 Phase 2 "Services injected at construction" | High |
| A4 | Wiring centralization partially complete | Phase 1.5 reviewer validation.md, S-5 and C-10 | High |

**Falsifiability Conditions**:
- If SHOULD fixes take more than 3 focused sessions (>6 hours), they should be time-boxed and remaining items deferred.
- If any SHOULD fix requires changes to more than 5 files, reassess scope.

---

### Position B -- The Pragmatist

**Core Recommendation**: Bundle the 3 highest-impact SHOULD fixes into the first TUI implementation step. Do not create a separate pre-Phase-2 sprint. Build the TUI tab-by-tab, shipping each tab as a working increment. Defer low-impact SHOULD items.

**Detailed Position**:

1. **Cherry-pick SHOULD fixes, do not batch all 7.** Three SHOULD items directly affect TUI quality; fix those. The rest are cosmetic or affect CLI-only paths:
   - **Fix S-1** (mutually exclusive flags) -- 5-line change in `cmd/reconcile.go`. Bundle into TUI step 1.
   - **Fix S-2** (missing documents check) -- needed for Checks tab accuracy. ~15 lines in `internal/validate/reconcile.go`. Bundle into Checks tab implementation.
   - **Fix `docs search` abstraction** -- needed for Docs tab search feature. Refactor to use DocRepo. Bundle into Docs tab implementation.
   - **Defer S-3** (transitive reasons) -- the Status tab can show "stale (transitive)" which is already informative. Edge-type-specific reasons at depth > 0 are a polish item.
   - **Defer `--project` rename** -- the TUI does not use CLI flags. Rename is a breaking change best done in a dedicated release.
   - **Defer GoDoc gaps** -- does not affect TUI functionality or testability.
   - **Defer wiring centralization** -- the TUI gets its own wiring path through `cmd/tui_cmd.go`. The CLI wiring in PersistentPreRunE already works.

2. **Tab-by-tab delivery with working increments.** Each tab should compile and run independently:
   - Start with `tui/app.go` + `tui/status.go` (minimum viable TUI).
   - Each subsequent tab is additive. The TUI is usable after Tab 1.
   - This matches how Bubble Tea applications are actually built -- incremental composition.

3. **Reuse Bubbles components aggressively.** The `bubbles` library provides table, list, viewport, spinner, text input, and progress bar components. BP-05 Section 3 references these. Do not build custom components where Bubbles already provides them:
   - `bubbles/table` for Iterations tab
   - `bubbles/list` for Docs tab (customized renderer)
   - `bubbles/viewport` for preview pane and scrollable content
   - `bubbles/spinner` for loading states
   - `bubbles/textinput` for search

4. **Testing strategy: test the service layer, not the TUI.** Bubble Tea models are difficult to unit test in isolation because they produce styled string output. The pragmatic approach:
   - **Service layer tests exist and are comprehensive** (246 tests, 80%+ coverage on domain and validate).
   - **TUI tests**: test `Update()` functions with known `tea.Msg` inputs and verify the resulting model state (not rendered output). Test that key messages produce correct state transitions.
   - **Do not test `View()` functions** -- they are pure rendering with no business logic. Visual correctness is verified by manual testing and the acceptance criteria in BP-05.
   - **Integration test**: a single `TestTUILaunchAndQuit` that verifies the TUI initializes without panic and handles `tea.Quit` cleanly.

5. **Scope management.** BP-05 specifies Watch TUI (Section 7) and Orchestration TUI (Section 8). These are Phase 3-4 deliverables, not Phase 2. Phase 2 scope is exactly: `mind tui` with 5 tabs + help overlay + responsive design.

**Evidence Registry**:

| ID | Claim | Source | Confidence |
|----|-------|--------|------------|
| B1 | Bubbles provides table, list, viewport, spinner, textinput | github.com/charmbracelet/bubbles README | High |
| B2 | Bubble Tea models are composable via delegation | Charm "Tutorials: Composing Models" | Medium [Single-source] |
| B3 | TUI View() functions are pure rendering | Bubble Tea architecture: View is a pure function of Model state | High |
| B4 | 246 tests pass with 80%+ on domain/validate | Phase 1.5 validation.md deterministic gate | Very High |
| B5 | Watch/Orchestration TUI are Phase 3-4 | BP-08 Sections 5-6, BP-05 Sections 7-8 | Very High |

**Falsifiability Conditions**:
- If the 3 cherry-picked SHOULD fixes cannot be cleanly integrated into TUI implementation steps (require architectural refactoring), escalate to full pre-Phase-2 sprint.
- If any Bubbles component requires more than 100 lines of customization to match BP-05 wireframes, evaluate building a custom component instead.

---

### Position C -- The Critic

**Core Recommendation**: The biggest risk to Phase 2 is not the SHOULD items or the Bubble Tea architecture -- it is the scope. BP-05 specifies an extremely detailed TUI with 30+ components, 3 responsive breakpoints, monochrome fallback, 5 tab-specific key binding sets, and a help overlay system. This is substantially more code than Phases 1 and 1.5 combined in the presentation layer. Challenge whether all of BP-05 is needed for Phase 2 MVP.

**Detailed Position**:

1. **Scope risk is the dominant threat.** BP-05 is 1000+ lines of specification. It defines:
   - 5 tabs with distinct data requirements
   - 30+ individual components (per Section 3 component hierarchy)
   - 3 layout modes (narrow < 80, standard 80-99, wide >= 100)
   - Monochrome fallback with alternative symbols
   - Preview pane with Glamour markdown rendering
   - ASCII line chart for Quality tab
   - Accordion with detail expansion for Checks tab
   - Zone and type filter bars with keyboard shortcuts
   - Help overlay with context-sensitive keybindings
   - Focus management across 3 layers (tab, component, modal)

   For comparison: Phase 1 delivered 20+ CLI commands in ~1800 lines of cmd/ code. Phase 2's TUI specification implies 2000-3000+ lines of presentation code plus custom components. This is a significant expansion.

2. **The testing gap is real.** The Pragmatist's suggestion to skip View() testing is reasonable but creates a coverage gap. TUI bugs tend to be visual: layout overflow, truncation errors, color bleed, focus trapping. These cannot be caught by state transition tests alone. The project has 246 tests and high coverage -- Phase 2 will add a large untested surface area.

3. **SHOULD items create a hidden dependency.** S-2 (missing documents check) means the Checks tab will show incomplete data. If this fix is deferred or botched, the TUI launches with a known data gap. The Pragmatist's approach of bundling fixes into tab implementation is pragmatic but risks entangling bug fixes with new feature code, making both harder to review.

4. **The ASCII chart (Quality tab) is a custom component with no Bubbles equivalent.** `bubbles` does not provide a charting component. The Quality tab requires a custom ASCII line chart that:
   - Scales to terminal width
   - Plots data points with `●` connected by lines (`─`, `╭`, `╯`, `╰`, `╮`)
   - Shows a dashed threshold line
   - Supports left-right navigation across data points

   This is a non-trivial custom component. It could consume a disproportionate amount of development time relative to its value. [Unsourced assertion -- no prior art for effort estimation on ASCII charts in Go TUI.]

5. **Service readiness check.** Not all services the TUI needs are fully implemented:
   - `QualityService.ReadLog()` -- QualityRepo and QualityService exist in the architecture but were not part of Phase 1 or 1.5 deliverables. The Quality tab depends on `quality-log.yml` parsing that may not be implemented yet.
   - `DocRepo.Read()` for preview -- exists but the Glamour rendering pipeline is new.
   - The Docs tab search requires DocRepo search which currently bypasses the abstraction (SHOULD item).

6. **Recommendation: Define a Phase 2 MVP subset.** Instead of implementing the full BP-05 specification:
   - **Phase 2a (MVP)**: App shell + Tab 1 (Status) + Tab 4 (Checks) + Tab 2 (Docs, no preview) + basic responsive design. This covers the highest-value surfaces.
   - **Phase 2b (Complete)**: Tab 3 (Iterations) + Tab 5 (Quality) + preview pane + help overlay + monochrome fallback + full responsive design.
   - This reduces risk by shipping a usable TUI faster and gathering feedback before building the less critical tabs.

**Evidence Registry**:

| ID | Claim | Source | Confidence |
|----|-------|--------|------------|
| C1 | BP-05 specifies 30+ components | BP-05 Section 3 Component Hierarchy (explicit tree) | Very High |
| C2 | Phase 1 cmd/ layer was ~1800 lines | Codebase: 14 cmd/ files at ~130 lines average | Medium [estimate] |
| C3 | Bubbles does not include a chart component | github.com/charmbracelet/bubbles -- no chart in component list | High |
| C4 | QualityService may not be implemented | BP-08 Phase 1 scope does not list QualityService; Phase 1 validation.md does not test it | High |
| C5 | Glamour markdown rendering is a new dependency | Not in Phase 1 or 1.5 go.mod dependencies | High |

**Falsifiability Conditions**:
- If QualityService and QualityRepo are already implemented with tests, the Quality tab scope concern is reduced.
- If the ASCII chart can be implemented in under 200 lines using a third-party library (e.g., `termui` or `asciigraph`), the effort concern is reduced.
- If BP-05 specifies features that the project brief does not require for Phase 2 acceptance, the MVP subset approach is valid.

---

### Position D -- The Researcher

**Core Recommendation**: Examine prior art in Bubble Tea TUI applications for architectural patterns, testing strategies, and scope management. Provide empirical grounding for the architectural and testing decisions.

**Detailed Position**:

1. **Bubble Tea architectural patterns from production applications.** Several production Bubble Tea applications demonstrate the per-tab delegation pattern at scale:

   - **Soft Serve** (github.com/charmbracelet/soft-serve): Git server TUI with multiple views. Uses a top-level model with delegated sub-models. Each view maintains independent state. Communication via custom message types. Source: Charm Labs open-source repository. [Single-source -- confidence capped at Medium]

   - **Glow** (github.com/charmbracelet/glow): Markdown reader with stash/local/settings views. Demonstrates tab-like navigation with independent view models. Uses `tea.Cmd` for async data loading (file reads, API calls). Source: Charm Labs open-source. [Single-source -- confidence capped at Medium]

   - **Lazygit** (github.com/jesseduffield/lazygit): Not Bubble Tea, but the most mature Go TUI. Uses a panel-based architecture where each panel has its own view and controller. 75K+ stars. Demonstrates that large TUI applications benefit from strong panel/view isolation. Source: GitHub repository. [Single-source -- confidence capped at Medium]

   Pattern observed: All three applications use delegated sub-models with message-based communication. None use a monolithic model for the entire application.

2. **Testing patterns for Bubble Tea applications.** The Charm ecosystem provides `teatest`, a testing library specifically for Bubble Tea:

   - **`teatest`** (github.com/charmbracelet/x/exp/teatest): Provides `NewModel()`, `WaitFor()`, and golden file testing for Bubble Tea output. Enables programmatic interaction: send key messages, wait for specific output, compare against golden files. Source: Charm Labs experimental library. [Single-source -- confidence capped at Medium]

   - **Pattern from Soft Serve**: Tests `Update()` with specific messages and asserts on model state. Does not test `View()` output directly for most components. Source: soft-serve test files. [Single-source]

   - **Golden file testing for View()**: `teatest` supports golden file comparison of `View()` output. This addresses the Critic's concern about visual testing coverage. However, golden files are fragile -- any styling change breaks tests. The trade-off: initial investment in golden files vs. ongoing maintenance cost.

3. **ASCII chart libraries in Go.** Several options exist for the Quality tab chart:

   - **`asciigraph`** (github.com/guptarohit/asciigraph): Dedicated ASCII line chart library. 2.7K stars. Renders to a string that can be embedded in a Bubble Tea View(). Supports custom dimensions, labels, and threshold lines. Would require minor adaptation for interactive navigation (data point selection). Source: GitHub repository, README examples. [Single-source -- confidence capped at Medium]

   - **Custom implementation**: BP-05 specifies exact chart characters (`●`, `─`, `╭`, `╯`, `╰`, `╮`). The `asciigraph` library uses different characters (`┤`, `╭`, `╮`, `╰`, `╯`, `─`). A custom implementation would match the spec exactly but requires more effort. [Unsourced assertion for effort estimate]

4. **Glamour for markdown preview.** Glamour (github.com/charmbracelet/glamour) is the standard Charm library for terminal markdown rendering:

   - Already part of the Charm ecosystem (compatible with Lip Gloss styles).
   - Returns a rendered string suitable for embedding in a `bubbles/viewport`.
   - The Docs tab preview pane maps directly to `glamour.Render()` + `viewport.Model`.
   - New dependency but within the already-committed Charm ecosystem. Source: Charm docs. [Single-source -- confidence capped at Medium]

5. **Scope estimation based on prior implementations.** Examining Glow's codebase (a comparable 5-view Bubble Tea application):
   - Glow's TUI code (excluding server/API logic) is approximately 3,500 lines across 15 files. [estimate from repository structure]
   - mind-cli's BP-05 specification is more detailed (5 tabs with more features per tab than Glow's views).
   - Estimated TUI code for full BP-05: 2,500-4,000 lines. This is substantial but not unprecedented for a Bubble Tea application.
   - The Critic's concern about scope is valid -- this is a significant implementation effort.

**Evidence Registry**:

| ID | Claim | Source | Confidence |
|----|-------|--------|------------|
| D1 | Soft Serve uses delegated sub-models | github.com/charmbracelet/soft-serve source | Medium [Single-source] |
| D2 | teatest provides golden file and programmatic testing | github.com/charmbracelet/x/exp/teatest README | Medium [Single-source] |
| D3 | asciigraph provides ASCII line charts in Go | github.com/guptarohit/asciigraph README, 2.7K stars | Medium [Single-source] |
| D4 | Glamour renders markdown for terminal display | github.com/charmbracelet/glamour README | Medium [Single-source] |
| D5 | Glow TUI is ~3,500 lines across 15 files | Approximate from glow repository structure | Low [estimate] |
| D6 | Lazygit uses panel isolation at 75K+ stars | github.com/jesseduffield/lazygit | Medium [Single-source] |

**Falsifiability Conditions**:
- If `teatest` does not support golden file testing for styled (Lip Gloss) output, the golden file approach is not viable.
- If `asciigraph` cannot be adapted for interactive data point selection, a custom chart is required.
- If Glow's actual line count is significantly different from the estimate, the scope projection should be revised.

---

## Phase 2.5: Diversity Audit

### Pairwise Similarity Assessment

| Pair | Similarity | Key Differentiator |
|------|-----------|-------------------|
| Architect -- Pragmatist | 0.55 | Agree on per-tab architecture and service reuse. Disagree on SHOULD fix timing (all-first vs cherry-pick). |
| Architect -- Critic | 0.35 | Both identify scope risks differently. Architect focuses on architectural cleanliness; Critic focuses on specification size. |
| Architect -- Researcher | 0.40 | Researcher provides evidence for Architect's architecture claims. No significant tension. |
| Pragmatist -- Critic | 0.30 | Genuine tension: Pragmatist wants incremental delivery; Critic argues the scope is too large even for incremental. |
| Pragmatist -- Researcher | 0.45 | Researcher's testing evidence supports Pragmatist's "test Update, not View" approach. |
| Critic -- Researcher | 0.25 | Most distinct pair. Critic challenges scope; Researcher provides comparative evidence. Researcher's scope estimate validates Critic's concern. |

### Effective Persona Count: 3.5

The Architect and Pragmatist share significant overlap on the fundamental architecture question (per-tab delegation) and service reuse strategy. Their disagreement on SHOULD fix timing is genuine but relatively narrow. The Critic provides the most unique perspective by challenging the scope assumption. The Researcher provides grounding evidence but does not introduce a distinct position -- more of an amplifier for other positions.

### Tension Matrix

| Tension | Poles | Severity |
|---------|-------|----------|
| T1: SHOULD fix timing | Architect (all-first) vs Pragmatist (cherry-pick) | Medium |
| T2: Scope management | Critic (MVP subset) vs Architect+Pragmatist (full BP-05) | High |
| T3: Testing depth | Pragmatist (skip View tests) vs Critic (coverage gap) | Medium |
| T4: Quality tab feasibility | Critic (custom chart is risky) vs Researcher (asciigraph exists) | Low |
| T5: Service readiness | Critic (QualityService may not exist) vs all others (assumed ready) | Medium |

---

## Phase 3: Cross-Examination

### Architect challenges Pragmatist

**Challenge A→B-1**: The Pragmatist recommends deferring the `--project` flag rename and wiring centralization. But BP-01 Section 4 specifies constructor injection in `main.go`, and the TUI entry point (`cmd/tui_cmd.go`) needs services injected at construction. How does the TUI get its services if wiring remains distributed across cmd/ handlers?

**Challenge A→B-2**: The Pragmatist says "the TUI gets its own wiring path through `cmd/tui_cmd.go`." This creates a second wiring point -- the CLI has PersistentPreRunE, and the TUI has its own constructor. When a new service is added in Phase 3, it must be wired in both places. This is the exact problem the Architect's centralization solves.

### Pragmatist challenges Architect

**Challenge B→A-1**: The Architect wants to fix all 7 SHOULD items before starting TUI work. Items like "5 exported methods lack GoDoc" and "Error() methods lack GoDoc" are documentation-only changes with zero functional impact on the TUI. How does the Architect justify blocking TUI development for GoDoc comments?

**Challenge B→A-2**: The Architect's 9-step implementation sequence treats all tabs equally. But Tab 1 (Status) surfaces the most critical data and validates the entire service integration pattern. If Tab 1 works, the architecture is proven. Why not ship Tab 1 and iterate, rather than planning all 9 steps upfront?

### Critic challenges Architect and Pragmatist

**Challenge C→AB-1**: Both the Architect and Pragmatist assume the full BP-05 specification is Phase 2 scope. But BP-08 Section 4 lists 14 acceptance criteria for Phase 2. Several BP-05 features (monochrome fallback, help overlay, responsive design at 3 breakpoints, preview pane with Glamour) go significantly beyond these acceptance criteria. Has anyone verified that BP-05 and BP-08 are aligned on scope?

**Challenge C→AB-2**: The Quality tab depends on `QualityService.ReadLog()` which reads `quality-log.yml`. Phase 1 delivered `QualityScore` domain types but the service implementation and repository may be incomplete. If Quality tab development reveals missing infrastructure, it becomes a blocking dependency that could delay the entire Phase 2. What is the contingency?

### Researcher challenges Critic

**Challenge D→C-1**: The Critic proposes splitting Phase 2 into 2a and 2b. But BP-08 Phase 2 acceptance criteria include "All 5 tabs render correctly at 80x24 minimum terminal size." If Phase 2a ships without Tabs 3 and 5, it fails the acceptance criteria. How does the Critic reconcile the MVP approach with the specification?

**Challenge D→C-2**: The Critic estimates TUI code will be 2000-3000+ lines. The Researcher's examination of Glow suggests 2,500-4,000 for a comparable application. This is significant but Phase 1 delivered ~4000+ lines across all packages in a single iteration. What specifically about the TUI makes this scope riskier than Phase 1?

---

## Phase 4: Rebuttals

### Pragmatist rebuttals

**Rebuttal to A→B-1 and A→B-2**: The Architect is correct that dual wiring points create maintenance burden. **Concession**: Wiring centralization should be included in the Phase 2 prerequisites, not deferred. However, this is a single focused task (move wiring from PersistentPreRunE to a `buildDeps()` function in `main.go` callable by both CLI and TUI paths), not a reason to batch all 7 SHOULD items.

**Revised position**: Fix S-1, S-2, wiring centralization, and `docs search` abstraction before TUI work. Defer S-3 (transitive reasons), `--project` rename, and GoDoc gaps.

### Architect rebuttals

**Rebuttal to B→A-1**: **Concession**: GoDoc comments on `Error()` methods and the 5 exported methods do not block TUI development. These can be addressed opportunistically during the TUI implementation. The Architect revises to: fix all functionally-impactful SHOULD items (S-1, S-2, S-3, wiring centralization, `docs search` abstraction), defer GoDoc and flag rename.

**Rebuttal to B→A-2**: The 9-step sequence is a dependency order, not a waterfall plan. Each step is independently shippable. Tab 1 (Status) is Step 3, which requires Steps 1-2 (styles, app shell). The Pragmatist's "ship Tab 1 and iterate" approach is exactly what the sequence describes.

### Critic rebuttals

**Rebuttal to D→C-1**: **Concession**: BP-08 acceptance criteria explicitly require all 5 tabs. The Phase 2a/2b split as proposed would fail the gate. Revised recommendation: instead of splitting into sub-phases, **define a minimum viable implementation per tab**. Tab 5 (Quality) can launch with the empty state message ("No quality data") if QualityService is not ready. Tab 3 (Iterations) can omit the detail expander in the first pass. This keeps all 5 tabs present while reducing scope for the initial delivery.

**Rebuttal to D→C-2**: The risk is not total lines of code but the ratio of new patterns to reused patterns. Phase 1 was predominantly Go stdlib + Cobra (familiar). Phase 2 introduces Bubble Tea + Lip Gloss + Bubbles + Glamour (new ecosystem). Every component in the TUI requires learning and validating new framework patterns. The Phase 1 developer had Go expertise; Phase 2 requires Charm ecosystem expertise.

### Researcher rebuttals

**Rebuttal to C→AB-1**: The Researcher examined BP-05 and BP-08 for scope alignment. **Finding**: BP-08 Section 4 acceptance criteria are a subset of BP-05 features. BP-05 includes features like the Watch TUI (Section 7) and Orchestration TUI (Section 8) that are explicitly Phase 3-4. However, within the 5-tab dashboard, BP-05's detail level (monochrome fallback, 3 responsive breakpoints, context-sensitive help overlay) goes beyond the BP-08 acceptance criteria. The acceptance criteria do not mention monochrome fallback or responsive breakpoints. The Critic's observation is validated.

---

## Phase 5: Convergence Synthesis

### Areas of Agreement (High Confidence)

1. **Per-tab delegated model architecture.** All four personas agree this is both the correct pattern and the specified approach (BP-05 Section 3). No dissent.

2. **Service integration through existing interfaces.** The TUI accesses data through the same service interfaces as the CLI. No new service methods needed (except potentially for the TUI command wiring). All personas agree.

3. **Watch TUI and Orchestration TUI are out of scope for Phase 2.** BP-08 and BP-05 clearly place these in Phases 3-4. All personas agree.

4. **Some SHOULD items must be fixed before TUI work.** The Architect and Pragmatist converged on: S-1, S-2, wiring centralization, and `docs search` abstraction must be fixed first. S-3, GoDoc gaps, and flag rename are deferrable.

### Areas of Disagreement (Requires Decision)

1. **S-3 (transitive reason strings) timing.** The Architect argues this affects Status tab quality; the Pragmatist says generic reasons are acceptable for MVP. The Critic notes this is a minor issue relative to scope risk.
   - **Resolution**: Defer S-3 to Phase 2 implementation. Fix during Status tab development if straightforward; otherwise defer to Phase 2 polish.

2. **BP-05 scope vs BP-08 acceptance criteria.** The Critic and Researcher identified that BP-05 specifies features beyond BP-08 acceptance criteria (monochrome fallback, responsive breakpoints, help overlay). The Architect treats BP-05 as the full specification; the Critic advocates MVP per tab.
   - **Resolution**: Implement to BP-08 acceptance criteria as the Phase 2 gate, with BP-05 as the target specification. Features in BP-05 but not in BP-08 acceptance criteria (monochrome fallback, 3-tier responsive layout) are SHOULD-level enhancements within Phase 2, not blocking requirements.

3. **Testing strategy for View() functions.** The Pragmatist says skip them; the Critic says the coverage gap is real; the Researcher proposes `teatest` golden files.
   - **Resolution**: Use `teatest` for model state testing and a small set of golden file tests for the 5 main tab views at standard (80x24) size. Do not test individual component View() output. Accept that visual correctness requires manual verification.

4. **Quality tab readiness.** The Critic flagged QualityService as potentially unimplemented. The Researcher confirmed QualityScore domain types exist but service implementation status is uncertain.
   - **Resolution**: Implement Quality tab with an empty state handler. If QualityService.ReadLog() is not ready, the tab shows "No quality data" (matching BP-05 empty state spec). Implement QualityService during Phase 2 if needed -- it is a straightforward CRUD service reading a YAML file.

### Decision Matrix

| Criterion | Weight | Option A: Fix-All-First + Full-BP-05 | Option B: Cherry-Pick + Tab-by-Tab | Option C: Cherry-Pick + MVP-per-Tab |
|-----------|--------|--------------------------------------|-------------------------------------|--------------------------------------|
| Specification compliance | 0.25 | 5.0 -- Full BP-05 compliance | 4.0 -- Full BP-05 but SHOULD items mixed in | 3.5 -- BP-08 criteria met, BP-05 partial |
| Risk management | 0.20 | 3.5 -- Low arch risk, high scope risk (full spec) | 3.0 -- Medium arch risk (deferred fixes), medium scope risk | 4.5 -- Low arch risk (key fixes done), low scope risk (MVP) |
| Time-to-value | 0.20 | 2.5 -- Delayed by full SHOULD sprint | 4.5 -- Fast increments, usable after Tab 1 | 4.0 -- Fast increments with lighter tabs |
| Architectural cleanliness | 0.15 | 5.0 -- All tech debt addressed | 3.0 -- Key debt fixed, rest deferred | 4.0 -- Key debt fixed, MVP reduces new debt surface |
| Testing confidence | 0.10 | 4.0 -- Full test surface | 3.0 -- State tests only | 4.0 -- State + golden file for key views |
| Scope predictability | 0.10 | 2.5 -- Full BP-05 is ambitious | 3.0 -- Full BP-05 tab-by-tab | 4.5 -- Reduced scope per tab |

**Scores**:
- Option A: (5.0 * 0.25) + (3.5 * 0.20) + (2.5 * 0.20) + (5.0 * 0.15) + (4.0 * 0.10) + (2.5 * 0.10) = 1.25 + 0.70 + 0.50 + 0.75 + 0.40 + 0.25 = **3.85**
- Option B: (4.0 * 0.25) + (3.0 * 0.20) + (4.5 * 0.20) + (3.0 * 0.15) + (3.0 * 0.10) + (3.0 * 0.10) = 1.00 + 0.60 + 0.90 + 0.45 + 0.30 + 0.30 = **3.55**
- Option C: (3.5 * 0.25) + (4.5 * 0.20) + (4.0 * 0.20) + (4.0 * 0.15) + (4.0 * 0.10) + (4.5 * 0.10) = 0.875 + 0.90 + 0.80 + 0.60 + 0.40 + 0.45 = **4.03**

**Winner: Option C -- Cherry-Pick SHOULD Fixes + MVP-per-Tab Implementation** (4.03/5.00)

### Recommendations

#### R1: Fix 4 critical SHOULD items before TUI implementation (Confidence: HIGH, 90%)

Fix these items in a focused pre-Phase-2 commit sequence:
1. **S-1**: Add `--check`/`--force` mutual exclusion guard in `cmd/reconcile.go` (~5 lines)
2. **S-2**: Add missing documents check to ReconcileSuite in `internal/validate/reconcile.go` (~15 lines)
3. **Wiring centralization**: Complete the `buildDeps()` pattern in `main.go`, callable from both PersistentPreRunE and TUI initialization
4. **`docs search` abstraction**: Refactor `cmd/docs.go:199-248` to use DocRepo instead of direct `filepath.WalkDir`

Defer: S-3 (transitive reasons), `--project` rename, GoDoc gaps. These do not affect TUI functionality or architecture.

*Falsifiability*: If any single fix requires changes to more than 8 files, it has become a refactoring task and should be time-boxed to 2 hours.

#### R2: Implement TUI tab-by-tab with MVP scope per tab (Confidence: HIGH, 85%)

Follow this implementation sequence, where each step produces a compilable, runnable increment:
1. **Foundation**: `tui/styles.go` (theme), `tui/keys.go` (keybindings)
2. **App shell**: `tui/app.go` (chrome, tab bar, delegation, service injection)
3. **Tab 1 -- Status**: Zone health bars, staleness panel, warnings, workflow state. Two-column at >=80, single-column below.
4. **Tab 2 -- Docs**: Document list with zone filter and status indicators. Defer preview pane to later step.
5. **Tab 3 -- Iterations**: Table with type filter. Defer detail expander to later step.
6. **Tab 4 -- Checks**: Accordion with lazy loading. Full implementation per BP-05.
7. **Tab 5 -- Quality**: Empty state message. Implement chart only if QualityService is ready.
8. **Polish**: Help overlay, preview pane (Docs), detail expander (Iterations), Quality chart, responsive breakpoints.
9. **Wiring**: `cmd/tui_cmd.go`, integration with `buildDeps()`

MVP exit criteria per BP-08: all 5 tabs render at 80x24, tab switching works, data loads from services, quit works cleanly. Full BP-05 features are target, not gate requirement.

*Falsifiability*: If Steps 1-6 exceed 2500 lines of code, reassess scope for Steps 7-8. If any individual tab exceeds 500 lines, evaluate decomposition into sub-components.

#### R3: Use Bubbles components + teatest for testing (Confidence: MEDIUM, 75%)

- Use `bubbles/table`, `bubbles/list`, `bubbles/viewport`, `bubbles/spinner`, `bubbles/textinput` for standard components.
- Use `glamour` for Docs tab preview pane (deferred to polish step).
- Evaluate `asciigraph` for Quality tab chart; if chart characters do not match BP-05 spec, build a custom 150-200 line chart component.
- Testing: `teatest` for model state assertions + golden file tests for the 5 main tab views at 80x24. Target: 1 golden file test per tab view + 3-5 state transition tests per tab model.

*Falsifiability*: If `teatest` is incompatible with the current Bubble Tea version or introduces excessive test brittleness (>50% golden file updates per styling change), fall back to state-only testing.

#### R4: Implement QualityService during Phase 2 if needed (Confidence: MEDIUM, 70%)

- `QualityService.ReadLog()` reads `quality-log.yml` and returns `[]QualityEntry`.
- `QualityRepo` needs a filesystem implementation that parses the YAML file.
- This is straightforward CRUD -- no complex business logic.
- If QualityService is already implemented, Tab 5 gets the full chart experience.
- If not, implement it as part of the Quality tab step, or launch Tab 5 with the empty state.

*Falsifiability*: If `quality-log.yml` format is not yet defined in BP-03, the Quality tab must launch with empty state and QualityService implementation defers to Phase 3.

#### R5: Preserve BP-05 as the complete specification; do not modify it (Confidence: HIGH, 80%)

BP-05 is a detailed specification that serves Phases 2 through 4 (Watch TUI and Orchestration TUI are in the same document). Phase 2 implements the 5-tab dashboard section of BP-05 to MVP level, with full BP-05 compliance as a polish target. Features beyond BP-08 acceptance criteria (monochrome fallback, 3-tier responsive, context-sensitive help) are SHOULD-level within Phase 2 and can be deferred to polish or Phase 3 without failing the Phase 2 gate.

*Falsifiability*: If the reviewer considers monochrome fallback or responsive design as MUST for Phase 2 acceptance (citing accessibility or minimum quality standards), elevate those features to the MVP scope.

---

## Quality Rubric Assessment

| Dimension | Score | Justification |
|-----------|-------|---------------|
| Perspective Diversity | 4 | Genuine tension between Critic (scope risk) and Architect+Pragmatist (implementation approach). Architect-Pragmatist overlap on architecture dilutes total diversity. Researcher amplifies but does not introduce independent position. |
| Evidence Quality | 3 | Mix of codebase references (Very High), Charm ecosystem documentation (Medium -- single-source), and comparative estimates (Low). No multi-source empirical claims. Evidence audit: 6 of 20 evidence items are estimates or unsourced assertions, all flagged. |
| Concession Depth | 4 | Both Architect and Pragmatist made substantive concessions. Architect dropped GoDoc and flag rename requirements. Pragmatist accepted wiring centralization. Critic conceded on Phase 2a/2b split. |
| Challenge Substantiveness | 4 | Challenges identified real weaknesses: dual wiring points (A→B), BP-05/BP-08 scope gap (C→AB), QualityService readiness (C→AB), acceptance criteria conflict with MVP (D→C). Counter-evidence provided in most cases. |
| Synthesis Quality | 4 | Decision matrix derived from context (spec compliance, risk, time-to-value). MVP-per-tab recommendation synthesizes Pragmatist's incremental approach with Critic's scope caution. Emergent insight: BP-05 as target spec vs BP-08 as gate criteria. |
| Actionability | 5 | 5 phased recommendations with falsifiability conditions, specific file references, line count estimates, and clear MVP exit criteria. Implementation sequence maps directly to development steps. |

**Overall Quality Score: 4.00 / 5.00**

Gate 0 requirement (>= 3.0): **PASS**

---

## Convergence Diff

**Compared against**: `docs/knowledge/reconciliation-engine-convergence.md` (Phase 1.5)

| Dimension | Phase 1.5 Score | Phase 2 Score | Delta | Notes |
|-----------|----------------|---------------|-------|-------|
| Perspective Diversity | 3 | 4 | +1 | 4 personas vs 3; added Researcher for empirical grounding |
| Evidence Quality | 3 | 3 | 0 | Similar mix of codebase refs + single-source external claims |
| Concession Depth | 5 | 4 | -1 | Phase 1.5 had a fundamental position revision (Architect dropped tech-debt-first). Phase 2 concessions are meaningful but narrower. |
| Challenge Substantiveness | 4 | 4 | 0 | Comparable challenge depth in both analyses |
| Synthesis Quality | 5 | 4 | -1 | Phase 1.5 produced a 12-step implementation sequence with specific file mappings. Phase 2 produces a 9-step sequence with MVP scoping. Slightly less specific due to TUI's presentation-layer focus. |
| Actionability | 5 | 5 | 0 | Both produce phased plans with falsifiability conditions |
| **Overall** | **4.33** | **4.00** | **-0.33** | Expected: Phase 2 is a presentation layer with less architectural novelty to analyze |

**Key differences from Phase 1.5 convergence**:
1. **Risk profile shifted.** Phase 1.5's dominant risk was blueprint inconsistencies (structural). Phase 2's dominant risk is scope (volume of UI specification vs implementation capacity).
2. **SHOULD fix strategy evolved.** Phase 1.5 recommended deferring all tech debt past the phase. Phase 2 selectively fixes 4 items that directly impact TUI quality while deferring 3 that do not.
3. **Testing strategy is new territory.** Phase 1.5 relied on the project's established testing patterns (table-driven, golden files for CLI output). Phase 2 introduces `teatest` and TUI-specific testing patterns that have no prior art in this codebase.
4. **Decision matrix winner pattern.** Phase 1.5: Option C "Blueprint-first + Test-driven" (4.20). Phase 2: Option C "Cherry-Pick + MVP-per-Tab" (4.03). Both winning options balance specification compliance with pragmatic risk management.

---

## Evidence Audit

All evidence items categorized by confidence level:

| Confidence | Count | Items |
|------------|-------|-------|
| Very High (>= 2 independent sources) | 5 | A2 (BP-05 spec), A3 (BP-01 + BP-08), B4 (validation.md gate), B5 (BP-08 + BP-05), C1 (BP-05 tree) |
| High (>= 2 sources, any type) | 5 | A4 (validation + codebase), B1 (bubbles README), C3 (bubbles README), C4 (BP-08 + validation.md), C5 (go.mod) |
| Medium (single source) | 8 | A1, B2, B3, D1, D2, D3, D4, D6 |
| Low (estimate/unsourced) | 2 | C2 (line count estimate), D5 (Glow line estimate) |

**Flagged assertions**: C2 and D5 are estimates marked as Low confidence. No unflagged unsourced assertions remain.
