package server

type RequestMeta struct {
	IfNoneMatch string
}

type ResponseMeta struct {
	Headers map[string]string
}

func NewResponseMeta() ResponseMeta {
	return ResponseMeta{Headers: make(map[string]string)}
}

func (m *ResponseMeta) SetHeader(key, value string) {
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	m.Headers[key] = value
}
