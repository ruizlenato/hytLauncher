package main

import (
	"embed"
)

//go:embed Aurora/Build/Aurora.so
var embeddedFiles embed.FS
