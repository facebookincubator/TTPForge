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

package files_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/file"
	"github.com/l50/goutils/v2/docs"
	"github.com/l50/goutils/v2/fileutils"
)

func TestExamples(t *testing.T) {
	t.Run("ExampleCreateDirIfNotExists", ExampleCreateDirIfNotExists)
	t.Run("ExamplePathExistsInInventory", ExamplePathExistsInInventory)
	t.Run("ExampleTemplateExists", ExampleTemplateExists)
	t.Run("ExampleTTPExists", ExampleTTPExists)
	t.Run("ExampleMkdirAllFS", ExampleMkdirAllFS)
}

func ExampleCreateDirIfNotExists(t *testing.T) {
	fsys := fileutils.MockFs{}
	dirPath := "path/to/directory"

	if err := file.CreateDirIfNotExists(fsys, dirPath); err != nil {
		fmt.Printf("failed to create directory: %v", err)
		return
	}
}

func ExamplePathExistsInInventory(t *testing.T) {
	fsys := fileutils.MockFs{} // Replace with your mocked filesystem implementation
	relFilePath := "templates/exampleTTP.yaml.tmpl"
	inventoryPaths := []string{"path/to/inventory1", "path/to/inventory2"}

	exists, err := file.PathExistsInInventory(fsys, relFilePath, inventoryPaths)
	if err != nil {
		fmt.Printf("failed to check file existence: %v", err)
		return
	}

	if exists {
		fmt.Printf("File %s found in the inventory directories\n", relFilePath)
	} else {
		fmt.Printf("File %s not found in the inventory directories\n", relFilePath)
	}
}

func ExampleTemplateExists(t *testing.T) {
	fsys := fileutils.MockFs{}
	templatePath := "bash"
	inventoryPaths := []string{"path/to/inventory1", "path/to/inventory2"}

	fullPath, err := file.TemplateExists(fsys, templatePath, inventoryPaths)
	if err != nil {
		fmt.Printf("failed to check template existence: %v", err)
		return
	}

	if fullPath != "" {
		fmt.Printf("Template %s found in the parent directory of the inventory directories\n", templatePath)
	} else {
		fmt.Printf("Template %s not found in the parent directory of the inventory directories\n", templatePath)
	}
}

func ExampleTTPExists(t *testing.T) {
	ttpName := "exampleTTP"
	inventoryPaths := []string{"path/to/inventory1", "path/to/inventory2"}
	fsys := fileutils.MockFs{}

	exists, err := docs.TTPExists(fsys, ttpName, inventoryPaths)
	if err != nil {
		log.Fatalf("failed to check TTP existence: %v", err)
	}

	if exists {
		log.Printf("TTP %s found in the inventory directories\n", ttpName)
	} else {
		log.Printf("TTP %s not found in the inventory directories\n", ttpName)
	}
}

func ExampleMkdirAllFS(t *testing.T) {
	fsys := fileutils.MockFs{} // Replace with your mocked filesystem implementation
	dirPath := "path/to/directory"
	if err := file.MkdirAllFS(fsys, dirPath, 0755); err != nil {
		fmt.Printf("failed to create directory: %v", err)
		return
	}
}

func ExampleFixCodeBlocks(t *testing.T) {
	input := `Driver represents an interface to Google Chrome using go.

It contains a context.Context associated with this Driver and
Options for the execution of Google Chrome.

` + "```go" + `
browser, err := cdpchrome.Init(true, true)

if err != nil {
    log.Fatalf("failed to initialize a chrome browser: %v", err)
}
` + "```"
	language := "go"

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "example.*.md")
	if err != nil {
		log.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	// Write the input to the temp file
	if _, err := tmpfile.Write([]byte(input)); err != nil {
		log.Fatalf("failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatalf("failed to close temp file: %v", err)
	}

	// Run the function
	file := fileutils.RealFile(tmpfile.Name())
	err = docs.FixCodeBlocks(file, language)
	if err != nil {
		log.Fatalf("failed to fix code blocks: %v", err)
	}

	// Read the modified content
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	// Print the result
	fmt.Println(strings.TrimSpace(string(content)))
	// Output:
	// Driver represents an interface to Google Chrome using go.
	//
	// It contains a context.Context associated with this Driver and
	// Options for the execution of Google Chrome.
	//
	// ```go
	// browser, err := cdpchrome.Init(true, true)
	//
	// if err != nil {
	//     log.Fatalf("failed to initialize a chrome browser: %v", err)
	// }
	// ```
}
