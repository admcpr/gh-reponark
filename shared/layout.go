package shared

// Titled models name themselves in the breadcrumb drawn along the top edge
// of the frame, e.g. "acme-corp" or "Filters".
type Titled interface {
	Breadcrumb() string
}

// StatusProvider models show a short status at the right of the frame's top
// edge, e.g. "37 of 148 repos".
type StatusProvider interface {
	Status() string
}
