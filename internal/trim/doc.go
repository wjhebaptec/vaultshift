// Package trim provides a Trimmer that enforces a maximum byte length on
// secret values before they are stored or transmitted.
//
// Some secret managers impose hard limits on value size. The Trimmer ensures
// values never exceed those limits, optionally appending a configurable
// suffix (e.g. "...") so consumers can detect that truncation occurred.
//
// Basic usage:
//
//	tr, err := trim.New(256, trim.WithSuffix("..."))
//	if err != nil {
//		log.Fatal(err)
//	}
//	safe := tr.Trim(rawValue)
//
// To process an entire secrets map at once:
//
//	trimmed := tr.TrimAll(secretsMap)
package trim
