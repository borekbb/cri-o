//go:build !windows
// +build !windows

/*
Copyright 2021 Mirantis

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cni

import (
	"fmt"

	"github.com/Mirantis/cri-dockerd/config"

	"github.com/containernetworking/cni/libcni"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"

	"github.com/Mirantis/cri-dockerd/network"
)

func getLoNetwork(binDirs []string) *cniNetwork {
	loConfig, err := libcni.ConfListFromBytes([]byte(`{
  "cniVersion": "0.2.0",
  "name": "cni-loopback",
  "plugins":[{
    "type": "loopback"
  }]
}`))
	if err != nil {
		// The hardcoded config above should always be valid and unit tests will
		// catch this
		panic(err)
	}
	loNetwork := &cniNetwork{
		name:          "lo",
		NetworkConfig: loConfig,
		CNIConfig:     &libcni.CNIConfig{Path: binDirs},
	}

	return loNetwork
}

func (plugin *cniNetworkPlugin) platformInit() error {
	var err error
	plugin.nsenterPath, err = plugin.execer.LookPath("nsenter")
	if err != nil {
		return err
	}
	return nil
}

// Also fix the runtime's call to Status function to be done only in the case that the IP is lost, no need to do periodic calls
func (plugin *cniNetworkPlugin) GetPodNetworkStatus(
	namespace string,
	name string,
	id config.ContainerID,
) (*network.PodNetworkStatus, error) {
	netnsPath, err := plugin.host.GetNetNS(id.ID)
	if err != nil {
		return nil, fmt.Errorf("CNI failed to retrieve network namespace path: %v", err)
	}
	if netnsPath == "" {
		return nil, fmt.Errorf(
			"cannot find the network namespace, skipping pod network status for container %q",
			id,
		)
	}

	ips, err := network.GetPodIPs(
		plugin.execer,
		plugin.nsenterPath,
		netnsPath,
		network.DefaultInterfaceName,
	)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf(
			"cannot find pod IPs in the network namespace, skipping pod network status for container %q",
			id,
		)
	}

	return &network.PodNetworkStatus{
		IP:  ips[0],
		IPs: ips,
	}, nil
}

// buildDNSCapabilities builds cniDNSConfig from runtimeapi.DNSConfig.
func buildDNSCapabilities(dnsConfig *runtimeapi.DNSConfig) *cniDNSConfig {
	return nil
}
