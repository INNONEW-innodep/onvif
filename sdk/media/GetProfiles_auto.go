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

// Call_GetProfiles forwards the call to dev.CallMethod() then parses the payload of the reply as a GetProfilesResponse.
func Call_GetProfiles(ctx context.Context, dev *onvif.Device, request media.GetProfiles) (media.GetProfilesResponse, error) {
	type Envelope struct {
		Header struct{}
		Body   struct {
			GetProfilesResponse media.GetProfilesResponse
		}
	}
	var reply Envelope
	if httpReply, err := dev.CallMethod(request); err != nil {
		return reply.Body.GetProfilesResponse, errors.Annotate(err, "call")
	} else {
		err = sdk.ReadAndParse(ctx, httpReply, &reply, "GetProfiles")
		return reply.Body.GetProfilesResponse, errors.Annotate(err, "reply")
	}
}
