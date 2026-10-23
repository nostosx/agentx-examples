---
name: commit-message
description: Write a Conventional Commits message for a described change. Use when the user asks for a commit message.
allowed-tools: []
---

# Commit messages

1. Pick a type: feat, fix, docs, refactor, test, chore or perf.
2. Add an optional scope in parentheses naming the affected package.
3. Write a subject in the imperative mood, at most 72 characters, without a
   trailing period.
4. If the change needs explanation, add a body after a blank line that says
   why the change was made, wrapped at 72 characters.

Example:

```
fix(runtime): keep tool results in call order

Concurrent read-only calls finished in arbitrary order, which changed
the prompt between runs and defeated prompt caching.
```
