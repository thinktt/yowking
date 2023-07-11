package main

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/move-req", func(c *gin.Context) {
		url := "https://yowking.localhost"
		method := "GET"

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
				ForceAttemptHTTP2: false,
			},
		}

		req, _ := http.NewRequest(method, url, nil)
		req.Header.Set("Connection", "close")

		res, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer res.Body.Close()

		body, _ := ioutil.ReadAll(res.Body)

		c.Data(res.StatusCode, res.Header.Get("Content-Type"), body)
	})

	router.NoRoute(func(c *gin.Context) {
		serveReverseProxy("https://yowking.localhost", c)
	})

	router.Run(":8083")
}

func serveReverseProxy(target string, c *gin.Context) {
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		ForceAttemptHTTP2: false,
	}

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Host = url.Host
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
