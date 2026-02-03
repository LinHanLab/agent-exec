#!/bin/bash

# Script to create a snake game using golang with agent-exec evolve

# Clean up and create test project directory
echo "Setting up test project directory..."

rm -rf test-project
mkdir -p test-project
cd test-project

# Initialize git repository
echo "Initializing git repository..."
git init
git config user.name "Test User"
git config user.email "test@example.com"

# Create initial commit
echo "Creating initial commit..."
echo "# Snake Game" > README.md
git add --all
git commit -m "Initial commit"

echo "Starting evolution process..."

# Define prompts as variables for readability
INITIAL_PROMPT="Create a snake game in golang with the following features:
- Terminal-based UI using a TUI library
- Snake movement controlled by arrow keys
- Food spawning at random positions
- Score tracking
- Game over detection when snake hits walls or itself
- Smooth gameplay with configurable speed"

IMPROVE_PROMPT="improve game performance, code quality, and user experience"

COMPARE_PROMPT="compare these two implementations and determine which has worse performance, code quality, or gameplay experience"

COMPARE_SYSTEM_PROMPT="You are a code comparison agent. Your task is to compare two implementations and determine which one is worse. You MUST output ONLY the branch name of the worse implementation, nothing else. No explanations, no reasoning, no additional text - just the branch name."

ITERATIONS=2

# Run agent-exec evolve command
agent-exec evolve "$INITIAL_PROMPT" \
  -i "$IMPROVE_PROMPT" \
  -c "$COMPARE_PROMPT" \
  --compare-system-prompt "$COMPARE_SYSTEM_PROMPT" \
  -n "$ITERATIONS"
