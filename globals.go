package globals

import (
	_ "embed"
)

//go:embed commands.csv
var CommandsCsv string
