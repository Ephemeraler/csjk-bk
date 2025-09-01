package options

var Databases = &Database{}

type Database struct {
	Postgres Postgres `group:"Database" namespace:"postgres" description:"PostgreSQL connection options"`
}

type Postgres struct {
	DSN string `long:"dsn" description:"PostgreSQL DSN"`
}
