package trace

import "net/http"

// RequestIdKey is the request id header, has a unique association to a trace id.
// The Request ID uses the UUID format because, first, it is more in line with industry standards,
// second, it guarantees the privacy and security of the Trace ID,
// and third, it provides standardization capabilities for subsequent distribution.
var RequestIdKey = http.CanonicalHeaderKey("x-request-id")
