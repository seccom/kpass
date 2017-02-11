package pkg

import (
	"time"

	"github.com/seccom/kpass/pkg/auth"
	"github.com/seccom/kpass/pkg/logger"
	"github.com/seccom/kpass/pkg/service"
	"github.com/teambition/gear"
	"github.com/teambition/gear/middleware/favicon"
	"github.com/teambition/gear/middleware/secure"
	"github.com/teambition/gear/middleware/static"
)

// Version is app version
const Version = "0.4.0"

// New returns a app instance
func New(dbPath string, env string) *gear.App {
	if env == "production" {
		logger.Init()
	}

	db, err := service.NewDB(dbPath)
	if err != nil {
		panic(err)
	}
	auth.Init(db.Salt, 10*time.Minute)

	app := gear.New()
	app.Use(secure.Default)
	app.Use(func(ctx *gear.Context) (err error) {
		if ctx.Path == "/" {
			return ctx.HTML(200, string(MustAsset("web/index.html")))
		}
		return nil
	})
	app.Use(favicon.NewWithIco(MustAsset("web/image/favicon.ico")))

	if env == "development" {
		app.Use(static.New(static.Options{
			Root:        "./web",
			Prefix:      "/dev",
			StripPrefix: true,
		}))
	}
	app.UseHandler(logger.Default())
	app.UseHandler(newRouter(db))

	return app
}
