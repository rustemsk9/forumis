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
	// writer.Header().Set("Strict-Transport-Security", "max-age=15768000 ; includeSubDomains")
	// var alsoid int
	// cook, err := request.Cookie("_cookie")
	// if err != nil {
	// 	http.Redirect(writer, request, "/", 302)
	// 	fmt.Println("Error") // or redirect
	// }
	// alsoid, _ = strconv.Atoi(cook.Value)
	// cook, err := request.Cookie("_cookie")
	// if err != nil {
	// fmt.Println("Error") // or redirect
	// } else {

	// 	fmt.Println(cooPart)
	// 	getUUid := cooPart[1]

	// 	if !data.SessionByUUID(getUUid) {
	// if err != http.ErrNoCookie {
	// 	cooPart := strings.Split(cook.Value, "&")
	// 	fmt.Println("Error1") // or redirect
	// 	session := data.Session{Uuid: cooPart[1]}
	// 	err2 := session.DeleteByUUID()
	// 	if err2 != nil {
	// 		fmt.Println("Error2") // or redirect
	// 		cook.MaxAge = -1
	// 		cook.Expires = time.Unix(1, 0)
	// 		http.SetCookie(writer, cook)
	// 		http.Redirect(writer, request, "/login", 302)
	// 		return
	// 	}
	// 	http.Redirect(writer, request, "/login", 302)
	// }

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
