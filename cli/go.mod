module github.com/dezhishen/now-and-again/cli

go 1.25

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/spf13/cobra v1.8.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/gorm v1.31.2 // indirect
)

require (
	github.com/dezhishen/now-and-again/backend v0.0.0
	github.com/dezhishen/now-and-again/sdk v0.0.0-00010101000000-000000000000
)

replace github.com/dezhishen/now-and-again/backend => ../backend

replace github.com/dezhishen/now-and-again/sdk => ../sdk
