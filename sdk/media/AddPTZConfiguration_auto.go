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

// Call_AddPTZConfiguration forwards the call to dev.CallMethod() then parses the payload of the reply as a AddPTZConfigurationResponse.
func Call_AddPTZConfiguration(ctx context.Context, dev *onvif.Device, request media.AddPTZConfiguration) (media.AddPTZConfigurationResponse, error) {
	type Envelope struct {
		Header struct{}
		Body   struct {
			AddPTZConfigurationResponse media.AddPTZConfigurationResponse
		}
	}
	var reply Envelope
	if httpReply, err := dev.CallMethod(request); err != nil {
		return reply.Body.AddPTZConfigurationResponse, errors.Annotate(err, "call")
	} else {
		err = sdk.ReadAndParse(ctx, httpReply, &reply, "AddPTZConfiguration")
		return reply.Body.AddPTZConfigurationResponse, errors.Annotate(err, "reply")
	}
}
