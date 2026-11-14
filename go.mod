module github.com/sdttttt/ressu

go 1.27.3

require (
	github.com/egoist/mygo v0.9.0
	github.com/sdttttt/ressu/core v0.0.0
	github.com/sdttttt/ressu/mygoui v0.0.0
)

replace github.com/sdttttt/ressu/mygoui => ./mygoui

require (
	github.com/ProtonMail/go-crypto v1.4.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/cloudflare/circl v1.6.2 // indirect
	github.com/ebitengine/purego v0.11.3 // indirect
	github.com/go-text/typesetting v0.3.5 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/microcosm-cc/bluemonday v1.0.27 // indirect
	github.com/mmcdole/gofeed v1.5.0 // indirect
	github.com/mmcdole/goxpp/v2 v2.0.0 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	go.digitalxero.dev/rpm v0.3.4 // indirect
	go.etcd.io/bbolt v1.5.0 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/image v0.48.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/text v0.44.0 // indirect
)

tool github.com/egoist/mygo/cmd/mygo

replace github.com/sdttttt/ressu/core => ./core
