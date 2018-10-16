package main

import "github.com/taskcluster/slugid-go/slugid"

var (
	// all tests can share taskGroupId so we can view all test tasks in same
	// graph later for troubleshooting
	taskGroupID = slugid.Nice()
)
