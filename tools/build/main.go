// Command build wraps the release build flow in a single entry point: it
// samples a fresh PGO profile with pgo-collect, then packages the app with
// mygo build for each requested platform. It replaces the sequence of
// `go run ./tools/pgo-collect -out default.pgo` and
// `go tool mygo build -platform ...` invocations documented in AGENTS.md.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	platforms := flag.String("platform", "", "comma separated mygo platforms, e.g. darwin/universal,linux/amd64,windows/amd64 (default: the host platform)")
	skipPGO := flag.Bool("skip-pgo", false, "reuse the existing default.pgo instead of sampling a fresh profile")
	flag.Parse()

	root, err := findRoot()
	if err != nil {
		log.Fatal(err)
	}

	if !*skipPGO {
		if err := run(root, "go", "run", "./tools/pgo-collect", "-out", "default.pgo"); err != nil {
			log.Fatalf("pgo-collect: %v", err)
		}
	}

	targets := splitPlatforms(*platforms)
	if len(targets) == 0 {
		targets = []string{runtime.GOOS + "/" + runtime.GOARCH}
	}
	for _, platform := range targets {
		log.Printf("packaging %s", platform)
		if err := run(root, "go", "tool", "mygo", "build", "-platform", platform); err != nil {
			log.Fatalf("mygo build -platform %s: %v", platform, err)
		}
	}
}

// run executes name with args in dir, streaming output to the terminal.
func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// findRoot locates the workspace root by walking up from the working
// directory until it finds the module that owns this file.
func findRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("locate tools/build source file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..")), nil
}

// splitPlatforms normalizes the -platform flag into a clean target list.
func splitPlatforms(value string) []string {
	var targets []string
	for _, part := range strings.Split(value, ",") {
		if platform := strings.TrimSpace(part); platform != "" {
			targets = append(targets, platform)
		}
	}
	return targets
}
