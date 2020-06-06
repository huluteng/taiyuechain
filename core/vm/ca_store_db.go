package vm

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/taiyuechain/taiyuechain/common"
	"github.com/taiyuechain/taiyuechain/consensus/tbft/help"
	"github.com/taiyuechain/taiyuechain/core/types"
	"github.com/taiyuechain/taiyuechain/log"
	"github.com/taiyuechain/taiyuechain/rlp"
	"io"
	"math/big"
)

func (ca *CACertList) LoadCACertList(state StateDB, preAddress common.Address) error {
	key := common.BytesToHash(preAddress[:])
	data := state.GetCAState(preAddress, key)
	lenght := len(data)
	if lenght == 0 {
		return errors.New("Load data = 0")
	}
	hash := types.RlpHash(data)
	var temp CACertList
	watch1 := help.NewTWatch(0.005, "Load impawn")
	if cc, ok := CASC.Cache.Get(hash); ok {
		caList := cc.(*CACertList)
		temp = *(CloneCaCache(caList))
	} else {
		if err := rlp.DecodeBytes(data, &temp); err != nil {
			watch1.EndWatch()
			watch1.Finish("DecodeBytes")
			log.Error(" Invalid CACertList entry RLP", "err", err)
			return errors.New(fmt.Sprintf("Invalid CACertList entry RLP %s", err.Error()))
		}
		//tmp := CloneCaCache(&temp)
		//
		//if tmp != nil {
		//	CASC.Cache.Add(hash, tmp)
		//}
	}

	for k, val := range temp.caCertMap {
		items := &CACert{
			val.CACert,
			val.IsStore,
		}
		ca.caCertMap[k] = items
	}

	for k, val := range temp.proposalMap {
		//log.Info("--clone 2","k",k,"val",val.CACert)
		item := &ProposalState{
			val.PHash,
			val.CACert,
			val.StartHeight,
			val.EndHeight,
			val.PState,
			val.NeedPconfirmNumber,
			val.PNeedDo,
			val.SignList,
			val.SignMap,
		}

		ca.proposalMap[k] = item
	}
	watch1.EndWatch()
	watch1.Finish("DecodeBytes")
	return nil
}

func (ca *CACertList) SaveCACertList(state StateDB, preAddress common.Address) error {
	key := common.BytesToHash(preAddress[:])
	watch1 := help.NewTWatch(0.005, "Save impawn")
	data, err := rlp.EncodeToBytes(ca)
	watch1.EndWatch()
	watch1.Finish("EncodeToBytes")

	if err != nil {
		log.Crit("Failed to RLP encode CACertList", "err", err)
	}
	for _, val := range ca.proposalMap {
		log.Info("save CA info", "Ce name", hex.EncodeToString(val.CACert), "is store", val.PHash, "caCertMap", len(ca.caCertMap), "ca", ca.caCertMap)
	}
	state.SetCAState(preAddress, key, data)

	//hash := types.RlpHash(data)
	//tmp := CloneCaCache(ca)
	//if tmp != nil {
	//	CASC.Cache.Add(hash, tmp)
	//}
	return err
}

// "external" CACertList encoding. used for pos staking.
type extCACertList struct {
	CACerts       []*CACert
	CAArray       []uint64
	Proposals     []*ProposalState
	ProposalArray []common.Hash
}

func (i *CACertList) DecodeRLP(s *rlp.Stream) error {
	var ei extCACertList
	if err := s.Decode(&ei); err != nil {
		return err
	}
	certs := make(map[uint64]*CACert)
	for i, cert := range ei.CACerts {
		certs[ei.CAArray[i]] = cert
	}
	proposals := make(map[common.Hash]*ProposalState)
	for i, proposal := range ei.Proposals {
		proposals[ei.ProposalArray[i]] = proposal
	}

	i.caCertMap, i.proposalMap = certs, proposals
	return nil
}

// EncodeRLP serializes b into the truechain RLP ImpawnImpl format.
func (i *CACertList) EncodeRLP(w io.Writer) error {
	var certs []*CACert
	var order []uint64
	for i, _ := range i.caCertMap {
		order = append(order, i)
	}
	for m := 0; m < len(order)-1; m++ {
		for n := 0; n < len(order)-1-m; n++ {
			if order[n] > order[n+1] {
				order[n], order[n+1] = order[n+1], order[n]
			}
		}
	}
	for _, index := range order {
		certs = append(certs, i.caCertMap[index])
	}

	var proposals []*ProposalState
	var proposalOrders []common.Hash
	for i, _ := range i.proposalMap {
		proposalOrders = append(proposalOrders, i)
	}
	for m := 0; m < len(proposalOrders)-1; m++ {
		for n := 0; n < len(proposalOrders)-1-m; n++ {
			if proposalOrders[n].Big().Cmp(proposalOrders[n+1].Big()) > 0 {
				proposalOrders[n], proposalOrders[n+1] = proposalOrders[n+1], proposalOrders[n]
			}
		}
	}
	for _, index := range proposalOrders {
		proposals = append(proposals, i.proposalMap[index])
	}
	return rlp.Encode(w, extCACertList{
		CACerts:       certs,
		CAArray:       order,
		Proposals:     proposals,
		ProposalArray: proposalOrders,
	})
}

// "external" ProposalState encoding. used for pos staking.
type extProposalState struct {
	PHash              common.Hash
	CACert             []byte
	StartHight         *big.Int
	EndHight           *big.Int
	PState             uint8
	NeedPconfirmNumber uint64 // muti need confir len
	PNeedDo            uint8  // only supprot add and del
	SignList           []common.Hash
	SignMap            []bool
	SignArray          []common.Hash
}

func (i *ProposalState) DecodeRLP(s *rlp.Stream) error {
	var ei extProposalState
	if err := s.Decode(&ei); err != nil {
		return err
	}
	proposals := make(map[common.Hash]bool)
	for i, proposal := range ei.SignMap {
		proposals[ei.SignArray[i]] = proposal
	}

	i.SignMap, i.PHash, i.CACert, i.StartHeight, i.EndHeight, i.PState, i.NeedPconfirmNumber, i.PNeedDo, i.SignList =
		proposals, ei.PHash, ei.CACert, ei.StartHight, ei.EndHight, ei.PState, ei.NeedPconfirmNumber, ei.PNeedDo, ei.SignList
	return nil
}

// EncodeRLP serializes b into the truechain RLP ImpawnImpl format.
func (i *ProposalState) EncodeRLP(w io.Writer) error {
	var proposals []bool
	var proposalOrders []common.Hash
	for i, _ := range i.SignMap {
		proposalOrders = append(proposalOrders, i)
	}
	for m := 0; m < len(proposalOrders)-1; m++ {
		for n := 0; n < len(proposalOrders)-1-m; n++ {
			if proposalOrders[n].Big().Cmp(proposalOrders[n+1].Big()) > 0 {
				proposalOrders[n], proposalOrders[n+1] = proposalOrders[n+1], proposalOrders[n]
			}
		}
	}
	for _, index := range proposalOrders {
		proposals = append(proposals, i.SignMap[index])
	}
	return rlp.Encode(w, extProposalState{
		PHash:              i.PHash,
		CACert:             i.CACert,
		StartHight:         i.StartHeight,
		EndHight:           i.EndHeight,
		PState:             i.PState,
		NeedPconfirmNumber: i.NeedPconfirmNumber,
		PNeedDo:            i.PNeedDo,
		SignList:           i.SignList,
		SignMap:            proposals,
		SignArray:          proposalOrders,
	})
}
