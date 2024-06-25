package config

var (
	VERSION string
	COMMIT  string
)

func Version() string {
	if len(VERSION) == 0 {
		return "0.0.0"
	}
	return VERSION
}

func Commit() string {
	if len(COMMIT) == 0 {
		return "N/A"
	}
	return COMMIT
}
