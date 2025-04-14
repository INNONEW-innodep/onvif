// Code generated : DO NOT EDIT.
// Copyright (c) 2022 Jean-Francois SMIGIELSKI
// Distributed under the MIT License

package ptz

import (
	"context"
	"github.com/juju/errors"
	"github.com/ljhljh127/onvif"
	"github.com/ljhljh127/onvifvif/sdk"
	"github.com/ljhljh127/onvifvif/ptz"
)

// Call_SetPreset forwards the call to dev.CallMethod() then parses the payload of the reply as a SetPresetResponse.
func Call_SetPreset(ctx context.Context, dev *onvif.Device, request ptz.SetPreset) (ptz.SetPresetResponse, error) {
	type Envelope struct {
		Header struct{}
		Body   struct {
			SetPresetResponse ptz.SetPresetResponse
		}
	}
	var reply Envelope
	if httpReply, err := dev.CallMethod(request); err != nil {
		return reply.Body.SetPresetResponse, errors.Annotate(err, "call")
	} else {
		err = sdk.ReadAndParse(ctx, httpReply, &reply, "SetPreset")
		return reply.Body.SetPresetResponse, errors.Annotate(err, "reply")
	}
}
