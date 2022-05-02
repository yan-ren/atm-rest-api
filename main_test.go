package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo"
)

const Login = "http://localhost:8080/login"
const Signup = "http://localhost:8080/signup"
const Users = "http://localhost:8080/users"
const Authorization = "x-authentication-token"

type TokenResponse struct {
	Token string `json:"token"`
}

func TestMain(t *testing.T) {
	// test POST /signup
	reqStr := `{
		"email": "test@axiomzen.co",
		"password": "axiomzen",
		"firstName": "Alex",
		"lastName": "Zimmerman"
	  }`
	req, err := http.NewRequest(echo.POST, Signup, strings.NewReader(reqStr))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	client := http.Client{}
	response, err := client.Do(req)
	assertNoError(t, err)
	assertEqual(t, http.StatusCreated, response.StatusCode)

	byteBody, err := ioutil.ReadAll(response.Body)
	assertNoError(t, err)

	var token TokenResponse
	if error := json.Unmarshal(byteBody, &token); err != nil {
		panic(error)
	}
	if token.Token == "" {
		t.Logf("Empty token")
		t.Fail()
	}

	// test GET /users
	req, err = http.NewRequest(echo.GET, Users, strings.NewReader(""))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(Authorization, token.Token)

	response, err = client.Do(req)
	assertNoError(t, err)
	assertEqual(t, http.StatusOK, response.StatusCode)

	byteBody, err = ioutil.ReadAll(response.Body)
	assertNoError(t, err)
	assertEqual(t, `{"users":[{"email":"test@axiomzen.co","firstName":"Alex","lastName":"Zimmerman"}]}`, strings.Trim(string(byteBody), "\n"))

	// test POST /login
	reqStr = `{
		"email": "test@axiomzen.co",
		"password": "axiomzen"
	  }`
	req, err = http.NewRequest(echo.POST, Login, strings.NewReader(reqStr))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	response, err = client.Do(req)
	assertNoError(t, err)
	assertEqual(t, http.StatusCreated, response.StatusCode)

	byteBody, err = ioutil.ReadAll(response.Body)
	assertNoError(t, err)

	if error := json.Unmarshal(byteBody, &token); err != nil {
		panic(error)
	}
	if token.Token == "" {
		t.Logf("Empty token")
		t.Fail()
	}

	// test PUT /users
	reqStr = `{
		"firstName": "NewFirstName",
		"lastName": "NewLastName"
	  }`
	req, err = http.NewRequest(echo.PUT, Users, strings.NewReader(reqStr))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(Authorization, token.Token)

	response, err = client.Do(req)
	assertNoError(t, err)
	assertEqual(t, http.StatusNoContent, response.StatusCode)

	// test GET /users, should retrieve latest value
	req, err = http.NewRequest(echo.GET, Users, strings.NewReader(""))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(Authorization, token.Token)

	response, err = client.Do(req)
	assertNoError(t, err)
	assertEqual(t, http.StatusOK, response.StatusCode)

	byteBody, err = ioutil.ReadAll(response.Body)
	assertNoError(t, err)
	assertEqual(t, `{"users":[{"email":"test@axiomzen.co","firstName":"NewFirstName","lastName":"NewLastName"}]}`, strings.Trim(string(byteBody), "\n"))

	// test GET /users, without token
	req, err = http.NewRequest(echo.GET, Users, strings.NewReader(""))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	response, err = client.Do(req)
	assertNoError(t, err)
	assertEqual(t, http.StatusUnauthorized, response.StatusCode)

	byteBody, err = ioutil.ReadAll(response.Body)
	assertNoError(t, err)
	assertEqual(t, `{"error":"token error"}`, strings.Trim(string(byteBody), "\n"))

	response.Body.Close()
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}
}
