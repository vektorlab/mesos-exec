package commands

import (
	"github.com/ghodss/yaml"
	"github.com/jawher/mow.cli"
	"github.com/vektorlab/mesos-cli/commands/options"
	"github.com/vektorlab/mesos-cli/config"
	"github.com/vektorlab/mesos-cli/runner"
	"io/ioutil"
	"os"
)

type Run struct{}

func (_ Run) Name() string { return "run" }
func (_ Run) Desc() string { return "Run tasks on Mesos" }

func (_ Run) Init(profile config.ProfileFn) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[OPTIONS] [FILE]"

		var (
			file     = cmd.StringArg("FILE", "Mesosfile", "File containing Mesos TaskInfos, - for stdin")
			hostname = cmd.StringOpt("m master", "", "Mesos Master")
			restart  = cmd.BoolOpt("restart", false, "Restart containers on failure")
			sync     = cmd.BoolOpt("s sync", false, "Run containers synchronously")
		)

		cmd.Action = func() {
			profile().With(
				config.Master(*hostname),
				config.Restart(*restart),
				config.Sync(*sync),
			)

			mf := &config.Mesosfile{}

			if *file == "-" {
				raw, err := ioutil.ReadAll(os.Stdin)
				failOnErr(err)
				failOnErr(yaml.Unmarshal(raw, &mf))
			} else {
				raw, err := ioutil.ReadFile(*file)
				failOnErr(err)
				failOnErr(yaml.Unmarshal(raw, &mf))
			}

			for _, task := range mf.Tasks() {
				options.Apply(
					task,
					options.WithDefaultResources(),
					options.WithPorts(),
				)
			}

			failOnErr(runner.Run(profile(), mf.Tasks()))
		}
	}
}
