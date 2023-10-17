package main

import "github.com/JackalLabs/sequoia/cmd"

func main() {
	root := cmd.RootCmd()

	cmd.Execute(root)
}
