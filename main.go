package main

import "calendar-sync/cmd"

func main() {
	if err := cmd.Main(); err != nil {
		println(err)
	}
}
