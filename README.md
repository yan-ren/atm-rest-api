## Problem statement

Implement a web service containing REST API for an ATM. The use cases that you must fulfill are:
- A customer can login to the ATM account by providing a 4 digit PIN number
- A customer can view their current balance
- A customer can deposit money
- A customer can withdraw money

## API Specs

### `POST /login`
Endpoint to log an user in. The payload should have the following fields:

```json
{
  "email": "test@axiomzen.co",
  "password": "1234"
}
```

The response body should return a JWT on success that can be used for other endpoints:

```json
{
  "token": "some_jwt_token"
}
```

### `GET /account`
Endpoint to retrieve a json of all accounts under costomer. This endpoint requires a valid `x-authentication-token` header to be passed in with the request.

The response body should look like:
```json
{
  "accounts": [1, 2]
}
```

### `GET /account/1`
Endpoint to retrieve a json of a specific account under costomer. This endpoint requires a valid `x-authentication-token` header to be passed in with the request.

The response body should look like:
```json
{
  "account": 
  {
    "id": 1,
    "balance": 100
  }
}
```

### `POST /account/1`
Endpoit withdraw/deposit to a specific account, account id should be in URL path

This endpoint requires a valid `x-authentication-token` header to be passed in with the request.

The payload should have the following fields:

```json
{
  "type": "withdraw/deposit",
  "amount": 5
}
```

The response body should look like:
```json
{
  "account": 
  {
    "id": 1,
    "balance": 95
  }
}
```

## Design Spec

To determine whether a request is from a logged in user or not, I use Json Web Tokens (https://jwt.io/). The frontend will be sending requests with the JWT in the x-authentication-token header.

For the database, I use PostgreSQL with go package https://github.com/go-pg/pg

## How to run the API

#### Run with Docker
This repository is provided with Dockerfile, docker-compose.yml and makefile. 

To run the API, simply run with
```
make up
```
Test API is accessible through end point
http://localhost:8080/ping

Other end points are:
- POST http://locahost:8080/singup
- POST http://locahost:8080/login
- GET http://locahost:8080/users
- PUT http://locahost:8080/users

To run integration test
```
make test -k
```