package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	b64 "encoding/base64"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/shopspring/decimal"
	"github.com/ybbus/jsonrpc"
)

type OrcaPair struct {
	Version                     uint8
	Is_inialized                uint8
	BumpSeed                    uint8
	TokenProgramId              solana.PublicKey
	TokenAccountA               solana.PublicKey
	TokenAccountB               solana.PublicKey
	TokenPool                   solana.PublicKey
	MintA                       solana.PublicKey
	MintB                       solana.PublicKey
	FeeAccount                  solana.PublicKey
	TradeFeeNumerator           uint64
	TradeFeeDenominator         uint64
	OwnerTradeFeeNumerator      uint64
	OwnerTradeFeeDenominator    uint64
	OwnerWithdrawFeeNumerator   uint64
	OwnerWithdrawFeeDenominator uint64
	HostFeeNumerator            uint64
	HostFeeDenominator          uint64
	CurveType                   uint8
	CurveParameters             [32]byte
	SwapAccount                 solana.PublicKey
	Authority                   solana.PublicKey
	Cointoken                   POOL_LAYOUT
	Pctoken                     POOL_LAYOUT
	Name                        string
}
type GetAuthority struct {
	Donotcare uint32
	Authority solana.PublicKey
}
type ORCA_API struct {
	Name           string  `json:"name"`
	Name2          string  `json:"name2"`
	Account        string  `json:"account"`
	MintAccount    string  `json:"mint_account"`
	Liquidity      float64 `json:"liquidity"`
	Price          float64 `json:"price"`
	Apy24H         float64 `json:"apy_24h"`
	Apy7D          float64 `json:"apy_7d"`
	Apy30D         float64 `json:"apy_30d"`
	Volume24H      float64 `json:"volume_24h"`
	Volume24HQuote float64 `json:"volume_24h_quote"`
	Volume7D       float64 `json:"volume_7d"`
	Volume7DQuote  float64 `json:"volume_7d_quote"`
	Volume30D      float64 `json:"volume_30d"`
	Volume30DQuote float64 `json:"volume_30d_quote"`
}
type POOL_LAYOUT struct {
	Useless [64]byte
	Ammount uint64
}
type ENCODING struct {
	Encoding    string `json:"encoding"`
	Commitement string `json:"commitment"`
}
type SLOT struct {
	Slot uint64 `json:"slot"`
}
type RPCVALUE struct {
	Data       [2]string `json:"data"`
	Executable bool      `json:"executable"`
	Lamports   uint64    `json:"lamports"`
	Owner      string    `json:"owner"`
	Rentepoch  uint64    `json:"rentEpoch"`
}
type RPCRESPONSE struct {
	Context SLOT     `json:"context"`
	Value   RPCVALUE `json:"value"`
}
type Path struct {
	Pair1    *OrcaPair
	Pair2    *OrcaPair
	Pair3    *OrcaPair
	Pair4    *OrcaPair
	Reverse1 bool
	Reverse2 bool
	Reverse3 bool
	Reverse4 bool
	Name     string
}
type Instruction_layout struct {
	Instruction  uint8
	AmountIn     uint64
	MinAmountOut uint64
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func pairInSlice(a *OrcaPair, list []*OrcaPair) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func get_init_accounts() map[string]string {
	raw_accounts_list := `{}`
	accounts_list := make(map[string]string)
	json.Unmarshal([]byte(raw_accounts_list), &accounts_list)
	return accounts_list
}
func init_orca_pairs(pair_list map[string]*OrcaPair, currencies_list []string, reference string) {
	//Get pairs info from the API
	rpcClient := rpc.New(rpc.MainNetBetaSerum_RPC)
	var info []ORCA_API
	resp, err := http.Get("https://api.orca.so/pools")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &info)
	for _, i := range info {
		if !strings.Contains(i.Name2, "[deprecated]") {
			tokens := strings.Split(i.Name2, "/")
			//Use only the pairs specified in the currencies_list
			if !stringInSlice(tokens[0], currencies_list) || !stringInSlice(tokens[1], currencies_list) {
				continue
			}
			fmt.Println(i.Name2)
			//Get generals info from the swap account
			pubkey := solana.MustPublicKeyFromBase58(i.Account)
			res, err := rpcClient.GetAccountInfo(context.TODO(), pubkey)
			if err != nil {
				panic(err)
			}
			var newpair OrcaPair
			swapdecoder := bin.NewBinDecoder(res.Value.Data.GetBinary())
			swapdecoder.Decode(&newpair)
			//Get Authority info from the MintAccount
			pubkey = solana.MustPublicKeyFromBase58(i.MintAccount)
			res, err = rpcClient.GetAccountInfo(context.TODO(), pubkey)
			if err != nil {
				panic(err)
			}
			var authority_info GetAuthority
			authoritydecoder := bin.NewBinDecoder(res.Value.Data.GetBinary())
			authoritydecoder.Decode(&authority_info)
			account := solana.MustPublicKeyFromBase58(i.Account)
			newpair.Authority = authority_info.Authority
			newpair.SwapAccount = account
			newpair.Name = i.Name2
			pair_list[i.Name2] = &newpair
		}
	}
}
func update_orca_pools(pair_list []*OrcaPair, rpcClient jsonrpc.RPCClient) {
	//client
	var input []*jsonrpc.RPCRequest
	//liste de requete
	for _, element := range pair_list {
		input = append(input, jsonrpc.NewRequest("getAccountInfo", element.TokenAccountA, ENCODING{"base64", "processed"}))
		input = append(input, jsonrpc.NewRequest("getAccountInfo", element.TokenAccountB, ENCODING{"base64", "processed"}))
	}
	//reponse
	response, errbatch := rpcClient.CallBatch(input)

	if errbatch != nil {
		fmt.Println(errbatch)
		return
	}
	responsesordered := response.AsMap()
	compteur := 0
	for _, element := range pair_list {
		var coinResponse *RPCRESPONSE
		var poolResponse *RPCRESPONSE
		err := responsesordered[compteur].GetObject(&coinResponse)
		if err != nil {
			fmt.Println(err)
			return
		}
		compteur++
		err = responsesordered[compteur].GetObject(&poolResponse)
		if err != nil {
			fmt.Println(err)
			return
		}
		compteur++
		if coinResponse == nil || poolResponse == nil {
			fmt.Println(err)
			return
		}
		b64coin, err := b64.StdEncoding.DecodeString(coinResponse.Value.Data[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		b64pool, err := b64.StdEncoding.DecodeString(poolResponse.Value.Data[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		coindecoder := bin.NewBinDecoder(b64coin)
		pooldecoder := bin.NewBinDecoder(b64pool)
		err = coindecoder.Decode(&element.Cointoken)
		if err != nil {
			panic(err)
		}
		err = pooldecoder.Decode(&element.Pctoken)
		if err != nil {
			panic(err)
		}
	}
}
func (pair OrcaPair) get_pools() [2]uint64 {
	coinpool := pair.Cointoken.Ammount
	pcpool := pair.Pctoken.Ammount
	return [2]uint64{coinpool, pcpool}
}
func get_amount_out(amount_in uint64, pools [2]uint64, reverse bool, string_fee string) uint64 {
	amoutused := decimal.NewFromInt(int64(amount_in))
	x := decimal.NewFromInt(int64((pools[0]))) //coin
	y := decimal.NewFromInt(int64((pools[1]))) //pc
	fees := decimal.RequireFromString(string_fee)
	res := (amoutused.Mul(x.Mul(fees))).Div(y.Add(fees.Mul(amoutused)))
	if reverse {
		res = (amoutused.Mul(y.Mul(fees))).Div(x.Add(fees.Mul(amoutused)))
	}
	return uint64((res.BigInt().Int64()))
}
func compute_arbitrage(possible_path []Path, reference_ammount uint64, reference string, privateKey solana.PrivateKey, rpcClient *rpc.Client, accounts_list map[string]string, lastTrade map[Path]time.Time) {
	for _, path := range possible_path {
		if time.Since(lastTrade[path]) < 5*time.Second {
			continue
		}
		if path.Pair1.Name == "mSOL/SOL" || path.Pair1.Name == "USDC/USDT" || path.Pair1.Name == "scnSOL/SOL" || path.Pair2.Name == "mSOL/SOL" || path.Pair2.Name == "USDC/USDT" || path.Pair2.Name == "scnSOL/SOL" || path.Pair3.Name == "mSOL/SOL" || path.Pair3.Name == "USDC/USDT" || path.Pair3.Name == "scnSOL/SOL" || path.Pair4.Name == "mSOL/SOL" || path.Pair4.Name == "USDC/USDT" || path.Pair4.Name == "scnSOL/SOL" {
			continue
		}
		//Output : (0.997*(0.997*(0.997*(0.997*a*y1)/(0.997*a + x1)*y2)/(0.997*(0.997*a*y1)/(0.997*a + x1) + x2)*y3)/(0.997*(0.997*(0.997*a*y1)/(0.997*a + x1)*y2)/(0.997*(0.997*a*y1)/(0.997*a + x1) + x2) + x3)*y4)/(0.997*(0.997*(0.997*(0.997*a*y1)/(0.997*a + x1)*y2)/(0.997*(0.997*a*y1)/(0.997*a + x1) + x2)*y3)/(0.997*(0.997*(0.997*a*y1)/(0.997*a + x1)*y2)/(0.997*(0.997*a*y1)/(0.997*a + x1) + x2) + x3) + x4)
		//Output : (b*d*f*h*x)/(1.01209*a*c*e*g+x*(b*(d*f+1.00301*d*g+1.00603*e*g)+1.00905*c*e*g))
		//Profit : (b * d * f * h * x) / (1.01209*a*c*e*g + x*(b*(d*f+1.00301*d*g+1.00603*e*g)+1.00905*c*e*g)) - x
		//dProfit = 0: (math.Sqrt(1.0121*a*b*c*d*e*f*g*h) - 1.0121*a*c*e*g)/(b*d*f + 1.003*b*d*g+1.006*b*e*g+1.009*c*e*g)
		pair_pool1 := path.Pair1.get_pools()
		pair_pool2 := path.Pair2.get_pools()
		pair_pool3 := path.Pair3.get_pools()
		pair_pool4 := path.Pair4.get_pools()
		x1 := float64(pair_pool1[1])
		y1 := float64(pair_pool1[0])
		x2 := float64(pair_pool2[1])
		y2 := float64(pair_pool2[0])
		x3 := float64(pair_pool3[1])
		y3 := float64(pair_pool3[0])
		x4 := float64(pair_pool4[1])
		y4 := float64(pair_pool4[0])
		if path.Reverse1 {
			tmp1 := x1
			x1 = y1
			y1 = tmp1
		}
		if path.Reverse2 {
			tmp2 := x2
			x2 = y2
			y2 = tmp2
		}
		if path.Reverse3 {
			tmp3 := x3
			x3 = y3
			y3 = tmp3
		}
		if path.Reverse4 {
			tmp4 := x4
			x4 = y4
			y4 = tmp4
		}

		a := x1
		b := y1
		c := x2
		d := y2
		e := x3
		f := y3
		g := x4
		h := y4

		//x := float64(100000000)
		//output := (b * d * f * h * x) / (1.01209*a*c*e*g + x*(b*(d*f+1.00301*d*g+1.00603*e*g)+1.00905*c*e*g))
		optimised_value := math.Abs(math.Sqrt(1.0121*a*b*c*d*e*f*g*h)-1.0121*a*c*e*g) / (b*d*f + 1.003*b*d*g + 1.006*b*e*g + 1.009*c*e*g)
		used_value := uint64(math.Min(float64(reference_ammount), optimised_value))
		res1 := get_amount_out(used_value, path.Pair1.get_pools(), path.Reverse1, "0.997")
		res2 := get_amount_out(uint64(res1), path.Pair2.get_pools(), path.Reverse2, "0.997")
		res3 := get_amount_out(uint64(res2), path.Pair3.get_pools(), path.Reverse3, "0.997")
		res4 := get_amount_out(uint64(res3), path.Pair4.get_pools(), path.Reverse4, "0.997")
		//fmt.Printf("%f | %v | %f |%v | %v | %v | %v | %v\n", output, res4, optimised_value, path.Name, path.Reverse1, path.Reverse2, path.Reverse3, path.Reverse4)
		res1 -= uint64(0.00001 * float64(res1))
		res2 -= uint64(0.00001 * float64(res2))
		res3 -= uint64(0.00001 * float64(res3))
		if int64(res4)-int64(used_value) > 2000 {
			nb_loop := 1
			if int64(res4)-int64(used_value) > 100000 {
				nb_loop = 5
			}
			for i := 1; i <= nb_loop; i++ {
				//fmt.Printf("%f | %v | %f |%v | %v | %v | %v | %v\n", output, res4, optimised_value, path.Name, path.Reverse1, path.Reverse2, path.Reverse3, path.Reverse4)
				fmt.Println(path.Name, int64(res4)-int64(used_value))
				sendQuadTransac(privateKey, rpcClient, accounts_list, path, used_value, res1, res2, res3)
				fmt.Println("-------------------------------------------------------------------")
			}
			lastTrade[path] = time.Now()
		}
	}
}
func sendQuadTransac(privateKey solana.PrivateKey, rpcClient *rpc.Client, accounts_list map[string]string, path Path, amountin uint64, amountout1 uint64, amountout2 uint64, amountout3 uint64) error {

	tokens := strings.Split(path.Name, "/")
	first := tokens[0]
	second := tokens[1]
	third := tokens[2]
	fourth := tokens[3]

	TOKEN_PROGRAM, _ := solana.PublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	first_account, _ := solana.PublicKeyFromBase58(accounts_list[first])
	second_account, _ := solana.PublicKeyFromBase58(accounts_list[second])
	third_account, _ := solana.PublicKeyFromBase58(accounts_list[third])
	fourth_account, _ := solana.PublicKeyFromBase58(accounts_list[fourth])
	Orca_swap, _ := solana.PublicKeyFromBase58("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP")

	//pair1
	pair1 := path.Pair1

	var pool_src1 solana.PublicKey = pair1.TokenAccountB
	var pool_dest1 solana.PublicKey = pair1.TokenAccountA
	if path.Reverse1 {
		pool_src1 = pair1.TokenAccountA
		pool_dest1 = pair1.TokenAccountB
	}

	meta_tokenSwapAccount1 := solana.NewAccountMeta(pair1.SwapAccount, false, false)
	meta_authority1 := solana.NewAccountMeta(pair1.Authority, false, false)
	meta_wallet1 := solana.NewAccountMeta(privateKey.PublicKey(), false, true)
	meta_userSource1 := solana.NewAccountMeta(first_account, true, false)
	meta_poolSource1 := solana.NewAccountMeta(pool_src1, true, false)
	meta_poolDestination1 := solana.NewAccountMeta(pool_dest1, true, false)
	meta_userDestination1 := solana.NewAccountMeta(second_account, true, false)
	meta_poolMint1 := solana.NewAccountMeta(pair1.TokenPool, true, false)
	meta_feeAccount1 := solana.NewAccountMeta(pair1.FeeAccount, true, false)
	meta_tokenProgramId1 := solana.NewAccountMeta(TOKEN_PROGRAM, false, false)

	var meta_list1 []*solana.AccountMeta
	meta_list1 = append(meta_list1, meta_tokenSwapAccount1, meta_authority1, meta_wallet1, meta_userSource1, meta_poolSource1, meta_poolDestination1, meta_userDestination1, meta_poolMint1, meta_feeAccount1, meta_tokenProgramId1)
	//Création des instructions1
	buf1 := new(bytes.Buffer)
	borshEncoder1 := bin.NewBorshEncoder(buf1)
	err := borshEncoder1.Encode(Instruction_layout{1, amountin, amountout1})
	if err != nil {
		return err
	}
	bytes_data1 := buf1.Bytes()
	instruction1 := solana.NewInstruction(Orca_swap, meta_list1, bytes_data1)

	//Création des comptes associés de la paire2
	pair2 := path.Pair2

	var pool_src2 solana.PublicKey = pair2.TokenAccountB
	var pool_dest2 solana.PublicKey = pair2.TokenAccountA
	if path.Reverse2 {
		pool_src2 = pair2.TokenAccountA
		pool_dest2 = pair2.TokenAccountB
	}

	meta_tokenSwapAccount2 := solana.NewAccountMeta(pair2.SwapAccount, false, false)
	meta_authority2 := solana.NewAccountMeta(pair2.Authority, false, false)
	meta_wallet2 := solana.NewAccountMeta(privateKey.PublicKey(), false, true)
	meta_userSource2 := solana.NewAccountMeta(second_account, true, false)
	meta_poolSource2 := solana.NewAccountMeta(pool_src2, true, false)
	meta_poolDestination2 := solana.NewAccountMeta(pool_dest2, true, false)
	meta_userDestination2 := solana.NewAccountMeta(third_account, true, false)
	meta_poolMint2 := solana.NewAccountMeta(pair2.TokenPool, true, false)
	meta_feeAccount2 := solana.NewAccountMeta(pair2.FeeAccount, true, false)
	meta_tokenProgramId2 := solana.NewAccountMeta(TOKEN_PROGRAM, false, false)

	var meta_list2 []*solana.AccountMeta
	meta_list2 = append(meta_list2, meta_tokenSwapAccount2, meta_authority2, meta_wallet2, meta_userSource2, meta_poolSource2, meta_poolDestination2, meta_userDestination2, meta_poolMint2, meta_feeAccount2, meta_tokenProgramId2)
	//Création des instructions2
	buf2 := new(bytes.Buffer)
	borshEncoder2 := bin.NewBorshEncoder(buf2)
	err = borshEncoder2.Encode(Instruction_layout{1, amountout1, amountout2})
	if err != nil {
		return err
	}
	bytes_data2 := buf2.Bytes()
	instruction2 := solana.NewInstruction(Orca_swap, meta_list2, bytes_data2)

	//Création des comptes associés de la paire 3
	pair3 := path.Pair3
	fmt.Println(pair3.Name, path.Reverse3)

	var pool_src3 solana.PublicKey = pair3.TokenAccountB
	var pool_dest3 solana.PublicKey = pair3.TokenAccountA
	if path.Reverse3 {
		pool_src3 = pair3.TokenAccountA
		pool_dest3 = pair3.TokenAccountB
	}

	meta_tokenSwapAccount3 := solana.NewAccountMeta(pair3.SwapAccount, false, false)
	meta_authority3 := solana.NewAccountMeta(pair3.Authority, false, false)
	meta_wallet3 := solana.NewAccountMeta(privateKey.PublicKey(), false, true)
	meta_userSource3 := solana.NewAccountMeta(third_account, true, false)
	meta_poolSource3 := solana.NewAccountMeta(pool_src3, true, false)
	meta_poolDestination3 := solana.NewAccountMeta(pool_dest3, true, false)
	meta_userDestination3 := solana.NewAccountMeta(fourth_account, true, false)
	meta_poolMint3 := solana.NewAccountMeta(pair3.TokenPool, true, false)
	meta_feeAccount3 := solana.NewAccountMeta(pair3.FeeAccount, true, false)
	meta_tokenProgramId3 := solana.NewAccountMeta(TOKEN_PROGRAM, false, false)

	var meta_list3 []*solana.AccountMeta
	meta_list3 = append(meta_list3, meta_tokenSwapAccount3, meta_authority3, meta_wallet3, meta_userSource3, meta_poolSource3, meta_poolDestination3, meta_userDestination3, meta_poolMint3, meta_feeAccount3, meta_tokenProgramId3)

	//Création des instructions3
	buf3 := new(bytes.Buffer)
	borshEncoder3 := bin.NewBorshEncoder(buf3)
	err = borshEncoder3.Encode(Instruction_layout{1, amountout2, amountout3})
	if err != nil {
		return err
	}
	bytes_data3 := buf3.Bytes()
	instruction3 := solana.NewInstruction(Orca_swap, meta_list3, bytes_data3)

	//Création des comptes associés de la paire 4
	pair4 := path.Pair4
	fmt.Println(pair4.Name, path.Reverse4)

	var pool_src4 solana.PublicKey = pair4.TokenAccountB
	var pool_dest4 solana.PublicKey = pair4.TokenAccountA
	if path.Reverse4 {
		pool_src4 = pair4.TokenAccountA
		pool_dest4 = pair4.TokenAccountB
	}

	meta_tokenSwapAccount4 := solana.NewAccountMeta(pair4.SwapAccount, false, false)
	meta_authority4 := solana.NewAccountMeta(pair4.Authority, false, false)
	meta_wallet4 := solana.NewAccountMeta(privateKey.PublicKey(), false, true)
	meta_userSource4 := solana.NewAccountMeta(fourth_account, true, false)
	meta_poolSource4 := solana.NewAccountMeta(pool_src4, true, false)
	meta_poolDestination4 := solana.NewAccountMeta(pool_dest4, true, false)
	meta_userDestination4 := solana.NewAccountMeta(first_account, true, false)
	meta_poolMint4 := solana.NewAccountMeta(pair4.TokenPool, true, false)
	meta_feeAccount4 := solana.NewAccountMeta(pair4.FeeAccount, true, false)
	meta_tokenProgramId4 := solana.NewAccountMeta(TOKEN_PROGRAM, false, false)

	var meta_list4 []*solana.AccountMeta
	meta_list4 = append(meta_list4, meta_tokenSwapAccount4, meta_authority4, meta_wallet4, meta_userSource4, meta_poolSource4, meta_poolDestination4, meta_userDestination4, meta_poolMint4, meta_feeAccount4, meta_tokenProgramId4)

	//Création des instructions4
	buf4 := new(bytes.Buffer)
	borshEncoder4 := bin.NewBorshEncoder(buf4)
	err = borshEncoder4.Encode(Instruction_layout{1, amountout3, amountin})
	if err != nil {
		return err
	}
	bytes_data4 := buf4.Bytes()
	instruction4 := solana.NewInstruction(Orca_swap, meta_list4, bytes_data4)

	//Construction de la transaction
	transac := solana.NewTransactionBuilder()
	transac = transac.AddInstruction(instruction1)
	transac = transac.AddInstruction(instruction2)
	transac = transac.AddInstruction(instruction3)
	transac = transac.AddInstruction(instruction4)
	transac = transac.SetFeePayer(privateKey.PublicKey())
	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentConfirmed)
	if err != nil {
		return err
	}
	transac = transac.SetRecentBlockHash(recent.Value.Blockhash)
	final_transac, err := transac.Build()
	_, err = final_transac.Sign( //Signature de la transaction avec la clef privée
		func(key solana.PublicKey) *solana.PrivateKey {
			if privateKey.PublicKey().Equals(key) {
				return &privateKey
			}
			return nil
		},
	)
	sig, err := rpcClient.SendTransactionWithOpts(context.TODO(), final_transac, true, rpc.CommitmentProcessed) //envois de la transaction
	fmt.Println(sig)
	return err
}
func websocket_ammount(client *ws.Client, token *POOL_LAYOUT, pubkey string) {

	token_sub, err := solana.PublicKeyFromBase58(pubkey)
	if err != nil {
		panic(err)
	}

	ammid_sub, errsub := client.AccountSubscribe(token_sub, rpc.CommitmentProcessed)
	if errsub != nil {
		panic(errsub)
	}
	go get_token_balance(ammid_sub, token, token_sub, client)
}
func get_token_balance(sub *ws.AccountSubscription, balance *POOL_LAYOUT, token_sub solana.PublicKey, client *ws.Client) {
	for {
		got, err := sub.Recv()
		if err != nil {
			fmt.Printf("%v : %v\n", time.Now().Format(time.ANSIC), err)
			continue
		}
		decoder := bin.NewBinDecoder(got.Value.Data.GetBinary())
		decoder.Decode(balance)
	}
}
func init_account_info(account *POOL_LAYOUT, pubkey string) {
	client := rpc.New(rpc.MainNetBeta_RPC)
	pubKey := solana.MustPublicKeyFromBase58(pubkey)
	resp, err := client.GetAccountInfo(context.TODO(), pubKey)
	if err != nil {
		panic(err)
	}
	decoder := bin.NewBinDecoder(resp.Value.Data.GetBinary())
	decoder.Decode(account)
}
func compute_possible_path(pair_list map[string]*OrcaPair, reference string, currencies_list []string, possible_path *[]Path) {
	for i := 1; i < len(currencies_list); i++ {
		for j := 1; j < len(currencies_list); j++ {
			for k := 1; k < len(currencies_list); k++ {
				if k == i {
					continue
				}
				//Pair1
				_, ok1 := pair_list[reference+"/"+currencies_list[i]]
				_, ok2 := pair_list[currencies_list[i]+"/"+reference]
				//Pair2
				_, ok3 := pair_list[currencies_list[j]+"/"+currencies_list[i]]
				_, ok4 := pair_list[currencies_list[i]+"/"+currencies_list[j]]
				//Pair3
				_, ok5 := pair_list[currencies_list[k]+"/"+currencies_list[j]]
				_, ok6 := pair_list[currencies_list[j]+"/"+currencies_list[k]]
				//Pair4
				_, ok7 := pair_list[reference+"/"+currencies_list[k]]
				_, ok8 := pair_list[currencies_list[k]+"/"+reference]

				if (ok1 || ok2) && (ok3 || ok4) && (ok5 || ok6) && (ok7 || ok8) {
					var newpath Path
					if ok1 {
						newpath.Pair1 = pair_list[reference+"/"+currencies_list[i]]
						newpath.Reverse1 = true
					} else {
						newpath.Pair1 = pair_list[currencies_list[i]+"/"+reference]
						newpath.Reverse1 = false
					}
					if ok3 {
						newpath.Pair2 = pair_list[currencies_list[j]+"/"+currencies_list[i]]
						newpath.Reverse2 = false
					} else {
						newpath.Pair2 = pair_list[currencies_list[i]+"/"+currencies_list[j]]
						newpath.Reverse2 = true
					}
					if ok5 {
						newpath.Pair3 = pair_list[currencies_list[k]+"/"+currencies_list[j]]
						newpath.Reverse3 = false
					} else {
						newpath.Pair3 = pair_list[currencies_list[j]+"/"+currencies_list[k]]
						newpath.Reverse3 = true
					}
					if ok7 {
						newpath.Pair4 = pair_list[reference+"/"+currencies_list[k]]
						newpath.Reverse4 = false
					} else {
						newpath.Pair4 = pair_list[currencies_list[k]+"/"+reference]
						newpath.Reverse4 = true
					}
					newpath.Name = reference + "/" + currencies_list[i] + "/" + currencies_list[j] + "/" + currencies_list[k] + "/" + reference
					*possible_path = append(*possible_path, newpath)
				}
			}
		}
	}
}

func main() {
	fmt.Println("BEGINNING INITIALISATION")
	account_list := get_init_accounts()
	privkey, err := solana.PrivateKeyFromSolanaKeygenFile("")
	if err != nil {
		panic(err)
	}
	updateClient := jsonrpc.NewClient("")
	rpcClient := rpc.New("")
	wsClient, err := ws.Connect(context.Background(), "")
	for err != nil {
		fmt.Println(err)
		wsClient, err = ws.Connect(context.Background(), rpc.MainNetBeta_WS)
	}
	reference := os.Args[1]
	var referenceToken POOL_LAYOUT
	init_account_info(&referenceToken, account_list[reference])
	websocket_ammount(wsClient, &referenceToken, account_list[reference])
	currencies_list := []string{}
	currencies_list = append(currencies_list, reference)
	for currencies := range account_list {
		if currencies != reference {
			currencies_list = append(currencies_list, currencies)
		}
	}
	//Init Orca pairs
	Orca_pair_list := make(map[string]*OrcaPair)
	init_orca_pairs(Orca_pair_list, currencies_list, reference)
	var possible_path []Path
	compute_possible_path(Orca_pair_list, reference, currencies_list, &possible_path)

	//Update only relevent pairs
	var Orca_orderedpairlist []*OrcaPair
	for _, value := range possible_path {
		if value.Pair1 == nil || value.Pair2 == nil || value.Pair3 == nil || value.Pair4 == nil {
			fmt.Println(value)
		}
		if !pairInSlice(value.Pair1, Orca_orderedpairlist) {
			Orca_orderedpairlist = append(Orca_orderedpairlist, value.Pair1)
		}
		if !pairInSlice(value.Pair2, Orca_orderedpairlist) {
			Orca_orderedpairlist = append(Orca_orderedpairlist, value.Pair2)
		}
		if !pairInSlice(value.Pair3, Orca_orderedpairlist) {
			Orca_orderedpairlist = append(Orca_orderedpairlist, value.Pair3)
		}
		if !pairInSlice(value.Pair4, Orca_orderedpairlist) {
			Orca_orderedpairlist = append(Orca_orderedpairlist, value.Pair4)
		}
	}

	lastTrade := make(map[Path]time.Time)
	for _, value := range possible_path {
		fmt.Println(value.Name)
		lastTrade[value] = time.Time{}
	}
	fmt.Println("ENDING INITIALISATION")
	start := time.Now()
	for {
		//start := time.Now()
		if time.Since(start) > (2 * time.Minute) {
			fmt.Println("STILL UP AND RUNNING")
			start = time.Now()
		}
		update_orca_pools(Orca_orderedpairlist, updateClient)
		//fmt.Println(time.Since(start))
		//fmt.Println(Orca_pair_list["USDC/USDT"].CurveParameters, Orca_pair_list["USDC/USDT"].get_pools())
		compute_arbitrage(possible_path, referenceToken.Ammount, reference, privkey, rpcClient, account_list, lastTrade)
	}
}
