package config

var defaultConfig = map[string]interface{}{
	"server.host": "localhost",
	"server.port": "8080",

	"db.host":     "localhost",
	"db.user":     "postgres",
	"db.password": "postgres",
	"db.port":     5432,
}
