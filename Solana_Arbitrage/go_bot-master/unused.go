package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func websocket_pairs(client *ws.Client, pair *pair, amm_key string, openorder_key string, coin_key string, pool_key string) {

	ammid, err := solana.PublicKeyFromBase58(amm_key)
	if err != nil {
		panic(err)
	}
	amm_Open_Orders, err := solana.PublicKeyFromBase58(openorder_key)
	if err != nil {
		panic(err)
	}
	pool_coin, err := solana.PublicKeyFromBase58(coin_key)
	if err != nil {
		panic(err)
	}
	pool_pc, err := solana.PublicKeyFromBase58(pool_key)
	if err != nil {
		panic(err)
	}

	ammid_sub, errsub := client.AccountSubscribe(ammid, rpc.CommitmentFinalized)
	if errsub != nil {
		panic(errsub)
	}
	amm_Open_Orders_sub, errsub := client.AccountSubscribe(amm_Open_Orders, rpc.CommitmentFinalized)
	if errsub != nil {
		panic(errsub)
	}
	pool_coin_sub, errsub := client.AccountSubscribe(pool_coin, rpc.CommitmentFinalized)
	if errsub != nil {
		panic(errsub)
	}
	pool_pc_sub, errsub := client.AccountSubscribe(pool_pc, rpc.CommitmentFinalized)
	if errsub != nil {
		panic(errsub)
	}
	second_pool_coin_sub, errsub := client.AccountSubscribe(pool_coin, rpc.CommitmentFinalized)
	if errsub != nil {
		panic(errsub)
	}
	second_pool_pc_sub, errsub := client.AccountSubscribe(pool_pc, rpc.CommitmentFinalized)
	if errsub != nil {
		panic(errsub)
	}
	third_pool_coin_sub, errsub := client.AccountSubscribe(pool_coin, rpc.CommitmentFinalized)
	if errsub != nil {
		panic(errsub)
	}
	third_pool_pc_sub, errsub := client.AccountSubscribe(pool_pc, rpc.CommitmentFinalized)
	if errsub != nil {
		panic(errsub)
	}
	go get_amm_info(third_pool_coin_sub, third_pool_pc_sub, ammid_sub, pair)
	go get_open_order_info(second_pool_coin_sub, second_pool_pc_sub, amm_Open_Orders_sub, pair)
	go get_double_balance(pool_coin_sub, pool_pc_sub, pair)
	//go get_balance(pool_coin_sub, pair, &pair.Coin_pool_balance)
	//go get_balance(pool_pc_sub, pair, &pair.Pc_pool_balance)
}
func sendtripleTransaction(privateKey solana.PrivateKey, rpcClient *rpc.Client, pair_list map[string]*pair, accounts_list map[string]string, pair1 *pair, reversed1 bool, amountin1 uint64, amoutout1 uint64, pair2 *pair, reversed2 bool, amountin2 uint64, amoutout2 uint64, pair3 *pair, reversed3 bool, amountin3 uint64, amoutout3 uint64) error {

	//Création des comptes associés de la paire 1
	TOKEN_PROGRAM, _ := solana.PublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	tokens1 := strings.Split(pair1.Name, "-")
	var from_account1 solana.PublicKey
	var to_account1 solana.PublicKey
	if reversed1 {
		from_account1, _ = solana.PublicKeyFromBase58(accounts_list[tokens1[0]])
		to_account1, _ = solana.PublicKeyFromBase58(accounts_list[tokens1[1]])
	} else {
		from_account1, _ = solana.PublicKeyFromBase58(accounts_list[tokens1[1]])
		to_account1, _ = solana.PublicKeyFromBase58(accounts_list[tokens1[0]])
	}
	meta_token_program1 := solana.NewAccountMeta(TOKEN_PROGRAM, false, false)
	meta_amm_id1 := solana.NewAccountMeta(pair1.Amm_id, true, false)
	meta_amm_authority1 := solana.NewAccountMeta(pair1.Amm_authority, true, false)
	meta_amm_open_order1 := solana.NewAccountMeta(pair1.Amm_info.AmmOpenOrders, true, false)
	meta_amm_target_order1 := solana.NewAccountMeta(pair1.Amm_info.AmmTargetOrders, true, false)
	meta_poolcoin1 := solana.NewAccountMeta(pair1.Amm_info.PoolCoinTokenAccount, true, false)
	meta_poolpc1 := solana.NewAccountMeta(pair1.Amm_info.PoolPcTokenAccount, true, false)
	meta_serum_program_id1 := solana.NewAccountMeta(pair1.Amm_info.SerumProgramId, false, false)
	meta_serum_market1 := solana.NewAccountMeta(pair1.Amm_info.SerumMarket, true, false)
	meta_serum_bids1 := solana.NewAccountMeta(pair1.SerumBids, true, false)
	meta_serum_asks1 := solana.NewAccountMeta(pair1.serumAsks, true, false)
	meta_serum_event1 := solana.NewAccountMeta(pair1.serumEventQueue, true, false)
	meta_serum_coinvault1 := solana.NewAccountMeta(pair1.serumCoinVaultAccount, true, false)
	meta_serum_pcvault1 := solana.NewAccountMeta(pair1.serumPcVaultAccount, true, false)
	meta_serum_vaultsigner1 := solana.NewAccountMeta(pair1.serumVaultSigner, true, false)
	meta_from_account1 := solana.NewAccountMeta(from_account1, true, false)
	meta_to_account1 := solana.NewAccountMeta(to_account1, true, false)
	meta_wallet1 := solana.NewAccountMeta(privateKey.PublicKey(), true, true)
	var meta_list1 []*solana.AccountMeta
	meta_list1 = append(meta_list1, meta_token_program1, meta_amm_id1, meta_amm_authority1, meta_amm_open_order1, meta_amm_target_order1, meta_poolcoin1, meta_poolpc1, meta_serum_program_id1, meta_serum_market1, meta_serum_bids1, meta_serum_asks1, meta_serum_event1, meta_serum_coinvault1, meta_serum_pcvault1, meta_serum_vaultsigner1, meta_from_account1, meta_to_account1, meta_wallet1)
	//Création des instructions1
	buf1 := new(bytes.Buffer)
	borshEncoder1 := bin.NewBorshEncoder(buf1)
	err := borshEncoder1.Encode(Instruction_data_layout{9, amountin1, amoutout1})
	if err != nil {
		panic(err)
	}
	bytes_data1 := buf1.Bytes()
	instruction1 := solana.NewInstruction(pair1.Program_id, meta_list1, bytes_data1)

	//Création des comptes associés de la paire 2
	tokens2 := strings.Split(pair2.Name, "-")
	var from_account2 solana.PublicKey
	var to_account2 solana.PublicKey
	if reversed2 {
		from_account2, _ = solana.PublicKeyFromBase58(accounts_list[tokens2[0]])
		to_account2, _ = solana.PublicKeyFromBase58(accounts_list[tokens2[1]])
	} else {
		from_account2, _ = solana.PublicKeyFromBase58(accounts_list[tokens2[1]])
		to_account2, _ = solana.PublicKeyFromBase58(accounts_list[tokens2[0]])
	}

	meta_token_program2 := solana.NewAccountMeta(TOKEN_PROGRAM, false, false)
	meta_amm_id2 := solana.NewAccountMeta(pair2.Amm_id, true, false)
	meta_amm_authority2 := solana.NewAccountMeta(pair2.Amm_authority, true, false)
	meta_amm_open_order2 := solana.NewAccountMeta(pair2.Amm_info.AmmOpenOrders, true, false)
	meta_amm_target_order2 := solana.NewAccountMeta(pair2.Amm_info.AmmTargetOrders, true, false)
	meta_poolcoin2 := solana.NewAccountMeta(pair2.Amm_info.PoolCoinTokenAccount, true, false)
	meta_poolpc2 := solana.NewAccountMeta(pair2.Amm_info.PoolPcTokenAccount, true, false)
	meta_serum_program_id2 := solana.NewAccountMeta(pair2.Amm_info.SerumProgramId, false, false)
	meta_serum_market2 := solana.NewAccountMeta(pair2.Amm_info.SerumMarket, true, false)
	meta_serum_bids2 := solana.NewAccountMeta(pair2.SerumBids, true, false)
	meta_serum_asks2 := solana.NewAccountMeta(pair2.serumAsks, true, false)
	meta_serum_event2 := solana.NewAccountMeta(pair2.serumEventQueue, true, false)
	meta_serum_coinvault2 := solana.NewAccountMeta(pair2.serumCoinVaultAccount, true, false)
	meta_serum_pcvault2 := solana.NewAccountMeta(pair2.serumPcVaultAccount, true, false)
	meta_serum_vaultsigner2 := solana.NewAccountMeta(pair2.serumVaultSigner, true, false)
	meta_from_account2 := solana.NewAccountMeta(from_account2, true, false)
	meta_to_account2 := solana.NewAccountMeta(to_account2, true, false)
	meta_wallet2 := solana.NewAccountMeta(privateKey.PublicKey(), true, true)

	var meta_list2 []*solana.AccountMeta
	meta_list2 = append(meta_list2, meta_token_program2, meta_amm_id2, meta_amm_authority2, meta_amm_open_order2, meta_amm_target_order2, meta_poolcoin2, meta_poolpc2, meta_serum_program_id2, meta_serum_market2, meta_serum_bids2, meta_serum_asks2, meta_serum_event2, meta_serum_coinvault2, meta_serum_pcvault2, meta_serum_vaultsigner2, meta_from_account2, meta_to_account2, meta_wallet2)
	//Création des instructions2
	buf2 := new(bytes.Buffer)
	borshEncoder2 := bin.NewBorshEncoder(buf2)
	err = borshEncoder2.Encode(Instruction_data_layout{9, amountin2, amoutout2})
	if err != nil {
		return err
	}
	bytes_data2 := buf2.Bytes()
	instruction2 := solana.NewInstruction(pair2.Program_id, meta_list2, bytes_data2)

	//Création des comptes associés de la paire 3
	tokens3 := strings.Split(pair3.Name, "-")
	var from_account3 solana.PublicKey
	var to_account3 solana.PublicKey
	if reversed3 {
		from_account3, _ = solana.PublicKeyFromBase58(accounts_list[tokens3[0]])
		to_account3, _ = solana.PublicKeyFromBase58(accounts_list[tokens3[1]])
	} else {
		from_account3, _ = solana.PublicKeyFromBase58(accounts_list[tokens3[1]])
		to_account3, _ = solana.PublicKeyFromBase58(accounts_list[tokens3[0]])
	}

	meta_token_program3 := solana.NewAccountMeta(TOKEN_PROGRAM, false, false)
	meta_amm_id3 := solana.NewAccountMeta(pair3.Amm_id, true, false)
	meta_amm_authority3 := solana.NewAccountMeta(pair3.Amm_authority, true, false)
	meta_amm_open_order3 := solana.NewAccountMeta(pair3.Amm_info.AmmOpenOrders, true, false)
	meta_amm_target_order3 := solana.NewAccountMeta(pair3.Amm_info.AmmTargetOrders, true, false)
	meta_poolcoin3 := solana.NewAccountMeta(pair3.Amm_info.PoolCoinTokenAccount, true, false)
	meta_poolpc3 := solana.NewAccountMeta(pair3.Amm_info.PoolPcTokenAccount, true, false)
	meta_serum_program_id3 := solana.NewAccountMeta(pair3.Amm_info.SerumProgramId, false, false)
	meta_serum_market3 := solana.NewAccountMeta(pair3.Amm_info.SerumMarket, true, false)
	meta_serum_bids3 := solana.NewAccountMeta(pair3.SerumBids, true, false)
	meta_serum_asks3 := solana.NewAccountMeta(pair3.serumAsks, true, false)
	meta_serum_event3 := solana.NewAccountMeta(pair3.serumEventQueue, true, false)
	meta_serum_coinvault3 := solana.NewAccountMeta(pair3.serumCoinVaultAccount, true, false)
	meta_serum_pcvault3 := solana.NewAccountMeta(pair3.serumPcVaultAccount, true, false)
	meta_serum_vaultsigner3 := solana.NewAccountMeta(pair3.serumVaultSigner, true, false)
	meta_from_account3 := solana.NewAccountMeta(from_account3, true, false)
	meta_to_account3 := solana.NewAccountMeta(to_account3, true, false)
	meta_wallet3 := solana.NewAccountMeta(privateKey.PublicKey(), true, true)

	var meta_list3 []*solana.AccountMeta
	meta_list3 = append(meta_list3, meta_token_program3, meta_amm_id3, meta_amm_authority3, meta_amm_open_order3, meta_amm_target_order3, meta_poolcoin3, meta_poolpc3, meta_serum_program_id3, meta_serum_market3, meta_serum_bids3, meta_serum_asks3, meta_serum_event3, meta_serum_coinvault3, meta_serum_pcvault3, meta_serum_vaultsigner3, meta_from_account3, meta_to_account3, meta_wallet3)
	//Création des instructions3
	buf3 := new(bytes.Buffer)
	borshEncoder3 := bin.NewBorshEncoder(buf3)
	err = borshEncoder3.Encode(Instruction_data_layout{9, amountin3, amoutout3})
	if err != nil {
		panic(err)
	}
	bytes_data3 := buf3.Bytes()
	instruction3 := solana.NewInstruction(pair3.Program_id, meta_list3, bytes_data3)

	//Construction de la transaction
	transac := solana.NewTransactionBuilder()
	transac = transac.AddInstruction(instruction1)
	transac = transac.AddInstruction(instruction2)
	transac = transac.AddInstruction(instruction3)
	transac = transac.SetFeePayer(privateKey.PublicKey())
	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
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
	if err != nil {
		return err
	}
	sig, err := rpcClient.SendTransactionWithOpts(context.TODO(), final_transac, false, rpc.CommitmentProcessed) //envois de la transaction
	if err != nil {
		return err
	}
	fmt.Println(sig)
	return nil
}
