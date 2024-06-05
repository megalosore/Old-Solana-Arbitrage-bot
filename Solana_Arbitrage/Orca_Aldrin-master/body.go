package main

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
type ALDRIN_POOL_LAYOUT struct {
	Name            string
	PoolMint        string
	PoolPublicKey   string
	BaseTokenMint   string
	QuoteTokenMint  string
	BaseTokenVault  string
	QuoteTokenVault string
	Curve           string
}
type Arbitrage_result struct {
	Token       string
	Order       bool //1 = RAY => ORCA
	In_amount   uint64
	Out_amount1 uint64
	Out_amount2 uint64
	Profit      int64
}
type Instruction_layout struct {
	Instruction  uint8
	AmountIn     uint64
	MinAmountOut uint64
}
type Side struct {
	Defined float64
}
type RSide struct {
	Defined int64
}
type Aldrin_Instruction_layout struct {
	Instruction [8]byte
	Tokens      uint64
	MinTokens   uint64
	Side        Side
}
type RAldrin_Instruction_layout struct {
	Instruction [8]byte
	Tokens      uint64
	MinTokens   uint64
	Side        RSide
}
type FEELAYOUT struct {
	TradeFeeNumerator           uint64
	TradeFeeDenominator         uint64
	OwnerTradeFeeNumerator      uint64
	OwnerTradeFeeDenominator    uint64
	OwnerWithdrawFeeNumerator   uint64
	OwnerWithdrawFeeDenominator uint64
}
type Aldrin_pairs struct {
	Padding             [8]byte
	LpTokenFreezeVault  solana.PublicKey
	PoolMint            solana.PublicKey
	BaseTokenVault      solana.PublicKey
	BaseTokenMint       solana.PublicKey
	QuoteTokenVault     solana.PublicKey
	QuoteTokenMint      solana.PublicKey
	PoolSigner          solana.PublicKey
	PoolSignerNonce     uint8
	Authority           solana.PublicKey
	InitializerAccount  solana.PublicKey
	FeeBaseAccount      solana.PublicKey
	FeeQuoteAccount     solana.PublicKey
	FeePoolTokenAccount solana.PublicKey
	FEES_LAYOUT         FEELAYOUT
	Cointoken           POOL_LAYOUT
	Pctoken             POOL_LAYOUT
	Name                string
	PoolPublicKey       solana.PublicKey
	Curve               solana.PublicKey
}

