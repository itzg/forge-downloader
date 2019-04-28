package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

var (
	version = "DEV"
	date    = "n/a"
	commit  = "HEAD"
)

func main() {
	app := cli.NewApp()

	app.Name = "forge-downloader"
	app.Usage = "A Forge downloader"
	app.Description = "Downloads a version of Forge that aligns with the Minecraft version given"
	app.ArgsUsage = "MC_VERSION"
	app.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)

	app.Action = run

	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "install", Usage: "run the installer after downloading"},
		cli.BoolFlag{Name: "keep", Usage: "keep the installer file after running it"},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if c.NArg() == 0 {
		cli.ShowAppHelpAndExit(c, 1)
	}

	version := c.Args()[0]
	fmt.Println("Downloading Forge for Minecraft version:", version)

	err, promoVersion := locatePromoVersion(version)
	if err != nil {
		log.Fatal(err)
	}

	err, filename := download(version, promoVersion)
	if err != nil {
		log.Fatal(err)
	}

	if !c.Bool("install") {
		fmt.Printf("Downloaded installer to %s\n", filename)
	} else {

		fmt.Println("Running installer")
		err = runInstaller(filename)

		if !c.Bool("keep") {
			fmt.Println("Removing installer file")
			err := os.Remove(filename)
			if err != nil {
				fmt.Printf("warning: unable to remove downloaded installer: %v\n", err)
			}
		}
	}

	return nil
}

func locatePromoVersion(mcVersion string) (error, string) {
	resp, err := http.Get("http://files.minecraftforge.net/maven/net/minecraftforge/forge/promotions_slim.json")
	if err != nil {
		return errors.Wrap(err, "unable to download promotions info"), ""
	}
	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("unexpected response code %d getting promotions", resp.StatusCode), ""
	}

	decoder := json.NewDecoder(resp.Body)
	var promotions Promotions
	err = decoder.Decode(&promotions)
	if err != nil {
		return errors.Wrap(err, "failed to decode promotions response"), ""
	}

	for _, flavor := range []string{"recommended", "latest"} {
		for ver, promo := range promotions.Promos {
			if strings.HasPrefix(ver, fmt.Sprintf("%s-%s", mcVersion, flavor)) {
				return nil, promo
			}
		}
	}

	return errors.Errorf("unable to find Forge version for minecraft version %s", mcVersion), ""
}

// download will download the Forge installer for the given MC and promo verison and returns
// the path to the downloaded file
func download(mcVersion, promoVersion string) (error, string) {
	combinedVer := fmt.Sprintf("%s-%s", mcVersion, promoVersion)

	url := fmt.Sprintf(
		"http://files.minecraftforge.net/maven/net/minecraftforge/forge/%s/forge-%s-%s.jar",
		combinedVer, combinedVer, "installer")

	fmt.Printf("Getting %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "failed to get forge file"), ""
	}
	//noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("unexpected response code %d while getting forge file", resp.StatusCode), ""
	}

	filename := path.Base(url)
	file, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "unable to open file for writing"), ""
	}
	//noinspection GoUnhandledErrorResult
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read content of forge file response"), ""
	}

	return nil, filename
}

func runInstaller(filename string) error {
	cmd := exec.Command("java", "-jar", filename)
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to run installer")
	}

	return nil
}
