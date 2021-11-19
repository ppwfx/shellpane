package communication

type Config struct {
	HttpAddr      string
	Listener      string
	UserIDHeader  string
	DefaultUserID string
	Router        RouterConfig
	Client        ClientConfig
}
