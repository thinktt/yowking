package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var lichessUrl = "https://lichess.org/api/account"
var jwtKey = os.Getenv("JWT_KEY")

// hard coded for now, will be deligated to a db or api later
var validUsers = []string{"thinktt"}

// would be good to run this through a validator since it's coming from outside
// even if lichess is  a trusted source
type LichessAccount struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type TokenRes struct {
	Token  string                 `json:"token"`
	Claims map[string]interface{} `json:"claims"`
}

func GetToken(lichessToken string) (TokenRes, error) {
	req, err := http.NewRequest("GET", lichessUrl, nil)
	if err != nil {
		fmt.Println("error creating lichess request: ", err)
		return TokenRes{}, &ServerError{"error creating lichess request"}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", lichessToken))

	client := &http.Client{
		Timeout: time.Second * 5,
	}
	res, err := client.Do(req)
	if err != nil {
		errMsg := "error contacting lichess"
		fmt.Println(errMsg, err)
		return TokenRes{}, &ServerError{errMsg}
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		errMsg := fmt.Sprintf("lichess auth failed with: %s", res.Status)
		fmt.Println(errMsg)
		return TokenRes{}, &AuthError{errMsg}
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error reading lichess response: ", err)
		return TokenRes{}, &ServerError{"error reading lichess response"}
	}

	var account LichessAccount
	err = json.Unmarshal(body, &account)
	if err != nil {
		fmt.Println("error parsing lichess response: ", err)
		return TokenRes{}, &ServerError{"error parsing lichess response"}
	}

	if !isValidUser(account.Username) {
		errMsg := fmt.Sprintf("no authorization found for user %s", account.Username)
		return TokenRes{}, &AuthError{errMsg}
	}

	claims := jwt.MapClaims{
		"iss":   "yeoldwizard.com",
		"sub":   account.Id,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
		"roles": []string{"mover", "admin"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		fmt.Println("error creating token:", err)
		return TokenRes{}, &ServerError{"error creating token"}
	}

	tokenRes := TokenRes{
		Token:  tokenStr,
		Claims: claims,
	}

	return tokenRes, nil
}

func isValidUser(username string) bool {
	for _, user := range validUsers {
		if user == username {
			return true
		}
	}
	return false
}

type ServerError struct {
	Message string
}

func (e *ServerError) Error() string {
	return e.Message
}

type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

type Claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func CheckToken(tokenStr string) (Claims, error) {
	claims := &Claims{}

	// parse the token
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtKey), nil
	})

	if err != nil {
		errMsg := "invalid token format"
		fmt.Println(errMsg, err)
		return Claims{}, &AuthError{errMsg}
	}

	if !token.Valid {
		return Claims{}, fmt.Errorf("token is not valid")
	}

	// If we get here, everything worked and we can return the Claims
	return *claims, nil
}
