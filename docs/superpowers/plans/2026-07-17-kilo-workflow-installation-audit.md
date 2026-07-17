# Kilo workflow installation audit

Status: complete

## Task

Explain where Kilo workflows belong and identify the active workflow surfaces in
this repository.

## Evidence

- `kilo.json` is the project-level Kilo configuration. It selects the provider,
  model, and project directory permissions.
- `.kilo/agent/developer.md` is the repository-local developer workflow that
  Kilo should apply while working in this project.
- The global Kilo configuration is installed at `~/.config/kilo`: its shared
  workflow is `global_workflow.md`, shared guardrails are `AGENTS.md`, reusable
  agents are under `agents/`, and slash commands are under `commands/`.
- Globally discoverable skills live under
  `~/.kilo/skills/<skill-name>/SKILL.md`.

## Outcome

Use `.kilo/` and `kilo.json` for behavior that is specific to this repository.
Use `~/.config/kilo/` for workflows and agents that should follow the user into
every Kilo project. Reload Kilo or restart VS Code after changing global files.
