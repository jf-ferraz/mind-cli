package orchestrate

import "github.com/jf-ferraz/mind-cli/domain"

// Classify routes a user request string to a domain.RequestType.
// This is a thin adapter re-exporting domain.Classify for package-level access.
func Classify(request string) domain.RequestType { return domain.Classify(request) }

// Slugify converts a request string to a URL-safe slug.
// This is a thin adapter re-exporting domain.Slugify for package-level access.
func Slugify(request string) string { return domain.Slugify(request) }
