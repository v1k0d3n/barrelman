package e2e

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"strconv"
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
	fmt.Fprintln(os.Stdout, "Waiting for atmost 120 seconds for the pods to get into 'Running' state")
	expectedPodCount = expectedPodCount+1
	for {
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
	}
	return nil
}
