package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/avahowell/masterkey/filelock"
	"github.com/avahowell/masterkey/repl"
	"github.com/avahowell/masterkey/secureclip"
	"github.com/avahowell/masterkey/vault"
	"github.com/howeyc/gopass"
)

const usage = `Usage: masterkey [-new] vault`

func die(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func setupRepl(v *vault.Vault, vaultPath string, timeout time.Duration) *repl.REPL {
	r := repl.New(fmt.Sprintf("masterkey [%v] > ", vaultPath), timeout)

	locations, _ := v.Locations()

	r.AddCommand(importCmd(v), []string{})
	r.AddCommand(listCmd(v), []string{})
	r.AddCommand(saveCmd(v, vaultPath), []string{})
	r.AddCommand(getCmd(v), locations)
	r.AddCommand(addCmd(v), []string{})
	r.AddCommand(genCmd(v), []string{})
	r.AddCommand(editCmd(v), locations)
	r.AddCommand(clipCmd(v), locations)
	r.AddCommand(searchCmd(v), []string{})
	r.AddCommand(addmetaCmd(v), locations)
	r.AddCommand(editmetaCmd(v), locations)
	r.AddCommand(deletemetaCmd(v), locations)
	r.AddCommand(deleteCmd(v), locations)
	r.AddCommand(changePasswordCmd(v), []string{})
	r.AddCommand(mergeCmd(v), []string{})

	r.OnStop(func() {
		fmt.Println("clearing clipboard and saving vault")
		secureclip.Clear()
		v.Save(vaultPath)
	})

	return r
}

func main() {
	createVault := flag.Bool("new", false, "whether to create a new vault at the specified location")
	timeout := flag.Duration("timeout", time.Minute*5, "how long to wait with no vault activity before exiting")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println(usage)
		flag.PrintDefaults()
		os.Exit(1)
	}

	vaultPath := flag.Args()[0]
	var v *vault.Vault

	if !*createVault {
		fmt.Print("Password for " + vaultPath + ": ")
		passphrase, err := gopass.GetPasswd()
		if err != nil {
			die(err)
		}
		fmt.Printf("Opening %v...\n", vaultPath)

		v, err = vault.Open(vaultPath, string(passphrase))
		if err != nil {
			if err == filelock.ErrLocked {
				die(fmt.Errorf("%v is open by another masterkey instance! exit that instance first, or remove %v before opening this vault.", vaultPath, vaultPath+".lck"))
			}
			die(err)
		}
		defer v.Close()
	} else {
		fmt.Print("Enter a passphrase for " + vaultPath + ": ")
		passphrase1, err := gopass.GetPasswd()
		if err != nil {
			die(err)
		}
		fmt.Print("Enter the same passphrase again: ")
		passphrase2, err := gopass.GetPasswd()
		if err != nil {
			die(err)
		}
		if string(passphrase1) != string(passphrase2) {
			die(fmt.Errorf("passphrases do not match"))
		}
		v, err = vault.New(string(passphrase1))
		if err != nil {
			die(err)
		}
		err = v.Save(vaultPath)
		if err != nil {
			die(err)
		}
	}

	r := setupRepl(v, vaultPath, *timeout)
	r.Loop()
}
