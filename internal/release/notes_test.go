package release_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var repo, script, cliff string

func run(args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), string(out))
	return strings.TrimSpace(string(out))
}
func notes(tag string) (string, error) {
	cmd := exec.Command("bash", script, tag)
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "PATH="+filepath.Dir(cliff)+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	return string(out), err
}

var _ = BeforeSuite(func() {
	var err error
	script, err = filepath.Abs("../../scripts/release-notes.sh")
	Expect(err).NotTo(HaveOccurred())
	repo = GinkgoT().TempDir()
	cliff = filepath.Join(GinkgoT().TempDir(), "git-cliff")
	// Render the selected real Git range without a network or GitHub API call.
	Expect(os.WriteFile(cliff, []byte("#!/bin/bash\nset -eu\nfor range; do :; done\ngit log --format='%s' \"$range\"\n"), 0700)).To(Succeed())
	run("init", "-b", "master")

	// Import a fixed DAG in one reference transaction. Fixtures need neither
	// personal signing keys nor interactive tag editors.
	var fixture strings.Builder
	commit := func(branch string, mark, timestamp int, message, parents string) {
		message += "\n"
		fmt.Fprintf(&fixture, "commit refs/heads/%s\nmark :%d\ncommitter Fixture Author <fixture@example.com> %d +0000\ndata %d\n%s%s\n", branch, mark, timestamp, len(message), message, parents)
	}
	commit("baseline", 1, 1735732800, "feat: baseline", "")
	commit("incoming", 2, 1577880000, "feat: old imported capability", "from :1\n")
	commit("fork", 3, 1738411200, "fix: fork improvement", "from :1\n")
	commit("merged", 4, 1738497600, "chore: integrate incoming", "from :3\nmerge :2\n")
	commit("event", 5, 1738584000, "fix: included in event release", "from :4\n")
	commit("master", 6, 1740830400, "fix: only in newer release", "from :5\n")
	for _, tag := range []struct {
		name string
		mark int
	}{{"v1.2.0", 1}, {"v99.0.0", 2}, {"v1.3.0", 5}, {"v1.4.0", 6}} {
		fmt.Fprintf(&fixture, "reset refs/tags/%s\nfrom :%d\n\n", tag.name, tag.mark)
	}
	cmd := exec.Command("git", "fast-import", "--quiet")
	cmd.Dir = repo
	cmd.Stdin = strings.NewReader(fixture.String())
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(out))
})

var _ = Describe("Release notes", func() {
	BeforeEach(func() { run("checkout", "--detach", "v1.3.0") })

	It("uses the event tag's range even when a newer release exists", func() {
		out, err := notes("v1.3.0")
		Expect(err).NotTo(HaveOccurred(), out)
		Expect(out).To(ContainSubstring("included in event release"))
		Expect(out).NotTo(ContainSubstring("only in newer release"))
		Expect(out).NotTo(ContainSubstring("feat: baseline"))
	})

	It("includes old upstream commits by ancestry and ignores side-branch tags", func() {
		out, err := notes("v1.3.0")
		Expect(err).NotTo(HaveOccurred(), out)
		Expect(out).To(ContainSubstring("old imported capability"))
		Expect(out).To(ContainSubstring("fork improvement"))
	})

	It("rejects a checkout that does not match the event tag", func() {
		out, err := notes("v1.4.0")
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("checkout does not match release tag"))
	})

	It("fails rather than guessing when the previous fork release is absent", func() {
		run("checkout", "--detach", "v1.2.0")
		out, err := notes("v1.2.0")
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("previous fork release"))
	})

	It("rejects invalid release tags", func() {
		out, err := notes("--all")
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("expected a vX.Y.Z release tag"))
	})
})
