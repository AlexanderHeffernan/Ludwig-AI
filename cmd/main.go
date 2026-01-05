package main

import (
	"flag"
	"fmt"
	"ludwig/internal/cli"
	"ludwig/internal/updater"
)

var version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "Print the version and exit")
	updateFlag := flag.Bool("update", false, "Check for and install updates")
	flag.Parse()

	// Apply any pending updates from previous run
	updated, err := updater.ApplyPendingUpdate()
	if err != nil {
		fmt.Println("Warning: " + err.Error())
	}
	if updated {
		fmt.Println("Update applied! Please run the command again.")
		return
	}

	if *versionFlag {
		fmt.Println("ludwig version " + version)
		return
	}

	if *updateFlag {
		if err := updater.DownloadAndInstall(version); err != nil {
			fmt.Println("Error: " + err.Error())
			return
		}
		return
	}

	cli.StartInteractive(version)
}
