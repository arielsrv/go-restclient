package rest_test

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
	"golang.org/x/oauth2"
)

var lastModifiedDate = time.Now()

type User struct {
	Name string `json:"name" xml:"name"`
	ID   int    `json:"id"   xml:"id"`
}

var (
	tmux   = http.NewServeMux()
	server = httptest.NewServer(tmux)
)

var users []User

var userList = []string{
	"Alice", "Bob", "Maria",
}

var rb = rest.Client{
	BaseURL: server.URL,
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	//	teardown()
	os.Exit(code)
}

func setup() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	users = make([]User, len(userList))
	for i, n := range userList {
		users[i] = User{ID: i + 1, Name: n}
	}

	// users
	tmux.HandleFunc("/user", allUsers)
	tmux.HandleFunc("/xml/user", usersXML)
	tmux.HandleFunc("/form/user", usersForm)
	tmux.HandleFunc("/bytes/user", usersBytes)
	tmux.HandleFunc("/cache/user", usersCache)
	tmux.HandleFunc("/cache/expires/user", usersCacheWithExpires)
	tmux.HandleFunc("/cache/etag/user", usersEtag)
	tmux.HandleFunc("/cache/lastmodified/user", usersLastModified)
	tmux.HandleFunc("/slow/cache/user", slowUsersCache)
	tmux.HandleFunc("/slow/user", slowUsers)
	tmux.HandleFunc("/auth", auth)
	tmux.HandleFunc("/auth/token", authToken)
	tmux.HandleFunc("/problem", problem)
	tmux.HandleFunc("/problem_err", problemErr)

	// One user
	tmux.HandleFunc("/user/", oneUser)

	// Header
	tmux.HandleFunc("/header", withHeader)
}

func withHeader(writer http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		if h := req.Header.Get("X-Test"); h == "test" {
			return
		}

		h1 := req.Header.Get("X-Params-Test")
		h2 := req.Header.Get("X-Default-Test")
		if h1 == "test" && h2 == "test" {
			return
		}
	}

	writer.WriteHeader(http.StatusBadRequest)
}

func slowUsersCache(writer http.ResponseWriter, req *http.Request) {
	time.Sleep(30 * time.Millisecond)
	usersCache(writer, req)
}

func slowUsers(writer http.ResponseWriter, req *http.Request) {
	time.Sleep(10 * time.Millisecond)
	allUsers(writer, req)
}

