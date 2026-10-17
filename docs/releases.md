# Releases

Release Please prepares draft release PRs on `master`. It owns version selection,
the version manifest, tags and GitHub Releases. Git-cliff fills the proposed
CHANGELOG.md entry from the full Git ancestry range before review, so imported
upstream features and fixes appear individually even when their commits are old.
It preserves the proposed version, release heading and all historical entries. This Go
library publishes no executable archives. Semantic-release and GoReleaser are
retired; do not enable a second publisher.

## Local checks

```sh
actionlint .github/workflows/release-please.yml
bash -n scripts/release-notes.sh
go test -race ./...
go vet ./...
```

Validate `release-please-config.json` against the schema linked in its `$schema`
field. Run `prek` on staged files only. Configuration checks do not prove that
GitHub credentials, required checks or release events work.

Historical release headings are normalized to H2 under a Changelog title so
Release Please prepends new entries correctly. Existing release text, links,
dates and versions are preserved.

The initial manifest records the existing v1.3.0 release. The bootstrap commit is
that release's full SHA. Do not choose the next version manually or replace fork
tags with upstream tags. Conventional Commits follow strict SemVer: compatible
features bump minor, fixes bump patch, and incompatible public API changes require
major. Finish compatible integration on v1 before a separate reviewed v2 change.

## Activation requires approval

Work sequentially in the existing checkout on local `master`; local commits
there are authorized. Do not create source PRs, feature branches or parallel
feature worktrees. The user performs every push of `master`; the assistant never
pushes and asks the user when a push is needed. Changes to secrets/settings,
manual remote workflow runs and publication require separate approval. The upstream merge waits until release automation is activated and a
draft release PR with its required checks is verified.

With approval, install the workflow on `master`. The `Release` environment must
provide the repository-scoped `GH_TOKEN` secret or an approved GitHub App token
with equivalent rights. Never copy credentials from another repository. Default
workflow tokens may prevent generated PRs and releases from triggering downstream
workflows; verify actual PR checks and release-event delivery. The workflow fails
without printing the token when that secret is absent.

Review the draft release PR's proposed version, changelog and CI results. Do not
merge it to test the setup. Creating a draft PR is not publication approval.

## Publication and notes

Only merge the release PR after explicit publication approval. Verify exactly one
fork tag and GitHub Release, unchanged historical changelog entries and successful
post-release notes.

The notes job checks out the release event's exact tag. `scripts/release-notes.sh`
uses its previous first-parent fork release tag and an ancestry range, so older
upstream commits merged into this release are included. It fails if the checkout
or previous release is wrong instead of guessing a latest release. Full fork
history and tags are required; fetch upstream source with `--no-tags`.

Git-cliff edits only that release's body, not CHANGELOG.md or tags. Commit links
point to commits retained in this fork and author names remain visible. Issue
numbers stay escaped text, and automatic issue/PR release comments are deliberately
omitted because imported numbers may refer to unrelated fork items.

To preview notes without publishing, check out the release tag in an isolated
worktree with full history and run:

```sh
bash scripts/release-notes.sh vX.Y.Z > /tmp/logseq-go-release-notes.md
```

This runs Git and git-cliff locally; it does not call GitHub. Do not assume other
release tools' dry-run flags prevent API writes.

## Rollback

Use reviewed corrective commits on master rather than discarding local work or
rewriting history. Abort an unfinished merge only after verifying that no unrelated
edits will be lost. Never restore the retired publisher alongside Release
Please, move an existing release tag or publish a version just to test rollback.
