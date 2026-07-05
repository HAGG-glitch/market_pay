package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    hash := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"
    passwords := []string{"Admin@1234", "password", "admin", "Admin@123", "Password123"}
    for _, p := range passwords {
        err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(p))
        if err == nil {
            fmt.Printf("MATCH: %q\n", p)
        } else {
            fmt.Printf("NO MATCH: %q\n", p)
        }
    }
}