func auth(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func authToken(writer http.ResponseWriter, _ *http.Request) {
	token := new(oauth2.Token)
	token.AccessToken = "access_token"
	token.RefreshToken = "refresh_token"
	token.Expiry = time.Now().Add(time.Duration(30) * time.Minute)
	token.TokenType = "Bearer"

	ub, err := json.Marshal(token)
	if err != nil {
		log.Fatal(err)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(ub)
}

func usersCache(writer http.ResponseWriter, req *http.Request) {
	// Get
	if req.Method == http.MethodGet {
		c := rand.Intn(2) + 1
		b, _ := json.Marshal(users)

		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Cache-Control", "max-age="+strconv.Itoa(c))
		writer.Write(b)
	}
}

func usersCacheWithExpires(writer http.ResponseWriter, req *http.Request) {
	// Get
	if req.Method == http.MethodGet {
		c := rand.Intn(2) + 1
		b, _ := json.Marshal(users)

		expires := time.Now().Add(time.Duration(c) * time.Second)

		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Expires", expires.Format(time.RFC1123))
		writer.Write(b)
	}
}

func usersEtag(writer http.ResponseWriter, req *http.Request) {
	// Get
	if req.Method == http.MethodGet {
		etag := req.Header.Get("If-None-Match")

		if etag == "1234" {
			writer.WriteHeader(http.StatusNotModified)
			return
		}

		b, _ := json.Marshal(users)

		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Etag", "1234")
		writer.Write(b)
	}
}

func usersLastModified(writer http.ResponseWriter, req *http.Request) {
	// Get
	if req.Method == http.MethodGet {
		ifModifiedSince, err := time.Parse(time.RFC1123, req.Header.Get("If-Modified-Since"))

		if err == nil && ifModifiedSince.Sub(lastModifiedDate) == 0 {
			writer.WriteHeader(http.StatusNotModified)
			return
		}

		b, _ := json.Marshal(users)

		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Last-Modified", lastModifiedDate.Format(time.RFC1123))
		writer.Write(b)
	}
}

func usersXML(writer http.ResponseWriter, req *http.Request) {
	// Get
	if req.Method == http.MethodGet {
		b, _ := xml.Marshal(users)

		writer.Header().Set("Content-Type", "application/xml")
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Write(b)
	}

	// Post
	if req.Method == http.MethodPost {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		u := new(User)
		if err = xml.Unmarshal(b, u); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		u.ID = 3
		ub, _ := json.Marshal(u)

		writer.Header().Set("Content-Type", "application/xml")
		writer.WriteHeader(http.StatusCreated)
		writer.Write(ub)

		return
	}
}

func usersForm(writer http.ResponseWriter, req *http.Request) {
	// Get
	if req.Method == http.MethodGet {
		b, _ := json.Marshal(users)

		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Write(b)
	}

	// Post
	if req.Method == http.MethodPost {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		form, err := url.ParseQuery(string(b))
		if err != nil {
			log.Error(err)
			return
		}

		u := new(User)
		u.ID = 3
		u.Name = form["name"][0]

		ub, err := json.Marshal(u)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		writer.Write(ub)

		return
	}
}

func usersBytes(writer http.ResponseWriter, req *http.Request) {
	// Post
	if req.Method == http.MethodPost {
		req.Body = http.MaxBytesReader(writer, req.Body, 10<<20) // 10 MB

		_, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(writer, "read file error", http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		writer.WriteHeader(http.StatusCreated)
		writer.Write([]byte("success"))
	}
}

func oneUser(writer http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		b, _ := json.Marshal(users[0])

		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Write(b)
		return
	}

	// Put
	if req.Method == http.MethodPut || req.Method == http.MethodPatch {
		b, _ := json.Marshal(users[0])

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(b)
		return
	}

	// Delete
	if req.Method == http.MethodDelete {
		return
	}
}

func allUsers(writer http.ResponseWriter, req *http.Request) {
	// Head
	if req.Method == http.MethodHead {
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Cache-Control", "no-cache")
		return
	}

	// Get
	if req.Method == http.MethodGet {
		b, _ := json.Marshal(users)

		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Write(b)
		return
	}

	// Post
	if req.Method == http.MethodPost {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		u := new(User)
		if err = json.Unmarshal(b, u); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		u.ID = 3
		ub, _ := json.Marshal(u)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		writer.Write(ub)

		return
	}

	// Options
	if req.Method == http.MethodOptions {
		b := []byte(`User resource
		id: ID of the user
		name: Name of the user`)

		writer.Header().Set("Content-Type", "text/plain")
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Write(b)
		return
	}
}

func problem(writer http.ResponseWriter, req *http.Request) {
	problemResponse := &rest.Problem{
		Type:     "https://httpstatuses.com/404",
		Title:    "Not Found",
		Detail:   "The requested resource was not found.",
		Status:   404,
		Instance: req.URL.String(),
	}
	b, _ := json.Marshal(problemResponse)

	writer.Header().Set("Content-Type", "application/problem+json")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.WriteHeader(problemResponse.Status)
	writer.Write(b)
}

func problemErr(writer http.ResponseWriter, _ *http.Request) {
	b, _ := xml.Marshal(`<invalid-request>`)

	writer.Header().Set("Content-Type", "application/problem+json")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.WriteHeader(http.StatusNotFound)
	writer.Write(b)
}
