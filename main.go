package main

import "net/http"

// important to remember - every time i change the code, it will be
// required to rebuild and restart the server

func main() {
	serverMux := http.ServeMux{}
	serverStruct := http.Server{
		Handler: &serverMux,
		Addr:    ":8080",
	}
	handler := http.FileServer(http.Dir("."))
	serverMux.Handle("/", handler)
	serverStruct.ListenAndServe()
}
