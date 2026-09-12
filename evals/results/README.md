# Committed outcome-benchmark results

Store every attempted outcome-benchmark trial, including failures, under:

```text
evals/results/<benchmark>/<agent>/<YYYY-MM-DD>-<condition>-<trial>.json
evals/results/<benchmark>/<agent>/<YYYY-MM-DD>-<condition>-<trial>-artifact.<ext>
```

Compact treatment records include their preflight profile so matched full and compact runs cannot collide:

```text
evals/results/<benchmark>/<agent>/<YYYY-MM-DD>-treatment-compact-<trial>.json
evals/results/<benchmark>/<agent>/<YYYY-MM-DD>-treatment-compact-<trial>-artifact.<ext>
```

Control, full-profile treatment, and treatment results created by older harness versions retain the original form.

Use `repo-knowledge-eval grade --results evals/results` to write the pair with exclusive-create semantics. The command refuses to replace an existing result or artifact. The preserved artifact must be a regular file containing the generated patch or output archive.

Result JSON records the benchmark and source revisions, condition, agent host version, model version, reasoning configuration, Repository Knowledge version or commit, trial number, treatment preflight profile, content-free hook payload bytes/characters/generation time, duration, optional provider-attributed input/output/cached/total token usage, preserved artifact, deterministic checks, and the independently supplied semantic score and reviewer. Payload bytes and characters are never described as tokens. A deterministic pass remains `pending_semantic_review` unless a reviewer explicitly supplies semantic status, score, and identity.

Commit failed trials as well as successful trials. Never edit a committed trial in place; correct metadata by adding a new trial number and explain the superseded record in the review or commit message.
