// Copyright (C) 2019 Cisco Systems Inc.
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

package main

import (
	"flag"
	"fmt"

	"github.com/projectcalico/vpp-dataplane/calico-vpp-agent/cni/storage"
	"github.com/projectcalico/vpp-dataplane/calico-vpp-agent/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	var fname string
	cniServerStateFile := fmt.Sprintf("%s%d", config.CniServerStateFile, storage.CniServerStateFileVersion)
	flag.StringVar(&fname, "f", cniServerStateFile, "Pod state path")
	flag.Parse()

	st, err := storage.LoadCniServerState(fname)
	if err != nil {
		log.Errorf("LoadCniServerState errored: %v", err)
		return
	}
	for i, s := range st {
		log.Infof("-------- Elem %d--------\n%s", i, s.FullString())
	}
	log.Infof("%d Elts", len(st))
}
