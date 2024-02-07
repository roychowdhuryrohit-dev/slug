package http

import "net/textproto"

type Header map[string][]string

func (h Header) Add(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	if _, ok := h[key]; ok {
		h[key] = append(h[key], value)
	} else {
		h.Set(key, value)
	}
}

func (h Header) Set(key, value string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	h[key] = []string{value}
}

func (h Header) Get(key string) (string, bool) {

	key = textproto.CanonicalMIMEHeaderKey(key)
	val, ok := h[key]
	if ok {
		return val[0], ok
	}
	return "", false
}

func (h Header) GetValues(key string) ([]string, bool) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	val, ok := h[key]
	return val, ok
}

func (h Header) Delete(key string) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	delete(h, key)
}
