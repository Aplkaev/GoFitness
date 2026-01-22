package helper

import (
	"fmt"

	"gopkg.in/telebot.v3"
)

func GetUserName(user *telebot.User) string {
	return fmt.Sprintf("%s %s %s", user.Username, user.FirstName, user.LastName)
}