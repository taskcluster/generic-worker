// This source code file is AUTO-GENERATED by github.com/taskcluster/jsonschema2go

package tcnotifyevents

import (
	"encoding/json"
)

type (
	// This can be pretty much anything you want it to be.
	//
	// See https://taskcluster-staging.net/schemas/notify/v1/notification-message.json#
	NotificationMessage struct {

		// Arbitrary message.
		//
		// Additional properties allowed
		//
		// See https://taskcluster-staging.net/schemas/notify/v1/notification-message.json#/properties/message
		Message json.RawMessage `json:"message"`

		// Message version
		//
		// Possible values:
		//   * 1
		//
		// See https://taskcluster-staging.net/schemas/notify/v1/notification-message.json#/properties/version
		Version int64 `json:"version,omitempty"`
	}
)
