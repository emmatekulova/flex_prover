package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"sign-tools/app"
	"sign-tools/base"
	"sign-tools/base/fccutils"

	"github.com/flare-foundation/go-flare-common/pkg/logger"
)

func main() {
	af := flag.String("a", base.DefaultAddressesFile, "file with deployed addresses")
	cf := flag.String("c", base.DefaultChainNodeURL, "chain node url")
	outFile := flag.String("o", "", "write deployed address to this file (optional)")
	verify := flag.Bool("verify", true, "verify contract on block explorer after deployment")
	verifyRequired := flag.Bool("verify-required", false, "fail if verification does not succeed")
	verifyAddress := flag.String("verify-address", "", "verify an already deployed InstructionSender address and exit")
	explorerURL := flag.String("explorer-url", "https://coston2-explorer.flare.network/api", "block explorer API URL for verification")
	flag.Parse()

	testSupport, err := base.DefaultSupport(*af, *cf)
	if err != nil {
		fccutils.FatalWithCause(err)
	}

	if *verifyAddress != "" {
		logger.Infof("Verifying existing InstructionSender at: %s", *verifyAddress)
		err := verifyContract(*verifyAddress, testSupport.Addresses, testSupport.ChainID.String(), *explorerURL)
		if err != nil {
			if *verifyRequired {
				fccutils.FatalWithCause(err)
			}
			logger.Warnf("Verification did not succeed: %v", err)
		}
		fmt.Println(*verifyAddress)
		return
	}

	logger.Infof("Deploying InstructionSender contract...")
	address, _, err := app.DeployInstructionSender(testSupport)
	if err != nil {
		fccutils.FatalWithCause(err)
	}

	logger.Infof("InstructionSender deployed at: %s", address.Hex())
	logger.Infof("  Explorer: https://coston2-explorer.flare.network/address/%s", address.Hex())

	// Optionally write address to file for script consumption.
	if *outFile != "" {
		os.MkdirAll(filepath.Dir(*outFile), 0755)
		os.WriteFile(*outFile, []byte(address.Hex()), 0644)
	}

	// Verify contract on block explorer.
	if *verify {
		logger.Infof("Verifying source code on block explorer (chain id %s)...", testSupport.ChainID.String())
		err := verifyContract(address.Hex(), testSupport.Addresses, testSupport.ChainID.String(), *explorerURL)
		if err != nil {
			if *verifyRequired {
				fccutils.FatalWithCause(err)
			}
			logger.Warnf("Verification did not succeed: %v", err)
		}
	}

	// Machine-readable output on stdout.
	fmt.Println(address.Hex())
}

func verifyContract(address string, addresses *base.Addresses, chainID string, explorerURL string) error {
	// Check if forge and cast are available.
	if _, err := exec.LookPath("forge"); err != nil {
		logger.Warnf("forge not found, skipping contract verification (install Foundry to enable)")
		logger.Infof("  You can verify manually at: https://coston2-explorer.flare.network/address/%s", address)
		return err
	}
	if _, err := exec.LookPath("cast"); err != nil {
		logger.Warnf("cast not found, skipping contract verification (install Foundry to enable)")
		logger.Infof("  You can verify manually at: https://coston2-explorer.flare.network/address/%s", address)
		return err
	}

	// Encode constructor args.
	castArgs := exec.Command("cast", "abi-encode",
		"constructor(address,address)",
		addresses.TeeExtensionRegistry.Hex(),
		addresses.TeeMachineRegistry.Hex(),
	)
	constructorArgs, err := castArgs.Output()
	if err != nil {
		logger.Warnf("Failed to encode constructor args: %v", err)
		logger.Infof("  You can verify manually at: https://coston2-explorer.flare.network/address/%s", address)
		return err
	}

	// Find the contract directory (relative to go/tools/).
	contractDir := "../../contract"
	if _, err := os.Stat(contractDir); err != nil {
		logger.Warnf("Contract directory not found at %s, skipping verification", contractDir)
		logger.Infof("  You can verify manually at: https://coston2-explorer.flare.network/address/%s", address)
		return err
	}

	cmd := exec.Command("forge", "verify-contract",
		"--verifier", "etherscan",
		"--verifier-url", explorerURL,
		"--etherscan-api-key", "placeholder",
		"--chain-id", chainID,
		"--compiler-version", "0.8.31",
		"--evm-version", "prague",
		"--constructor-args", string(constructorArgs[:len(constructorArgs)-1]), // trim newline
		"--watch",
		"--root", contractDir,
		address,
		"InstructionSender.sol:InstructionSender",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Warnf("Contract verification failed: %v", err)
		logger.Infof("  Contract is deployed but not yet verified. Check status at:")
		logger.Infof("  https://coston2-explorer.flare.network/address/%s", address)
		return err
	}
	logger.Infof("✓ Contract source code verified on block explorer!")
	logger.Infof("  View at: https://coston2-explorer.flare.network/address/%s", address)
	return nil
}
