package e2e

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func fullPath() string {
	barrelmanPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(barrelmanPath)
	return barrelmanPath
}

func WaitForPodsToBeInRunningState(manifestNS string, expectedPodCount int) error {
	tStart := time.Now()
	fmt.Fprint(os.Stdout, "Waiting for atmost 120seconds for the pods to get into 'Running' state'")
	for {
		tProgress := time.Now()
		out, err := exec.Command("kubectl", "-n", manifestNS, "get", "pods", "|", "grep", "Running", "|", "wc", "-l").CombinedOutput()
		time.Sleep(100 * time.Millisecond)
		fmt.Fprint(os.Stdout, ".")
		if string(out) == string(expectedPodCount) {
			fmt.Fprint(os.Stdout, "Pod is/are in 'Running' state after performing 'barrelman apply'")
			break
		}
		if err != nil {
			return err
		}
		elapsed := tProgress.Sub(tStart)
		if elapsed > 12000 {
			log.Fatal("Timed out to list the pod after the barrelman manifest is applied")
		}
	}
	return nil
}
