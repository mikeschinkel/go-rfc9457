package rfc9457

type ErrorTypeURI string

// Error type URLs (RFC 9457 requires absolute URIs)
const (
	ErrorTypeRootURI  ErrorTypeURI = "https://schema.xmlui.org/errors"
	TestServerAPIPath              = "/test-server/api"
)
const (
	InvalidParameterErrorType    = uri + path + "/validation/invalid-parameter-type"
	ConstraintViolationErrorType = uri + path + "/validation/constraint-violation"
	MissingParametersErrorType   = uri + path + "/validation/missing-required-parameters"
	CurrentlyUnhandledErrorType  = uri + path + "/validation/currently-unhandled"
	UnauthorizedErrorType        = uri + path + "/validation/unauthorized"
	InvalidBodyFormatErrorType   = uri + path + "/validation/invalid-body-format"
	InvalidURLFormatErrorType    = uri + path + "/validation/invalid-url-format"
	InvalidURLParameterErrorType = uri + path + "/validation/invalid-url-parameter"
	InvalidDBQueryErrorType      = uri + path + "/validation/invalid-database-query"
	InternalServerErrorType      = uri + path + "/server/internal"
	EndpointNotMatchedErrorType  = uri + path + "/routing/endpoint-not-matched"
	NoResultsErrorType           = uri + path + "/database/no-results"

	CardinalityMismatchErrorType = uri + path + "/routing/cardinality-mismatch"
	MethodNotAllowedErrorType    = uri + path + "/routing/method-not-allowed"
	QueryFailedErrorType         = uri + path + "/database/query-failed"
)

// Private convenience constants
const (
	uri  = ErrorTypeRootURI
	path = ErrorTypeURI(TestServerAPIPath)
)
