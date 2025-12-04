package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	plainPassword := "fanyo_1234"
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		// handle error
	}
	fmt.Println(string(hashed))
}
