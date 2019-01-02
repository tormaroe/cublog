package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// BlogAdmin represents credentials of a user
type BlogAdmin struct {
	Username string
	Password string
	IsParent bool
}

func LoadAdminsFromFile(path string) ([]BlogAdmin, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	admins := &[]BlogAdmin{}
	err = json.Unmarshal(bytes, admins)
	fmt.Printf("Loaded %d users\n", len(*admins))
	return *admins, err
}
