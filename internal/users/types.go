package users

import "time"

// Definition of the creation of a new user.
type createUserParams struct {
	First_name string `json:"first_name"`
	Last_name  string `json:"last_name"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Email      string `json:"email"`
}

// Definition of logging in.
type loginUserParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRes struct {
	Session_ID               string    `json:"session_id"`
	Access_token             string    `json:"access_token"`
	Refresh_token            string    `json:"refresh_token"`
	Access_token_expires_at  time.Time `json:"access_token_expires_at"`
	Refresh_token_expires_at time.Time `json:"refresh_token_expires_at"`
	Email                    string    `json:"email"`
}

type createSessionParams struct {
	ID            string    `json:"id"`
	User_email    string    `json:"user_email"`
	Refresh_token string    `json:"refresh_token"`
	Is_revoked    bool      `json:"is_revoked"`
	Expires_at    time.Time `json:"expires_at"`
}

type RenewAccessTokenReq struct {
	Refresh_token string `json:"refresh_token"`
}

type RenewAccessTokenRes struct {
	Access_token            string    `json:"access_token"`
	Access_token_expires_at time.Time `json:"access_token_expires_at"`
}
