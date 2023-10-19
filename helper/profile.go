package helper

type (
	Profile struct {
		Avatar        string
		Certificates  []Certificate
		FirstName     string
		LastName      string
		LastSeen      string
		Me            bool
		PasswordFound bool
		UserID        string
		UserName      string
		CreatedAt     string
	}

	Certificate struct {
		ExpireAt  string
		CreatedAt string
	}
)
