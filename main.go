package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"sort"

	"github.com/gorcon/rcon"
	"github.com/lmorg/readline/v4"
	"github.com/renatopp/go-cli"
	"github.com/renatopp/x/strx"
)

//go:embed commands.csv
var commandsCsv string

type command struct {
	defaultValue string
	flags        string
	description  string
}

var (
	commandNames []string
	commandMap   map[string]command
)

func main() {
	cli.Name("rcon")
	cli.Description(`RCON client.
The REPL is enabled when no commands are provided.`)
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
		repl(conn, !noAutocomplete.Value())
	} else {
		send(conn, commands.Values()...)
	}
	cli.FatalIf(err)
}

func repl(conn *rcon.Conn, autocomplete bool) {
	rl := readline.NewInstance()
	rl.SetPrompt(" > ")
	rl.MaxTabCompleterRows = 12
	if autocomplete {
		parseCommands()
		rl.TabCompleter = func(line []rune, pos int, _ readline.DelayedTabContext) *readline.TabCompleterReturnT {
			command := string(line[:pos])
			suggestions := []string{}
			descriptions := map[string]string{}
			total := 0
			for _, name := range commandNames {
				if strx.HasPrefix(name, command) {
					suffix := name[len(command):]
					suggestions = append(suggestions, suffix)
					if cmd, ok := commandMap[name]; ok && cmd.description != "" {
						descriptions[suffix] = cmd.description
					}
					total++
					if total > rl.MaxTabCompleterRows {
						break
					}
				}
			}
			return &readline.TabCompleterReturnT{
				Prefix:       command,
				Suggestions:  suggestions,
				Descriptions: descriptions,
				DisplayType:  readline.TabDisplayList,
			}
		}
		rl.HintText = func(line []rune, pos int) []rune {
			fields := strx.Fields(string(line))
			if len(fields) == 0 {
				return nil
			}
			name := fields[0]
			if cmd, ok := commandMap[name]; ok {
				hint := cmd.description
				if hint == "" {
					hint = "(no description)"
				}
				if cmd.defaultValue != "" && cmd.defaultValue != "cmd" {
					hint += "  [default: " + cmd.defaultValue + "]"
				}
				if cmd.flags != "" {
					hint += "  [flags: " + cmd.flags + "]"
				}
				return []rune(hint)
			}
			return nil
		}

		for ch := rune(32); ch <= 126; ch++ {
			rl.AddEvent(string(ch), func(_ int, state *readline.EventState) *readline.EventReturn {
				return &readline.EventReturn{
					SetLine:  []rune(state.Line),
					SetPos:   state.CursorPos,
					Actions:  []func(*readline.Instance){readline.HkFnModeAutocomplete},
					Continue: true,
				}
			})
		}
	}

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

		if line == "clear" {
			println("\033c")
			continue
		}
	}
}

func send(conn *rcon.Conn, cmds ...string) {
	for _, cmd := range cmds {
		output, err := conn.Execute(cmd)
		cli.FatalIf(err)
		cli.Print("%s", output)
	}
}

func parseCommands() {
	commandMap = make(map[string]command)

	r := csv.NewReader(bytes.NewBufferString(commandsCsv))
	records, err := r.ReadAll()
	if err != nil || len(records) < 2 {
		return
	}

	for _, row := range records[1:] { // skip header
		if len(row) < 4 {
			continue
		}
		name := strx.TrimSpace(row[0])
		commandMap[name] = command{
			defaultValue: row[1],
			flags:        row[2],
			description:  row[3],
		}
		commandNames = append(commandNames, name)
	}
	sort.Strings(commandNames)
}
