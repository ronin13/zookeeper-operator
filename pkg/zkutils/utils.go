package zkutils

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	"k8s.io/apimachinery/pkg/labels"

	wnohangv1alpha1 "github.com/ronin13/zookeeper-operator/pkg/apis/wnohang/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	log = logf.Log.WithName("Zookeeper Utils")
	// ErrInvalidInput is used invalid paramter is supplied.
	ErrInvalidInput = errors.New("Invalid parameter supplied")
)

// IsZKReady is used to determine if all nodes in cluster are Ready
// by definition of 'readiness' (either follower or leader)
func IsZKReady(zkClient client.Client, instance *wnohangv1alpha1.Zookeeper) (bool, error) {
	var isReady bool
	var stateCommand string
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"app": instance.Name})
	listOps := &client.ListOptions{
		Namespace:     instance.Namespace,
		LabelSelector: labelSelector,
	}
	err := zkClient.List(context.TODO(), listOps, podList)
	if err != nil {
		log.Error(err, "Failed to list pods")
		return false, err
	}
	isReady = true
	log.Info("Waiting for pods to become ready")
	for _, pod := range podList.Items {

		isReady = isReady && podutil.IsPodReady(&pod)
		if isReady {
			stateCommand = fmt.Sprintf("echo mntr | nc %s 2181 | grep zk_server_state | grep -q leader", pod.Status.PodIP)
			_, err := exec.Command("sh", "-c", stateCommand).Output()
			if err == nil {
				log.Info(fmt.Sprintf("Found leader with name %s", pod.Name))
			}
		}

	}
	return isReady, nil

}

// GetZooIds returns Zookeeper server id corresponding to number of nodes
func GetZooIds(numNodes uint32) (string, error) {
	if numNodes == 0 {
		return "", ErrInvalidInput
	}
	zooIDs := "1"

	for nd := 2; nd <= int(numNodes); nd++ {
		zooIDs += "," + strconv.Itoa(nd)
	}

	return zooIDs, nil
}
