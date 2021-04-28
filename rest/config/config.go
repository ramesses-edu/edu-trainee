package config

type DBCfg struct {
	UserDB string
	PassDB string
	HostDB string
	PortDB string
	NameDB string
}
type GoogleAuthCfg struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
}
type FacebookAuthCfg struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
}
type TwitterAuthCfg struct {
	TwitterAPIKey      string
	TwitterAPISecret   string
	TwitterTokenKey    string
	TwitterTokenSecret string
	RedirectURL        string
	ReqTokenURL        string
	AuthURL            string
	TokenURL           string
}

type Config struct {
	DB       DBCfg
	Google   GoogleAuthCfg
	Facebook FacebookAuthCfg
	Twitter  TwitterAuthCfg
	HostAddr string
}

func New() *Config {
	return &Config{}
}
