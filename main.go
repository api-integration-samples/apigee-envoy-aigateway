package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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

var prompt_regexp, _ = regexp.Compile("\"prompt_tokens\":\\s*\\d+")
var completion_regexp, _ = regexp.Compile("\"completion_tokens\":\\s*\\d+")
var total_regexp, _ = regexp.Compile("\"total_tokens\":\\s*\\d+")

// The callbacks in the filter, like `DecodeHeaders`, can be implemented on demand.
// Because api.PassThroughStreamFilter provides a default implementation.
type filter struct {
	api.PassThroughStreamFilter

	callbacks api.FilterCallbackHandler
	path      string
	key       string
	config    *config
}

// Callbacks which are called in request path
// The endStream is true if the request doesn't have body
func (f *filter) DecodeHeaders(header api.RequestHeaderMap, endStream bool) api.StatusType {
	f.path, _ = header.Get(":path")
	api.LogDebugf("get path %s", f.path)
	key, found := header.Get("x-api-key")
	if found {
		f.key = key
	}

	// Add Google auth token for backend
	var token *oauth2.Token
	scopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}

	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx, scopes...)

	if err == nil {
		token, err = credentials.TokenSource.Token()

		if err == nil {
			header.Set("Authorization", "Bearer "+token.AccessToken)
		} else {
			fmt.Println(err.Error())
		}
	}

	return api.Continue
}

// DecodeData might be called multiple times during handling the request body.
// The endStream is true when handling the last piece of the body.
func (f *filter) DecodeData(buffer api.BufferInstance, endStream bool) api.StatusType {
	// support suspending & resuming the filter in a background goroutine
	return api.Continue
}

func (f *filter) DecodeTrailers(trailers api.RequestTrailerMap) api.StatusType {
	// support suspending & resuming the filter in a background goroutine
	return api.Continue
}

// Callbacks which are called in response path
// The endStream is true if the response doesn't have body
func (f *filter) EncodeHeaders(header api.ResponseHeaderMap, endStream bool) api.StatusType {

	// support suspending & resuming the filter in a background goroutine
	return api.Continue
}

// EncodeData might be called multiple times during handling the response body.
// The endStream is true when handling the last piece of the body.
func (f *filter) EncodeData(buffer api.BufferInstance, endStream bool) api.StatusType {

	start := time.Now()

	bufferContent := buffer.String()

	if strings.Contains(bufferContent, "prompt_tokens") {
		prompt_tokens := prompt_regexp.FindString(bufferContent)
		prompt_tokens = strings.Trim(strings.Replace(prompt_tokens, "\"prompt_tokens\":", "", -1), " ")
		completion_tokens := completion_regexp.FindString(bufferContent)
		completion_tokens = strings.Trim(strings.Replace(completion_tokens, "\"completion_tokens\":", "", -1), " ")
		total_tokens := total_regexp.FindString(bufferContent)
		total_tokens = strings.Trim(strings.Replace(total_tokens, "\"total_tokens\":", "", -1), " ")
		fmt.Println("found " + prompt_tokens + " prompt tokens, " + completion_tokens + " completion tokens, " + total_tokens + " total tokens - sending to Apigee X Analytics.")

		go sendAnalyticsToApigee(f.path, prompt_tokens, completion_tokens, total_tokens, f.key, f.config.apigeeEndpoint)
	}

	elapsed := time.Since(start)
	fmt.Printf("apigee token analytics checks took %s for buffer length %d\n", elapsed, len(bufferContent))

	return api.Continue
}

func (f *filter) EncodeTrailers(trailers api.ResponseTrailerMap) api.StatusType {
	return api.Continue
}

// OnLog is called when the HTTP stream is ended on HTTP Connection Manager filter.
func (f *filter) OnLog(reqHeader api.RequestHeaderMap, reqTrailer api.RequestTrailerMap, respHeader api.ResponseHeaderMap, respTrailer api.ResponseTrailerMap) {
	code, _ := f.callbacks.StreamInfo().ResponseCode()
	respCode := strconv.Itoa(int(code))
	api.LogDebug(respCode)
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

func sendAnalyticsToApigee(modelName string, promptTokens string, completionTokens string, totalTokens string, apiKey string, apigeeEndpoint string) {
	postBody, _ := json.Marshal(map[string]string{
		"model_name":        modelName,
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
		"total_tokens":      totalTokens,
	})

	requestBody := bytes.NewBuffer(postBody)
	resp, _ := http.NewRequest(http.MethodPost, apigeeEndpoint+"/genai/token-analytics", requestBody)
	resp.Header.Add("x-api-key", apiKey)
	resp.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(resp)
	if err != nil {
		fmt.Println("Error sending data to Apigee - " + err.Error())
		// handle error, log to retry later...
	} else if res.StatusCode != 200 {
		fmt.Println("Error sending data to Apigee - response code " + strconv.Itoa(res.StatusCode))
		// handle error, log to retry later...
	}
}
