package rfc9457

type Extension interface{}

var registeredExtensions = make([]Extension, 0)

func RegisterExtension(ext Extension) {
	registeredExtensions = append(registeredExtensions, ext)
}
