package communication

type Config struct {
	HttpAddr string
	Listener string
	Router   RouterConfig
	Client   ClientConfig
}
