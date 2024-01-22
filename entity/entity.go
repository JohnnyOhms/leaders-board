package entity

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required,min=5"`
	UserId   string `json:"user_id"`
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password" binding:"required,min=5"`
}

type DiscordToken struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type UserData struct {
	ID                   string `json:"id"`
	Username             string `json:"username"`
	Avatar               string `json:"avatar"`
	Discriminator        string `json:"discriminator"`
	PublicFlags          int    `json:"public_flags"`
	PremiumType          int    `json:"premium_type"`
	Flags                int    `json:"flags"`
	Banner               string `json:"banner"`
	AccentColor          string `json:"accent_color"`
	GlobalName           string `json:"global_name"`
	AvatarDecorationData string `json:"avatar_decoration_data"`
	BannerColor          string `json:"banner_color"`
	MFAEnabled           bool   `json:"mfa_enabled"`
	Locale               string `json:"locale"`
	Email                string `json:"email"`
	Verified             bool   `json:"verified"`
}
