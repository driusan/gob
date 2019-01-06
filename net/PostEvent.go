package net

import (
	"net/url"
)

// An Event which causes a Post. This generally needs to be sent
// to the window's event queue to be handled, because it needs to
// trigger a complete new page load.
type PostEvent struct {
	URL    *url.URL
	Values url.Values
}
