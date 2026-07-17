# Opt-in Kilo `/plan-task` command

Status: proposed

## Goal

Add a global, opt-in Kilo command that turns an objective into a durable,
granular execution backlog. It must not replace Kilo's normal coordinator or
change default agent routing.

## Invocation and boundary

`/plan-task <objective>` performs planning only. It reads repository authority,
current state, and relevant source, then writes durable task records. It ends
by presenting the ordered backlog and waits for the user to say `go`.

After `go`, the normal Kilo coordinator executes the selected active task:

1. assign the bounded implementation lane to an executor;
2. obtain an independent read-only review;
3. assign a fixer only for a named review finding;
4. obtain reviewer re-check evidence when a fix was made;
5. record verification and durable-document triage, then set task status.

The command contains no ADK-specific routing. ADK graph work remains a
separate command or workflow.

## Parallel implementation dispatch

The coordinator determines whether implementation subagents can run in
parallel. It inspects the granular task records, owned files, package
boundaries, dependency inputs, and verification commands. Parallel dispatch is
allowed only when all lanes have no overlapping owned files, no
producer-consumer dependency, and no unresolved shared contract.

If any condition is not met, the coordinator sequences the lanes and does not
dispatch them concurrently. It records the resulting `parallel` or
`sequential` decision and short rationale in the parent task for auditability.
Each subagent returns its own completion report; the coordinator checks the
reports and changed-file sets before starting independent review. The reviewer
then reviews the integrated result, not merely the individual handoffs.

## Granular task contract

The planner must split work at independently auditable boundaries. It may not
create an umbrella task whose acceptance target is a feature composed of
multiple unrelated changes.

Every task record must include:

- an ordered identifier and a one-sentence objective;
- dependencies and the condition that unlocks the task;
- owned files or package boundaries, plus explicit non-goals;
- an implementation-sized acceptance target;
- the coordinator's parallel-dispatch decision and short rationale;
- focused verification commands and expected evidence;
- required review scope;
- a completion report with changed files, commands/results, concerns, and
  external-mutation status;
- durable-document triage targets and reason.

Examples of separate tasks: add a migration, add repository behavior, add
service behavior, add HTTP boundary and contract regeneration, add focused
tests, or revise operations documentation. Merge tasks only when they are too
small to verify independently and share the same files and acceptance check.

## Durable locations

The command uses the repository's existing `docs/tasks/` convention when it
exists. If absent, it creates `docs/tasks/` before writing the first task. The
task record is the source of truth for status and receipts; chat and agent
handoffs are supporting evidence only.

## Agent roles

- Planner: creates the ordered granular tasks; does not edit implementation.
- Coordinator: chooses the next ready task, dispatches roles, and is the only
  role allowed to update task status or durable documentation receipts.
- Executor: changes only the active task lane and runs its declared checks.
- Reviewer: read-only, independent, and evidence-based.
- Fixer: changes only one named reviewer finding and reruns its affected check.

## Completion rules

A task is `DONE` only after declared checks, independent review, any required
review re-check, and document triage are recorded. Otherwise it is `BLOCKED`
or `NEEDS_CONTEXT` with the exact missing condition.
