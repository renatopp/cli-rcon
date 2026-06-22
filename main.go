package main

import (
	"github.com/gorcon/rcon"
	"github.com/lmorg/readline/v4"
	"github.com/renatopp/go-cli"
)

func main() {
	cli.Name("rcon")
	cli.Description("RCON client")
	address := cli.Pos("address", "The host and port of the RCON server").AsRequired()
	commands := cli.Pos("commands", "The RCON commands to execute").AsVariadic()
	password := cli.Flag("pass", "p", "Server RCON password")

	cli.Parse()

	conn, err := rcon.Dial(address.Value(), password.Value())
	cli.FatalIf(err)
	defer conn.Close()

	if len(commands.Values()) == 0 {
		repr(conn)
	} else {
		send(conn, commands.Values()...)
	}
	cli.FatalIf(err)
}

func repr(conn *rcon.Conn) {
	rl := readline.NewInstance()

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrCtrlC {
				return
			}
			cli.FatalIf(err)
		}

		if len(line) == 0 {
			continue
		}

		if line == "exit" {
			return
		}

		send(conn, line)
	}
}

func send(conn *rcon.Conn, cmds ...string) {
	for _, cmd := range cmds {
		output, err := conn.Execute(cmd)
		cli.FatalIf(err)
		cli.Print("%s", output)
	}
}
