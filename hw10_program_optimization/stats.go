package hw10programoptimization

import (
	json "encoding/json"
	"fmt"
	"io"
	"strings"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

func getUsers(r io.Reader) ([]User, error) {
	var user User
	decoder := json.NewDecoder(r)
	result := make([]User, 0, 100_000)
	for {
		if err := decoder.Decode(&user); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		result = append(result, user)
	}
	return result, nil
}

func countDomains(u []User, domain string) (DomainStat, error) {
	result := make(DomainStat)

	for _, user := range u {
		if strings.Contains(user.Email, "."+domain) {
			result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]++
		}
	}
	return result, nil
}
