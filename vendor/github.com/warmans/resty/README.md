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
func (h *SomeResource) Before(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(rw, "ALWAYS -> ")
}
func (h *SomeResource) HandleGet(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(rw, "GET")
}
func (h *SomeResource) HandlePost(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(rw, "POST")
}

func main() {

	mux := http.NewServeMux()
	mux.Handle("/resource", ctxhandler.Ctx(resty.Restful(&SomeResource{})))

	log.Printf("Listening on localhost:8080. Try and send a POST or GET")
	log.Fatal(http.ListenAndServe(":8080", mux))

}
```
