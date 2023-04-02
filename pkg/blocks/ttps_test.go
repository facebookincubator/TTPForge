package blocks_test

import (
	"testing"

	"github.com/facebookincubator/TTP-Runner/pkg/blocks"
	"github.com/facebookincubator/TTP-Runner/pkg/logging"

	"gopkg.in/yaml.v3"
)

func init() {
	logging.ToggleDebug()
}

func TestUnmarshalSimpleCleanupLarge(t *testing.T) {

	var ttps blocks.TTP

	content := `name: test
description: this is a test
steps:
  - name: testinline
    inline: |
      ls
    cleanup:
      name: test_cleanup
      inline: |
        ls -la
  - name: test_cleanup_two
    inline: |
      ls
    cleanup:
      name: test_cleanup
      inline: |
        ls -la
  - name: test_cleanup_three
    inline: |
      ls
    cleanup:
      name: test_cleanup
      inline: |
        ls -la

  `

	err := yaml.Unmarshal([]byte(content), &ttps)
	if err != nil {
		t.Errorf("failed to unmarshal basic inline %v", err)
	}

	t.Logf("successfully unmarshalled data: %v", ttps)

}

func TestUnmarshalScenario(t *testing.T) {

	var ttps blocks.TTP

	content := `
name: FBPkg Privesc
description: |
  Privesc via malicious fbpkg backdoor
steps:
  - name: generate_evil_entrypoint
    file: ~/security-ttpcode/ttps/privilege-escalation/credential-theft/fbpkg-backdoor-tls-cert-theft/generate-evil-entrypoint.sh
    cleanup:
      name: cleanup_entrypoint
      inline: |
        rm entrypoint.sh
  - name: overwrite_fbpkg
    inline: |
      ~/security-ttpcode/ttps/utils/tupperware/sandbox-job/build-entrypoint-fbpkg.sh "tw_ttp_entrypoint_fbpkg_${USER}"
  - name: provision_tupperware_job
    file: ~/security-ttpcode/ttps/utils/tupperware/sandbox-job/launch-sandbox.sh
    args:
      - --skip-fbpkg
    cleanup:
      name: cleanup_script
      file: ~/security-ttpcode/ttps/utils/tupperware/sandbox-job/cleanup.sh
  - name: receive_certificate
    inline: |
      nc -6 -lnvp 8888 | gzip -d > fbpkg-stolen-tls-cert.pem
    cleanup:
      name: remove_stolen_tls
      inline: |
        rm -f fbpkg-stolen-tls-cert.pem
  `

	err := yaml.Unmarshal([]byte(content), &ttps)
	if err != nil {
		t.Errorf("failed to unmarshal basic inline %v", err)
	}

	t.Logf("successfully unmarshalled data: %v", ttps)

}
