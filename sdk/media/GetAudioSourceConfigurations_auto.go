// Code generated : DO NOT EDIT.
// Copyright (c) 2022 Jean-Francois SMIGIELSKI
// Distributed under the MIT License

package media

import (
	"context"
	"github.com/juju/errors"
	"github.com/ljhljh127/onvif"
	"github.com/ljhljh127/onvif/sdk"
	"github.com/ljhljh127/onvif/media"
)

// Call_GetAudioSourceConfigurations forwards the call to dev.CallMethod() then parses the payload of the reply as a GetAudioSourceConfigurationsResponse.
func Call_GetAudioSourceConfigurations(ctx context.Context, dev *onvif.Device, request media.GetAudioSourceConfigurations) (media.GetAudioSourceConfigurationsResponse, error) {
	type Envelope struct {
		Header struct{}
		Body   struct {
			GetAudioSourceConfigurationsResponse media.GetAudioSourceConfigurationsResponse
		}
	}
	var reply Envelope
	if httpReply, err := dev.CallMethod(request); err != nil {
		return reply.Body.GetAudioSourceConfigurationsResponse, errors.Annotate(err, "call")
	} else {
		err = sdk.ReadAndParse(ctx, httpReply, &reply, "GetAudioSourceConfigurations")
		return reply.Body.GetAudioSourceConfigurationsResponse, errors.Annotate(err, "reply")
	}
}
