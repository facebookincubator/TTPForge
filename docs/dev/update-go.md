# Updating the Golang Version for TTPForge

---

If you want to change the supported/recommended Golang version(s) for TTPForge,
you must do the following:

* Update minimum required Golang version specified in `go.mod`
* Make a corresponding update to `magefiles/go.mod`
* Update recommended Golang version for asdf in `.tool-versions`
* Update the [developer documentation](README.md) to indicate the new versions