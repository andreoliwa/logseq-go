const { execFileSync } = require("node:child_process");

function git(...args) {
    return execFileSync("git", args, { encoding: "utf8" });
}

function renderChangelog(baseSha, version, proposed) {
    if (!/^\d+\.\d+\.\d+(?:-[\w.-]+)?$/.test(version)) {
        throw new Error("invalid proposed release version");
    }
    const manifest = JSON.parse(
        git("show", `${baseSha}:.release-please-manifest.json`),
    );
    const previousTag = `v${manifest["."]}`;
    const nearestTag = git(
        "describe",
        "--tags",
        "--abbrev=0",
        "--first-parent",
        "--match",
        "v[0-9]*",
        baseSha,
    ).trim();
    if (previousTag !== nearestTag) {
        throw new Error(
            "manifest does not match the previous fork release tag",
        );
    }
    git("merge-base", "--is-ancestor", previousTag, baseSha);

    const historical = git("show", `${baseSha}:CHANGELOG.md`);
    const headings = [...proposed.matchAll(/^## \[([^\]]+)\][^\n]*\n/gm)];
    const oldHeading = /^## \[/m.exec(historical);
    if (
        !oldHeading ||
        headings.length < 2 ||
        headings[0][1] !== version ||
        headings[1][1] !== manifest["."] ||
        proposed.slice(headings[1].index) !==
            historical.slice(oldHeading.index) ||
        proposed.slice(0, headings[0].index) !==
            historical.slice(0, oldHeading.index)
    ) {
        throw new Error(
            "proposed release changed historical changelog entries or headers",
        );
    }

    // The explicit ancestry range includes old commits merged after the release.
    // Tags on side branches must not split this into several release sections.
    const notes = execFileSync(
        "git-cliff",
        [
            "--config",
            "cliff.toml",
            "--strip",
            "header",
            "--ignore-tags",
            ".*",
            `${previousTag}..${baseSha}`,
        ],
        { encoding: "utf8" },
    )
        .replace(/^## What's Changed\r?\n/, "")
        .replace(/\n{3,}/g, "\n\n")
        .trim();
    if (!notes) throw new Error("generated changelog is empty");
    return (
        proposed.slice(0, headings[0].index + headings[0][0].length) +
        "\n" +
        notes +
        "\n\n" +
        historical.slice(oldHeading.index)
    );
}

module.exports = async function updateChangelog({ github, context, baseSha }) {
    const { owner, repo } = context.repo;
    const headBranch = "release-please--branches--master";
    // Release Please emits no PR output on an unchanged retry. Find the pending
    // release branch directly so a failed enrichment can still be retried.
    const { data: pulls } = await github.rest.pulls.list({
        owner,
        repo,
        state: "open",
        head: `${owner}:${headBranch}`,
        base: "master",
    });
    for (const pull of pulls) {
        if (
            pull.state !== "open" ||
            pull.base.ref !== "master" ||
            pull.head.ref !== headBranch ||
            pull.head.repo.full_name !== `${owner}/${repo}`
        ) {
            throw new Error(
                "release PR no longer matches the expected repository and branches",
            );
        }
        if (pull.base.sha !== baseSha) {
            throw new Error(
                "source push is outdated; retry the latest master workflow run",
            );
        }
        const ref = pull.head.sha;
        const read = async (path) => {
            const { data } = await github.rest.repos.getContent({
                owner,
                repo,
                path,
                ref,
            });
            if (data.type !== "file" || data.encoding !== "base64") {
                throw new Error(`cannot read release PR file: ${path}`);
            }
            return Buffer.from(data.content, "base64").toString("utf8");
        };
        const manifest = JSON.parse(
            await read(".release-please-manifest.json"),
        );
        const original = await read("CHANGELOG.md");
        const content = renderChangelog(baseSha, manifest["."], original);
        if (content === original) continue;

        const { data: parent } = await github.rest.git.getCommit({
            owner,
            repo,
            commit_sha: ref,
        });
        const { data: tree } = await github.rest.git.createTree({
            owner,
            repo,
            base_tree: parent.tree.sha,
            tree: [
                { path: "CHANGELOG.md", mode: "100644", type: "blob", content },
            ],
        });
        const { data: commit } = await github.rest.git.createCommit({
            owner,
            repo,
            tree: tree.sha,
            parents: [ref],
            message: "chore(release): include merged history in changelog",
        });
        // A concurrent PR update makes this non-fast-forward and fails safely.
        await github.rest.git.updateRef({
            owner,
            repo,
            ref: `heads/${headBranch}`,
            sha: commit.sha,
            force: false,
        });
    }
};

module.exports.renderChangelog = renderChangelog;
