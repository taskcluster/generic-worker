// +build !docker

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/xeipuuv/gojsonschema"
)

func (task *TaskRun) validatePayload() *CommandExecutionError {
	jsonPayload := task.Definition.Payload
	log.Printf("JSON payload: %s", jsonPayload)
	schemaLoader := gojsonschema.NewStringLoader(taskPayloadSchema())
	docLoader := gojsonschema.NewStringLoader(string(jsonPayload))
	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return MalformedPayloadError(err)
	}
	if !result.Valid() {
		task.Error("TASK FAIL since the task payload is invalid. See errors:")
		for _, desc := range result.Errors() {
			task.Errorf("- %s", desc)
		}
		// Dealing with Invalid Task Payloads
		// ----------------------------------
		// If the task payload is malformed or invalid, keep in mind that the
		// queue doesn't validate the contents of the `task.payload` property,
		// the worker may resolve the current run by reporting an exception.
		// When reporting an exception, using `tcqueue.ReportException` the
		// worker should give a `reason`. If the worker is unable execute the
		// task specific payload/code/logic, it should report exception with
		// the reason `malformed-payload`.
		//
		// This can also be used if an external resource that is referenced in
		// a declarative nature doesn't exist. Generally, it should be used if
		// we can be certain that another run of the task will have the same
		// result. This differs from `tcqueue.ReportFailed` in the sense that we
		// report a failure if the task specific code failed.
		//
		// Most tasks includes a lot of declarative steps, such as poll a
		// docker image, create cache folder, decrypt encrypted environment
		// variables, set environment variables and etc. Clearly, if decryption
		// of environment variables fail, there is no reason to retry the task.
		// Nor can it be said that the task failed, because the error wasn't
		// cause by execution of Turing complete code.
		//
		// If however, we run some executable code referenced in `task.payload`
		// and the code crashes or exists non-zero, then the task is said to be
		// failed. The difference is whether or not the unexpected behavior
		// happened before or after the execution of task specific Turing
		// complete code.
		return MalformedPayloadError(fmt.Errorf("Validation of payload failed for task %v", task.TaskID))
	}
	err = json.Unmarshal(jsonPayload, &task.Payload)
	if err != nil {
		return MalformedPayloadError(err)
	}
	for _, artifact := range task.Payload.Artifacts {
		// The default artifact expiry is task expiry, but is only applied when
		// the task artifacts are resolved. We intentionally don't modify
		// task.Payload otherwise it no longer reflects the real data defined
		// in the task.
		if !time.Time(artifact.Expires).IsZero() {
			// Don't be too strict: allow 1s discrepancy to account for
			// possible timestamp rounding on upstream systems
			if time.Time(artifact.Expires).Add(time.Second).Before(time.Time(task.Definition.Deadline)) {
				return MalformedPayloadError(fmt.Errorf("Malformed payload: artifact '%v' expires before task deadline (%v is before %v)", artifact.Path, artifact.Expires, task.Definition.Deadline))
			}
			// Don't be too strict: allow 1s discrepancy to account for
			// possible timestamp rounding on upstream systems
			if time.Time(artifact.Expires).After(time.Time(task.Definition.Expires).Add(time.Second)) {
				return MalformedPayloadError(fmt.Errorf("Malformed payload: artifact '%v' expires after task expiry (%v is after %v)", artifact.Path, artifact.Expires, task.Definition.Expires))
			}
		}
	}
	return nil
}
