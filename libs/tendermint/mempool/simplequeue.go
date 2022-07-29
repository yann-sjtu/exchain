package mempool

import (
	"crypto/sha256"
	"sync"

	"github.com/okex/exchain/libs/tendermint/libs/clist"
	"github.com/okex/exchain/libs/tendermint/types"
)

type SimpleTxQueue struct {
	txs    *clist.CList // FIFO list
	txsMap sync.Map     //txKey -> CElement

	mockAddressRecord
}

func NewSimpleTxQueue() *SimpleTxQueue {
	return &SimpleTxQueue{
		txs: clist.New(),
	}
}

func (q *SimpleTxQueue) Len() int {
	return q.txs.Len()
}

func (q *SimpleTxQueue) Insert(tx *mempoolTx) error {
	/*
		1. insert tx list
		2. insert address record
		3. insert tx map
	*/
	ele := q.txs.PushBack(tx)
	//ele.Address = tx.from
	//ele.Nonce = tx.realTx.GetNonce()

	q.txsMap.Store(txKey(ele.Value.(*mempoolTx).tx), ele)
	return nil
}

func (q *SimpleTxQueue) Remove(element *clist.CElement) {
	q.removeElement(element)
}

func (q *SimpleTxQueue) RemoveByKey(key [sha256.Size]byte) *clist.CElement {
	return q.removeElementByKey(key)

}

func (q *SimpleTxQueue) Front() *clist.CElement {
	return q.txs.Front()
}

func (q *SimpleTxQueue) Back() *clist.CElement {
	return q.txs.Back()
}

func (q *SimpleTxQueue) BroadcastFront() *clist.CElement {
	return q.txs.Front()
}

func (q *SimpleTxQueue) BroadcastLen() int {
	return q.txs.Len()
}

func (q *SimpleTxQueue) TxsWaitChan() <-chan struct{} {
	return q.txs.WaitChan()
}

func (q *SimpleTxQueue) Load(hash [sha256.Size]byte) (*clist.CElement, bool) {
	v, ok := q.txsMap.Load(hash)
	if !ok {
		return nil, false
	}
	return v.(*clist.CElement), true
}

func (q *SimpleTxQueue) removeElement(element *clist.CElement) {
	q.txs.Remove(element)
	element.DetachPrev()

	tx := element.Value.(*mempoolTx).tx
	txHash := txKey(tx)
	q.txsMap.Delete(txHash)
}

func (q *SimpleTxQueue) removeElementByKey(key [32]byte) *clist.CElement {
	if v, ok := q.txsMap.LoadAndDelete(key); ok {
		element := v.(*clist.CElement)
		q.txs.Remove(element)
		element.DetachPrev()
		return element
	}
	return nil
}

func (q *SimpleTxQueue) CleanItems(address string, nonce uint64) {}

type mockAddressRecord struct{}

func (m mockAddressRecord) GetAddressList() []string {
	return nil
}

func (m mockAddressRecord) GetAddressNonce(address string) (uint64, bool) {
	return 0, false
}

func (m mockAddressRecord) GetAddressTxsCnt(address string) int {
	return 0
}

func (m mockAddressRecord) GetAddressTxs(address string, max int) types.Txs {
	return nil
}

func (m mockAddressRecord) CleanItems(address string, nonce uint64) {}
