package rfc9457

import (
	"encoding/json"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"fmt"
	"net/http"
	"reflect"
)

var _ ResponsePayload = (*Response)(nil)
var _ error = (*Response)(nil)

var ResponseArchetype = reflect.TypeOf((*Response)(nil))

type Response struct {
	Type       ErrorTypeURI `json:"type"`
	Title      string       `json:"title"`
	Status     int          `json:"status"`
	Detail     string       `json:"detail,omitempty"`
	Instance   string       `json:"instance,omitempty"`
	Extensions []Extension  `json:"extensions,omitempty"`
}

func (r *Response) AddExtension(ext Extension) {
	r.Extensions = append(r.Extensions, ext)
}

func (r *Response) Error() string {
	return SprintfMany(r.Title, "\n",
		"title=%s", r.Title,
		"error_type=%s", r.Type,
		"error_detail=%s", r.Detail,
		"http_status=%d", r.Status,
		"instance=%s", r.Instance,
	)
}

func (r *Response) MIMEType() MIMEType {
	return ApplicationProblemJSON
}

func (r *Response) HTTPStatusCode() int {
	return r.Status
}

func (*Response) ResponsePayload() {}

func NewResponse(args ResponseArgs) *Response {
	return &Response{
		Type:       args.Type,
		Title:      args.Title,
		Status:     args.Status,
		Detail:     args.Detail,
		Instance:   args.Instance,
		Extensions: args.Extensions,
	}
}

type ResponseArgs struct {
	Type       ErrorTypeURI `json:"type"`
	Title      string       `json:"title"`
	Status     int          `json:"status"`
	Detail     string       `json:"detail"`
	Instance   string       `json:"instance"`
	Extensions []Extension  `json:"extensions"`
}

func (r *ResponseArgs) AddExtension(ext Extension) {
	r.Extensions = append(r.Extensions, ext)
}

func (r *Response) Write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/problem+json") // RFC 9457 media type
	w.WriteHeader(r.Status)
	return json.NewEncoder(w).Encode(r)
}

func (r *Response) UnmarshalJSON(data []byte) error {
	// Use alias to avoid recursion during unmarshaling
	type responseAlias struct {
		Type       ErrorTypeURI     `json:"type"`
		Title      string           `json:"title"`
		Status     int              `json:"status"`
		Detail     string           `json:"detail,omitempty"`
		Instance   string           `json:"instance,omitempty"`
		Extensions []jsontext.Value `json:"extensions,omitempty"`
	}

	var temp responseAlias
	if err := jsonv2.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy standard fields
	r.Type = temp.Type
	r.Title = temp.Title
	r.Status = temp.Status
	r.Detail = temp.Detail
	r.Instance = temp.Instance

	// Unmarshal extensions into their concrete types
	r.Extensions = make([]Extension, 0, len(temp.Extensions))
	for i, raw := range temp.Extensions {
		ext, err := unmarshalExtension(raw, i)
		if err != nil {
			Logger().Error("Failed to unmarshal extension",
				"index", i,
				"error", err,
			)
		}
		r.Extensions = append(r.Extensions, ext)
	}

	return nil
}

func unmarshalExtension(raw jsontext.Value, index int) (Extension, error) {
	// Try each registered extension type
	for _, registeredExt := range registeredExtensions {
		registeredType := reflect.TypeOf(registeredExt)

		// Determine the underlying value type
		valueType := registeredType
		if registeredType.Kind() == reflect.Ptr {
			valueType = registeredType.Elem()
		}

		// Create a new instance of the value type (newExt is always a pointer)
		newExt := reflect.New(valueType)

		// Try to unmarshal into this type
		if err := jsonv2.Unmarshal(raw, newExt.Interface()); err == nil {
			// Success! Return as value type (dereference)
			// This handles the case where the extension was registered as (*Type)(nil)
			// but we need to return Type (value) not *Type (pointer)
			return newExt.Elem().Interface().(Extension), nil
		}
	}

	// No registered type matched - fall back to map[string]any
	var fallback map[string]any
	if err := jsonv2.Unmarshal(raw, &fallback); err != nil {
		return nil, fmt.Errorf("failed to unmarshal extension at index %d: %w", index, err)
	}

	Logger().Error("Extension did not match any registered type, using map[string]any",
		"index", index,
		"data", fallback,
	)

	return fallback, nil
}
