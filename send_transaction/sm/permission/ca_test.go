package test

import (
	"fmt"
	"github.com/taiyuechain/taiyuechain/cim"
	"github.com/taiyuechain/taiyuechain/core/vm"
	"github.com/taiyuechain/taiyuechain/crypto"
	"math/big"
	"os"
	"testing"

	"github.com/taiyuechain/taiyuechain/core"
	"github.com/taiyuechain/taiyuechain/core/state"
	"github.com/taiyuechain/taiyuechain/core/types"
	"github.com/taiyuechain/taiyuechain/log"
)

func init() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(false))))
}

//neo test cacert contract
func TestGrantPermission(t *testing.T) {
	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(number uint64, gen *core.BlockGen, fastChain *core.BlockChain, header *types.Header, statedb *state.StateDB, cimList *cim.CimList) {
		sendTranction(number, gen, statedb, saddr1, saddr2, big.NewInt(6000000000000000000), priKey, signer, nil, header, pbft1Byte, cimList)
		sendTranction(number-1, gen, statedb, saddr2, paddr4, big.NewInt(5000000000000000000), prikey2, signer, nil, header, pbft2Byte, cimList)
		sendGrantPermissionTranscation(number, gen, saddr2, paddr4,new(big.Int).SetInt64(int64(vm.ModifyPerminType_AddSendTxPerm)), prikey2, signer, statedb, fastChain, abiCA, nil, pbft2Byte)

		sendTranction(number-1-25, gen, statedb, paddr4, saddr2, big.NewInt(1000000000000000000), pkey4, signer, nil, header, p2p4Byte, cimList)
		sendRevokePermissionTranscation(number, gen, saddr2, paddr4, prikey2, signer, statedb, fastChain, abiCA, nil, pbft2Byte)
		sendTranction(number-40, gen, statedb, paddr4, saddr2, big.NewInt(1000000000000000000), pkey4, signer, nil, header, p2p4Byte, cimList)
	}
	newTestPOSManager(6, executable)
}

//neo test cacert contract
func TestCreateGroupPermission(t *testing.T) {
	gropAddr := crypto.CreateGroupkey(paddr4, 3)
	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(number uint64, gen *core.BlockGen, fastChain *core.BlockChain, header *types.Header, statedb *state.StateDB, cimList *cim.CimList) {
		sendTranction(number, gen, statedb, saddr1, saddr2, big.NewInt(6000000000000000000), priKey, signer, nil, header, pbft1Byte, cimList)
		sendTranction(number-1, gen, statedb, saddr2, paddr4, big.NewInt(5000000000000000000), prikey2, signer, nil, header, pbft2Byte, cimList)
		sendGrantPermissionTranscation(number, gen, saddr2, paddr4,new(big.Int).SetInt64(int64(vm.ModifyPerminType_AddSendTxPerm)), prikey2, signer, statedb, fastChain, abiCA, nil, pbft2Byte)
		//sendGrantPermissionTranscation(number -1, gen, saddr2, paddr4,new(big.Int).SetInt64(int64(vm.ModifyPerminType_AddGropMemberPerm)), prikey2, signer, statedb, fastChain, abiCA, nil, pbft2Byte)

		sendCreateGroupPermissionTranscation(number, gen, paddr4, "CA", pkey4, signer, statedb, fastChain, abiCA, nil, p2p4Byte)
		sendDelGroupPermissionTranscation(number, gen, paddr4, gropAddr, pkey4, signer, statedb, fastChain, abiCA, nil, p2p4Byte)
	}
	newTestPOSManager(6, executable)
}

func TestGetAddress(t *testing.T) {
	fmt.Println("saddr1", crypto.AddressToHex(saddr1), "saddr2", crypto.AddressToHex(saddr2), "\n saddr3", crypto.AddressToHex(saddr3), "saddr4 ", crypto.AddressToHex(saddr4))
	fmt.Println("paddr1", crypto.AddressToHex(paddr4))
}