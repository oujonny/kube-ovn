package framework

import (
	"context"
	"errors"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/gomega"

	apiv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	v1 "github.com/kubeovn/kube-ovn/pkg/client/clientset/versioned/typed/kubeovn/v1"
	"github.com/kubeovn/kube-ovn/pkg/util"
)

// IpClient is a struct for Ip client.
type IpClient struct {
	f *Framework
	v1.IPInterface
}

func (f *Framework) IpClient() *IpClient {
	return &IpClient{
		f:           f,
		IPInterface: f.KubeOVNClientSet.KubeovnV1().IPs(),
	}
}

func (c *IpClient) Get(name string) *apiv1.IP {
	Ip, err := c.IPInterface.Get(context.TODO(), name, metav1.GetOptions{})
	ExpectNoError(err)
	return Ip.DeepCopy()
}

// Create creates a new Ip according to the framework specifications
func (c *IpClient) Create(Ip *apiv1.IP) *apiv1.IP {
	Ip, err := c.IPInterface.Create(context.TODO(), Ip, metav1.CreateOptions{})
	ExpectNoError(err, "Error creating Ip")
	return Ip.DeepCopy()
}

// CreateSync creates a new IP according to the framework specifications, and waits for it to be ready.
func (c *IpClient) CreateSync(Ip *apiv1.IP) *apiv1.IP {
	Ip = c.Create(Ip)
	ExpectTrue(c.WaitToBeReady(Ip.Name, timeout))
	// Get the newest IP after it becomes ready
	return c.Get(Ip.Name).DeepCopy()
}

// WaitToBeReady returns whether the IP is ready within timeout.
func (c *IpClient) WaitToBeReady(name string, timeout time.Duration) bool {
	Logf("Waiting up to %v for IP %s to be ready", timeout, name)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(poll) {
		ip := c.Get(name)
		if ip.Spec.V4IPAddress != "" || ip.Spec.V6IPAddress != "" {
			Logf("IP %s is ready ", name)
			return true
		}
		Logf("IP %s is not ready ", name)
	}
	Logf("IP %s was not ready within %v", name, timeout)
	return false
}

// Patch patches the Ip
func (c *IpClient) Patch(original, modified *apiv1.IP, timeout time.Duration) *apiv1.IP {
	patch, err := util.GenerateMergePatchPayload(original, modified)
	ExpectNoError(err)

	var patchedIp *apiv1.IP
	err = wait.PollUntilContextTimeout(context.Background(), 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		p, err := c.IPInterface.Patch(ctx, original.Name, types.MergePatchType, patch, metav1.PatchOptions{}, "")
		if err != nil {
			return handleWaitingAPIError(err, false, "patch Ip %q", original.Name)
		}
		patchedIp = p
		return true, nil
	})
	if err == nil {
		return patchedIp.DeepCopy()
	}

	if errors.Is(err, context.DeadlineExceeded) {
		Failf("timed out while retrying to patch IP %s", original.Name)
	}
	Failf("error occurred while retrying to patch IP %s: %v", original.Name, err)

	return nil
}

// Delete deletes a Ip if the Ip exists
func (c *IpClient) Delete(name string) {
	err := c.IPInterface.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		Failf("Failed to delete Ip %q: %v", name, err)
	}
}

// DeleteSync deletes the IP and waits for the IP to disappear for `timeout`.
// If the IP doesn't disappear before the timeout, it will fail the test.
func (c *IpClient) DeleteSync(name string) {
	c.Delete(name)
	gomega.Expect(c.WaitToDisappear(name, 2*time.Second, timeout)).To(gomega.Succeed(), "wait for ovn eip %q to disappear", name)
}

// WaitToDisappear waits the given timeout duration for the specified IP to disappear.
func (c *IpClient) WaitToDisappear(name string, interval, timeout time.Duration) error {
	err := framework.Gomega().Eventually(context.Background(), framework.HandleRetry(func(ctx context.Context) (*apiv1.IP, error) {
		ip, err := c.IPInterface.Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return ip, err
	})).WithTimeout(timeout).Should(gomega.BeNil())
	if err != nil {
		return fmt.Errorf("expected IP %s to not be found: %w", name, err)
	}
	return nil
}

func MakeIp(name, ns, subnet string) *apiv1.IP {
	// pod ip name should including: pod name and namespace
	// node ip name: only node name
	Ip := &apiv1.IP{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiv1.IPSpec{
			Namespace: ns,
			Subnet:    subnet,
		},
	}
	return Ip
}
