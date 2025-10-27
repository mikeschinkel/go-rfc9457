package rfc9457

type ContentGetter interface {
	Content() any
}

type MIMEType string

type Selector string

func Selectors[S ~string](ss []S) (ids []Selector) {
	ids = make([]Selector, len(ss))
	for i, s := range ss {
		ids[i] = Selector(s)
	}
	return ids
}

type ResponsePayload interface {
	ResponsePayload()
	HTTPStatusCode() int
	MIMEType() MIMEType
	error
}
