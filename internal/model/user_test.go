package model

import "testing"

func TestUserInit(t *testing.T) {
	u := User{
		Id:       1,
		UserId:   "u123",
		Nickname: "testuser",
		Password: "pass",
		Email:    "test@example.com",
	}
	if u.UserId != "u123" || u.Nickname != "testuser" {
		t.Errorf("User fields not set correctly")
	}
}

func TestUserTableName(t *testing.T) {
	u := User{}
	if u.TableName() != "users" {
		t.Errorf("TableName should return 'users'")
	}
}
