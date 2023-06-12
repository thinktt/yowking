package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var lichessUrl = "https://lichess.org/api/account"

// hard coded for now, will be deligated to a db or api later
var validUsers = []string{"thinktt"}

// would be good to run this through a validator since it's coming from outside
// even if lichess is  a trusted source
type LichessAccount struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

func GetToken(lichessToken string) (string, error) {
	req, err := http.NewRequest("GET", lichessUrl, nil)
	if err != nil {
		fmt.Println("error creating lichess request: ", err)
		return "", &ServerError{"error creating lichess request"}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", lichessToken))

	client := &http.Client{
		Timeout: time.Second * 5,
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error contacting lichess: ", err)
		return "", &ServerError{"error contactiing lichess"}
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		errMsg := fmt.Sprintf("lichess auth failed with: %s", res.Status)
		fmt.Println(errMsg)
		return "", &AuthError{errMsg}
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error reading lichess response: ", err)
		return "", &ServerError{"error reading lichess response"}
	}

	var account LichessAccount
	err = json.Unmarshal(body, &account)
	if err != nil {
		fmt.Println("error parsing lichess response: ", err)
		return "", &ServerError{"error parsing lichess response"}
	}

	if !isValidUser(account.Username) {
		errMsg := fmt.Sprintf("no authorization found for user %s", account.Username)
		return "", &AuthError{errMsg}
	}

	key := []byte("fake-key")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss":   "yeoldwizard.com",
			"sub":   account.Id,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
			"roles": []string{"mover"},
		})

	tokenStr, err := token.SignedString(key)
	if err != nil {
		fmt.Println("error creating token:", err)
		return "", &ServerError{"error creating token"}
	}

	return tokenStr, nil
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
