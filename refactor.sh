#!/bin/bash

# Coding standards to follow
CODING_STANDARDS="
## Architecture Design

Modularity & Composition:
- Break systems into independent modules.
- Group related functionality together.
- Compose components to build features.
- Design components that can be tested in isolation.
- Avoid tight coupling between modules.

YAGNI (You Aren't Gonna Need It):
- Never write code for future requirements.
- Build only what is needed now.

DRY (Don't Repeat Yourself):
- Eliminate code duplication.
- Extract common logic into reusable components.

## Code Details

Simplicity:
- Prefer simple over clever.
- Avoid premature abstraction.
- Keep functions small and focused.
- Return early - reduce nesting.

Input Validation:
- Validate at system boundaries (API endpoints, user input).
- Fail fast with clear error messages.

Error Handling:
- Handle errors at boundaries (API, external calls).
- Fail fast with clear messages, return immediately.
- Provide context (what failed, why).
- Never swallow errors silently.
- Log with sufficient detail.

Comments:
- Avoid comments - make code self-explanatory instead.
- Write comments only for non-obvious business rules or explaining WHY decisions were made.
- Comments become outdated as code changes and mislead developers. Delete wrong or outdated comments immediately.
"

SYSTEM_PROMPT_FOR_REFACTOR="Keep ALL existing software features when refactoring. Refactor based on the coding standards:
$CODING_STANDARDS"


COMPARE_PROMPT="compare these two git branches and determine which one is worse"

COMPARE_SYSTEM_PROMPT="You are a code comparison agent. Your task is to compare two implementations and determine which one is worse. You MUST output ONLY the branch name of the worse implementation, nothing else. No explanations, no reasoning, no additional text - just the branch name."

# STEP 1: Generate refactor plan (3 iterations)
PLAN_INITIAL_PROMPT="Analyze the pkg/display/ package and create a detailed refactor plan for it. 
Keep ALL existing features. Save the plan to a file: 'pkg/display/refactor_plan.md'"

PLAN_IMPROVE_PROMPT="Refine and improve the refactor plan ('pkg/display/refactor_plan.md')"

echo "=== STEP 1: Generating refactor plan (3 iterations) ==="
agent-exec evolve "$PLAN_INITIAL_PROMPT" \
  --system-prompt "$SYSTEM_PROMPT_FOR_REFACTOR"\
  -i "$PLAN_IMPROVE_PROMPT" \
  --improve-system-prompt "$SYSTEM_PROMPT_FOR_REFACTOR"\
  -c "$COMPARE_PROMPT" \
  --compare-system-prompt "$COMPARE_SYSTEM_PROMPT" \
  -n 3

# STEP 2: Implement the refactor (2 iterations)
IMPL_INITIAL_PROMPT="Read the refactor plan from 'pkg/display/refactor_plan.md' and implement it for the 'pkg/display/' package."

IMPL_IMPROVE_PROMPT="According to 'pkg/display/refactor_plan.md', improve the refactored code while maintaining all existing features."

echo ""
echo "=== STEP 2: Implementing refactor (2 iterations) ==="
agent-exec evolve "$IMPL_INITIAL_PROMPT" \
  --system-prompt "$SYSTEM_PROMPT_FOR_REFACTOR"\
  -i "$IMPL_IMPROVE_PROMPT" \
  --improve-system-prompt "$SYSTEM_PROMPT_FOR_REFACTOR"\
  -c "$COMPARE_PROMPT" \
  --compare-system-prompt "$COMPARE_SYSTEM_PROMPT" \
  -n 2
