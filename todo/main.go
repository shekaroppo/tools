package todo

import (
	"os"

	"github.com/urfave/cli"
)

func GetApp() *cli.App {
	app := cli.NewApp()
	InitConfigCommands(app)
	return app
}

func main() {
	app := GetApp()
	app.Run(os.Args)
}
