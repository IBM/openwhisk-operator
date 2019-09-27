/*
 * Copyright 2019 IBM Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package wsk

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// Server mocks whisk backend
type Server struct {
	server *httptest.Server
	URL    string

	// Record incoming requests
	Requests []Request

	responses map[string]*Response
}

// Request record incoming requests
type Request struct {
	Body string
}

// Response is the repsonse to send back for a given request
type Response struct {
	Path   string
	Method string
	Body   []byte
}

// NewServer creates a new fake whisk server
func NewServer(responses ...*Response) *Server {
	server := &Server{
		Requests:  make([]Request, 0),
		responses: make(map[string]*Response),
	}
	server.server = httptest.NewServer(server)
	server.URL = server.server.URL

	for _, r := range responses {
		server.responses[r.Path] = r
	}

	return server
}

// ServeHTTP handles whisk request
func (w *Server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(400)
		return
	}

	w.Requests = append(w.Requests, Request{Body: string(b)})

	r, ok := w.responses[req.URL.Path]

	if ok {
		resp.Write(r.Body)
	} else {
		resp.Write([]byte("{}"))
	}
}

// Close the whisk server connection
func (w *Server) Close() {
	w.server.Close()
}
