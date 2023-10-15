package main

import "sequoia/cmd"

func main() {
	root := cmd.RootCmd()

	cmd.Execute(root)
}
