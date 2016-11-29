Resty
========

This library just makes the http.Handler a bit RESTy-er and
provides compatibility middlesware to allow a RESTHandler
to be used with anything that supports http.Handler.

For example this library works fine with both the standard
go route matching or gorilla mux.


```
type SomeResource struct {
	resty.DefaultRESTHandler
}

func (h *SomeResource) HandleGet(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "GET")
}
func (h *SomeResource) HandlePost(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "POST")
}

func main() {

	mux := http.NewServeMux()
	mux.Handle("/resource", resty.Restful(&SomeResource{}))

	log.Print("Listening on localhost:8080. Try and send a POST or GET")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```
