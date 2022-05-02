package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Declare the expiration time of the token
// here, we have kept it as 5 minutes
const TokenExpiration = 5 * time.Minute

type Customer struct {
	Id        int
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type UserGetResponse struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type UserUpdateRequest struct {
	FirstName string
	LastName  string
}

// Create the JWT key used to create the signature
var jwtKey = []byte("my_secret_key")

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	Id int `json:"id"`
	jwt.StandardClaims
}

const (
	host     = "db-postgres"
	port     = 5432
	user     = "admin"
	password = "admin123"
	dbname   = "dev"
)

func main() {
	dbInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, user, password, dbname, port)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		panic(err)
	}

	log.Printf("Postgres started at %d PORT", port)
	defer db.Close()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, "success")
	})

	r.POST("/login", func(c *gin.Context) {
		var user Customer
		c.BindJSON(&user)
		token, err := login(user, db)
		if err != nil {
			log.Panicln(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"token": token})
	})

	r.GET("/account", func(c *gin.Context) {
		tknStr := c.GetHeader("x-authentication-token")
		// Initialize a new instance of `Claims`
		claims := &Claims{}

		// Parse the JWT string and store the result in `claims`.
		// Note that we are passing the key in this method as well. This method will return an error
		// if the token is invalid (if it has expired according to the expiry time we set on sign in),
		// or if the signature does not match
		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token error"})
			return
		}
		if !tkn.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		//
		accounts, err := getUserAccounts(db, claims.Id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"accounts": accounts})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getUserAccounts(db *sql.DB, customerId int) ([]int, error) {
	accounts := make([]int, 0)
	rows, err := db.Query(`SELECT account_id FROM test.customer_account WHERE customer_id= $1`, customerId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var accountId int
		err := rows.Scan(&accountId)
		if err != nil {
			log.Panicln(err)
		}
		accounts = append(accounts, accountId)
	}

	err = rows.Err()
	if err != nil {
		log.Panicln(err)
	}

	return accounts, nil
}

func login(user Customer, db *sql.DB) (string, error) {
	// check if customer exixts
	var result Customer
	row := db.QueryRow(`SELECT id FROM test.customer WHERE email = $1 AND password = $2`, user.Email, user.Password)
	err := row.Scan(&result.Id)

	switch err {
	case sql.ErrNoRows:
		return "", errors.New("user does not exist or invalid password")
	case nil:
		token, err := getJWT(result.Id)
		if err != nil {
			return "", err
		}
		return token, nil
	default:
		return "", err
	}
}

func getJWT(customerId int) (string, error) {
	expirationTime := time.Now().Add(TokenExpiration)
	// Create the JWT claims, which includes the email and expiry time
	claims := &Claims{
		Id: customerId,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Panicln(err)
		return "", errors.New("token generation error")
	}

	return tokenString, nil
}
