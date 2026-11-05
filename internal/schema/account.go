package schema

// Health is the body of GET /healthz.
type Health struct {
	Status string `json:"status"`
}

// User is the account of the token, from GET /auth/me (UserResponse).
type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}
