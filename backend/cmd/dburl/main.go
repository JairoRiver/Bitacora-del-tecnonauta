// dburl prints the database DSN read from config.yaml.
// Used by Taskfile tasks that need the connection URL (e.g. migrate).
package main

import (
	"fmt"
	"os"

	"github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading config:", err)
		os.Exit(1)
	}
	fmt.Print(cfg.Database.DSN())
}
