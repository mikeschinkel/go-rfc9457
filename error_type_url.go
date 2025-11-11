package rfc9457

type ErrorTypeURI string

// Error type URLs (RFC 9457 requires absolute URIs)
const (
	ErrorTypeRootURI  ErrorTypeURI = "https://schema.xmlui.org/errors"
	TestServerAPIPath ErrorTypeURI = "/test-server/api"
)
const (
	InvalidParameterErrorType    ErrorTypeURI = uri + path + "/validation/invalid-parameter-type"
	ConstraintViolationErrorType ErrorTypeURI = uri + path + "/validation/constraint-violation"
	MissingParametersErrorType   ErrorTypeURI = uri + path + "/validation/missing-required-parameters"
	CurrentlyUnhandledErrorType  ErrorTypeURI = uri + path + "/validation/currently-unhandled"
	UnauthorizedErrorType        ErrorTypeURI = uri + path + "/validation/unauthorized"
	InvalidBodyFormatErrorType   ErrorTypeURI = uri + path + "/validation/invalid-body-format"
	InvalidURLFormatErrorType    ErrorTypeURI = uri + path + "/validation/invalid-url-format"
	InvalidURLParameterErrorType ErrorTypeURI = uri + path + "/validation/invalid-url-parameter"
	InvalidDBQueryErrorType      ErrorTypeURI = uri + path + "/validation/invalid-database-query"
	InternalServerErrorType      ErrorTypeURI = uri + path + "/server/internal"
	EndpointNotMatchedErrorType  ErrorTypeURI = uri + path + "/routing/endpoint-not-matched"
	NoResultsErrorType           ErrorTypeURI = uri + path + "/database/no-results"

	CardinalityMismatchErrorType ErrorTypeURI = uri + path + "/routing/cardinality-mismatch"
	MethodNotAllowedErrorType    ErrorTypeURI = uri + path + "/routing/method-not-allowed"
	QueryFailedErrorType         ErrorTypeURI = uri + path + "/database/query-failed"
)

// Private convenience constants
const (
	uri  = ErrorTypeRootURI
	path = TestServerAPIPath
)
