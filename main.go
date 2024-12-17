package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Result struct {
	UsageMetadata UsageMetadata `json:"usageMetadata"`
}

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

var UpdateUpstreamBody = "upstream response body updated by the simple plugin"
var requestBody = ""

// The callbacks in the filter, like `DecodeHeaders`, can be implemented on demand.
// Because api.PassThroughStreamFilter provides a default implementation.
type filter struct {
	api.PassThroughStreamFilter

	callbacks api.FilterCallbackHandler
	path      string
	config    *config
}

func (f *filter) sendLocalReplyInternal() api.StatusType {
	body := fmt.Sprintf("%s, path: %s\r\n", f.config.echoBody, f.path)
	f.callbacks.DecoderFilterCallbacks().SendLocalReply(200, body, nil, 0, "")
	// Remember to return LocalReply when the request is replied locally
	return api.LocalReply
}

// Callbacks which are called in request path
// The endStream is true if the request doesn't have body
func (f *filter) DecodeHeaders(header api.RequestHeaderMap, endStream bool) api.StatusType {
	f.path, _ = header.Get(":path")
	api.LogDebugf("get path %s", f.path)

	fmt.Println("DECODEHEADERS")

	var token *oauth2.Token
	scopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}

	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx, scopes...)

	if err == nil {
		token, err = credentials.TokenSource.Token()

		if err == nil {
			// flags.Token = token.AccessToken
			fmt.Println("SETTING AUTHORIZATION HEADER")
			header.Set("Authorization", "Bearer "+token.AccessToken)
		} else {
			fmt.Println(err.Error())
		}
	} else {
		fmt.Println("Credentials file : " + os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

		keyFile, err := os.Open(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
		if err == nil {
			byteValue, _ := io.ReadAll(keyFile)
			fmt.Println(string(byteValue))
		} else {
			fmt.Println(err.Error())
		}
		defer keyFile.Close()

		fmt.Println("COULD NOT FIND CREDENTIALS")
	}

	if f.path == "/localreply_by_config" {
		return f.sendLocalReplyInternal()
	}
	return api.Continue
	/*
		// If the code is time-consuming, to avoid blocking the Envoy,
		// we need to run the code in a background goroutine
		// and suspend & resume the filter
		go func() {
			defer f.callbacks.DecoderFilterCallbacks().RecoverPanic()
			// do time-consuming jobs

			// resume the filter
			f.callbacks.DecoderFilterCallbacks().Continue(status)
		}()

		// suspend the filter
		return api.Running
	*/
}

// DecodeData might be called multiple times during handling the request body.
// The endStream is true when handling the last piece of the body.
func (f *filter) DecodeData(buffer api.BufferInstance, endStream bool) api.StatusType {
	// support suspending & resuming the filter in a background goroutine
	fmt.Println("DECODEDATA")
	return api.Continue
}

func (f *filter) DecodeTrailers(trailers api.RequestTrailerMap) api.StatusType {
	// support suspending & resuming the filter in a background goroutine
	return api.Continue
}

// Callbacks which are called in response path
// The endStream is true if the response doesn't have body
func (f *filter) EncodeHeaders(header api.ResponseHeaderMap, endStream bool) api.StatusType {
	if f.path == "/update_upstream_response" {
		fmt.Println("SETTING Content-Length to " + strconv.Itoa(len(UpdateUpstreamBody)))
		header.Set("Content-Length", strconv.Itoa(len(UpdateUpstreamBody)))
	}
	header.Set("Rsp-Header-From-Go", "bar-test")
	// support suspending & resuming the filter in a background goroutine
	return api.Continue
}

// EncodeData might be called multiple times during handling the response body.
// The endStream is true when handling the last piece of the body.
func (f *filter) EncodeData(buffer api.BufferInstance, endStream bool) api.StatusType {
	// if f.path == "/update_upstream_response" {
	// 	if endStream {
	// 		buffer.SetString(UpdateUpstreamBody)
	// 	} else {
	// 		buffer.Reset()
	// 	}
	// } else {

	fmt.Println("RESPONSE IS " + buffer.String())

	if endStream {
		fmt.Println("RESPONSE IS " + buffer.String())
		var response []Result
		json.Unmarshal(buffer.Bytes(), &response)

		if len(response) > 0 {
			fmt.Println("Token length: " + strconv.Itoa(response[len(response)-1].UsageMetadata.CandidatesTokenCount))
		}
	}

	//}
	// support suspending & resuming the filter in a background goroutine
	return api.Continue
}

func (f *filter) EncodeTrailers(trailers api.ResponseTrailerMap) api.StatusType {
	fmt.Println("ENCODETRAILERS")
	return api.Continue
}

// OnLog is called when the HTTP stream is ended on HTTP Connection Manager filter.
func (f *filter) OnLog(reqHeader api.RequestHeaderMap, reqTrailer api.RequestTrailerMap, respHeader api.ResponseHeaderMap, respTrailer api.ResponseTrailerMap) {
	code, _ := f.callbacks.StreamInfo().ResponseCode()
	respCode := strconv.Itoa(int(code))
	api.LogDebug(respCode)

	/*
		// It's possible to kick off a goroutine here.
		// But it's unsafe to access the f.callbacks because the FilterCallbackHandler
		// may be already released when the goroutine is scheduled.
		go func() {
			defer func() {
				if p := recover(); p != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					fmt.Printf("http: panic serving: %v\n%s", p, buf)
				}
			}()

			// do time-consuming jobs
		}()
	*/
}

// OnLogDownstreamStart is called when HTTP Connection Manager filter receives a new HTTP request
// (required the corresponding access log type is enabled)
func (f *filter) OnLogDownstreamStart(reqHeader api.RequestHeaderMap) {
	// also support kicking off a goroutine here, like OnLog.
}

// OnLogDownstreamPeriodic is called on any HTTP Connection Manager periodic log record
// (required the corresponding access log type is enabled)
func (f *filter) OnLogDownstreamPeriodic(reqHeader api.RequestHeaderMap, reqTrailer api.RequestTrailerMap, respHeader api.ResponseHeaderMap, respTrailer api.ResponseTrailerMap) {
	// also support kicking off a goroutine here, like OnLog.
}

func (f *filter) OnDestroy(reason api.DestroyReason) {
	// One should not access f.callbacks here because the FilterCallbackHandler
	// is released. But we can still access other Go fields in the filter f.

	// goroutine can be used everywhere.
}
