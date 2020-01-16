package main

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/peterpla/lead-expert/pkg/request"
)

func TestTaggingQAGDLPReorgMatchedTags(t *testing.T) {

	type test struct {
		name         string
		req          *request.Request
		expectedTags map[string]request.Tags
	}

	var customerID = 1234567
	var media = "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3"
	var timeNow = time.Now().UTC().Format(time.RFC3339Nano)

	reqSingleTag, expSingleTag := createReqWithSingleTag(t, customerID, media, timeNow)
	reqSameLikelihood, expSameLikelihood := createReqWithSameLikelihood(t, customerID, media, timeNow)
	reqHigherFirst, expHigherFirst := createReqWithHigherLikelihoodFirst(t, customerID, media, timeNow)
	reqHigherLast, expHigherLast := createReqWithHigherLikelihoodLast(t, customerID, media, timeNow)
	reqNoTags, expNoTags := createReqWithNoTags(t, customerID, media, timeNow)

	tests := []test{
		// valid
		{name: "single Tag",
			req:          reqSingleTag,
			expectedTags: expSingleTag},
		{name: "same likelihood",
			req:          reqSameLikelihood,
			expectedTags: expSameLikelihood},
		{name: "lower likelihood",
			req:          reqHigherFirst,
			expectedTags: expHigherFirst},
		{name: "higher likelihood",
			req:          reqHigherLast,
			expectedTags: expHigherLast},
		{name: "no MatchedTags",
			req:          reqNoTags,
			expectedTags: expNoTags},
	}

	for _, tc := range tests {

		gDLPReorgMatchedTags(tc.req)

		if !cmp.Equal(tc.expectedTags, tc.req.MatchedTags) {
			t.Errorf("%s: expected tags %v, got %v", tc.name, tc.expectedTags, tc.req.MatchedTags)
		}
	}
}

// createReqWithSingleTag returns a request with a single Tag, and a map with the expected result
func createReqWithSingleTag(t *testing.T, customerID int, media string, timeNow string) (*request.Request, map[string]request.Tags) {

	req := createBasicRequest(customerID, media, timeNow)

	var m = make(map[string]request.Tags)
	var k = "123 Main Street"
	var it = "ADDRESS"
	m[k] = request.Tags{Quote: k, InfoType: it, Likelihood: 4, BeginByteOffset: 0, EndByteOffset: len(k)}
	req.MatchedTags = m

	var exp = make(map[string]request.Tags)
	exp[it] = request.Tags{Quote: k, InfoType: it, Likelihood: 4, BeginByteOffset: 0, EndByteOffset: len(k)}

	return req, exp
}

// createReqWithSameLikelihood returns a request with two tags with the same likelihood, and a map with the expected result
func createReqWithSameLikelihood(t *testing.T, customerID int, media string, timeNow string) (*request.Request, map[string]request.Tags) {

	req := createBasicRequest(customerID, media, timeNow)

	var m = make(map[string]request.Tags)
	var k1 = "123 Main Street Southeast"
	var k2 = "123 Main Street"
	var it = "STREET_ADDRESS"
	m[k1] = request.Tags{
		Quote:           k1,
		InfoType:        it,
		Likelihood:      4,
		BeginByteOffset: 0,
		EndByteOffset:   len(k1)}
	m[k2] = request.Tags{ // the m[k2] entry should be removed by gDLPReorgMatchedTags()
		Quote:           k2,
		InfoType:        it,
		Likelihood:      4,
		BeginByteOffset: len(k1) + 10,
		EndByteOffset:   len(k1) + 10 + len(k2)}
	req.MatchedTags = m

	var exp = make(map[string]request.Tags)
	exp[it] = request.Tags{
		Quote:           k1,
		InfoType:        it,
		Likelihood:      4,
		BeginByteOffset: 0,
		EndByteOffset:   len(k1)}

	return req, exp
}

// createReqWithHigherLikelihoodFirst returns a request with the two tags, the first having higher likelihood
func createReqWithHigherLikelihoodFirst(t *testing.T, customerID int, media string, timeNow string) (*request.Request, map[string]request.Tags) {

	req := createBasicRequest(customerID, media, timeNow)

	var m = make(map[string]request.Tags)
	var k1 = "123 Main Street Southeast"
	var k2 = "123 Main Street"
	var it = "STREET_ADDRESS"
	m[k1] = request.Tags{
		Quote:           k1,
		InfoType:        it,
		Likelihood:      5,
		BeginByteOffset: 0,
		EndByteOffset:   len(k1)}
	m[k2] = request.Tags{ // the m[k2] entry should be removed by gDLPReorgMatchedTags()
		Quote:           k2,
		InfoType:        it,
		Likelihood:      4,
		BeginByteOffset: len(k1) + 10,
		EndByteOffset:   len(k1) + 10 + len(k2)}
	req.MatchedTags = m

	var exp = make(map[string]request.Tags)
	exp[it] = request.Tags{
		Quote:           k1,
		InfoType:        it,
		Likelihood:      5,
		BeginByteOffset: 0,
		EndByteOffset:   len(k1)}

	return req, exp
}

// createReqWithHigherLikelihoodLast returns a Request with the two tags, the second having higher likelihood
func createReqWithHigherLikelihoodLast(t *testing.T, customerID int, media string, timeNow string) (*request.Request, map[string]request.Tags) {

	req := createBasicRequest(customerID, media, timeNow)

	var m = make(map[string]request.Tags)
	var k1 = "123 Main Street"
	var k2 = "123 Main Street Southeast"
	var it = "STREET_ADDRESS"
	m[k1] = request.Tags{ // the m[k1] entry should be removed by gDLPReorgMatchedTags()
		Quote:           k1,
		InfoType:        it,
		Likelihood:      4,
		BeginByteOffset: 0,
		EndByteOffset:   len(k1)}
	m[k2] = request.Tags{
		Quote:           k2,
		InfoType:        it,
		Likelihood:      5,
		BeginByteOffset: len(k1) + 10,
		EndByteOffset:   len(k1) + 10 + len(k2)}

	req.MatchedTags = m

	var exp = make(map[string]request.Tags)
	exp[it] = request.Tags{
		Quote:           k2,
		InfoType:        it,
		Likelihood:      5,
		BeginByteOffset: len(k1) + 10,
		EndByteOffset:   len(k1) + 10 + len(k2)}

	return req, exp
}

// createReqWithNoTags returns a Request with the two tags, the second having higher likelihood
func createReqWithNoTags(t *testing.T, customerID int, media string, timeNow string) (*request.Request, map[string]request.Tags) {

	req := createBasicRequest(customerID, media, timeNow)
	exp := make(map[string]request.Tags)

	return req, exp
}

// createBasicRequest returns a Request with common fields set
func createBasicRequest(customerID int, media string, timeNow string) *request.Request {
	req := request.Request{}
	req.RequestID = uuid.New()
	req.CustomerID = customerID
	req.MediaFileURI = media
	req.AcceptedAt = timeNow

	return &req
}
