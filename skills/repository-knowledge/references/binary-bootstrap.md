# Missing binary bootstrap

Use this procedure only when the installed skill is available but `repo-knowledge` is not executable by name and the requested work needs a deterministic CLI operation. Documentation reasoning and source inspection do not require the binary by themselves.

## Resolve a pinned release

1. Check `command -v repo-knowledge` on macOS/Linux or `Get-Command repo-knowledge` on Windows.
2. If the command is missing, read `.repo-knowledge/toolkit.json` from the repository root.
3. Use its exact semantic `ref`, such as `v0.11.1`. Never substitute `latest`, a branch, or an unpinned ref.
4. Accept `https://github.com/rustedzone/repository-knowledge` as the canonical source. Treat the legacy source value `release-binary` as this canonical repository. If a manifest names a different source, explain it and obtain explicit confirmation before downloading from that source; the bundled bootstrap scripts accept only an `OWNER/REPO` GitHub identity.
5. When working on the toolkit repository itself, use the root `VERSION` file and its canonical GitHub source. If neither a valid manifest ref nor the toolkit `VERSION` file exists, stop and ask the user which exact release to trust.

The manifest is repository configuration, not proof that a remote artifact is safe. The bootstrapper also verifies the selected artifact against the release's `SHA256SUMS` and verifies its GitHub artifact attestation when a compatible `gh` command is available.

## Preview and request permission

Locate the script under this skill's `scripts/` directory. Preview macOS/Linux resolution without network access or writes:

```bash
sh scripts/install-binary.sh \
  --version v0.11.1 \
  --repository rustedzone/repository-knowledge \
  --dry-run
```

On Windows, preview the default user-local destination and user PATH change:

```powershell
& scripts/install-binary.ps1 `
  -Version v0.11.1 `
  -Repository rustedzone/repository-knowledge `
  -AddToPath `
  -DryRun
```

Before any download, replacement, user-directory write, or PATH change, tell the user:

- that `repo-knowledge` is missing and why the current task needs it;
- the exact source repository, pinned tag, platform artifact, and destination;
- whether an existing destination will be replaced;
- that checksum verification is mandatory and attestation verification runs when supported;
- whether the Windows user PATH will change; and
- that the installer does not use `sudo` or edit Unix shell startup files.

Ask for explicit permission. Do not interpret the original repository task as installation permission, and do not run the bootstrapper with an affirmative flag until permission is given.

## Install after approval

On macOS or Linux, the default is `~/.local/bin` or `~/bin`, but only when the chosen directory is already on `PATH`:

```bash
sh scripts/install-binary.sh \
  --version v0.11.1 \
  --repository rustedzone/repository-knowledge \
  --yes
```

The script refuses to edit shell profiles. If neither default directory is on `PATH`, ask the user whether to use another user-owned directory already on `PATH` or let them configure PATH themselves. Never fall back to `sudo` or a system directory without a separate explicit request.

On Windows, install to `%LOCALAPPDATA%\Programs\repo-knowledge\bin` and add that directory to the current user's PATH only when the permission request explicitly included the PATH change:

```powershell
& scripts/install-binary.ps1 `
  -Version v0.11.1 `
  -Repository rustedzone/repository-knowledge `
  -AddToPath `
  -Yes
```

The helper executes the verified download before replacing the destination and requires its reported version to match the pinned manifest ref. After installation, resolve the command by name and run `repo-knowledge --version`. On Windows, a newly persisted user PATH may require the host terminal or agent to restart; until then, use the helper's version result and the exact installed path as verification. If verification fails, stop and report the exact failure; do not use or install the downloaded artifact.

If the user declines installation, continue without the CLI where the requested result can still be produced safely. Clearly identify any deterministic operation that could not be completed.
