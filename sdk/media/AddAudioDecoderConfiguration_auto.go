// Code generated : DO NOT EDIT.
// Copyright (c) 2022 Jean-Francois SMIGIELSKI
// Distributed under the MIT License

package media

import (
	"context"
	"github.com/juju/errors"
	"github.com/ljhljh127/onvif"
	"github.com/ljhljh127/onvifvif/sdk"
	"github.com/ljhljh127/onvifvif/media"
)

// Call_AddAudioDecoderConfiguration forwards the call to dev.CallMethod() then parses the payload of the reply as a AddAudioDecoderConfigurationResponse.
func Call_AddAudioDecoderConfiguration(ctx context.Context, dev *onvif.Device, request media.AddAudioDecoderConfiguration) (media.AddAudioDecoderConfigurationResponse, error) {
	type Envelope struct {
		Header struct{}
		Body   struct {
			AddAudioDecoderConfigurationResponse media.AddAudioDecoderConfigurationResponse
		}
	}
	var reply Envelope
	if httpReply, err := dev.CallMethod(request); err != nil {
		return reply.Body.AddAudioDecoderConfigurationResponse, errors.Annotate(err, "call")
	} else {
		err = sdk.ReadAndParse(ctx, httpReply, &reply, "AddAudioDecoderConfiguration")
		return reply.Body.AddAudioDecoderConfigurationResponse, errors.Annotate(err, "reply")
	}
}
