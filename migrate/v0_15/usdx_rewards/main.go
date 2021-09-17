// package main calculates missing rewards from kava-7 so they can be include in the migration to kava-8.
// In the migration folder run `go run ./usdx_rewards` to compute the output file. The output is a go file so it can be included in the migrations.
// Data is taken from an export at height 829296 of kava-7 using v0.14.3.
package main

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	v0_15cdp "github.com/kava-labs/kava/x/cdp/types"
	v0_15incentive "github.com/kava-labs/kava/x/incentive"
)

var (
	outFileName = "rewards.go"

	dirPath         = "./usdx_rewards"
	cdpsFileName    = filepath.Join(dirPath, "cdp-cdps-829296.json")
	indexesFileName = filepath.Join(dirPath, "incentive-usdx-indexes-829296.json")
	claimsFileName  = filepath.Join(dirPath, "incentive-usdx-claims-829296.json")
)

func main() {
	app.SetBech32AddressPrefixes(sdk.GetConfig())
	cdc := app.MakeCodec()

	var cdps v0_15cdp.CDPs
	if err := fetchFromJSONFile(cdc, cdpsFileName, &cdps); err != nil {
		log.Fatal(err.Error())
	}
	var globalIndexes v0_15incentive.MultiRewardIndexes
	if err := fetchFromJSONFile(cdc, indexesFileName, &globalIndexes); err != nil {
		log.Fatal(err.Error())
	}
	var claims v0_15incentive.USDXMintingClaims
	if err := fetchFromJSONFile(cdc, claimsFileName, &claims); err != nil {
		log.Fatal(err.Error())
	}

	calculator := NewRewardsCalculator(claims, convertRewardIndexesToUSDXMintingIndexes(globalIndexes), cdps)
	rewards, err := calculator.Calculate()
	if err != nil {
		log.Fatal(err.Error())
	}

	out := createOutputString(cdc, rewards)

	if err = ioutil.WriteFile(outFileName, []byte(out), 0644); err != nil {
		log.Fatal(err.Error())
	}

}

// fetchFromJSONFile reads and unmarshals the contents of a json file.
func fetchFromJSONFile(cdc *codec.Codec, filePath string, pointer interface{}) error {

	bz, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err = cdc.UnmarshalJSON(bz, pointer); err != nil {
		return err
	}

	return nil
}

func createOutputString(cdc *codec.Codec, rewards rewards) string {
	bz := cdc.MustMarshalJSON(rewards)

	// sort json keys so make changes easier to compare
	bz = sdk.MustSortJSON(bz)

	out := `// Code generated by package ./usdx_rewards DO NOT EDIT.

package v0_15

var missedUSDXMintingRewards = ` + "`" + string(bz) + "`\n"

	return out
}