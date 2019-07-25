package e2e

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func WaitForPodsRunningState(manifestNS string, expectedPodCount int) error {
	if expectedPodCount == 0 {
		for now := time.Now(); ; {
			outCountString, err := kubeCmdExec(manifestNS)
                        if err != nil {
                                log.Panic("Kubectl listing pods execution error")
				return err
                        }
			if len(strings.SplitAfter(outCountString, "\n"))==2 {
				expectedPodCountStr := strconv.Itoa(expectedPodCount)
				outCountStringForNoResources := strings.SplitAfter(outCountString, "\n")
				if strings.Compare(expectedPodCountStr, outCountStringForNoResources[1]) == 0 {
					break
				}
			}
			timeout(now, 20, "waiting for pods to result in Running State")
		}
	} else {
		expectedPodCount = expectedPodCount + 1
		for now := time.Now(); ; {
			outCountString, err := kubeCmdExec(manifestNS)
			if err != nil {
				log.Panic("Kubectl listing pods execution error")
				return err
			}
			if strings.Compare(strconv.Itoa(expectedPodCount), outCountString) == 0 {
				break
			}
			timeout(now, 20, "waiting for pods to result in Running State")
		}
	}
	return nil
}

func kubeCmdExec(manifestNS string) (string, error) {
	kubecmd := "kubectl get pods -n " + manifestNS + " --field-selector status.phase=Running | wc -l"
        outCount, err := exec.Command("/bin/bash", "-c", kubecmd).CombinedOutput()
	return strings.TrimSpace(string(outCount)), err
}

func timeout(currentTime time.Time, sec int, msg string) {
	if time.Since(currentTime) > time.Second*time.Duration(sec) {
		log.Panic("Timed-out:", sec, "seconds, ", msg)
        }
}
