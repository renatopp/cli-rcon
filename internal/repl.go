package internal

import (
	"bytes"
	"encoding/csv"
	"sort"

	"github.com/gorcon/rcon"
	"github.com/lmorg/readline/v4"
	globals "github.com/renatopp/go-rcon"
	"github.com/renatopp/x/fmtx"
	"github.com/renatopp/x/strx"
)

type command struct {
	defaultValue string
	flags        string
	description  string
}

var (
	commandNames []string
	commandMap   map[string]command
)

func StartRepl(conn *rcon.Conn, autocomplete bool) error {
	rl := readline.NewInstance()
	rl.SetPrompt(" > ")
	rl.MaxTabCompleterRows = 12

	var currentLine string
	trackLine := func(line []rune) { currentLine = string(line) }

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
			trackLine(line)
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
	} else {
		rl.HintText = func(line []rune, _ int) []rune {
			trackLine(line)
			return nil
		}
	}

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrCtrlC {
				if currentLine != "" {
					currentLine = ""
					continue
				}
				return nil
			}
			return err
		}

		if len(line) == 0 {
			continue
		}

		if line == "exit" {
			return nil
		}

		output, err := conn.Execute(line)
		if err != nil {
			return err
		}
		if output != "" {
			fmtx.Println("%s", fmtx.Dim(output))
		}
		if line == "clear" {
			fmtx.Print("\033c")
			continue
		}
	}
}

func parseCommands() {
	commandMap = make(map[string]command)

	r := csv.NewReader(bytes.NewBufferString(globals.CommandsCsv))
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
