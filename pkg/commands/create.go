package commands

import (
	"fmt"
	"os"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/juju/errors"
	"github.com/online-net/c14-cli/pkg/api"
)

type create struct {
	Base
	createFlags
}

type createFlags struct {
	flName   string
	flDesc   string
	flSafe   string
	flQuiet  bool
	flParity string
	flLarge  bool
	flCrypto bool
}

// Create returns a new command "create"
func Create() Command {
	ret := &create{}
	ret.Init(Config{
		UsageLine:   "create [OPTIONS]",
		Description: "Create a new archive",
		Help:        "Create a new archive, by default with a random name, standard storage (0.0002€/GB/month), automatic locked in 7 days and your datas will be stored at DC2.",
		Examples: `
        $ c14 create
        $ c14 create --name "MyBooks" --description "hardware books"
        $ c14 create --name "MyBooks" --description "hardware books" --safe "Bookshelf"
`,
	})
	ret.Flags.StringVarP(&ret.flName, "name", "n", "", "Assigns a name")
	ret.Flags.StringVarP(&ret.flDesc, "description", "d", "", "Assigns a description")
	ret.Flags.BoolVarP(&ret.flQuiet, "quiet", "q", false, "Don't display the waiting loop")
	ret.Flags.StringVarP(&ret.flSafe, "save", "s", "", "Name of the safe to use. If it doesn't exists it will be created.")
	ret.Flags.StringVarP(&ret.flParity, "parity", "p", "standard", "Specify a parity to use")
	ret.Flags.BoolVarP(&ret.flLarge, "large", "l", false, "Ask for a large bucket")
	ret.Flags.BoolVarP(&ret.flCrypto, "crypto", "c", true, "Enable aes-256-bc cryptography, enabled by default.")
	return ret
}

func (c *create) GetName() string {
	return "create"
}

func (c *create) CheckFlags(args []string) (err error) {
	if len(args) != 0 {
		c.PrintUsage()
		os.Exit(1)
	}

	if c.flName == "" {
		c.flName = namesgenerator.GetRandomName(0)
	}
	if c.flDesc == "" {
		c.flDesc = " "
	}
	return
}

func (c *create) Run(args []string) (err error) {
	if err = c.InitAPI(); err != nil {
		return
	}
	var (
		uuidArchive string
		safeName    string
		keys        []api.OnlineGetSSHKey
		crypto      string
	)

	if keys, err = c.OnlineAPI.GetSSHKeys(); err != nil {
		err = errors.Annotate(err, "Run:GetSSHKey")
		return
	}
	if len(keys) == 0 {
		err = errors.New("Please add an SSH Key here: https://console.online.net/en/account/ssh-keys")
		return
	}

	safeName = c.flSafe

	if safeName == "" {
		safeName = fmt.Sprintf("%s_safe", c.flName)
	}

	if c.flCrypto == false {
		crypto = "none"
	} else {
		crypto = "aes-256-cbc"
	}

	if _, uuidArchive, _, err = c.OnlineAPI.CreateSSHBucketFromScratch(api.ConfigCreateSSHBucketFromScratch{
		SafeName:    safeName,
		ArchiveName: c.flName,
		Desc:        c.flDesc,
		UUIDSSHKeys: []string{keys[0].UUIDRef},
		Platforms:   []string{"1"},
		Days:        7,
		Quiet:       c.flQuiet,
		Parity:      c.flParity,
		LargeBucket: c.flLarge,
		Crypto:      crypto,
	}); err != nil {
		err = errors.Annotate(err, "Run:CreateSSHBucketFromScratch")
		return
	}
	fmt.Printf("%s\n", uuidArchive)
	return
}