func stringInSlice(a string, list []string) bool {
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
func init_aldrin_pool() map[string]ALDRIN_POOL_LAYOUT {
	ALL_POOLS := `{
		"PANDA_USDC": {
			"name" : "PANDA_USDC",
			"poolMint": "7AMQffX5RnwQMfWZJ7jxxp8zhmntLyziJ2Za7eNHAR8t",
			"poolPublicKey": "7EchbXqY3d7vqEs7zHEnZEhvMdGBvgWKRxjBYnA2smyJ",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "2U8uGQDNxYGK7N8oje1pJ2H1aenK8PtgwKVK5sqYbXhS",
			"quoteTokenVault": "ZGUxR916YkdCkWpQZqHYDBq64PMyQ95S3XYJGAqM4QT",
			"curve" : "5SLMNL3r75YeYyWyxTdVCqNvxSMEBbENtMviqka7RPfW"
		},
		"SOLA_USDC": {
			"name" : "SOLA_USDC",
			"poolMint": "C8brpX2n8Peuhpts1Cpo55HvHdFsnXbU26F1yn9WGnuC",
			"poolPublicKey": "5qEkhwFvymhSzbQR6Qqns6rAS9SXS79g7a4FtRxTXts2",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "FoRPexeBMC9dAKPBHD465FhnvXUMugVwK3MNYuEZPAUp",
			"quoteTokenVault": "9fgf4JeyXUNof7BJEr49eDvF8m1TVHkKTFYbg6rSmpfC",
			"curve" : "7HVf95mWddqWenDyXK43ZEh6XHzxLGqz65sapHHCZzaW"
		},
		"BABY_USDC": {
			"name" : "BABY_USDC",
			"poolMint": "3K2S5WPRDDJ3NSvYyjwt71rRdnZFv3oDJHLUs2Lnc1fv",
			"poolPublicKey": "3s5MnHpfMQE5yYaLWfPiKWWEeKuHejgTXyD4PWAPNzvM",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "38GdJotkZH55BMuzqVMHiUgdrmiMij64ZSMoLYAV31a3",
			"quoteTokenVault": "62HRkTskR8ZbNiWcsph88aM2kaXTDEj4QgNJPXrzgKRg",
			"curve" : "No"
		},
		"JUNGLE_USDC": {
			"name" : "JUNGLE_USDC",
			"poolMint": "9TS4qcHgTuAJ5JY4e3PvXp3b84cJ3PHZaoqg4cNEXBMF",
			"poolPublicKey": "2AiUX9wcMUJh8vcnx2fHY9kHrw2f1mamV5fnriLZQDHv",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "5KiPd7vTx2y8yNDQeHeSQn6FCjmF28Gzz9tF1XTF5RHE",
			"quoteTokenVault": "9oZSaBCvLVV4GStGwaZhzTGpX5yGhzZXwbYVbX4BAc5A",
			"curve" : "No"
		},
		"HIMA_USDC": {
			"name" : "HIMA_USDC",
			"poolMint": "7bDRyQyvsfjXy9kPvj6fj2W6bX6cgEH5gCPa9Ln5fCKU",
			"poolPublicKey": "DyoosXiS4F1ujmY4XX6t5i1EzXQ3vtuqXfKp4atrSCqR",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "B28c5qZBhLMy3PKMNfbWGjYNfi2C9fEPXyUQzFmmj6hR",
			"quoteTokenVault": "4vUZ5eQPqvKMjr1SevQjBi5mwp8MHxsKozKYJ7uwmbs3",
			"curve" : "No"
		},
		"PUFF_USDC": {
			"name" : "PUFF_USDC",
			"poolMint": "6UVQCtF8rRTBC27u1wL6YEYUsHcic3VPPkxr5TVtanHZ",
			"poolPublicKey": "2ozfF94ZmTaAqXdoaqekzhM2VX5ZdNA1pEeLAkPWu4Lx",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "J6Yhn6BQRKZiShjNukth3e43GoE8zFdBPgC3ejSjxYvz",
			"quoteTokenVault": "9vTbm7fvX78vxpDJxaGxyoL7Drboq9s64zPPJnBa2s3U",
			"curve" : "No"
		},
		"FUM_USDC": {
			"name" : "FUM_USDC",
			"poolMint": "3ZFTVkCrnVvm7jqmVqYn1NupgG9k8BRqaui2BjMUMzwQ",
			"poolPublicKey": "HT7TQn9io3AwTt7hpBTJxBuzf27ugB1ZrD9G8UecUzbx",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "GX2UoDCGh11GuQV59MDoHPNeSx1Qz49W2n3RiRpZ24Ww",
			"quoteTokenVault": "8TtkzA63RPmX777PGrcLRYdtWowA5E7pMr8dbPX3nfSU",
			"curve" : "4YrCyUPJgkjYCuhbBo6LXe4DSTqJB87KffeWki7PGbJt"
		},
		"SOULO_USDC": {
			"name" : "SOULO_USDC",
			"poolMint": "89MSBKMWUB8qwKNVF2LuAwoKiNqqPKwyZH9x3USALak8",
			"poolPublicKey": "zbyUkk999LwQdSLMAvg3iAEU8DRvidvzxzinv35n2gW",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "Cwp75j1bfXP4rWqiHgh2azTj33wsjCCz9k1PRXPFKeKz",
			"quoteTokenVault": "EkBzuLfUsV2bJP8rF3GYUKxA9fveSzeuRjP1KMX6eRmp",
			"curve" : "HsoMGs7nBErd9CxB8xwfGRkEbaeodDmVo9S1v3wC8fqD"
		},
		"TICKET_USDC": {
			"name" : "TICKET_USDC",
			"poolMint": "FKQX91BgH633Ww72Jpeq2g11MC3MTP97k4Nd2bk2W2Vn",
			"poolPublicKey": "AFu6fk1de2XbmvS16m3E2zA1zANR6Zwg9TBDmo8xDpDu",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "6hEZrVBkpNt7xCZCFquDNcammi4hy5TDLrsDcnZP53zv",
			"quoteTokenVault": "4C9gL7KJkriRR3UqRLgZMJ2R4MR9A3qLJ8LBDwdNzGae",
			"curve" : "BEB9isDBZuAR53yy3Ycpi5evm3YP3wZhNfDPxqKot2Qu"
		},
		"RIN_USDC": {
			"name" : "RIN_USDC",
			"poolMint": "Gathk79qZfJ4G36M7hiL3Ef1P5SDt7Xhm2C1vPhtWkrw",
			"poolPublicKey": "Gubmyfw5Ekdp4pkXk9be5yNckSgCdgd7JEThx8SFzCQQ",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "8YuEKfvSwcfNKvdoHijzrUAgEeevj4529m8SddSYQ8FV",
			"quoteTokenVault": "5P7J5sPvJmdnNX4JuhGDsNRnTihVMY8q4dHHbbmQUouJ",
			"curve" : "No"
		},
		"RIN_SOL": {
			"name" : "RIN_SOL",
			"poolMint": "HFNv9CeUtKFKm7gPoX1QG1NnrPnDhA5W6xqHGxmV6kxX",
			"poolPublicKey": "7nrkzur3LUxgfxr3TBj9GpUsiABCwMgzwtNhHKG4hPYz",
			"baseTokenMint": "E5ndSkaB17Dm7CsD22dvcjfrYSDLCxFcMd6z8ddCk5wp",
			"quoteTokenMint": "So11111111111111111111111111111111111111112",
			"baseTokenVault": "3reyueV93V8CxXakMk4FF96uqBibDS9Di7zWgjxhkqt7",
			"quoteTokenVault": "3LX2NHkUux6gGjiQXY2nMCnLTr9QuCjguqh7KTwaupV5",
			"curve" : "No"
		},
		"mSOL_USDT": {
			"name" : "mSOL_USDT",
			"poolMint": "77qHkg6TEe4FuZAr35bthTEadmT4ueWe1xomFFZkwiGQ",
			"poolPublicKey": "FC4sYMpsMvdsq8hHMEtmWA8xN25W71t2c7RycU5juX35",
			"baseTokenMint": "mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So",
			"quoteTokenMint": "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
			"baseTokenVault": "4aiKnDHFmNnLsopVsDyRBh8sbVohZYgdGzh3P9orpqNB",
			"quoteTokenVault": "HFHGsYQyni5gFMGudaHWpRzN5CejNpHr42PfQ4D6aGZM",
			"curve" : "No"
		},
		"mSOL_ETH": {
			"name" : "mSOL_ETH",
			"poolMint": "4KeZGuXPq9fyZdt5sfzHMM36mxTf3oSkDaa4Y4gHm9Hz",
			"poolPublicKey": "2JANvFVV2M8dv7twzL1EF3PhEJaoJpvSt6PhnjW6AHG6",
			"baseTokenMint": "mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So",
			"quoteTokenMint": "2FPyTwcZLUg1MDrwsyoP4D6s1tM7hAkHYRjkNb5w6Pxk",
			"baseTokenVault": "9MaVbwbZw3LgFTNAPfDj4viRAffXFGaAdJWfX3ifouHf",
			"quoteTokenVault": "6YwwwDQcQz5qAEipFJXHe3vBMKDcs9nfZXLitEubMxFc",
			"curve" : "No"
		},
		"mSOL_BTC": {
			"name" : "mSOL_BTC",
			"poolMint": "9hkYqNM8QSx2vTwspaNg5VvW1LBxKWWgud8pCVdxKYZU",
			"poolPublicKey": "13FjT6LMUH9LQLQn6KGjJ1GNXKvgzoDSdxvHvAd4hcan",
			"baseTokenMint": "mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So",
			"quoteTokenMint": "9n4nbM75f5Ui33ZbPYXn59EwSgE8CGsHtAeTH5YFeJ9E",
			"baseTokenVault": "EhWAErmyrX8nT1eT8HVFw37amsEkm5VjKZH4ZUreDRCs",
			"quoteTokenVault": "Fpy5DXqdz7mfLDF8PYKVzxQrYtsaiQu36fLpv6gmGseH",
			"curve" : "No"
		},
		"mSOL_USDC": {
			"name" : "mSOL_USDC",
			"poolMint": "H37kHxy82uLoF8t86wK414KzpVJy7uVJ9Kvt5wYsTGPh",
			"poolPublicKey": "Af4TpzGpo8Yc61bCNwactPKH9F951tHPzp8XGxWRLNE1",
			"baseTokenMint": "mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "BEPiCaDinG2uLSBKjiVGAdDV32dwiemANKJYejtpbT2h",
			"quoteTokenVault": "9CDfE5NfRcQomM7bZ2fCBLe9XKebmu8QY5tBHzojS8d8",
			"curve" : "No"
		},
		"SOL_USDC": {
			"name" : "SOL_USDC",
			"poolMint": "3sbMDzGtyHAzJqzxE7DPdLMhrsxQASYoKLkHMYJPuWkp",
			"poolPublicKey": "4GUniSDrCAZR3sKtLa1AWC8oyYubZeKJQ8KraQmy3Wt5",
			"baseTokenMint": "So11111111111111111111111111111111111111112",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "CLt1DtCioiByTizqLhxLAXweXr2g9D4ZEAStibACBg4L",
			"quoteTokenVault": "2M1JTZsc71V6FhRNjCDSttcs17HewC4KNNNkkc81L3gB",
			"curve" : "No"
		},
		"mSOL_UST": {
			"name" : "mSOL_UST",
			"poolMint": "BE7eTJ8DB7xTu6sKsch4gWDCXbD48PLGesRLx7E1Qce4",
			"poolPublicKey": "EnKhda5n5LYbZjPv7d7WChkSXzo5RgV8eSVVkGCXsQUn",
			"baseTokenMint": "mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So",
			"quoteTokenMint": "9vMJfxuKxXBoEa7rM12mYLMwTacLMLDJqHozw96WQL8i",
			"baseTokenVault": "29jNBEn9VEvM5ppVLGThrGc7ExnT3WyNhYyqbizpyNFK",
			"quoteTokenVault": "6RmiUpwLquyQWVMeYx4oktQvtCuUH48fzRMwbC5kUa4h",
			"curve" : "No"
		},
		"mSOL_MNGO": {
			"name" : "mSOL_MNGO",
			"poolMint": "EotLYRsnRVqR3euN24P9PMXCqJv1WLsV8kJxR9o1y4U7",
			"poolPublicKey": "CAHchWN1xoxNvXmqmmj6U834ip585rXZbh9NkvE9vTea",
			"baseTokenMint": "mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So",
			"quoteTokenMint": "MangoCzJ36AjZyKwVj3VnYU4GTonjfVEnJmvvWaxLac",
			"baseTokenVault": "FE3PR8sbojxrxWoTzuLHhDX5hAXfPocS9wCruSJ2y7BF",
			"quoteTokenVault": "CJzgYvbf2pv6HiTu13ymSVDSRmVQFoF8rkFYvwDNWVJL",
			"curve" : "No"
		},
		"SLX_USDC": {
			"name" : "SLX_USDC",
			"poolMint": "E3XeF4QCTMMo8P5yrgqNMvoRJMyVPTNHhWkbRCgoeAfC",
			"poolPublicKey": "Hv5F48Br7dbZvUpKFuyxxuaC4v95C1uyDGhdkFFCc9Gf",
			"baseTokenMint": "AASdD9rAefJ4PP7iM89MYUsQEyCQwvBofhceZUGDh5HZ",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "BWwtpKxkKJSe4Vv9Ha6iGxdgWwxvy2k6qwooE6WjwUhG",
			"quoteTokenVault": "EF9M6hDSSTZhwdSrpKvkHn33EafZfRtaFQ9rq6MfqALm",
			"curve" : "No"
		},
		"DATE_USDC": {
			"name" : "DATE_USDC",
			"poolMint": "3gigDvmgCuz2gWRhr6iSxH1gCd1k4LpYoUsxEjLWJcLB",
			"poolPublicKey": "F5MWosWE681D32N5QHbWWaJrXaMAD2PHhDsEr2Sac56X",
			"baseTokenMint": "Ce3PSQfkxT5ua4r2JqCoWYrMwKWC5hEzwsrT9Hb7mAz9",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "A9FRb9MyAipfzEwkbF5euJ2dE7Bcj1GHBhdSdN8KBkoE",
			"quoteTokenVault": "EVdVfqSEhumJmL3adAs4qCpEGGBGUoTdqG1krqMDboi4",
			"curve" : "No"
		},
		"OOGI_USDC": {
			"name" : "OOGI_USDC",
			"poolMint": "46EsyeSzs6tBoTRmFiGfDzGQe13LP337C7mMtdNMkgcU",
			"poolPublicKey": "6sKC96Z35vCNcDmA3ZbBd9Syx5gnTJdyNKVEdzpBE5uX",
			"baseTokenMint": "H7Qc9APCWWGDVxGD5fJHmLTmdEgT9GFatAKFNg6sHh8A",
			"quoteTokenMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			"baseTokenVault": "8fJUjr1o5i48R4URoQXDVbXmjrAUcRZWFkc2U2TMvgEd",
			"quoteTokenVault": "AdgHaAavPxwjBVw966dZPA9h9MbkbYXAXViGiQRBXMNJ",
			"curve" : "No"
		}
	}`
	ALL_POOLS_MAPPED := make(map[string]ALDRIN_POOL_LAYOUT)
	json.Unmarshal([]byte(ALL_POOLS), &ALL_POOLS_MAPPED)
	return ALL_POOLS_MAPPED
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
			if tokens[0] == reference && !stringInSlice(tokens[1], currencies_list) {
				continue
			} else if tokens[1] == reference && !stringInSlice(tokens[0], currencies_list) {
				continue
			} else if tokens[1] != reference && tokens[0] != reference {
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
func init_aldrin_pairs(currencies_list []string, pair_list map[string]*Aldrin_pairs, reference string) {
	aldrin_pools := init_aldrin_pool()
	rpcClient := rpc.New("https://ssc-dao.genesysgo.net/")
	for key, element := range aldrin_pools { //On initialise toute les paires utilisables en demandant au serveur leurs valeurs
		tokens := strings.Split(key, "_")
		if tokens[0] == reference && !stringInSlice(tokens[1], currencies_list) {
			continue
		} else if tokens[1] == reference && !stringInSlice(tokens[0], currencies_list) {
			continue
		} else if tokens[1] != reference && tokens[0] != reference {
			continue
		}
		fmt.Println(key)
		pubkey := solana.MustPublicKeyFromBase58(element.PoolPublicKey)
		res, _ := rpcClient.GetAccountInfo(context.TODO(), pubkey)
		var newpair Aldrin_pairs
		swapdecoder := bin.NewBinDecoder(res.Value.Data.GetBinary())
		swapdecoder.Decode(&newpair)
		newpair.Name = element.Name
		newpair.PoolPublicKey = pubkey
		if element.Curve != "No" {
			curvekey := solana.MustPublicKeyFromBase58(element.Curve)
			newpair.Curve = curvekey
		}
		pair_list[element.Name] = &newpair
	}
}

func (pair Aldrin_pairs) get_pools() [2]uint64 {
	totalcoin := pair.Cointoken.Ammount
	totalpc := pair.Pctoken.Ammount
	return [2]uint64{totalcoin, totalpc}
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
func update_pools(orca_pair []*OrcaPair, aldrin_pair []*Aldrin_pairs, rpcClient jsonrpc.RPCClient) {
	//client
	//rpcClient := jsonrpc.NewClient("https://ssc-dao.genesysgo.net/")
	var input []*jsonrpc.RPCRequest
	//liste de requete
	for _, element := range aldrin_pair {
		input = append(input, jsonrpc.NewRequest("getAccountInfo", element.BaseTokenVault, ENCODING{"base64", "processed"}))
		input = append(input, jsonrpc.NewRequest("getAccountInfo", element.QuoteTokenVault, ENCODING{"base64", "processed"}))
	}
	for _, element := range orca_pair {
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
	for _, element := range aldrin_pair {
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
	for _, element := range orca_pair {
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
func compute_arbitrage(Aldrin_pair_list map[string]*Aldrin_pairs, Orca_pair_list map[string]*OrcaPair, reference_ammount uint64, reference string, privateKey solana.PrivateKey, rpcClient *rpc.Client, accounts_list map[string]string, lastTrade map[*Aldrin_pairs]time.Time) {
	for index, aldrin_pair := range Aldrin_pair_list {
		if time.Now().Sub(lastTrade[aldrin_pair]) < 10*time.Second {
			continue
		}
		orca_fee := "0.997"
		aldrin_reversed := false
		tokens := strings.Split(index, "_")
		token := tokens[0]
		if token == reference {
			token = tokens[1]
			aldrin_reversed = true
		}
		orca_reversed := false
		orca_pair, ok := Orca_pair_list[token+"/"+reference]
		if !ok {
			orca_pair = Orca_pair_list[reference+"/"+token]
			orca_reversed = true
		}

		aldrin_pools := aldrin_pair.get_pools()
		orca_pools := orca_pair.get_pools()

		//Aldrin => ORCA
		var amount_in uint64
		a := float64(aldrin_pools[1])
		b := float64(aldrin_pools[0])
		c := float64(orca_pools[1])
		d := float64(orca_pools[0])
		z := 0.9975
		y, _ := strconv.ParseFloat(orca_fee, 64)

		if aldrin_reversed {
			tmp := a
			a = b
			b = tmp
		}
		if orca_reversed {
			tmp := c
			c = d
			d = tmp
		}

		max_in := uint64(math.Abs((-a*d + math.Sqrt(a*b*c*d*z*y)) / (b*y*z + d*y)))
		//x := float64(1000000)
		//test := ((b * c * x * y * z) / (a*d + b*x*y*z + d*x*z)) - x
		//fmt.Println("RO", token, aldrin_pools, orca_pools, test)

		if max_in > reference_ammount {
			amount_in = reference_ammount
		} else {
			amount_in = max_in
		}

		amountout1 := get_amount_out(amount_in, aldrin_pools, aldrin_reversed, "0.997")
		amountout2 := get_amount_out(amountout1, orca_pools, !orca_reversed, orca_fee)
		profit := int64(amountout2) - int64(amount_in)

		if profit > 2000 {
			risk_lvl := 0.001 //Account for imprecision in calculations
			amountout1 -= uint64(risk_lvl * float64(amountout1))
			resend := 1
			if profit > 100000 {
				resend = 5
			}
			for i := 0; i < resend; i++ {
				sendAldrinOrca(privateKey, rpcClient, accounts_list, aldrin_pair, orca_pair, aldrin_reversed, !orca_reversed, amount_in, amountout1)
			}
			lastTrade[aldrin_pair] = time.Now()
			fmt.Printf("Transaction: %v | Profit : %v\n", token, profit)
			fmt.Println(amount_in, amountout1, amountout2, a, b, c, d, max_in)
			fmt.Println("---------------------------------------------------------------")
		}
		//ORCA => Aldrin
		var reverse_amount_in uint64
		a = float64(orca_pools[1])
		b = float64(orca_pools[0])
		c = float64(aldrin_pools[1])
		d = float64(aldrin_pools[0])
		z, _ = strconv.ParseFloat(orca_fee, 64)
		y = 0.9975

		if orca_reversed {
			tmp := a
			a = b
			b = tmp
		}
		if aldrin_reversed {
			tmp := c
			c = d
			d = tmp
		}

		max_in = uint64(math.Abs((-a*d + math.Sqrt(a*b*c*d*z*y)) / (b*y*z + d*y)))
		//test = ((b * c * x * y * z) / (a*d + b*x*y*z + d*x*z)) - x
		//fmt.Println("OR", token, aldrin_pools, orca_pools, test)
		if max_in > reference_ammount {
			reverse_amount_in = reference_ammount
		} else {
			reverse_amount_in = max_in
		}

		reverseamountout1 := get_amount_out(reverse_amount_in, orca_pools, orca_reversed, orca_fee)
		reverseamountout2 := get_amount_out(reverseamountout1, aldrin_pools, !aldrin_reversed, "0.997")
		reverseprofit := int64(reverseamountout2) - int64(reverse_amount_in)
		//fmt.Println("OR", token, reverse_amount_in, reverseamountout1, reverseamountout2, reverseprofit)

		if reverseprofit > 2000 {
			risk_lvl := 0.001
			reverseamountout1 -= uint64(risk_lvl * float64(reverseamountout1))
			resend := 1
			if reverseprofit > 100000 {
				resend = 5
			}
			for i := 0; i < resend; i++ {
				sendOrcaAldrin(privateKey, rpcClient, accounts_list, orca_pair, aldrin_pair, orca_reversed, !aldrin_reversed, reverse_amount_in, reverseamountout1)
			}
			lastTrade[aldrin_pair] = time.Now()
			fmt.Printf("Transaction: %v | Profit : %v\n", token, reverseprofit)
			fmt.Println(reverse_amount_in, reverseamountout1, reverseamountout2, a, b, c, d, max_in)
			fmt.Println("---------------------------------------------------------------")
		}
	}
}
func sendOrcaAldrin(privateKey solana.PrivateKey, rpcClient *rpc.Client, accounts_list map[string]string, pair1 *OrcaPair, pair2 *Aldrin_pairs, reverse1 bool, reverse2 bool, amountin uint64, amountout1 uint64) error {

	tokens := strings.Split(pair1.Name, "/")
	first := tokens[1]
	second := tokens[0]

	TOKEN_PROGRAM, _ := solana.PublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	first_account, _ := solana.PublicKeyFromBase58(accounts_list[first])
	second_account, _ := solana.PublicKeyFromBase58(accounts_list[second])
	Orca_swap, _ := solana.PublicKeyFromBase58("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP")
	Aldrin_swap, _ := solana.PublicKeyFromBase58("AMM55ShdkoGRB5jVYPjWziwk8m5MpwyDgsMWHaMSQWH6")
	Aldrin_swapV2, _ := solana.PublicKeyFromBase58("CURVGoZn8zycx6FXwwevgBTB2gVvdbGTEpvMJDbgs2t4")

	//pair1
	var pool_src1 solana.PublicKey = pair1.TokenAccountB
	var pool_dest1 solana.PublicKey = pair1.TokenAccountA
	if reverse1 {
		pool_src1 = pair1.TokenAccountA
		pool_dest1 = pair1.TokenAccountB
		first_account, _ = solana.PublicKeyFromBase58(accounts_list[second])
		second_account, _ = solana.PublicKeyFromBase58(accounts_list[first])
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
	tokens = strings.Split(pair2.Name, "_")
	first = tokens[0]
	second = tokens[1]
	base_account, _ := solana.PublicKeyFromBase58(accounts_list[first])
	quote_account, _ := solana.PublicKeyFromBase58(accounts_list[second])

	meta_pool2 := solana.NewAccountMeta(pair2.PoolPublicKey, false, false)
	meta_pool_signer2 := solana.NewAccountMeta(pair2.PoolSigner, false, false)
	meta_pool_mint2 := solana.NewAccountMeta(pair2.PoolMint, true, false)
	meta_base_vault2 := solana.NewAccountMeta(pair2.BaseTokenVault, true, false)
	meta_quote_vault2 := solana.NewAccountMeta(pair2.QuoteTokenVault, true, false)
	meta_fee2 := solana.NewAccountMeta(pair2.FeePoolTokenAccount, true, false)
	meata_wallet2 := solana.NewAccountMeta(privateKey.PublicKey(), false, true)
	meta_user_base2 := solana.NewAccountMeta(base_account, false, false)
	meta_user_quote2 := solana.NewAccountMeta(quote_account, false, false)
	meta_token_program2 := solana.NewAccountMeta(TOKEN_PROGRAM, false, false)

	var meta_list2 []*solana.AccountMeta
	meta_list2 = append(meta_list2, meta_pool2, meta_pool_signer2, meta_pool_mint2, meta_base_vault2, meta_quote_vault2, meta_fee2, meata_wallet2, meta_user_base2, meta_user_quote2)
	var default_value solana.PublicKey
	var swap_address = Aldrin_swap
	if pair2.Curve != default_value {
		meta_curve := solana.NewAccountMeta(pair2.Curve, false, false)
		meta_list2 = append(meta_list2, meta_curve)
		swap_address = Aldrin_swapV2
	}
	meta_list2 = append(meta_list2, meta_token_program2)

	//Création des instructions2
	buf2 := new(bytes.Buffer)
	borshEncoder2 := bin.NewBorshEncoder(buf2)
	swap_instruction := [8]byte{248, 198, 158, 145, 225, 117, 135, 200}
	if !reverse2 {
		err = borshEncoder2.Encode(Aldrin_Instruction_layout{swap_instruction, amountout1, amountin, Side{float64(-1)}})
	}
	if reverse2 {
		err = borshEncoder2.Encode(RAldrin_Instruction_layout{swap_instruction, amountout1, amountin, RSide{int64(1)}})
	}
	fmt.Println(amountout1)
	if err != nil {
		return err
	}
	bytes_data2 := buf2.Bytes()
	instruction2 := solana.NewInstruction(swap_address, meta_list2, bytes_data2) //Création des instructions2

	//Construction de la transaction
	transac := solana.NewTransactionBuilder()
	transac = transac.AddInstruction(instruction1)
	transac = transac.AddInstruction(instruction2)
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
func sendAldrinOrca(privateKey solana.PrivateKey, rpcClient *rpc.Client, accounts_list map[string]string, pair1 *Aldrin_pairs, pair2 *OrcaPair, reverse1 bool, reverse2 bool, amountin uint64, amountout1 uint64) error {

	TOKEN_PROGRAM, _ := solana.PublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	Orca_swap, _ := solana.PublicKeyFromBase58("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP")
	Aldrin_swap, _ := solana.PublicKeyFromBase58("AMM55ShdkoGRB5jVYPjWziwk8m5MpwyDgsMWHaMSQWH6")
	Aldrin_swapV2, _ := solana.PublicKeyFromBase58("CURVGoZn8zycx6FXwwevgBTB2gVvdbGTEpvMJDbgs2t4")

	//Création des comptes associés de la paire1
	tokens := strings.Split(pair1.Name, "_")
	first := tokens[0]
	second := tokens[1]
	base_account, _ := solana.PublicKeyFromBase58(accounts_list[first])
	quote_account, _ := solana.PublicKeyFromBase58(accounts_list[second])

	meta_pool1 := solana.NewAccountMeta(pair1.PoolPublicKey, false, false)
	meta_pool_signer1 := solana.NewAccountMeta(pair1.PoolSigner, false, false)
	meta_pool_mint1 := solana.NewAccountMeta(pair1.PoolMint, true, false)
	meta_base_vault1 := solana.NewAccountMeta(pair1.BaseTokenVault, true, false)
	meta_quote_vault1 := solana.NewAccountMeta(pair1.QuoteTokenVault, true, false)
	meta_fee1 := solana.NewAccountMeta(pair1.FeePoolTokenAccount, true, false)
	meta_wallet1 := solana.NewAccountMeta(privateKey.PublicKey(), false, true)
	meta_user_base1 := solana.NewAccountMeta(base_account, false, false)
	meta_user_quote1 := solana.NewAccountMeta(quote_account, false, false)
	meta_token_program1 := solana.NewAccountMeta(TOKEN_PROGRAM, false, false)

	var meta_list1 []*solana.AccountMeta
	meta_list1 = append(meta_list1, meta_pool1, meta_pool_signer1, meta_pool_mint1, meta_base_vault1, meta_quote_vault1, meta_fee1, meta_wallet1, meta_user_base1, meta_user_quote1)
	var default_value solana.PublicKey
	var swap_address = Aldrin_swap
	if pair1.Curve != default_value {
		meta_curve := solana.NewAccountMeta(pair1.Curve, false, false)
		meta_list1 = append(meta_list1, meta_curve)
		swap_address = Aldrin_swapV2
	}
	meta_list1 = append(meta_list1, meta_token_program1)
	//Création des instructions1
	buf1 := new(bytes.Buffer)
	borshEncoder1 := bin.NewBorshEncoder(buf1)
	swap_instruction := [8]byte{248, 198, 158, 145, 225, 117, 135, 200}
	var err error
	if !reverse1 {
		err = borshEncoder1.Encode(Aldrin_Instruction_layout{swap_instruction, amountin, amountout1, Side{float64(-1)}})
	}
	if reverse1 {
		err = borshEncoder1.Encode(RAldrin_Instruction_layout{swap_instruction, amountin, amountout1, RSide{int64(1)}})
	}
	if err != nil {
		return err
	}
	bytes_data1 := buf1.Bytes()
	instruction1 := solana.NewInstruction(swap_address, meta_list1, bytes_data1) //Création des instructions2

	//pair2
	tokens = strings.Split(pair2.Name, "/")
	first = tokens[1]
	second = tokens[0]
	first_account, _ := solana.PublicKeyFromBase58(accounts_list[first])
	second_account, _ := solana.PublicKeyFromBase58(accounts_list[second])

	var pool_src2 solana.PublicKey = pair2.TokenAccountB
	var pool_dest2 solana.PublicKey = pair2.TokenAccountA
	if reverse2 {
		pool_src2 = pair2.TokenAccountA
		pool_dest2 = pair2.TokenAccountB
		first_account, _ = solana.PublicKeyFromBase58(accounts_list[second])
		second_account, _ = solana.PublicKeyFromBase58(accounts_list[first])
	}

	meta_tokenSwapAccount2 := solana.NewAccountMeta(pair2.SwapAccount, false, false)
	meta_authority2 := solana.NewAccountMeta(pair2.Authority, false, false)
	meta_wallet2 := solana.NewAccountMeta(privateKey.PublicKey(), false, true)
	meta_userSource2 := solana.NewAccountMeta(first_account, true, false)
	meta_poolSource2 := solana.NewAccountMeta(pool_src2, true, false)
	meta_poolDestination2 := solana.NewAccountMeta(pool_dest2, true, false)
	meta_userDestination2 := solana.NewAccountMeta(second_account, true, false)
	meta_poolMint2 := solana.NewAccountMeta(pair2.TokenPool, true, false)
	meta_feeAccount2 := solana.NewAccountMeta(pair2.FeeAccount, true, false)
	meta_tokenProgramId2 := solana.NewAccountMeta(TOKEN_PROGRAM, false, false)

	var meta_list2 []*solana.AccountMeta
	meta_list2 = append(meta_list2, meta_tokenSwapAccount2, meta_authority2, meta_wallet2, meta_userSource2, meta_poolSource2, meta_poolDestination2, meta_userDestination2, meta_poolMint2, meta_feeAccount2, meta_tokenProgramId2)
	//Création des instructions1
	buf2 := new(bytes.Buffer)
	borshEncoder2 := bin.NewBorshEncoder(buf2)
	err = borshEncoder2.Encode(Instruction_layout{1, amountout1, amountin})
	if err != nil {
		return err
	}
	bytes_data2 := buf2.Bytes()
	instruction2 := solana.NewInstruction(Orca_swap, meta_list2, bytes_data2)

	//Construction de la transaction
	transac := solana.NewTransactionBuilder()
	transac = transac.AddInstruction(instruction1)
	transac = transac.AddInstruction(instruction2)
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
	for currencies := range account_list {
		if currencies != reference {
			currencies_list = append(currencies_list, currencies)
		}
	}
	//Init Orca pairs
	Orca_pair_list := make(map[string]*OrcaPair)
	init_orca_pairs(Orca_pair_list, currencies_list, reference)

	//Init Aldrin pairs
	Aldrin_pair_list := make(map[string]*Aldrin_pairs)
	init_aldrin_pairs(currencies_list, Aldrin_pair_list, reference)
	//fmt.Println(Aldrin_pair_list["OOGI_USDC"])
	//init_raydium_pairs(currencies_list, Raydium_pair_list, reference)

	//Finding common tokens
	commontokens := []string{}
	for pairs := range Orca_pair_list {
		tokens := strings.Split(pairs, "/")
		var ok1 bool
		var ok2 bool
		if tokens[0] != reference {
			_, ok1 = Aldrin_pair_list[tokens[0]+"_"+reference]
			_, ok2 = Aldrin_pair_list[reference+"_"+tokens[0]]
			if ok1 || ok2 {
				commontokens = append(commontokens, tokens[0])
			}
		} else {
			_, ok1 = Aldrin_pair_list[tokens[1]+"_"+reference]
			_, ok2 = Aldrin_pair_list[reference+"_"+tokens[1]]
			if ok1 || ok2 {
				commontokens = append(commontokens, tokens[1])
			}
		}
	}
	//Removing unusable pairs
	for pairs := range Aldrin_pair_list {
		tokens := strings.Split(pairs, "_")
		var token string
		if tokens[0] != reference {
			token = tokens[0]
		} else {
			token = tokens[1]
		}
		if !stringInSlice(token, commontokens) {
			delete(Aldrin_pair_list, pairs)
		}
	}
	for pairs := range Orca_pair_list {
		tokens := strings.Split(pairs, "/")
		var token string
		if tokens[0] != reference {
			token = tokens[0]
		} else {
			token = tokens[1]
		}
		if !stringInSlice(token, commontokens) {
			delete(Orca_pair_list, pairs)
		}
	}
	//Ordered pairs for updating the prices
	var Orca_orderedpairlist []*OrcaPair
	for _, value := range Orca_pair_list {
		Orca_orderedpairlist = append(Orca_orderedpairlist, value)
	}
	var Aldrin_orderedpairlist []*Aldrin_pairs
	lastTrade := make(map[*Aldrin_pairs]time.Time)
	for _, value := range Aldrin_pair_list {
		lastTrade[value] = time.Time{}
		Aldrin_orderedpairlist = append(Aldrin_orderedpairlist, value)
	}
	fmt.Println(commontokens)
	fmt.Println("ENDING INITIALISATION")
	for {
		update_pools(Orca_orderedpairlist, Aldrin_orderedpairlist, updateClient)
		compute_arbitrage(Aldrin_pair_list, Orca_pair_list, referenceToken.Ammount, reference, privkey, rpcClient, account_list, lastTrade)
	}
}
