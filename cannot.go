// XXX???FIXME
//	- Is "errs" a good package name?
//		- was "utilerrs"

// Package errs provides a generic error type, and some utility functions that
// use that error type.
//
// It provides generic errors which are (pointers to) errs.CannotError, a struct
// type suitable for English-language error messages of the form
//	cannot <verb>[ <adjective>]  <o-q-noun>[ <suffix>[: <base-error>]
// where <o-q-noun> is a noun phrase that can be put in double quotes (using the
//  "fmt" package’s %q verb).
//
// Package errs also provides some utility functions for making simple HTTP GET
// and HEAD requests and dealing with related errors.
//
package errs

import (
	"fmt"
	urlpkg "net/url"
	"os"
	"strings"
)

//
// A CannotError holds details of a problem suitable for messages of the form
//	cannot <verb>[ <adjective>]  <o-q-noun>[ <suffix>[: <base-error>]
// which assumes English (sorry!).
//
type CannotError struct {
	Verb      string // A present-tense verb
	Adjective string // What kind of thing the action was on, or ""
	Noun      string // Which thing the action was on
	QuoteNoun bool   // Whether to put .Noun in double quotes
	Suffix    string // Text to go after the noun, or ""
	BaseError error  // The underlying error, if any
}

//
// Cannot() is a convenience function to produce a (pointer to a) CannotError value.
//
func Cannot(
	verb, adjective, noun string,
	quoteNoun bool, suffix string, baseError error,
) *CannotError {
	return &CannotError{
		Verb:      verb,
		Adjective: adjective,
		Noun:      noun,
		QuoteNoun: quoteNoun,
		Suffix:    suffix,
		BaseError: baseError,
	}
}

// Pointers to CannotError values satisfy the error interface.
func (ce *CannotError) Error() string {
	var b strings.Builder
	b.WriteString("cannot " + ce.Verb + " ")
	if ce.Adjective != "" {
		b.WriteString(ce.Adjective + " ")
	}
	if ce.QuoteNoun {
		fmt.Fprintf(&b, "%q", ce.Noun)
	} else {
		b.WriteString(ce.Noun)
	}
	if ce.Suffix != "" {
		b.WriteString(" " + ce.Suffix)
	}
	if ce.BaseError != nil {
		b.WriteString(": ")
		b.WriteString(TidyError(ce.BaseError))
	}
	return b.String()
}

// A CannotError may specify an underlying error.
func (ce *CannotError) Unwrap() error {
	return ce.BaseError
}

// TidyError(e) is equivalent to e.Error() except for a few special cases,
// in which it gives something users of command-line programs will (IMO) find easier
// to understand.
//
// Many Go library packages (notably "os" and "net/url") produce errors with
// terse message texts that seem to be intended for server operators. This
// function recognizes a few of those case.
//
// TO-DO: How does unwrapping os.PathError look on Windows??? Might want
//	  platform-specific code here.
//
func TidyError(e error) string {
	// os.PathError.Error() is just "<Operation>: <Path>: <BaseError>"
	// (with no added quotation marks for easier parsing!).
	// This package assumes you provide the first two items in English, so
	// we want the base error’s text instead of the PathError’s.
	if pe, isPathErr := e.(*os.PathError); isPathErr {
		return pe.Unwrap().Error()
	}

	// Errors from the net.url package can generate texts of the form
	//	net/url: <problem in url>
	// which I do not want.
	if ue, isNetUrlErr := e.(*urlpkg.Error); isNetUrlErr && ue.Op == "parse" {
		text := ue.Unwrap().Error()
		if strings.HasPrefix(text, "net/url: ") {
			text = text[len("net/url"):]
		}
		return text
	}

	// If whoever wrote archive/zip has used a zip-specific error type, we
	// could look at type info instead of playing dubious tricks with the
	// generated string.
	text := e.Error()
	if text[:5] == "zip: " {
		text = text[len("zip: "):]
	}

	return text
}
