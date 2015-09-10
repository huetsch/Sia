package main

import (
	"fmt"
	"math/big"

	"github.com/bgentry/speakeasy"
	"github.com/spf13/cobra"

	"github.com/NebulousLabs/Sia/api"
	"github.com/NebulousLabs/Sia/types"
)

var (
	walletCmd = &cobra.Command{
		Use:   "wallet",
		Short: "Perform wallet actions",
		Long: `Generate a new address, send coins to another wallet, or view info about the wallet.

Units:
The smallest unit of siacoins is the hasting. One siacoin is 10^24 hastings. Other supported units are:
  pS (pico,  10^-12 SC)
  nS (nano,  10^-9 SC)
  uS (micro, 10^-6 SC)
  mS (milli, 10^-3 SC)
  SC
  KS (kilo, 10^3 SC)
  MS (mega, 10^6 SC)
  GS (giga, 10^9 SC)
  TS (tera, 10^12 SC)`,
		Run: wrap(walletstatuscmd),
	}

	walletAddressCmd = &cobra.Command{
		Use:   "address",
		Short: "Get a new wallet address",
		Long:  "Generate a new wallet address.",
		Run:   wrap(walletaddresscmd),
	}

	walletAddressesCmd = &cobra.Command{
		Use:   "addresses",
		Short: "List all addresses",
		Long:  "List all addresses that have been generated by the wallet",
		Run:   wrap(walletaddressescmd),
	}

	walletInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize and encrypt a new wallet",
		Long: `Generate a new wallet from a seed string, and encrypt it.
The seed string, which is also the encryption password, will be returned.`,
		Run: wrap(walletinitcmd),
	}

	walletLoadCmd = &cobra.Command{
		Use:   "load",
		Short: "Load a wallet seed, v0.3.3.x wallet, or siag keyset",
		Long:  "Load a wallet seed, v0.3.3.x wallet, or siag keyset",
		Run:   walletloadcmd,
	}

	walletLoad033xCmd = &cobra.Command{
		Use:   "033x [filepath]",
		Short: "Load a v0.3.3.x wallet",
		Long:  "Load a v0.3.3.x wallet into the current wallet",
		Run:   wrap(walletload033xcmd),
	}

	walletLoadSeedCmd = &cobra.Command{
		Use:   `seed`,
		Short: "Add a seed to the wallet",
		Long:  "Uses the given password to create a new wallet with that as the primary seed",
		Run:   wrap(walletloadseedcmd),
	}

	walletLoadSiagCmd = &cobra.Command{
		Use:   `siag [filepaths]`,
		Short: "Load a siag keyset into the wallet",
		Long: `Load a set of siag keys into the wallet - typically used for siafunds.
Example: 'siac wallet load siag key1.siakey,key2.siakey'`,
		Run: wrap(walletloadsiagcmd),
	}

	walletLockCmd = &cobra.Command{
		Use:   "lock",
		Short: "Lock the wallet",
		Long:  "Lock the wallet, preventing further use",
		Run:   wrap(walletlockcmd),
	}

	walletSeedsCmd = &cobra.Command{
		Use:   "seeds",
		Short: "Retrieve information about your seeds",
		Long:  "Retrieves the current seed, how many addresses are remaining, and the rest of your seeds from the wallet",
		Run:   wrap(walletseedscmd),
	}

	walletSendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send either siacoins or siafunds to an address",
		Long:  "Send either siacoins or siafunds to an address",
		Run:   walletsendcmd,
	}

	walletSendSiacoinsCmd = &cobra.Command{
		Use:   "siacoins [amount] [dest]",
		Short: "Send siacoins to an address",
		Long: `Send siacoins to an address. 'dest' must be a 76-byte hexadecimal address.
'amount' can be specified in units, e.g. 1.23KS. Run 'wallet --help' for a list of units.
If no unit is supplied, hastings will be assumed.

A miner fee of 10 SC is levied on all transactions.`,
		Run: wrap(walletsendsiacoinscmd),
	}

	walletSendSiafundsCmd = &cobra.Command{
		Use:   "siafunds [amount] [dest] [keyfiles]",
		Short: "Send siafunds",
		Long: `Send siafunds to an address, and transfer the claim siacoins to your wallet.
Run 'wallet send --help' to see a list of available units.`,
		Run: wrap(walletsendsiafundscmd),
	}

	walletStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "View wallet status",
		Long:  "View wallet status, including the current balance and number of addresses.",
		Run:   wrap(walletstatuscmd),
	}

	walletTransactionsCmd = &cobra.Command{
		Use:   "transactions",
		Short: "View transactions",
		Long:  "View transactions related to addresses spendable by the wallet, providing a net flow of siacoins and siafunds for each transaction",
		Run:   wrap(wallettransactionscmd),
	}

	walletUnlockCmd = &cobra.Command{
		Use:   `unlock`,
		Short: "Unlock the wallet",
		Long:  "Decrypt and load the wallet into memory",
		Run:   wrap(walletunlockcmd),
	}
)

