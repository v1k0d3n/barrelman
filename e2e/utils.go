package e2e

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"time"
)

func WaitForPodsToBeInRunningState(manifestNS string, expectedPodCount int) error {
	fmt.Fprintln(os.Stdout, "Waiting for atmost 120 seconds for the pods to get into 'Running' state")
	expectedPodCount = expectedPodCount+1
	for now := time.Now(); ; {
		kubecmd := "kubectl get pods -n " + manifestNS + " --field-selector status.phase=Running"
		countcmd := kubecmd + "| wc -l"
		out, _ := exec.Command("/bin/bash", "-c", kubecmd).CombinedOutput()
		fmt.Fprintln(os.Stdout, string(out))
		expectedString := strconv.Itoa(expectedPodCount)
		fmt.Fprintln(os.Stdout, "Expected Pod Count:", expectedString)
		outCount, _ := exec.Command("/bin/bash", "-c", countcmd).CombinedOutput()
		outString := string(outCount)
		fmt.Fprintln(os.Stdout, "Actual Pod Count:", outString)
		if strings.Contains(outString, expectedString) {
			fmt.Fprintln(os.Stdout, "Pod is/are in 'Running' state after performing 'barrelman apply'")
			break
		}
		timeout(now, 20)
	}
	return nil
}

func timeout(currentTime time.Time, sec int) {
	if time.Since(currentTime) > time.Second*time.Duration(sec) {
		log.Panic("Timed-out:",sec,"seconds")
        }
}
