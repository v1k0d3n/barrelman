package e2e

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func WaitForPodsRunningState(manifestNS string, expectedPodCount int) error {
	expectedPodCount = expectedPodCount+1
	for now := time.Now(); ; {
		kubecmd := "kubectl get pods -n " + manifestNS + " --field-selector status.phase=Running | wc -l"
		outCount, err := exec.Command("/bin/bash", "-c", kubecmd).CombinedOutput()
		if err != nil {
			return err
		}
		if strings.Compare(strconv.Itoa(expectedPodCount), strings.TrimSpace(string(outCount))) == 0 {
			break
		}
		timeout(now, 20, "waiting for pods to result in Running State")
	}
	return nil
}

func timeout(currentTime time.Time, sec int, msg string) {
	if time.Since(currentTime) > time.Second*time.Duration(sec) {
		log.Panic("Timed-out:", sec, "seconds, ", msg)
        }
}
