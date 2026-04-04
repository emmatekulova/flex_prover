package main

import (
	"flag"
	"os"

	"sign-tools/app"
	"sign-tools/base"
	"sign-tools/base/fccutils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/flare-foundation/go-flare-common/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	defaultInstructionSender := os.Getenv("INSTRUCTION_SENDER")

	af := flag.String("a", base.DefaultAddressesFile, "deployed addresses JSON file")
	cf := flag.String("c", base.DefaultChainNodeURL, "chain RPC URL")
	isf := flag.String("instructionSender", defaultInstructionSender, "InstructionSender contract address")
	flag.Parse()

	if *isf == "" {
		logger.Fatal("--instructionSender is required (or set INSTRUCTION_SENDER in .env)")
	}

	instructionSenderAddr := common.HexToAddress(*isf)

	s, err := base.DefaultSupport(*af, *cf)
	if err != nil {
		fccutils.FatalWithCause(err)
	}

	logger.Infof("Setting extension ID on InstructionSender %s...", instructionSenderAddr.Hex())
	if err := app.SetExtensionId(s, instructionSenderAddr); err != nil {
		fccutils.FatalWithCause(err)
	}

	logger.Infof("Extension ID set successfully!")
}