// walletaddresscmd fetches a new address from the wallet that will be able to
// recieve coins.
func walletaddresscmd() {
	addr := new(api.WalletAddressGET)
	err := getAPI("/wallet/address", addr)
	if err != nil {
		fmt.Println("Could not generate new address:", err)
		return
	}
	fmt.Printf("Created new address: %s\n", addr.Address)
}

// walletaddressescmd fetches the list of addresses that the wallet knows.
func walletaddressescmd() {
	addrs := new(api.WalletAddressesGET)
	err := getAPI("/wallet/addresses", addrs)
	if err != nil {
		fmt.Println("Failed to fetch addresses:", err)
		return
	}
	for _, addr := range addrs.Addresses {
		fmt.Println(addr.Address)
	}
}

// walletinitcmd encrypts the wallet with the given password
func walletinitcmd() {
	var er api.WalletEncryptPOST
	qs := fmt.Sprintf("dictionary=%s", "english")
	if initPassword {
		password, err := speakeasy.Ask("Wallet password: ")
		if err != nil {
			fmt.Println("Reading password failed")
			return
		}
		qs += fmt.Sprintf("&encryptionpassword=%s", password)
	}
	err := postResp("/wallet/encrypt", qs, &er)
	if err != nil {
		fmt.Println("Error when encrypting wallet:", err)
		return
	}
	fmt.Printf("Seed is:\n %s\n\n", er.PrimarySeed)
	if initPassword {
		fmt.Printf("Wallet encrypted with given password\n")
	} else {
		fmt.Printf("Wallet encrypted with password: %s\n", er.PrimarySeed)
	}
}

// walletloadcmd is a no-op; it only has subcommands.
func walletloadcmd(cmd *cobra.Command, args []string) { cmd.Usage() }

// walletload033xcmd loads a v0.3.3.x wallet into the current wallet.
func walletload033xcmd(filepath string) {
	password, err := speakeasy.Ask("Wallet password: ")
	if err != nil {
		fmt.Println("Reading password failed")
		return
	}
	qs := fmt.Sprintf("filepath=%s&encryptionpassword=%s", filepath, password)
	err = post("/wallet/load/033x", qs)
	if err != nil {
		fmt.Println("loading error:", err)
		return
	}
	fmt.Println("Wallet loading successful.")
}

// walletloadseedcmd adds a seed to the wallet's list of seeds
func walletloadseedcmd() {
	password, err := speakeasy.Ask("Wallet password: ")
	if err != nil {
		fmt.Println("Reading password failed")
		return
	}
	seed, err := speakeasy.Ask("New Seed: ")
	if err != nil {
		fmt.Println("Reading seed failed")
		return
	}
	qs := fmt.Sprintf("encryptionpassword=%s&seed=%s&dictionary=%s", password, seed, "english")
	err = post("/wallet/load/seed", qs)
	if err != nil {
		fmt.Println("Could not add seed:", err)
		return
	}
	fmt.Println("Added Key")
}

// walletloadsiagcmd loads a siag key set into the wallet.
func walletloadsiagcmd(keyfiles string) {
	password, err := speakeasy.Ask("Wallet password: ")
	if err != nil {
		fmt.Println("Reading password failed")
		return
	}
	qs := fmt.Sprintf("keyfiles=%s&encryptionpassword=%s", keyfiles, password)
	err = post("/wallet/load/siag", qs)
	if err != nil {
		fmt.Println("loading error:", err)
		return
	}
	fmt.Println("Wallet loading successful.")
}

// walletlockcmd locks the wallet
func walletlockcmd() {
	err := post("/wallet/lock", "")
	if err != nil {
		fmt.Println("Could not lock wallet:", err)
	}
}

// walletseedcmd returns the current seed {
func walletseedscmd() {
	var seedInfo api.WalletSeedsGET
	err := getAPI("/wallet/seeds", &seedInfo)
	if err != nil {
		fmt.Println("Error retrieving the current seed:", err)
		return
	}
	fmt.Printf("Primary Seed: %s\n"+
		"Addresses Remaining %d\n"+
		"All Seeds:\n", seedInfo.PrimarySeed, seedInfo.AddressesRemaining)
	for _, seed := range seedInfo.AllSeeds {
		fmt.Println(seed)
	}
}

// walletsendcmd is a noop, it has only subcommands.
func walletsendcmd(cmd *cobra.Command, args []string) { cmd.Usage() }

// walletsendsiacoinscmd sends siacoins to a destination address.
func walletsendsiacoinscmd(amount, dest string) {
	adjAmount, err := coinUnits(amount)
	if err != nil {
		fmt.Println("Could not parse amount:", err)
		return
	}
	err = post("/wallet/siacoins", fmt.Sprintf("amount=%s&destination=%s", adjAmount, dest))
	if err != nil {
		fmt.Println("Could not send:", err)
		return
	}
	fmt.Printf("Sent %s hastings to %s\n", adjAmount, dest)
}

