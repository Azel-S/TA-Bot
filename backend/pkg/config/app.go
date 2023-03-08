package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	course "TA-Bot/backend/pkg/models/course"
	user "TA-Bot/backend/pkg/models/user"
)

type App struct {
	DB     *gorm.DB
	Router *mux.Router
}

// Opens a connection with the database
func (a *App) Connect(cPath string) {
	d, err := gorm.Open("mysql", cPath)
	if err != nil {
		panic(err)
	}
	a.DB = d
}

// Opens database according to given paramaters
// Sets routes for server
// Executes table creation queries
func (a *App) Initialize(username, password, dbname string) {
	a.Connect(username + ":" + password + "@tcp(localhost:3306)/" + dbname + "?charset=utf8&parseTime=True&loc=Local")
	a.Router = mux.NewRouter()
	a.initializeRoutes()
	a.DB.Exec(user.UsersCreationQuery)
	a.DB.Exec(course.CoursesCreationQuery)
	a.DB.AutoMigrate(&user.User{})
}

// Listens for incoming requests from Angular
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

// Returns pointer to database
func (a *App) GetDB() *gorm.DB {
	return a.DB
}

// Returns pointer to database
func (a *App) GetRTR() *mux.Router {
	return a.Router
}

/*

	USER FUNCTIONS

*/

// RETURNS A USER AS A JSON OBJECT
func (a *App) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                 // Make a map of string to strings from the http Request
	id, err := strconv.Atoi(vars["id"]) // convert the value of key "id" to readable integer
	if err != nil {                     // error handling
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	u := user.User{ID: id} // create a user with ID of the 'id' grabbed earlier
	if err := u.GetUser(a.DB); err != nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		// should have a check for error type and a respondWithError(w, http.StatusInternalServerError, err.Error()), but it's causing some issues
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

// RETURNS AN ARRAY OF USERS AS A JSON OBJECT
func (a *App) GetManyUsers(w http.ResponseWriter, r *http.Request) {
	// convert the HTTP request's 'count' and 'start' values into readable integers
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 { // edge cases
		count = 10
	}
	if start < 0 {
		start = 0
	}
	users, err := user.GetManyUsers(a.DB, start, count) // call function to get multiple users
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, users)
}

// CREATES A USER INTO DATABASE
func (a *App) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u user.User
	decoder := json.NewDecoder(r.Body)         // Grab decoding data from json to put into a user struct later
	if err := decoder.Decode(&u); err != nil { // Attempts to decode data into user struct
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close() // close http body at end of function call

	if err := u.CreateUser(a.DB); err != nil { // Attempts to add user into database
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, u)
}

// UPDATES USER INFORMATION IN DATABASE
func (a *App) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                 // Make a map of string to strings from the http Request with various data
	id, err := strconv.Atoi(vars["id"]) // convert the value of key "id" to readable integer
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	var u user.User
	decoder := json.NewDecoder(r.Body)         // Get decode data like in CreateUser()
	if err := decoder.Decode(&u); err != nil { // Attempt to decode info into user struct
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	u.ID = id
	// Attempts to find the user with given id above and update information with the decoded data earlier
	if err := u.UpdateUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

// DELETES A USER IN DATABASE
func (a *App) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // Same pattern as in UpdateUser() and GetUser()
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	u := user.User{ID: id} // Create user struct with given ID
	// Attempts to find the User row with matching ID as created user struct above and delete it
	if err := u.DeleteUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

/*

	COURSE FUNCTIONS

*/

// RETURNS A COURSE AS A JSON OBJECT
func (a *App) GetCourse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid course identifier")
		return
	}
	c := course.Course{ID: id}
	if err := c.GetCourse(a.DB); err != nil {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}
	respondWithJSON(w, http.StatusOK, c)
}

// RETURNS AN ARRAY OF COURSES AS A JSON OBJECT
func (a *App) GetManyCourses(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}
	courses, err := course.GetManyCourses(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, courses)
}

// CREATES A COURSE INTO DATABASE
func (a *App) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var c course.Course
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := c.CreateCourse(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, c)
}

// UPDATES A COURSE INFORMATION IN DATABASE
func (a *App) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid course identifier")
		return
	}
	var c course.Course
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	c.ID = id

	if err := c.UpdateCourse(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, c)
}

// DELETES A COURSE IN DATABASE
func (a *App) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid course identifier")
		return
	}
	c := course.Course{ID: id}
	if err := c.DeleteCourse(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

/*

	HELPER + TESTING FUNCTIONS

*/

// HELPER FUNCTION TO RETURN PAYLOAD AS A JSON OBJECT AS WELL AS RETURN APPROPRIATE ERROR CODES IF ANY
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// HELPER FUNCTION TO RETURN AN ERROR AS A JSON OBJECT
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// THESE ARE ALL TEST FUNCTIONS WHEN TRYING TO GET FRONTEND AND BACKEND WORKING TOGETHER
func (a *App) TestPrintComm(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Test Successful")
}

func (a *App) TestGET(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get recieved...")
	setCORSHeader(&w, r)

	if (*r).Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"username": "successGet"})
}

func (a *App) TestPOST(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Post recieved...")
	setCORSHeader(&w, r)

	if (*r).Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"username": "successPost"})
}

func (a *App) TestDialogFlow(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DialogFlow recieved...")
	setCORSHeader(&w, r)

	if (*r).Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"username": "successPost"})
}

// Sets header for CORS. Allows for communication between Angular and GO on different ports.
func setCORSHeader(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// Sets up routes that need handling -> WHEN ROUTER SEES A HTTP REQUEST MATCHING THE TYPE AND URL, EXECUTE A GIVEN FUNCTION
func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/userstest", a.TestGET).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/userstest", a.TestPOST).Methods("POST", "OPTIONS")

	a.Router.HandleFunc("/users", a.GetManyUsers).Methods("GET")
	a.Router.HandleFunc("/users", a.CreateUser).Methods("POST")
	a.Router.HandleFunc("/users/{id:[0-9]}", a.GetUser).Methods("GET")
	a.Router.HandleFunc("/users/{id:[0-9]}", a.UpdateUser).Methods("PUT")
	a.Router.HandleFunc("/users/{id:[0-9]}", a.DeleteUser).Methods("DELETE")

	a.Router.HandleFunc("/dialogflow/", a.TestDialogFlow).Methods("POST")
}
