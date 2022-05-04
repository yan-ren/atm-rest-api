package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

const LoginEndpoint = "http://localhost:8080/login"
const Authorization = "x-authentication-token"

type TokenResponse struct {
	Token string `json:"token"`
}

type GetAccountResponse struct {
	Accounts []int `json:"accounts"`
}

type GetAccountByIdResponse struct {
	Account Account `json:"account"`
}

func TestMain(t *testing.T) {
	client := http.Client{}
	var token TokenResponse

	// test POST /login
	reqStr := `{
		"email": "test@axiomzen.co",
		"password": "1234"
	  }`
	req, err := http.NewRequest(echo.POST, LoginEndpoint, strings.NewReader(reqStr))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	response, err := client.Do(req)
	assert.NilError(t, err)
	assert.Equal(t, http.StatusCreated, response.StatusCode)

	byteBody, err := ioutil.ReadAll(response.Body)
	assert.NilError(t, err)

	if error := json.Unmarshal(byteBody, &token); err != nil {
		panic(error)
	}
	if token.Token == "" {
		t.Logf("Empty token")
		t.Fail()
	}

	// test GET /account
	req, err = http.NewRequest(echo.GET, "http://localhost:8080/account", strings.NewReader(""))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(Authorization, token.Token)

	response, err = client.Do(req)
	assert.NilError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	byteBody, err = ioutil.ReadAll(response.Body)
	assert.NilError(t, err)

	var getAccountResponse GetAccountResponse
	if error := json.Unmarshal(byteBody, &getAccountResponse); err != nil {
		panic(error)
	}
	assert.Assert(t, is.Len(getAccountResponse.Accounts, 2))
	assert.Assert(t, is.Contains(getAccountResponse.Accounts, 1))
	assert.Assert(t, is.Contains(getAccountResponse.Accounts, 2))

	// test GET /account/1
	req, err = http.NewRequest(echo.GET, "http://localhost:8080/account/1", strings.NewReader(""))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(Authorization, token.Token)

	response, err = client.Do(req)
	assert.NilError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	byteBody, err = ioutil.ReadAll(response.Body)
	assert.NilError(t, err)

	var getAccountByIdResponse GetAccountByIdResponse
	if error := json.Unmarshal(byteBody, &getAccountByIdResponse); err != nil {
		panic(error)
	}
	assert.DeepEqual(t, getAccountByIdResponse.Account, Account{Id: 1, Balance: 100})

	// test withdraw
	reqStr = `{
		"type": "withdraw",
		"amount": 5
	}`
	req, err = http.NewRequest(echo.POST, "http://localhost:8080/account/1", strings.NewReader(reqStr))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(Authorization, token.Token)

	response, err = client.Do(req)
	assert.NilError(t, err)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)

	// verify new balance
	req, err = http.NewRequest(echo.GET, "http://localhost:8080/account/1", strings.NewReader(""))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(Authorization, token.Token)

	response, err = client.Do(req)
	assert.NilError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	byteBody, err = ioutil.ReadAll(response.Body)
	assert.NilError(t, err)

	if error := json.Unmarshal(byteBody, &getAccountByIdResponse); err != nil {
		panic(error)
	}
	assert.DeepEqual(t, getAccountByIdResponse.Account, Account{Id: 1, Balance: 95})

	// test deposit
	reqStr = `{
		"type": "deposit",
		"amount": 10
	}`
	req, err = http.NewRequest(echo.POST, "http://localhost:8080/account/1", strings.NewReader(reqStr))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(Authorization, token.Token)

	response, err = client.Do(req)
	assert.NilError(t, err)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)

	// verify new balance
	req, err = http.NewRequest(echo.GET, "http://localhost:8080/account/1", strings.NewReader(""))
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(Authorization, token.Token)

	response, err = client.Do(req)
	assert.NilError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	byteBody, err = ioutil.ReadAll(response.Body)
	assert.NilError(t, err)

	if error := json.Unmarshal(byteBody, &getAccountByIdResponse); err != nil {
		panic(error)
	}
	assert.DeepEqual(t, getAccountByIdResponse.Account, Account{Id: 1, Balance: 105})

	response.Body.Close()
}
