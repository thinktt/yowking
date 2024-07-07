package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/thinktt/yowking/pkg/auth"
	"github.com/thinktt/yowking/pkg/db"
	"github.com/thinktt/yowking/pkg/kingcheck"
	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/moveque"
)

var cmpMap = make(map[string]models.Cmp)

func main() {
	loadCmps()
	fmt.Println(cmpMap["Ash"])

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	config.AllowHeaders = []string{"*"}
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("No PORT environment variable detected, defaulting to 8080")
		port = "8080"
	}

	r := gin.New()
	r.Use(cors.New(config))

	// catch and remove trailing slashes on routes and start router again
	// this let's us handle trailing slashes without redirecting
	r.RedirectTrailingSlash = false
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) > 1 && path[len(path)-1] == '/' {
			c.Request.URL.Path = path[:len(path)-1]
			r.HandleContext(c)
			c.Abort()
			return
		}
		c.Next()
	})

	healthWasCalled := false
	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/health" && healthWasCalled {
			c.Next()
			return
		}
		gin.Logger()(c)
		c.Next()
	})

	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		healthWasCalled = true
		c.JSON(http.StatusOK, gin.H{
			"message": "API is healthy",
		})
	})

	r.POST("/users", func(c *gin.Context) {
		var userReq models.UserRequest
		if err := c.ShouldBindJSON(&userReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !userIsValid(userReq.ID) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID, does not match user regex"})
			return
		}

		// checking settings for user count limit and king requirement
		settings, err := db.GetSettings()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "DB Error getting settings: " + err.Error()})
			return
		}

		users, err := db.GetAllUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "DB Error getting user count: " + err.Error()})
			return
		}

		if users.Count >= settings.UserLimit {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "user limit reached, no new users can be added at this time",
			})
			return
		}

		kingIsRequired := settings.KingIsRequired

		var kingCmVersion string
		if kingIsRequired {
			kingCmVersion, err = kingcheck.GetVersion(userReq.KingBlob)
			if err != nil || kingCmVersion == "" {
				c.JSON(http.StatusBadRequest, gin.H{"message": "king blob is not valid"})
				return
			}
		} else {
			// king is not required, we will set this user as Beta user
			kingCmVersion = "B"
		}

		user := models.User{
			ID:                    userReq.ID,
			KingCmVersion:         kingCmVersion,
			HasAcceptedDisclaimer: userReq.HasAcceptedDisclaimer,
		}

		result, err := db.CreateUser(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "DB Error: " + err.Error()})
			return
		}

		if result.MatchedCount > 0 {
			c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("user %s already exist, no new creation", user.ID)})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("user %s successfully added", user.ID)})
	})

	r.GET("/users/:id", func(c *gin.Context) {
		userID := c.Param("id")

		result, err := db.GetUser(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error: " + err.Error()})
			return
		}

		if result == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no user found for id %s", userID)})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	r.POST("/games", func(c *gin.Context) {
		var game models.Game

		// Binding and validation
		if err := c.ShouldBindJSON(&game); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !userIsValid(game.User) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID, does not match user regex"})
			return
		}

		result, err := db.CreateGame(game)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "DB Error: " + err.Error()})
			return
		}

		// MongoDB already had this ID in the DB, so it didn't create a new one
		if result.MatchedCount > 0 {
			c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("game %s already exist, no new creation", game.ID)})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("game %s successfully added", game.ID)})
	})

	r.GET("/games/:id", func(c *gin.Context) {
		id := c.Param("id")

		game, err := db.GetGame(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error: " + err.Error()})
			return
		}

		if game == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no game found for id %s", id)})
			return
		}

		c.JSON(http.StatusOK, game)
	})

	r.POST("/games2", func(c *gin.Context) {
		var game models.Game2

		// Binding and validation
		if err := c.ShouldBindJSON(&game); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// if !userIsValid(game.User) {
		// 	c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID, does not match user regex"})
		// 	return
		// }

		result, err := db.CreateGame2(game)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "DB Error: " + err.Error()})
			return
		}

		// MongoDB already had this ID in the DB, so it didn't create a new one
		if result.MatchedCount > 0 {
			c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("game %s already exist, no new creation", game.ID)})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("game %s successfully added", game.ID)})
	})

	r.GET("/games2/:id", func(c *gin.Context) {
		id := c.Param("id")

		game, err := db.GetGame2(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error: " + err.Error()})
			return
		}

		emptyGame := models.Game2{}
		if game == emptyGame {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("no game found for id %s", id)})
			return
		}

		c.JSON(http.StatusOK, game)
	})

	r.Use(PullToken())

	// get a yow jwt token by sending a lichess token
	r.GET("/token", func(c *gin.Context) {
		lichessToken := c.GetString("token")

		tokenRes, err := auth.GetToken(lichessToken)
		if err != nil {
			switch err.(type) {
			case *auth.AuthError:
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, tokenRes)
	})

	//......................................
	// ..... Authed routes start here......
	//.....................................

	r.Use(Auth())

	r.GET("/users", CheckRole("admin"), func(c *gin.Context) {
		usersResponse, err := db.GetAllUsers()
		if err != nil {
			// Handle the error appropriately
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count": usersResponse.Count,
			"ids":   usersResponse.IDs,
		})
	})

	r.DELETE("/users/:id", CheckRole("admin"), func(c *gin.Context) {
		id := c.Param("id")

		result, err := db.DeleteUser(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error: " + err.Error()})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "No user found with given ID"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	})

	r.GET("/games", CheckRole("admin"), func(c *gin.Context) {
		userId := c.Query("userId")

		allGames, err := db.GetAllGames(userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, allGames)
	})

	r.DELETE("/games/:id", CheckRole("admin"), func(c *gin.Context) {
		id := c.Param("id")

		result, err := db.DeleteGame(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error: " + err.Error()})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "Game not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Game deleted successfully"})
	})

	r.POST("/settings", CheckRole("admin"), func(c *gin.Context) {
		var settings models.Settings
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "DB Error: " + err.Error()})
			return
		}

		if err := db.UpdateSettings(settings); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
	})

	r.GET("/settings", CheckRole("admin"), func(c *gin.Context) {
		settings, err := db.GetSettings()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Error: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, settings)
	})

	r.POST("/move-req", CheckRole("mover"), func(c *gin.Context) {
		var moveReq models.MoveReq
		if err := c.ShouldBindJSON(&moveReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, ok := cmpMap[moveReq.CmpName]
		if !ok {
			errMsg := fmt.Sprintf("%s is not a valid personality", moveReq.CmpName)
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}

		moveData, err := moveque.GetMove(moveReq)
		if err != nil {
			fmt.Println("There was ane error getting the move: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"messagge": "queue error"})
			return
		}

		// worker failed to process move requests, return a 400 error
		if moveData.Err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": *moveData.Err})
			return
		}

		c.JSON(http.StatusOK, moveData)
	})

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"messagge": "404 Not Found"})
	})

	if port == "8443" {
		r.RunTLS(":8443", "../certs/cert.pem", "../certs/key.pem")
		return
	}

	r.Run(":" + port)
}

func PullToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "No Authorization header provided",
			})
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization header",
			})
			c.Abort()
			return
		}
		token := parts[1]
		c.Set("token", token)
		c.Next()
	}
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetString("token")

		claims, err := auth.CheckToken(tokenStr)
		if err != nil {
			switch err.(type) {
			case *auth.AuthError:
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				c.Abort()
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		}

		c.Set("roles", claims.Roles)
		c.Next()
	}
}

func CheckRole(allowedRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := c.GetStringSlice("roles")

		for _, role := range roles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect role for this action"})
		c.Abort()
	}
}

func userIsValid(userId string) bool {
	regexPattern := `^[a-zA-Z0-9][a-zA-Z0-9_-]{0,28}[a-zA-Z0-9]$`
	matched, _ := regexp.MatchString(regexPattern, userId)
	return matched
}

func loadCmps() {
	file, err := os.Open("personalities.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&cmpMap)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
	}
}
