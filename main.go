package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type WrappedBody struct {
	closeCount int
	name       string
	io.ReadCloser
}

func (wb *WrappedBody) Read(p []byte) (n int, err error) {
	return wb.ReadCloser.Read(p)
}

func (wb *WrappedBody) Close() error {
	wb.closeCount += 1
	println(wb.name, " Called Closed:", wb.closeCount)
	err := wb.ReadCloser.Close()
	if err != nil {
		println("Error closing body:", err)
	}
	return err
}

func main() {
	res, err := http.Get("https://www.google.com")
	if err != nil {
		panic(err)
	}

	res.Body = &WrappedBody{
		ReadCloser: res.Body,
		name:       "original",
	}

	fmt.Printf("Body pointer: %p\n", res.Body)
	// Output when this is uncommented:
	//
	// Body pointer: 0xc00039ede0
	// original  Called Closed: 1
	// Post Dump ponter: 0xc00039ef00
	// original  Called Closed: 2
	//
	//	defer func(body io.Closer) {
	//		body.Close()
	//	}(res.Body)

	//Correct Output
	//   Body pointer: 0xc00007f320
	//   original  Called Closed: 1
	//   Post Dump ponter: 0xc00007f440
	//    post-dump  Called Closed: 1
	defer func() {
		res.Body.Close()
	}()

	dump, err := httputil.DumpResponse(res, true)
	if err != nil {
		panic(err)
	}

	res.Body = &WrappedBody{
		ReadCloser: res.Body,
		name:       "post-dump",
	}

	fmt.Printf("Post Dump ponter: %p\n", res.Body)
	_ = dump

	// let's read the body.
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		println("\n Error reading body:", err.Error())
	}

	return
}
