/*
Copyright 2017 The Kubernetes Authors.

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

package connection

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"google.golang.org/grpc"
)

// CSIConnection is gRPC connection to a remote CSI driver and abstracts all
// CSI calls.
type CSIConnection interface {
	// GetDriverName returns driver name as discovered by GetPluginInfo() gRPC
	// call.
	GetDriverName(ctx context.Context) (string, error)

	// IsAttachRequired returns a boolean that says whether the driver
	// requires attachment of volumes.
	IsAttachRequired(ctx context.Context) (bool, error)

	// Close the connection
	Close() error
}

type csiConnection struct {
	conn *grpc.ClientConn
}

var (
	_ CSIConnection = &csiConnection{}
)

func NewConnection(
	address string) (CSIConnection, error) {
	conn, err := connection.Connect(address)
	if err != nil {
		return nil, err
	}
	return &csiConnection{
		conn: conn,
	}, nil
}

func (c *csiConnection) GetDriverName(ctx context.Context) (string, error) {
	client := csi.NewIdentityClient(c.conn)

	req := csi.GetPluginInfoRequest{}

	rsp, err := client.GetPluginInfo(ctx, &req)
	if err != nil {
		return "", err
	}
	name := rsp.GetName()
	if name == "" {
		return "", fmt.Errorf("name is empty")
	}
	return name, nil
}

func (c *csiConnection) IsAttachRequired(ctx context.Context) (bool, error) {
	client := csi.NewControllerClient(c.conn)

	req := csi.ControllerGetCapabilitiesRequest{}

	rsp, err := client.ControllerGetCapabilities(ctx, &req)
	if err != nil {
		return false, err
	}

	caps := rsp.GetCapabilities()

	for _, cap := range caps {
		if cap.GetRpc().GetType() == csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME {
			return true, nil
		}
	}

	return false, nil
}

func (c *csiConnection) Close() error {
	return c.conn.Close()
}
