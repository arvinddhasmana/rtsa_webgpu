<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA — GitHub Copilot Custom AI Agents

This directory contains the **custom AI Agent profiles** for the RTSA project, configured for the GitHub Copilot agent framework.

---

## Available Agents

### 1. Greatest Ever Developer

| Property | Value |
|---|---|
| **File** | [greatest-ever-developer.prompt.md](greatest-ever-developer.prompt.md) |
| **Mode** | `agent` |
| **Scope** | End-to-end feature implementation |

**Purpose:** Performs the complete development lifecycle for an assigned feature:

- Reads SDLC guidelines and architecture documentation
- Performs impact analysis before writing any code
- Implements the feature following all RTSA coding standards
- Generates unit, integration, and E2E tests
- Runs an optimized validation cycle (fix individual issues before full test runs)
- Creates a PR against `main` with full documentation

**How to invoke in Copilot Chat:**

1. Open the Copilot Chat panel (`Ctrl+Alt+I`)
2. Click the agent selector (paperclip / `@`) and choose `greatest-ever-developer`
3. Provide the feature description or paste the issue/use-case reference:

   ```
   @greatest-ever-developer Implement UC009 anomaly detection alert threshold configuration.
   Issue: #42
   ```

---

### 2. Meanest Ever Reviewer

| Property | Value |
|---|---|
| **File** | [meanest-ever-reviewer.prompt.md](meanest-ever-reviewer.prompt.md) |
| **Mode** | `agent` |
| **Scope** | PR code review and merge/block decision |

**Purpose:** Performs exhaustive code review on a PR:

- Loads all relevant SDLC guidelines and security controls
- Runs automated build, vet, and test checks
- Reviews for security violations, classification compliance, architecture conformance, and code quality
- Checks test completeness (unit, integration, E2E)
- Either **merges** (if all checks pass) or **blocks with detailed actionable comments**
- Posts `Handover to Human` if merge conflicts cannot be auto-resolved

**How to invoke in Copilot Chat:**

1. Open the Copilot Chat panel (`Ctrl+Alt+I`)
2. Click the agent selector and choose `meanest-ever-reviewer`
3. Provide the PR number or URL:

   ```
   @meanest-ever-reviewer Review PR #57 for UC009 anomaly detection feature.
   ```

---

## Agent Interaction Flow

```
Developer / Issue Assignment
         │
         ▼
┌─────────────────────────┐
│   Greatest Ever         │
│   Developer             │
│                         │
│ 1. Load SDLC guidelines │
│ 2. Impact analysis      │
│ 3. Create feature branch│
│ 4. Implement feature    │
│ 5. Generate tests       │
│    (unit/integ/e2e)     │
│ 6. Validation cycle     │
│ 7. Raise PR → main      │
└────────────┬────────────┘
             │  PR Created
             ▼
┌─────────────────────────┐
│   Meanest Ever          │
│   Reviewer              │
│                         │
│ 1. Load SDLC guidelines │
│ 2. Run CI checks        │
│ 3. Security review      │
│ 4. Architecture review  │
│ 5. Code quality review  │
│ 6. Test quality review  │
│ 7. Audit trail check    │
│                         │
│  ┌─────────┬──────────┐ │
│  │ PASS    │  FAIL    │ │
│  │ Merge   │ Block +  │ │
│  │ squash  │ Comments │ │
│  └─────────┴──────────┘ │
│  Merge conflict?        │
│  → "Handover to Human"  │
└─────────────────────────┘
```

---

## Configuration

### VS Code Settings

Agent mode and prompt files are enabled in [.vscode/settings.json](../.vscode/settings.json):

```json
{
  "github.copilot.chat.experimental.agentDesktop": true,
  "github.copilot.chat.experimental.prompt-files": true,
  "github.copilot.chat.agent.enabled": true,
  "chat.agent.enableTerminalTool": true
}
```

### Repository-Level Instructions

Both agents inherit the global RTSA policy from [.github/copilot-instructions.md](../copilot-instructions.md), which is automatically loaded by GitHub Copilot for every interaction.

---

## Security Notes

- Both agents operate exclusively on UNCLASSIFIED code artifacts
- Neither agent will commit classified data, credentials, or PII — any such output is **self-rejected**
- All agent actions that modify the repository (commits, PRs, reviews, merges) must be auditable via the standard Git history
- Agent-assisted PRs are still subject to human review per team policy (at minimum the Meanest Ever Reviewer agent acts as first-pass gate)

---

## Copilot Framework Compatibility

| Feature | Required Version |
|---|---|
| GitHub Copilot Chat | ≥ 0.22 |
| VS Code | ≥ 1.97 |
| GitHub Copilot Extension | Latest stable |
| Prompt Files Feature Flag | `github.copilot.chat.experimental.prompt-files: true` |
| Agent Mode | `github.copilot.chat.agent.enabled: true` |
