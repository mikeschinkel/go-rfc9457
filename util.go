package rfc9457

import (
	"fmt"
	"strings"
)

// SprintfMany allows printing many name-value pairs starting with a prefix, e.g
//
//	func (r *Response) Error() string {
//		return SprintfMany(r.Title, "\n",
//			"title=%s", r.Title,
//			"error_type=%s", r.Type,
//			"error_detail=%s", r.Detail,
//			"http_status=%d", r.Status,
//			"instance=%s", r.Instance,
//			"parameter=%s", r.Parameter,
//			"expected_type=%s", r.ExpectedType,
//			"received_value=%s", r.ReceivedValue,
//			"error_location=%s", r.Location,
//			"constraint=%v", r.Constraint,
//			"suggestion=%s", r.Suggestion,
//			"validation_errors=%v", r.ValidationErrors,
//			"received_value=%s", r.ReceivedValue,
//		)
//	}
func SprintfMany(prefix, sep string, params ...any) (s string) {
	var sb strings.Builder
	var args []any
	if len(params) == 0 {
		s = prefix
		goto end
	}
	if len(params)%2 != 0 {
		panic(fmt.Sprintf("SprintfMany() needs an even number of arguments: %v", params))
	}
	sb.WriteString(prefix)
	sb.WriteString(sep)
	args = make([]any, 0, len(params)/2+1)
	for i := 0; i < len(params)-1; i += 2 {
		sb.WriteString(params[i].(string))
		sb.WriteString(sep)
		args = append(args, params[i+1])
	}
	s = fmt.Sprintf(sb.String(), args...)
end:
	return s
}
