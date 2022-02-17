package mempool

import (
	"sync"

	"github.com/okex/exchain/libs/tendermint/libs/clist"
	tmmath "github.com/okex/exchain/libs/tendermint/libs/math"
	"github.com/okex/exchain/libs/tendermint/types"
)

type AddressRecord2 struct {
	m sync.Map // address -> map[txHash]*addrMap
}

type addrMap struct {
	sync.RWMutex

	items map[string]*clist.CElement
	minNonce uint64
}

func newAddressRecord2() *AddressRecord2 {
	return &AddressRecord2{}
}

func (ar *AddressRecord2) AddItem(address string, txHash string, cElement *clist.CElement) {
	v, ok := ar.m.Load(address)
	if !ok {
		am := &addrMap{items: make(map[string]*clist.CElement)}
		am.items[txHash] = cElement
		ar.m.Store(address, am)
	}
	am := v.(*addrMap)
	am.Lock()
	defer am.Unlock()
	am.items[txHash] = cElement
}

func (ar *AddressRecord2) checkRepeatedElement(info ExTxInfo) (ele *clist.CElement, valid bool) {
	v, ok := ar.m.Load(info.Sender)
	if !ok {
		return nil, true
	}
	am := v.(*addrMap)
	am.Lock()
	defer am.Unlock()
	if info.Nonce < am.minNonce {
		return nil, false
	}
	for _, ele = range am.items {
		if ele.Nonce == info.Nonce {
			return ele, true
		}
	}
	return nil, true
}

func (ar *AddressRecord2) CleanItem(address string, nonce uint64) {
	v, ok := ar.m.Load(address)
	if !ok {
		return
	}
	am := v.(*addrMap)
	am.Lock()
	defer am.Unlock()
	if nonce < am.minNonce {
		return
	}
	am.minNonce = nonce+1
	for k, v := range am.items {
		if v.Nonce <= nonce {
			delete(am.items, k)
		}
	}
}

func (ar *AddressRecord2) GetItem(address string) (map[string]*clist.CElement, bool) {
	v, ok := ar.m.Load(address)
	if !ok {
		return nil, false
	}
	am := v.(*addrMap)
	m := make(map[string]*clist.CElement)
	am.RLock()
	defer am.RUnlock()
	for k, v := range am.items {
		m[k] = v
	}
	return m, true
}

func (ar *AddressRecord2) DeleteItem(e *clist.CElement) {
	if v, ok := ar.m.Load(e.Address); ok {
		am := v.(*addrMap)
		memTx := e.Value.(*mempoolTx)
		txHash := txID(memTx.tx, memTx.height)
		am.Lock()
		defer am.Unlock()
		delete(am.items, txHash)
		if len(am.items) == 0 {
			ar.m.Delete(e.Address)
		}
	}
}

func (ar *AddressRecord2) GetAddressList() []string {
	var addrList []string
	ar.m.Range(func(k, v interface{}) bool {
		addrList = append(addrList, k.(string))
		return true
	})
	return addrList
}

func (ar *AddressRecord2) GetAddressTxsCnt(address string) int {
	v, ok := ar.m.Load(address)
	if !ok {
		return 0
	}
	am := v.(*addrMap)
	am.RLock()
	defer am.RUnlock()
	return len(am.items)
}

func (ar *AddressRecord2) GetAddressNonce(address string) uint64 {
	v, ok := ar.m.Load(address)
	if !ok {
		return 0
	}
	am := v.(*addrMap)
	am.RLock()
	defer am.RUnlock()
	var nonce uint64
	for _, e := range am.items {
		if e.Nonce > nonce {
			nonce = e.Nonce
		}
	}
	return nonce
}

func (ar *AddressRecord2) GetAddressTxs(address string, txCount int, max int) types.Txs {
	v, ok := ar.m.Load(address)
	if !ok {
		return nil
	}
	am := v.(*addrMap)
	am.RLock()
	defer am.RUnlock()
	if max <= 0 || max > len(am.items) {
		max = len(am.items)
	}
	txs := make([]types.Tx, 0, tmmath.MinInt(txCount, max))
	for _, e := range am.items {
		if len(txs) == cap(txs) {
			break
		}
		txs = append(txs, e.Value.(*mempoolTx).tx)
	}
	return txs
}

