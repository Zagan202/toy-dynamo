/* app_test.go
 *
 * CMPS 128 Fall 2018
 *
 * Lawrence Lawson     lelawson
 * Pete Wilcox         pcwilcox
 *
 * Unit test definitions for app.go
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

// Define some constants. These can be reconfigured as needed.
const (
	domain       = "http://localhost"
	port         = "8080"
	root         = "/keyValue-store"
	hostname     = domain + ":" + port + root
	keyExists    = "KEY_EXISTS"
	KeyNotExists = "KEY_DOESN'T_EXIST"
	valExists    = "VAL_EXISTS"
)

type resp struct {
	buf       []uint8
	off       int
	bootstrap []uint8
	lastread  int
}

type TestKVS struct {
	key       string
	valExists string
	service   bool
}

/* This stub returns true for the key which exists and false for the one which doesn't */
func (t *TestKVS) Contains(key string) bool {
	if strings.Compare(key, t.key) == 0 {
		return true
	}
	return false
}

/* This stub returns the valExistsue associated with the key which exists, and returns nil for the key which doesn't */
func (t *TestKVS) Get(key string) string {
	if key == t.key {
		return t.valExists
	}
	return ""
}

/* This stub returns true for the key which exists and false for the one which doesn't */
func (t *TestKVS) Delete(key string) bool {
	if key == t.key {
		return true
	}
	return false
}

// Returns the value of the service bool
func (t *TestKVS) ServiceUp() bool {
	return t.service
}

// idk lets try this
func (t *TestKVS) Put(key, valExists string) {
}

// TestPutRequestKeyExists should return that the key has been replaced/updated successfully
func TestPutRequestKeyExists(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db, ":5000"}

	l, err := net.Listen("tcp", "127.0.0.1:5000")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "PUT"
	reqBody := strings.NewReader(valExists)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"replaced": "True",
		"msg":      "Updated successfully",
	}

	equals(t, expectedBody, gotBody)
}

// TestPutRequestKeyDoesntExist should return that the key has been created
func TestPutRequestKeyDoesntExist(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db, ":5000"}

	l, err := net.Listen("tcp", "127.0.0.1:5000")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := KeyNotExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "PUT"
	//reqBody := strings.NewReader(valExists)
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusCreated // code 201
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"replaced": "False",
		"msg":      "Added successfully",
	}

	equals(t, expectedBody, gotBody)
}

// TestPutRequestInvalExistsidKey makes a key with length == 201 and tests it for failure
func TestPutRequestInvalExistsidKey(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db, ":5000"}

	l, err := net.Listen("tcp", "127.0.0.1:5000")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject needs to be very long
	subject := ""
	for i := 0; i < 201; i++ {
		subject = subject + "a"
	}

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "PUT"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusUnprocessableEntity // code 422
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":    "Key not valid",
		"result": "Error",
	}

	equals(t, expectedBody, gotBody)
}

// TestPutRequestInvalExistsidValue tests for values that are too large
func TestPutRequestInvalExistsidValue(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db, ":5000"}

	l, err := net.Listen("tcp", "127.0.0.1:5000")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject doesn't really matter
	subject := keyExists

	// The value needs to be > 1MB
	val := "val="
	big := "a"
	for i := 1; i < 1048577; i *= 2 {
		big = big + big
	}
	val = val + big

	// Convert it to a reader
	reader := strings.NewReader(val)
	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "PUT"
	req, err := http.NewRequest(method, url, reader)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusUnprocessableEntity // code 422
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	var gotBody map[string]interface{}
	err = json.Unmarshal([]byte(recorder.Body.String()), &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":    "Object too large. Size limit is 1MB",
		"result": "Error",
	}

	equals(t, expectedBody, gotBody)
}

// TestGetRequestKeyExists should return success with the "VAL_EXISTS" string
func TestGetRequestKeyExists(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db, ":5000"}

	l, err := net.Listen("tcp", "127.0.0.1:5000")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.GetHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}
	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"result": "Success",
		"value":  valExists,
	}

	equals(t, expectedBody, gotBody)
}

// TestGetRequestKeyNotExists should return that the key has been replaced/updated successfully
func TestGetRequestKeyNotExists(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db, ":5000"}

	l, err := net.Listen("tcp", "127.0.0.1:5000")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.GetHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := KeyNotExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotFound // code 404
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	var gotBody map[string]interface{}
	err = json.Unmarshal([]byte(recorder.Body.String()), &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"result": "Error",
		"value":  "Not Found",
	}

	equals(t, expectedBody, gotBody)
}

// TestDeleteKeyExists should return success
func TestDeleteKeyExists(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db, ":5000"}

	l, err := net.Listen("tcp", "127.0.0.1:5000")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.DeleteHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "DELETE"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 404
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}
	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"result": "Success",
	}

	equals(t, expectedBody, gotBody)
}

/* These functions were taken from Ben Johnson's post here:
 * https://medium.com/@benbjohnson/structuring-tests-in-go-46ddee7a25c
 */

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
