package gosnowth

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// resolveURL - given a URL and a string reference, it will resolve the
// address of the URL plus the reference
func resolveURL(baseURL *url.URL, ref string) string {
	refURL, _ := url.Parse(ref)
	return baseURL.ResolveReference(refURL).String()
}

// multiError - sometimes you need to keep track of multiple errors,
// and this struct will allow you to keep track of multiple errors.
type multiError struct {
	errs []error
}

// newMultiError - will allow you to create a new multiError instance
func newMultiError() *multiError {
	return &multiError{
		errs: []error{},
	}
}

// Add - add an error to the list of errors
func (me *multiError) Add(err error) {
	if err != nil {
		me.errs = append(me.errs, err)
	}
}

// HasError - Do we have any errors in our list of errors
func (me *multiError) HasError() bool {
	if len(me.errs) > 0 {
		return true
	}
	return false
}

// Error - implement the error interface, provide string representation
// of the errors in our list of errors
func (me *multiError) Error() string {
	var errStrs []string
	for _, err := range me.errs {
		errStrs = append(errStrs, err.Error())
	}
	return strings.Join(errStrs, "; ")
}

// moveNode - move a url from a slice to a new slice, if this is used for
// SnowthInstances' active or inactive slices wrap in a write lock
func moveNode(from, dest *[]*SnowthNode, u *SnowthNode) {
	// put this url in active
	*dest = append(*dest, u)

	// find the item index in the deactive list
	var index = -1
	for i, v := range *from {
		if v.url.String() == u.url.String() {
			index = i
		}
	}
	if index != -1 {
		// remove from deactive
		*from = removeNode(*from, index)
	}
}

// removeNode - remove a url from a slice, if this is used for
// SnowthInstances' active or inactive slices wrap in a write lock
func removeNode(a []*SnowthNode, index int) []*SnowthNode {
	copy(a[index:], a[index+1:])
	a[len(a)-1] = nil // or the zero value of T
	a = a[:len(a)-1]
	return a
}

// decodeJSONFromResponse - given a response decode the body as json
func decodeJSONFromResponse(v interface{}, reader io.Reader) error {
	dec := json.NewDecoder(reader)

	if err := dec.Decode(v); err != nil {
		return errors.Wrap(err, "failed to decode response body")
	}
	return nil
}

// encodeXML - produce a reader which when read will be the xml
// representation of the interface provided
func encodeXML(v interface{}) (io.Reader, error) {
	buf := bytes.NewBuffer([]byte{})
	dec := xml.NewEncoder(buf)

	if err := dec.Encode(v); err != nil {
		return nil, errors.Wrap(err, "failed to encode")
	}
	return buf, nil
}

// decodeXMLFromResponse - Decode the response body as xml
func decodeXMLFromResponse(v interface{}, reader io.Reader) error {
	dec := xml.NewDecoder(reader)

	if err := dec.Decode(v); err != nil {
		return errors.Wrap(err, "failed to decode response body")
	}
	return nil
}
