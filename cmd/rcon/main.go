package main

import (
	"github.com/gorcon/rcon"
	"github.com/renatopp/go-cli"
	"github.com/renatopp/go-rcon/internal"
	"github.com/renatopp/x/fmtx"
)

func main() {
	cli.Name("rcon")
	cli.Description("RCON client.\nThe REPL is enabled when no commands are provided.")
	cli.AutoHelp(true)
	address := cli.Pos("address", "The host and port of the RCON server").AsRequired()
	commands := cli.Pos("commands", "The RCON commands to execute").AsVariadic()
	password := cli.Flag("pass", "p", "Server RCON password")
	noAutocomplete := cli.FlagBool("no-autocomplete", "", "Disable autocomplete")

	cli.Parse()

	conn, err := rcon.Dial(address.Value(), password.Value())
	cli.FatalIf(err)
	defer conn.Close()

	if len(commands.Values()) == 0 {
		err := internal.StartRepl(conn, !noAutocomplete.Value())
		cli.FatalIf(err)
	} else {
		for _, cmd := range commands.Values() {
			output, err := conn.Execute(cmd)
			cli.FatalIf(err)
			if output != "" {
				fmtx.Println(fmtx.Dim(output))
			}
		}
	}
	cli.FatalIf(err)
}
