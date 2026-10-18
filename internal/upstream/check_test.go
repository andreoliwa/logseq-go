package upstream_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUpstream(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Upstream notice") }

func mustLookPath(name string) string {
	path, err := exec.LookPath(name)
	Expect(err).NotTo(HaveOccurred())
	return path
}

var curlStub string

var _ = BeforeSuite(func() {
	directory := GinkgoT().TempDir()
	curlStub = filepath.Join(directory, "curl")
	if runtime.GOOS == "windows" {
		curlStub += ".exe"
	}
	command := exec.Command("go", "build", "-o", curlStub, "./testdata/curl")
	output, err := command.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(output))
})

var _ = Describe("Cached upstream notices", func() {
	var directory, script, counter, cache, reviewed string
	var environment []string
	invoke := func() string {
		command := exec.Command("python3", "-c", `import os,runpy,sys; m=runpy.run_path(sys.argv[1]); m["time"].time=lambda:float(os.environ["TEST_NOW"]); sys.exit(m["main"]())`, script)
		command.Env = environment
		output, err := command.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		return string(output)
	}
	calls := func() int { data, _ := os.ReadFile(counter); return strings.Count(string(data), "lookup") }
	state := func(value map[string]any) {
		data, err := json.Marshal(value)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.MkdirAll(filepath.Dir(cache), 0755)).To(Succeed())
		Expect(os.WriteFile(cache, data, 0644)).To(Succeed())
	}
	BeforeEach(func() {
		directory = GinkgoT().TempDir()
		source, err := os.ReadFile("../../scripts/check-upstream.py")
		Expect(err).NotTo(HaveOccurred())
		reviewedBytes, err := os.ReadFile("../../scripts/upstream-reviewed.sha")
		Expect(err).NotTo(HaveOccurred())
		reviewed = strings.TrimSpace(string(reviewedBytes))
		scripts := filepath.Join(directory, "path with spaces")
		Expect(os.MkdirAll(scripts, 0755)).To(Succeed())
		script = filepath.Join(scripts, "check-upstream.py")
		Expect(os.WriteFile(script, source, 0644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(scripts, "upstream-reviewed.sha"), reviewedBytes, 0644)).To(Succeed())
		bin := filepath.Join(directory, "bin")
		Expect(os.MkdirAll(bin, 0755)).To(Succeed())
		stub, err := os.ReadFile(curlStub)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(bin, filepath.Base(curlStub)), stub, 0755)).To(Succeed())
		counter = filepath.Join(directory, "calls")
		cache = filepath.Join(directory, "cache", "logseq-go", "upstream.json")
		environment = []string{"PATH=" + bin, "HOME=" + directory, "XDG_CACHE_HOME=" + filepath.Join(directory, "cache"), "TEST_NOW=1000", "COUNTER=" + counter, `RESPONSE={"object":{"sha":"` + reviewed + `"}}`}
		if runtime.GOOS == "windows" {
			environment = append(environment, "SYSTEMROOT="+os.Getenv("SYSTEMROOT"), "OS=Windows_NT")
		}
	})
	It("is quiet at the reviewed tip and throttles repeated checks", func() { Expect(invoke()).To(BeEmpty()); Expect(invoke()).To(BeEmpty()); Expect(calls()).To(Equal(1)) })
	It("retains a pending notice from cache without another lookup", func() {
		environment = append(environment, `RESPONSE={"object":{"sha":"`+strings.Repeat("a", 40)+`"}}`)
		Expect(invoke()).To(ContainSubstring("/compare/"))
		Expect(invoke()).To(ContainSubstring("/compare/"))
		Expect(calls()).To(Equal(1))
	})
	It("checks at the interval boundary but not before it", func() {
		state(map[string]any{"reviewed": reviewed, "attempt": 900})
		environment = append(environment, "UPSTREAM_CHECK_INTERVAL_SECONDS=101")
		Expect(invoke()).To(BeEmpty())
		Expect(calls()).To(BeZero())
		environment = append(environment, "UPSTREAM_CHECK_INTERVAL_SECONDS=100")
		invoke()
		Expect(calls()).To(Equal(1))
	})
	It("allows zero to force repeated lookups", func() {
		environment = append(environment, "UPSTREAM_CHECK_INTERVAL_SECONDS=0")
		invoke()
		invoke()
		Expect(calls()).To(Equal(2))
	})
	It("throttles failures without marking them successful", func() {
		environment = append(environment, "CHECK_FAIL=1")
		Expect(invoke()).To(ContainSubstring("warning:"))
		invoke()
		Expect(calls()).To(Equal(1))
		data, err := os.ReadFile(cache)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("success"))
	})
	It("invalidates comparisons when the reviewed baseline changes", func() {
		state(map[string]any{"reviewed": strings.Repeat("b", 40), "attempt": 1000, "observed": strings.Repeat("c", 40)})
		Expect(invoke()).To(BeEmpty())
		Expect(calls()).To(Equal(1))
	})
	It("recovers from corrupt cache data and stale lock files", func() {
		Expect(os.MkdirAll(filepath.Dir(cache), 0755)).To(Succeed())
		Expect(os.WriteFile(cache, []byte("broken"), 0644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(filepath.Dir(cache), "upstream.lock"), nil, 0644)).To(Succeed())
		Expect(invoke()).To(BeEmpty())
		Expect(calls()).To(Equal(1))
	})
	It("warns on an unusable cache without making a lookup", func() {
		Expect(os.WriteFile(filepath.Join(directory, "not-a-directory"), nil, 0644)).To(Succeed())
		environment = append(environment, "XDG_CACHE_HOME="+filepath.Join(directory, "not-a-directory"))
		Expect(invoke()).To(ContainSubstring("warning:"))
		Expect(calls()).To(BeZero())
	})
	It("rejects unexpected responses and keeps the reviewed file unchanged", func() {
		environment = append(environment, `RESPONSE={"object":{"sha":"oops"}}`)
		Expect(invoke()).To(ContainSubstring("warning:"))
		data, err := os.ReadFile(filepath.Join(filepath.Dir(script), "upstream-reviewed.sha"))
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.TrimSpace(string(data))).To(Equal(reviewed))
	})
	It("skips platform-specific locking setup under CI", func() {
		command := exec.Command("python3", "-c", `import builtins,runpy,sys
original = builtins.__import__
def no_lock_modules(name, *args, **kwargs):
    if name == "fcntl":
        raise ImportError("fcntl unavailable")
    return original(name, *args, **kwargs)
builtins.__import__ = no_lock_modules
runpy.run_path(sys.argv[1], run_name="__main__")`, script)
		command.Env = append(environment, "CI=1")
		output, err := command.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		Expect(string(output)).To(BeEmpty())
		Expect(calls()).To(BeZero())
	})
	It("skips all network work under CI", func() {
		environment = append(environment, "CI=1")
		Expect(invoke()).To(BeEmpty())
		Expect(calls()).To(BeZero())
	})
	It("does not duplicate concurrent lookups", func() {
		environment = append(environment, "CHECK_SLEEP=1")
		command := exec.Command("python3", script)
		command.Env = environment
		Expect(command.Start()).To(Succeed())
		Eventually(calls, 3*time.Second, 10*time.Millisecond).Should(Equal(1))
		Expect(invoke()).To(BeEmpty())
		Expect(command.Wait()).To(Succeed())
		Expect(calls()).To(Equal(1))
	})
	It("bounds a stalled lookup and cleans up its process group", func() {
		environment = append(environment, "CHECK_SLEEP=30")
		start := time.Now()
		Expect(invoke()).To(ContainSubstring("timed out"))
		Expect(time.Since(start)).To(BeNumerically("<", 9*time.Second))
		Expect(calls()).To(Equal(1))
	})
	It("works through make with a minimal environment", func() {
		python, err := exec.LookPath("python3")
		Expect(err).NotTo(HaveOccurred())
		environment = append(environment, "CI=1", "PATH="+filepath.Dir(python)+string(os.PathListSeparator)+filepath.Dir(mustLookPath("make")))
		command := exec.Command("make", "-s", "-C", "../..", "check-upstream")
		command.Env = environment
		output, err := command.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		Expect(string(output)).To(BeEmpty())
		Expect(calls()).To(BeZero())
	})
	It("warns without failing when curl is missing", func() {
		environment = append(environment, "PATH=/no-tools")
		Expect(invoke()).To(ContainSubstring("warning:"))
		Expect(calls()).To(BeZero())
	})
	It("runs make from a checkout path containing spaces", func() {
		checkout := filepath.Join(directory, "checkout with spaces")
		Expect(os.MkdirAll(filepath.Join(checkout, "scripts"), 0755)).To(Succeed())
		for _, item := range []struct{ source, target string }{{"../../Makefile", "Makefile"}, {script, "scripts/check-upstream.py"}, {filepath.Join(filepath.Dir(script), "upstream-reviewed.sha"), "scripts/upstream-reviewed.sha"}} {
			data, err := os.ReadFile(item.source)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(filepath.Join(checkout, item.target), data, 0644)).To(Succeed())
		}
		python, err := exec.LookPath("python3")
		Expect(err).NotTo(HaveOccurred())
		command := exec.Command("make", "-s", "-C", checkout, "check-upstream", "OS=Windows_NT")
		command.Env = append(environment, "CI=1", "PATH="+filepath.Dir(python)+string(os.PathListSeparator)+filepath.Dir(mustLookPath("make")))
		output, err := command.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
		Expect(string(output)).To(BeEmpty())
		Expect(calls()).To(BeZero())
	})

})
