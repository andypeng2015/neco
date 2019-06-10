package dctest

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cybozu-go/sabakan/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

func isNodeNumEqual(num int) error {
	stdout, stderr, err := execAt(boot0, "kubectl", "get", "nodes", "-o", "json")
	if err != nil {
		return fmt.Errorf("kubectl get nodes -o json failed. err: %v, stdout: %s, stderr: %s", err, stdout, stderr)
	}
	var nl corev1.NodeList
	err = json.Unmarshal(stdout, &nl)
	if err != nil {
		return fmt.Errorf("unmarshal failed. err: %v", err)
	}
	if len(nl.Items) != num {
		return fmt.Errorf("cluster node should be %d, but %d", num, len(nl.Items))
	}
	return nil
}

// TestPartsFailure test parts failure scenario
func TestPartsFailure() {
	var targetIP string

	It("transition machine state to unhealthy", func() {
		stdout, stderr, err := execAt(boot0, "ckecli", "cluster", "get")
		Expect(err).ShouldNot(HaveOccurred(), "stderr=%s", stderr)

		cluster := new(ckeCluster)
		err = yaml.Unmarshal(stdout, cluster)
		Expect(err).ShouldNot(HaveOccurred())

		for _, n := range cluster.Nodes {
			if !n.ControlPlane {
				targetIP = n.Address
				break
			}
		}
		Expect(targetIP).NotTo(Equal(""))

		By("copying dummy warning redfish data to " + targetIP + "@" + time.Now().String())
		Eventually(func() error {
			return copyDummyWarningRedfishDataToWorker(targetIP)
		}).Should(Succeed())

		By("checking machine state" + "@" + time.Now().String())
		Eventually(func() error {
			stdout, _, err := execAt(boot0, "sabactl", "machines", "get", "--ipv4", targetIP)
			if err != nil {
				return err
			}
			var machines []sabakan.Machine
			err = json.Unmarshal(stdout, &machines)
			if err != nil {
				return err
			}
			for _, m := range machines {
				if m.Status.State.String() != "unhealthy" {
					return errors.New(m.Spec.Serial + " is not unhealthy:" + m.Status.State.String())
				}
			}
			return nil
		}).Should(Succeed())

		By("checking the number of cluster nodes" + "@" + time.Now().String())
		Eventually(func() error {
			return isNodeNumEqual(6)
		}).Should(Succeed())
	})
}