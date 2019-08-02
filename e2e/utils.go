package e2e

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

type WrongNumberOfPods struct {
}

func (e *WrongNumberOfPods) Error() string {
	return fmt.Sprintf("Wrong number of pods running in the cluster.")
}

func isRetryableError(e error, retryableErrors []string) bool {
	for _, retryableError := range retryableErrors {
		if strings.Contains(reflect.TypeOf(e).String(), retryableError) {
			return true
		}
	}
	return false
}

func retry(f func() error, retryCount int, interval int, retryableErrors []string) error {
	for i := 0; i <= retryCount; i++ {
		err := f()
		if err != nil {
			if isRetryableError(err, retryableErrors) {
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			} else {
				return fmt.Errorf("Non retryable error returned, %s", err)
			}
		}

		return nil
	}

	return fmt.Errorf("Retry limit exceeded")
}

type KubectlOutput struct {
	Items []interface{}
}

func getPodCount(ns, podName string) (int, error) {
	kubecmd := fmt.Sprintf("kubectl get pods -n %s --field-selector status.phase=Running -o json", ns)
	out, err := exec.Command("/bin/bash", "-c", kubecmd).CombinedOutput()
	if err != nil {
		return -1, fmt.Errorf("Failed to run command, %s", err)
	}

	var kubectlOutput KubectlOutput
	err = json.Unmarshal(out, &kubectlOutput)
	if err != nil {
		return -1, fmt.Errorf("Failed to unmarshal, %s", err)
	}

	return len(kubectlOutput.Items), nil
}

func checkPodCount(ns, podName string, expectedPodCount int) error {
	count, err := getPodCount(ns, podName)
	if err != nil {
		return fmt.Errorf("Failed to get the pod count, %s", err)
	}

	if expectedPodCount != count {
		return &WrongNumberOfPods{}
	}

	return nil
}

