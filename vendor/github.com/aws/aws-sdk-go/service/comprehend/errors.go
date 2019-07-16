// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

package comprehend

const (

	// ErrCodeBatchSizeLimitExceededException for service response error code
	// "BatchSizeLimitExceededException".
	//
	// The number of documents in the request exceeds the limit of 25. Try your
	// request again with fewer documents.
	ErrCodeBatchSizeLimitExceededException = "BatchSizeLimitExceededException"

	// ErrCodeInternalServerException for service response error code
	// "InternalServerException".
	//
	// An internal server error occurred. Retry your request.
	ErrCodeInternalServerException = "InternalServerException"

	// ErrCodeInvalidFilterException for service response error code
	// "InvalidFilterException".
	//
	// The filter specified for the ListDocumentClassificationJobs operation is
	// invalid. Specify a different filter.
	ErrCodeInvalidFilterException = "InvalidFilterException"

	// ErrCodeInvalidRequestException for service response error code
	// "InvalidRequestException".
	//
	// The request is invalid.
	ErrCodeInvalidRequestException = "InvalidRequestException"

	// ErrCodeJobNotFoundException for service response error code
	// "JobNotFoundException".
	//
	// The specified job was not found. Check the job ID and try again.
	ErrCodeJobNotFoundException = "JobNotFoundException"

	// ErrCodeKmsKeyValidationException for service response error code
	// "KmsKeyValidationException".
	//
	// The KMS customer managed key (CMK) entered cannot be validated. Verify the
	// key and re-enter it.
	ErrCodeKmsKeyValidationException = "KmsKeyValidationException"

	// ErrCodeResourceInUseException for service response error code
	// "ResourceInUseException".
	//
	// The specified name is already in use. Use a different name and try your request
	// again.
	ErrCodeResourceInUseException = "ResourceInUseException"

	// ErrCodeResourceLimitExceededException for service response error code
	// "ResourceLimitExceededException".
	//
	// The maximum number of recognizers per account has been exceeded. Review the
	// recognizers, perform cleanup, and then try your request again.
	ErrCodeResourceLimitExceededException = "ResourceLimitExceededException"

	// ErrCodeResourceNotFoundException for service response error code
	// "ResourceNotFoundException".
	//
	// The specified resource ARN was not found. Check the ARN and try your request
	// again.
	ErrCodeResourceNotFoundException = "ResourceNotFoundException"

	// ErrCodeResourceUnavailableException for service response error code
	// "ResourceUnavailableException".
	//
	// The specified resource is not available. Check to see if the resource is
	// in the TRAINED state and try your request again.
	ErrCodeResourceUnavailableException = "ResourceUnavailableException"

	// ErrCodeTextSizeLimitExceededException for service response error code
	// "TextSizeLimitExceededException".
	//
	// The size of the input text exceeds the limit. Use a smaller document.
	ErrCodeTextSizeLimitExceededException = "TextSizeLimitExceededException"

	// ErrCodeTooManyRequestsException for service response error code
	// "TooManyRequestsException".
	//
	// The number of requests exceeds the limit. Resubmit your request later.
	ErrCodeTooManyRequestsException = "TooManyRequestsException"

	// ErrCodeUnsupportedLanguageException for service response error code
	// "UnsupportedLanguageException".
	//
	// Amazon Comprehend can't process the language of the input text. For all custom
	// entity recognition APIs (such as CreateEntityRecognizer), only English is
	// accepted. For most other APIs, Amazon Comprehend accepts only English or
	// Spanish text.
	ErrCodeUnsupportedLanguageException = "UnsupportedLanguageException"
)
