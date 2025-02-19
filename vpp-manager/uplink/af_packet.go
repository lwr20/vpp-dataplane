// Copyright (C) 2020 Cisco Systems Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package uplink

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/projectcalico/vpp-dataplane/vpp-manager/config"
	"github.com/projectcalico/vpp-dataplane/vpp-manager/utils"
	"github.com/projectcalico/vpp-dataplane/vpplink"
	"github.com/projectcalico/vpp-dataplane/vpplink/types"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type AFPacketDriver struct {
	UplinkDriverData
}

func (d *AFPacketDriver) IsSupported(warn bool) bool {
	return true
}

func (d *AFPacketDriver) PreconfigureLinux() error {
	link, err := netlink.LinkByName(d.spec.InterfaceName)
	if err != nil {
		return errors.Wrapf(err, "Error finding link %s", d.spec.InterfaceName)
	}
	err = netlink.SetPromiscOn(link)
	if err != nil {
		return errors.Wrapf(err, "Error set link %s promisc on", d.spec.InterfaceName)
	}
	return nil
}

func (d *AFPacketDriver) RestoreLinux(allInterfacesPhysical bool) {
	if !allInterfacesPhysical {
		err := d.moveInterfaceFromNS(d.spec.InterfaceName)
		if err != nil {
			log.Warnf("Moving uplink back from NS failed %s", err)
		}
	}

	if !d.conf.IsUp {
		return
	}
	// Interface should pop back in root ns once vpp exits
	link, err := utils.SafeSetInterfaceUpByName(d.spec.InterfaceName)
	if err != nil {
		log.Warnf("Error setting %s up: %v", d.spec.InterfaceName, err)
		return
	}

	if !d.conf.PromiscOn {
		log.Infof("Setting promisc off")
		err = netlink.SetPromiscOff(link)
		if err != nil {
			log.Errorf("Error setting link %s promisc off %v", d.spec.InterfaceName, err)
		}
	}

	// Re-add all adresses and routes
	d.restoreLinuxIfConf(link)
}

func (d *AFPacketDriver) CreateMainVppInterface(vpp *vpplink.VppLink, vppPid int) (err error) {
	err = d.moveInterfaceToNS(d.spec.InterfaceName, vppPid)
	if err != nil {
		return errors.Wrap(err, "Moving uplink in NS failed")
	}

	intf := types.AfPacketInterface{
		GenericVppInterface: d.getGenericVppInterface(),
	}
	swIfIndex, err := vpp.CreateAfPacket(&intf)
	if err != nil {
		return errors.Wrapf(err, "Error creating AF_PACKET interface")
	}
	log.Infof("Created AF_PACKET interface %d", swIfIndex)

	if d.spec.IsMain && swIfIndex != config.DataInterfaceSwIfIndex {
		return fmt.Errorf("Created AF_PACKET interface has wrong swIfIndex %d!", swIfIndex)
	}
	d.spec.SwIfIndex = swIfIndex
	err = d.TagMainInterface(vpp, swIfIndex, d.spec.InterfaceName)
	if err != nil {
		return err
	}
	return nil
}

func NewAFPacketDriver(params *config.VppManagerParams, conf *config.LinuxInterfaceState, spec *config.InterfaceSpec) *AFPacketDriver {
	d := &AFPacketDriver{}
	d.name = NATIVE_DRIVER_AF_PACKET
	d.conf = conf
	d.params = params
	d.spec = spec
	return d
}
