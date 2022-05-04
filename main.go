package main

import (
	"context"
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

type AccountUpdateRequest struct {
	Type   string
	Amount int
}

type Account struct {
	Id      int `json:"id"`
	Balance int `json:"balance"`
}

// Create the JWT key used to create the signature
var jwtKey = []byte("my_secret_key")

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	UserId int `json:"userId"`
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
		accounts, err := getUserAccounts(db, claims.UserId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"accounts": accounts})
	})

	r.GET("/account/:id", func(c *gin.Context) {
		tknStr := c.GetHeader("x-authentication-token")
		accountId := c.Param("id")
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
		account, err := getUserAccountById(db, claims.UserId, accountId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"account": account})
	})

	r.POST("/account/:id", func(c *gin.Context) {
		tknStr := c.GetHeader("x-authentication-token")
		accountId := c.Param("id")
		var payload AccountUpdateRequest

		err := c.BindJSON(&payload)
		if err != nil {
			log.Panicln(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		validPayload(payload)
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

		// perform update
		err = updateUserAccountById(db, claims.UserId, accountId, payload)
		if err != nil {
			log.Panicln(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusNoContent, gin.H{})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func updateUserAccountById(db *sql.DB, customerId int, accountId string, payload AccountUpdateRequest) error {
	var queryCustomerId int
	// verify user has this account
	row := db.QueryRow(`SELECT customer_id FROM test.customer_account WHERE customer_id = $1 AND account_id = $2`,
		customerId, accountId)
	err := row.Scan(&queryCustomerId)

	switch err {
	case sql.ErrNoRows:
		return errors.New("account not found")
	case nil:
		ctx := context.Background()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Panicln(err)
		}
		defer tx.Rollback()

		var balance int
		if err = tx.QueryRowContext(ctx, `SELECT balance FROM test.account WHERE id = $1`,
			accountId).Scan(&balance); err != nil {
			return err
		}

		if payload.Type == "withdraw" {
			if balance-payload.Amount < 0 {
				return errors.New("not enough fund for withdraw")
			}
			balance -= payload.Amount
		} else if payload.Type == "deposit" {
			balance += payload.Amount
		}

		_, err = tx.ExecContext(ctx, `UPDATE test.account SET balance = $1 WHERE id = $2`,
			balance, accountId)
		if err != nil {
			log.Panicln(err)
		}

		err = tx.Commit()
		if err != nil {
			log.Panicln(err)
		}

		return nil
	default:
		return err
	}
}

func getUserAccountById(db *sql.DB, customerId int, accountId string) (Account, error) {
	var account Account
	var queryCustomerId int
	// verify user has this account
	row := db.QueryRow(`SELECT customer_id FROM test.customer_account WHERE customer_id = $1 AND account_id = $2`,
		customerId, accountId)
	err := row.Scan(&queryCustomerId)

	switch err {
	case sql.ErrNoRows:
		return account, errors.New("account not found")
	case nil:
		row = db.QueryRow(`SELECT id, balance FROM test.account WHERE id = $1`, accountId)
		err = row.Scan(&account.Id, &account.Balance)

		if err != nil {
			return account, err
		}

		return account, nil
	default:
		return account, err
	}
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
		UserId: customerId,
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

func validPayload(payload AccountUpdateRequest) bool {
	if payload.Type != "withdraw" && payload.Type != "deposit" {
		return false
	}
	if payload.Amount < 0 {
		return false
	}

	return true
}
