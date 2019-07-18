package main

import (
	"fmt"
	"time"

	"github.com/samwho/livesplit"
)

func main() {
	client := livesplit.NewClient()
	defer client.Close()

	phase, err := client.GetCurrentTimerPhase()
	if err != nil {
		panic(err)
	}

	if phase != livesplit.NotRunning {
		if err := client.Reset(); err != nil {
			panic(err)
		}
	}

	if err := client.StartTimer(); err != nil {
		panic(err)
	}

	for {
		<-time.After(2 * time.Second)

		if err := client.Split(); err != nil {
			panic(err)
		}

		phase, err := client.GetCurrentTimerPhase()
		if err != nil {
			panic(err)
		}

		if phase == livesplit.Ended {
			break
		}

		splitTime, err := client.GetLastSplitTime()
		if err != nil {
			panic(err)
		}

		splitName, err := client.GetCurrentSplitName()
		if err != nil {
			panic(err)
		}

		timerPhase, err := client.GetCurrentTimerPhase()
		if err != nil {
			panic(err)
		}

		comparisonSplitTime, err := client.GetComparisonSplitTime()
		if err != nil {
			panic(err)
		}

		splitIndex, err := client.GetSplitIndex()
		if err != nil {
			panic(err)
		}

		bestPossibleTime, err := client.GetBestPossibleTime()
		if err != nil {
			panic(err)
		}

		finalTime, err := client.GetFinalTime("")
		if err != nil {
			panic(err)
		}

		fmt.Printf("------------------------------\n")
		fmt.Printf("split name: %v\n", splitName)
		fmt.Printf("split index: %v\n", splitIndex)
		fmt.Printf("split time: %v\n", livesplit.DurationToString(splitTime))
		fmt.Printf("comparison split time:  %v\n", livesplit.DurationToString(comparisonSplitTime))
		fmt.Printf("timer phase: %v\n", timerPhase)
		fmt.Printf("best possible time: %v\n", livesplit.DurationToString(bestPossibleTime))
		fmt.Printf("final time: %v\n", livesplit.DurationToString(finalTime))
	}
}
