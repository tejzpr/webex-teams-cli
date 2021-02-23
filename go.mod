module webex-teams-cli

go 1.14

require (
	github.com/gabriel-vasile/mimetype v1.0.5
	github.com/gammazero/workerpool v0.0.0-20200311205957-7b00833861c6
	github.com/go-chi/chi v4.1.1+incompatible
	github.com/go-ozzo/ozzo-validation/v4 v4.1.0
	github.com/go-resty/resty v0.0.0-00010101000000-000000000000 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/h2non/filetype v1.0.12 // indirect
	github.com/jbogarin/go-cisco-webex-teams v0.3.0
	github.com/jroimartin/gocui v0.4.0
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/nsf/termbox-go v0.0.0-20200418040025-38ba6e5628f1 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/peterhellberg/link v1.1.0 // indirect
	github.com/sirupsen/logrus v1.5.0
	github.com/urfave/cli/v2 v2.2.0
)

replace github.com/go-resty/resty => gopkg.in/resty.v1 v1.12.0
