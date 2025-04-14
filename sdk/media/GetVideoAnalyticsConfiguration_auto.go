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

// Call_GetVideoAnalyticsConfiguration forwards the call to dev.CallMethod() then parses the payload of the reply as a GetVideoAnalyticsConfigurationResponse.
func Call_GetVideoAnalyticsConfiguration(ctx context.Context, dev *onvif.Device, request media.GetVideoAnalyticsConfiguration) (media.GetVideoAnalyticsConfigurationResponse, error) {
	type Envelope struct {
		Header struct{}
		Body   struct {
			GetVideoAnalyticsConfigurationResponse media.GetVideoAnalyticsConfigurationResponse
		}
	}
	var reply Envelope
	if httpReply, err := dev.CallMethod(request); err != nil {
		return reply.Body.GetVideoAnalyticsConfigurationResponse, errors.Annotate(err, "call")
	} else {
		err = sdk.ReadAndParse(ctx, httpReply, &reply, "GetVideoAnalyticsConfiguration")
		return reply.Body.GetVideoAnalyticsConfigurationResponse, errors.Annotate(err, "reply")
	}
}
