package routes

import (
	"net/http"

	"forum/data"
	"forum/utils"
)

// GET /err?msg=
// shows the error message page
func Err(writer http.ResponseWriter, request *http.Request) {
	vals := request.URL.Query()
	_, err := data.SessionCheck(writer, request)
	if err != nil {
		utils.GenerateHTML(writer, vals.Get("msg"), "layout", "public.navbar", "error")
	} else {
		utils.GenerateHTML(writer, vals.Get("msg"), "layout", "private.navbar", "error")
	}
}

func Index(writer http.ResponseWriter, request *http.Request) {
	threads, err := data.Threads()

	// threads[0].SessionId = alsoid
	if err == nil {
		_, err := data.SessionCheck(writer, request)
		if err != nil {
			utils.GenerateHTML(writer, threads, "layout", "public.navbar", "index")
		} else {
			utils.GenerateHTML(writer, threads, "layout", "private.navbar", "index")
		}
	} else {
		utils.ErrorMessage(writer, request, "Cannot get threads")
	}
}
