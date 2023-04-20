package cmd

import "testing"

func runE2ETest(testFile string) {
	ExecuteYAML("e2e-tests/" + testFile)
}

func TestInline(t *testing.T) {
	runE2ETest("test_inline.yaml")
}