// walletsendsiafundscmd sends siafunds to a destination address.
func walletsendsiafundscmd(amount, dest string) {
	err := post("/wallet/siafunds", fmt.Sprintf("amount=%s&destination=%s", amount, dest))
	if err != nil {
		fmt.Println("Could not send:", err)
		return
	}
	fmt.Printf("Sent %s siafunds to %s\n", amount, dest)
}

// walletstatuscmd retrieves and displays information about the wallet
func walletstatuscmd() {
	status := new(api.WalletGET)
	err := getAPI("/wallet", status)
	if err != nil {
		fmt.Println("Could not get wallet status:", err)
		return
	}
	encStatus := "Unencrypted"
	if status.Encrypted {
		encStatus = "Encrypted"
	}
	lockStatus := "Locked"
	if status.Unlocked {
		lockStatus = "Unlocked"
	}
	// divide by 1e24 to get SC
	r := new(big.Rat).SetFrac(status.ConfirmedSiacoinBalance.Big(), new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil))
	sc, _ := r.Float64()
	unconfirmedBalance := status.ConfirmedSiacoinBalance.Add(status.UnconfirmedIncomingSiacoins).Sub(status.UnconfirmedOutgoingSiacoins)
	unconfirmedDifference := new(big.Int).Sub(unconfirmedBalance.Big(), status.ConfirmedSiacoinBalance.Big())
	r = new(big.Rat).SetFrac(unconfirmedDifference, new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil))
	usc, _ := r.Float64()
	fmt.Printf(`Wallet status:
%s, %s
Confirmed Balance:   %.2f SC
Unconfirmed Delta:  %+.2f SC
Exact:               %v H
Siafunds:            %v SF
Siafund Claims:      %v SC
`, encStatus, lockStatus, sc, usc, status.ConfirmedSiacoinBalance, status.SiafundBalance, status.SiacoinClaimBalance)
}

// wallettransactionscmd lists all of the transactions related to the wallet,
// providing a net flow of siacoins and siafunds for each.
func wallettransactionscmd() {
	wtg := new(api.WalletTransactionsGET)
	err := getAPI("/wallet/transactions?startheight=0&endheight=10000000", wtg)
	if err != nil {
		fmt.Println("Could not fetch transaction history:", err)
		return
	}

	fmt.Println("    [height]                                                   [transaction id]    [net siacoins]   [net siafunds]")
	txns := append(wtg.ConfirmedTransactions, wtg.UnconfirmedTransactions...)
	for _, txn := range txns {
		// Determine the number of outgoing siacoins and siafunds.
		var outgoingSiacoins types.Currency
		var outgoingSiafunds types.Currency
		for _, input := range txn.Inputs {
			if input.FundType == types.SpecifierSiacoinInput && input.WalletAddress {
				outgoingSiacoins = outgoingSiacoins.Add(input.Value)
			}
			if input.FundType == types.SpecifierSiafundInput && input.WalletAddress {
				outgoingSiafunds = outgoingSiafunds.Add(input.Value)
			}
		}

		// Determine the number of incoming siacoins and siafunds.
		var incomingSiacoins types.Currency
		var incomingSiafunds types.Currency
		for _, output := range txn.Outputs {
			if output.FundType == types.SpecifierMinerPayout {
				incomingSiacoins = incomingSiacoins.Add(output.Value)
			}
			if output.FundType == types.SpecifierSiacoinOutput && output.WalletAddress {
				incomingSiacoins = incomingSiacoins.Add(output.Value)
			}
			if output.FundType == types.SpecifierSiafundOutput && output.WalletAddress {
				incomingSiafunds = incomingSiafunds.Add(output.Value)
			}
		}

		// Convert the siacoins to a float.
		incomingSiacoinsFloat, _ := new(big.Rat).SetFrac(incomingSiacoins.Big(), types.SiacoinPrecision.Big()).Float64()
		outgoingSiacoinsFloat, _ := new(big.Rat).SetFrac(outgoingSiacoins.Big(), types.SiacoinPrecision.Big()).Float64()

		// Print the results.
		if txn.ConfirmationHeight < 1e9 {
			fmt.Printf("%12v", txn.ConfirmationHeight)
		} else {
			fmt.Printf(" unconfirmed")
		}
		fmt.Printf("%67v%15.2f SC", txn.TransactionID, incomingSiacoinsFloat-outgoingSiacoinsFloat)
		// For siafunds, need to avoid having a negative types.Currency.
		if incomingSiafunds.Cmp(outgoingSiafunds) >= 0 {
			fmt.Printf("%14v SF\n", incomingSiafunds.Sub(outgoingSiafunds))
		} else {
			fmt.Printf("-%14v SF\n", outgoingSiafunds.Sub(incomingSiafunds))
		}
	}
}

// walletunlockcmd unlocks a saved wallet
func walletunlockcmd() {
	password, err := speakeasy.Ask("Wallet password: ")
	if err != nil {
		fmt.Println("Reading password failed")
		return
	}
	qs := fmt.Sprintf("encryptionpassword=%s&dictonary=%s", password, "english")
	err = post("/wallet/unlock", qs)
	if err != nil {
		fmt.Println("Could not unlock wallet:", err)
		return
	}
	fmt.Println("Wallet unlocked")
}
