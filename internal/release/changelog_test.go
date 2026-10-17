package release_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Release PR changelog", func() {
	var scriptPath, proposed, history string

	BeforeEach(func() {
		var err error
		scriptPath, err = filepath.Abs("../../scripts/release-changelog.cjs")
		Expect(err).NotTo(HaveOccurred())
		run("checkout", "--detach", "merged")
		history = run("show", "HEAD:CHANGELOG.md") + "\n"
		proposed = strings.Replace(history, "# Changelog\n\n", "# Changelog\n\n## [1.3.0](https://example.com/v1.3.0) (2025-02-03)\n\n### Features\n\n- integrate incoming\n\n", 1)
	})

	// Only the remote API and git-cliff executable are replaced. Git traversal,
	// rendering, history validation and commit payloads use the real script.
	publish := func(changelog, headBranch string, liveBase ...string) (map[string]interface{}, string, error) {
		baseSha := run("rev-parse", "merged")
		liveBaseSha := baseSha
		if len(liveBase) > 0 {
			liveBaseSha = liveBase[0]
		}
		input, err := json.Marshal(map[string]interface{}{
			"changelog": changelog, "headBranch": headBranch, "baseSha": baseSha, "liveBaseSha": liveBaseSha,
		})
		Expect(err).NotTo(HaveOccurred())
		cmd := exec.Command("node", "-e", `
const fs = require('node:fs');
const input = JSON.parse(fs.readFileSync(0, 'utf8'));
const result = {writes: []};
const owner = 'fixture', repo = 'library', sha = 'a'.repeat(40);
const pull = {state:'open',head:{ref:input.headBranch,sha,repo:{full_name:owner+'/'+repo}},base:{ref:'master',sha:input.liveBaseSha}};
const github = {rest: {
  pulls: {
    list: async ({head,base,state}) => {
      if (head !== owner+':release-please--branches--master' || base !== 'master' || state !== 'open') {
        throw Error('lookup must select the existing pending release PR');
      }
      return {data:[pull]};
    },
  },
  repos: {getContent: async ({path,ref}) => {
    if (ref !== sha) throw Error('content must be read at the captured PR SHA');
    return {data:{type:'file',encoding:'base64',content:Buffer.from(path === 'CHANGELOG.md' ? input.changelog : '{".":"1.3.0"}').toString('base64')}};
  }},
  git: {
    getCommit: async () => ({data:{tree:{sha:'b'.repeat(40)}}}),
    createTree: async (value) => {result.tree=value; result.writes.push('tree'); return {data:{sha:'c'.repeat(40)}};},
    createCommit: async (value) => {result.commit=value; result.writes.push('commit'); return {data:{sha:'d'.repeat(40)}};},
    updateRef: async (value) => {result.ref=value; result.writes.push('ref');},
  }
}};
require(process.argv[1])({github, context:{repo:{owner,repo}}, baseSha:input.baseSha})
  .then(() => process.stdout.write(JSON.stringify(result)))
  .catch(error => {process.stderr.write(error.message); process.exitCode=1;});
`, scriptPath)
		cmd.Dir = repo
		cmd.Stdin = strings.NewReader(string(input))
		cmd.Env = append(os.Environ(), "PATH="+filepath.Dir(cliff)+string(os.PathListSeparator)+os.Getenv("PATH"))
		out, commandErr := cmd.CombinedOutput()
		var result map[string]interface{}
		if commandErr == nil {
			Expect(json.Unmarshal(out, &result)).To(Succeed(), string(out))
		}
		return result, string(out), commandErr
	}

	It("enriches an existing PR without action outputs and preserves individual merged features and history", func() {
		result, out, err := publish(proposed, "release-please--branches--master")
		Expect(err).NotTo(HaveOccurred(), out)
		Expect(result).To(HaveKey("tree"))
		tree := result["tree"].(map[string]interface{})["tree"].([]interface{})
		Expect(tree).To(HaveLen(1))
		file := tree[0].(map[string]interface{})
		Expect(file["path"]).To(Equal("CHANGELOG.md"))
		text := file["content"].(string)
		Expect(text).To(ContainSubstring("feat: old imported capability"))
		Expect(text).To(ContainSubstring("fix: fork improvement"))
		Expect(text).NotTo(ContainSubstring("only in newer release"))
		Expect(text).NotTo(ContainSubstring("feat: baseline"))
		Expect(text).To(HaveSuffix(strings.TrimPrefix(history, "# Changelog\n\n")))
		Expect(text).To(ContainSubstring("## [1.3.0](https://example.com/v1.3.0) (2025-02-03)"))
		Expect(result["ref"].(map[string]interface{})["force"]).To(BeFalse())
		Expect(result["commit"].(map[string]interface{})["parents"]).To(Equal([]interface{}{strings.Repeat("a", 40)}))
	})

	It("does not write a commit when the generated content already matches", func() {
		first, out, err := publish(proposed, "release-please--branches--master")
		Expect(err).NotTo(HaveOccurred(), out)
		generated := first["tree"].(map[string]interface{})["tree"].([]interface{})[0].(map[string]interface{})["content"].(string)
		second, out, err := publish(generated, "release-please--branches--master")
		Expect(err).NotTo(HaveOccurred(), out)
		Expect(second["writes"]).To(BeEmpty())
	})

	It("refuses an outdated source push rather than drop newer release entries", func() {
		_, out, err := publish(proposed, "release-please--branches--master", strings.Repeat("e", 40))
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("source push"))
	})

	It("refuses changed historical entries", func() {
		_, out, err := publish(strings.Replace(proposed, "Historical text", "Rewritten history", 1), "release-please--branches--master")
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("historical changelog"))
	})

	It("refuses a different PR head branch", func() {
		_, out, err := publish(proposed, "unrelated")
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("release PR"))
	})
})
