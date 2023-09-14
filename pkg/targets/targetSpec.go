/*
Copyright © 2023-present, Meta Platforms, Inc. and affiliates
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package targets

// MitreAttack represents mappings to the MITRE ATT&CK framework.
//
// **Attributes:**
//
// Tactics: A string slice containing the MITRE ATT&CK tactic(s) associated with the TTP.
// Techniques: A string slice containing the MITRE ATT&CK technique(s) associated with the TTP.
// SubTechniques: A string slice containing the MITRE ATT&CK sub-technique(s) associated with the TTP.
type MitreAttack struct {
	Tactics       []string `yaml:"tactics,omitempty"`
	Techniques    []string `yaml:"techniques,omitempty"`
	SubTechniques []string `yaml:"subtechniques,omitempty"`
}

// TargetSpec represents the specifications for valid targets within a TTP.
//
// Attributes:
// OS: A slice containing the operating systems valid for this TTP.
// Arch: A slice containing the architectures valid for this TTP.
// Cloud: A slice of Cloud structs, each representing a cloud provider and region valid for this TTP.
type TargetSpec struct {
	OS    []string `yaml:"os,omitempty"`
	Arch  []string `yaml:"arch,omitempty"`
	Cloud []Cloud  `yaml:"cloud,omitempty"`
}

// Cloud represents the cloud provider structure for a TTP.
//
// **Attributes:**
//
// Provider: The name of the cloud provider to be targeted.
// Region: The name of the cloud region to be targeted.
type Cloud struct {
	Provider string `yaml:"provider,omitempty"`
	Region   string `yaml:"region,omitempty"`
}

// ParseAndValidateTargets takes a TargetSpec and processes the targets specified in it.
// It then returns a map containing the processed targets. If the TargetSpec does not have any
// valid targets specified, the resulting map will be empty.
//
// Parameters:
// targetSpec: The specifications for valid targets, extracted from the YAML configuration.
//
// Returns:
// processedTargets: A map of target names (like 'os', 'arch', etc.) to their values.
// err: An error if any issues arise during the processing.
func ParseAndValidateTargets(targetSpec TargetSpec) (map[string]interface{}, error) {
	processedTargets := make(map[string]interface{})

	if len(targetSpec.OS) > 0 {
		processedTargets["os"] = targetSpec.OS
	}

	if len(targetSpec.Arch) > 0 {
		processedTargets["arch"] = targetSpec.Arch
	}

	if len(targetSpec.Cloud) > 0 {
		var cloudStrs []string
		for _, cloud := range targetSpec.Cloud {
			cloudStrs = append(cloudStrs, cloud.Provider+":"+cloud.Region)
		}
		processedTargets["cloud"] = cloudStrs
	}

	return processedTargets, nil
}
