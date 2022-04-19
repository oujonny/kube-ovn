package ovs

// SetInterfaceBandwidth set ingress/egress qos for given pod, annotation values are for node/pod
// but ingress/egress parameters here are from the point of ovs port/interface view, so reverse input parameters when call func SetInterfaceBandwidth
func SetInterfaceBandwidth(podName, podNamespace, iface, ingress, egress, podPriority string) error {
	// TODO
	return nil
}

func ClearHtbQosQueue(podName, podNamespace, iface string) error {
	//TODO
	return nil
}

func IsHtbQos(iface string) (bool, error) {
	// TODO
	return false, nil
}

func SetHtbQosQueueRecord(podName, podNamespace, iface, priority string, maxRateBPS int, queueIfaceUidMap map[string]string) (string, error) {
	//TODO
	return "", nil
}

// SetPodQosPriority set qos to this pod port.
func SetPodQosPriority(podName, podNamespace, ifaceID, priority string, qosIfaceUidMap, queueIfaceUidMap map[string]string) error {
	//TODO
	return nil
}

// SetQosQueueBinding set qos related to queue record.
func SetQosQueueBinding(podName, podNamespace, ifName, iface, queueUid string, qosIfaceUidMap map[string]string) error {
	// TODO
	return nil
}

// The latency value expressed in us.
func SetNetemQos(podName, podNamespace, iface, latency, limit, loss string) error {
	// TODO
	return nil
}

func IsUserspaceDataPath() (is bool, err error) {
	dp, err := ovsFind("bridge", "datapath_type", "name=br-int")
	if err != nil {
		return false, err
	}
	return len(dp) > 0 && dp[0] == "netdev", nil
}
