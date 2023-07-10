package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/thinktt/yowking/pkg/auth"
	"github.com/thinktt/yowking/pkg/books"
	"github.com/thinktt/yowking/pkg/engine"
	"github.com/thinktt/yowking/pkg/models"
	"github.com/thinktt/yowking/pkg/personalities"
)

func main() {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	config.AllowHeaders = []string{"*"}

	r := gin.New()
	r.Use(cors.New(config))

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

	r.Use(Auth())

	r.POST("/move-req", CheckRole("mover"), func(c *gin.Context) {
		var moveReq models.MoveReq
		if err := c.ShouldBindJSON(&moveReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cmp, ok := personalities.CmpMap[moveReq.CmpName]
		if !ok {
			errMsg := fmt.Sprintf("%s is not a valid personality", moveReq.CmpName)
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}

		bookMove, err := books.GetMove(moveReq.Moves, cmp.Book)
		// if no err we have a book move and can just return the move
		if err == nil {
			c.JSON(http.StatusOK, bookMove)
			return
		}

		settings := models.Settings{
			Moves:     moveReq.Moves,
			CmpVals:   cmp.Vals,
			ClockTime: personalities.GetClockTime(cmp),
		}

		moveData, err := engine.GetMove(settings)
		if err != nil {
			fmt.Println("There was ane error getting the move: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"messagge": "engine error"})
			return
		}

		// engine didn't accept the input, return a 400 error
		if moveData.Err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": *moveData.Err})
			return
		}

		moveData.WillAcceptDraw = personalities.GetDrawEval(moveData.Eval, settings)
		moveData.Type = "engine"
		c.JSON(http.StatusOK, moveData)
	})

	r.Run(":8082")
	// r.RunTLS(":8443", "../certs/cert.pem", "../certs/key.pem")
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
