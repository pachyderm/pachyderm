package csi

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
)

const (
	// pluginName is the name this plugin uses when identifying itself via the
	// GetPluginInfo API endpoint. The CSI spec provides examples in which plugin
	// names are reversed domain names, but more recent examples of CSI plugins
	// use regular domain names (see e.g.
	// http://pach-links.s3-website-us-west-1.amazonaws.com/csi-example)
	pluginName = "csi.pachyderm.com"

	// pluginVersion indicates the version of this plugin binary. The Pachyderm
	// CSI plugin is versioned separately from Pachyderm itself in case we need
	// to release bug fixes for the CSI plugin without releasing pachd itself
	pluginVersion = "v1.0"
)

type identitySvc struct{}

func (s *identitySvc) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Name:          pluginName,
		VendorVersion: pluginVersion,
	}, nil
}

func (s *identitySvc) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				// Advertise the CONTROLLER_SERVICE capability, as we do have a minimal
				// controller service
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}

func (s *identitySvc) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{
		Ready: &wrappers.BoolValue{Value: true},
	}, nil
}
