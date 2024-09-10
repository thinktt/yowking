package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/thinktt/yowking/pkg/auth"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: jot <username> [comma-separated-roles]")
		return
	}

	username := os.Args[1]
	var roles []string

	if len(os.Args) > 2 {
		roles = strings.Split(os.Args[2], ",")
	} else {
		roles = []string{"user"}
	}

	tokenStr, _, err := auth.MakeToken(username, roles)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(tokenStr)
}
