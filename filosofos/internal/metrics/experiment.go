package metrics

import "fmt"

func PrintExperimentHeader(
	strategy string,
	run int,
) {
	fmt.Printf(
		"Strategy: %s | Run: %d\n",
		strategy,
		run,
	)
}