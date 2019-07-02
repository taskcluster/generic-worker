package runtime

import (
	"errors"
	"time"
)

func (user *OSUser) CreateNew(okIfExists bool) (err error) {
	return errors.New("(*(runtime.OSUser)).CreateNew(bool) not implemented on Linux")
}

func DeleteUser(username string) (err error) {
	return errors.New("runtime.DeleteUser(string) not implemented on Linux")
}

func ListUserAccounts() (usernames []string, err error) {
	return nil, errors.New("runtime.ListUserAccounts() not implemented on Linux")
}

func UserHomeDirectoriesParent() string {
	return "/home"
}

func WaitForLoginCompletion(timeout time.Duration) error {
	return errors.New("runtime.WaitForLoginCompletion(time.Duration) not implemented on Linux")
}

func InteractiveUsername() (string, error) {
	return "", errors.New("runtime.InteractiveUsername() not implemented on Linux")
}
