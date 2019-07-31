package e2e

import (
	"errors"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func retryUntilExpectedPodCount(retryCount int, interval int, manifestNS string, podName string, expectedPodCount int) error {
	for i:=0; i<=retryCount; i++ {
		err := checkPodCount(manifestNS, podName, expectedPodCount)
		if err != nil {
			if err.Error() == "retry" {
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			} else {
				log.Panic("Kubectl listing pods execution error")
			}
		}
		return err
	}
	return errors.New("Out of retries")
}

func checkPodCount(manifestNS string, podName string, expectedPodCount int) error {
	if expectedPodCount == 0 {
		outCountString, err := kubeCmdExec(manifestNS, podName)
		if len(strings.SplitAfter(outCountString, "\n"))==2 {
			expectedPodCountStr := strconv.Itoa(expectedPodCount)
			outCountStringForNoResources := strings.SplitAfter(outCountString, "\n")
			if strings.Compare(expectedPodCountStr, outCountStringForNoResources[1]) == 0 {
				return err
		        }
	        }
		return errors.New("retry")
	} else {
		outCountString, err := kubeCmdExec(manifestNS, podName)
		if strings.Compare(strconv.Itoa(expectedPodCount), outCountString) == 0 {
			return err
                }
		return errors.New("retry")
	}
}

func kubeCmdExec(manifestNS string, podName string) (string, error) {
	kubecmd := "kubectl get pods -n " + manifestNS + " --field-selector status.phase=Running | grep " + podName + "| wc -l"
        outCount, err := exec.Command("/bin/bash", "-c", kubecmd).CombinedOutput()
	return strings.TrimSpace(string(outCount)), err
}
