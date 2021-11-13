package communication

type Config struct {
	HttpAddr     string
	Listener     string
	UserIDHeader string
	Router       RouterConfig
	Client       ClientConfig
}
