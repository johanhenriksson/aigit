package aigit

type Cli struct {
	model Model
}

func NewCli(model Model) *Cli {
	return &Cli{
		model: model,
	}
}

func (cli *Cli) Run() {

}
