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

// Call_GetOSDOptions forwards the call to dev.CallMethod() then parses the payload of the reply as a GetOSDOptionsResponse.
func Call_GetOSDOptions(ctx context.Context, dev *onvif.Device, request media.GetOSDOptions) (media.GetOSDOptionsResponse, error) {
	type Envelope struct {
		Header struct{}
		Body   struct {
			GetOSDOptionsResponse media.GetOSDOptionsResponse
		}
	}
	var reply Envelope
	if httpReply, err := dev.CallMethod(request); err != nil {
		return reply.Body.GetOSDOptionsResponse, errors.Annotate(err, "call")
	} else {
		err = sdk.ReadAndParse(ctx, httpReply, &reply, "GetOSDOptions")
		return reply.Body.GetOSDOptionsResponse, errors.Annotate(err, "reply")
	}
}
