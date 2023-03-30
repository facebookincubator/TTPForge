//go:build embed

package main

import (
	"embed"

	"github.com/facebookincubator/TTP-Runner/cmd"
)

//go:embed all:.generated_ttps/**/*
var embeddedTTPs embed.FS

func init() {
	cmd.CfgYAML(&embeddedTTPs)
}
