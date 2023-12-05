# TTP Requirements Specification

Suppose that a TTPForge user attempts to run a linux-only TTP on a macOS system.
It is better for them to get an immediate, clear error of the form "this TTP can
only be run on macOS" than for them to receive an obscure error such as "XYZ
command not found" that pops up halfway through the execution of their TTP. The
TTPForge `requirements` section solves this problem. By declaring a
`requirements` section, you can clearly indicate to users of your TTP:

- With which platforms your TTP is compatible.
- Whether your TTP requires superuser privileges.

## Specifying Compatible Platforms

The example below demonstrates how to tell TTPForge (and your TTP's users) that
your TTP is only compatible with (and should therefore only be run) on certain
platforms:

https://github.com/facebookincubator/TTPForge/blob/7634dc65879ec43a108a4b2d44d7eb2105a2a4b1/example-ttps/requirements/os-and-superuser.yaml#L1-L21

Each entry in the `platforms` list should contain one or both of the following
fields:

- `os`: the target operating system.
- `arch`: the target architecture.

Since TTPForge is written in [Go](https://go.dev/), you may specify any valid
`GOOS` or `GOARCH` value for these fields - see
[here](https://go.dev/doc/install/source#environment) for the full list of
allowed values.

Note that it is quite common to specify entries containing `os` but not `arch` -
this simply means that the TTP is compatible with the specified `os` and that
its implementation is not architecture specific. Conversely, the following entry
indicates that the TTP is compatible with ARM-based macOS systems and x64-based
Windows systems, but not with e.g. legacy intel-based macOS systems:

```yml
platforms:
  - os: darwin
    arch: arm64
  - os: windows
    arch: amd64
```

### How TTPForge Determines the Current OS/Arch

TTPForge consults its `runtime.GOOS` and `runtime.GOARCH`
[constants](https://pkg.go.dev/runtime#pkg-constants) to determine the OS and
architecture on which it is currently running. These values are based on the
`GOOS` and `GOARCH` settings with which that particular TTPForge executable was
compiled. **Note that these values may sometimes be misleading** - for instance,
if TTPForge was compiled with `GOARCH=amd64`, it will consider itself to be
running on `amd64` even if it is being run on an `arm64` system using
[Rosetta](https://support.apple.com/en-us/HT211861).

## Specifying Required Privileges

If your TTP requires superuser privileges to run, you should specify
`superuser: true` in your `requirements` section, as demonstrated by the TTP
above. This will ensure that your users get an immediate and unambiguous error
message if they attempt to execute your TTP without the required privileges,
rather than a "Permission Denied..." error midway through TTP execution.
