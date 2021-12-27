package types

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/ethereum/go-ethereum/ethdb"
	"os"
	"sync"
)

func init() {
	dbCreator := func(name string, dir string) (ethdb.KeyValueStore, error) {
		fmt.Println("badger:dbCreator")
		return NewWrapBadgerDB(dir), nil
	}
	registerDBCreator(BadgerDBBackend, dbCreator, false)
}

type WrapBadgerDB struct {
	*BadgerDB
}

func (w WrapBadgerDB) Stat(property string) (string, error) {
	//TODO implement me
	panic("implement me stat")
}

func (w WrapBadgerDB) Compact(start []byte, limit []byte) error {
	//TODO implement me
	panic("implement me compact")
}

func NewWrapBadgerDB(path string) ethdb.KeyValueStore {
	db, _ := NewBadgerDB(&Config{
		DataDir: path,
	})
	return &WrapBadgerDB{
		db,
	}

}

// BadgerDB contains directory path to data and db instance
type BadgerDB struct {
	config *Config
	db     *badger.DB
	lock   sync.RWMutex
}

type KVList = badger.KVList

// Config defines configurations for BadgerDB instance
type Config struct {
	DataDir  string
	InMemory bool
	Compress bool
}

// NewBadgerDB initializes badgerDB instance
func NewBadgerDB(cfg *Config) (*BadgerDB, error) {
	opts := badger.DefaultOptions(cfg.DataDir)
	opts.ValueDir = cfg.DataDir
	opts.Logger = nil
	opts.WithSyncWrites(false)
	opts.WithNumCompactors(20)
	// opts.WithBlockCacheSize(1 << 16) // TODO: add caching
	//opts.WithInMemory(cfg.InMemory)
	//
	//if cfg.Compress {
	//	opts.WithCompression(options.Snappy)
	//}
	if err := os.MkdirAll(cfg.DataDir, os.ModePerm); err != nil {
		return nil, err
	}
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerDB{
		config: cfg,
		db:     db,
	}, nil
}

// Path returns the path to the database directory.
func (db *BadgerDB) Path() string {
	return db.config.DataDir
}

// Batch struct contains a database instance, key-value mapping for batch writes and length of item value for batch write
type batchWriter struct {
	db      *BadgerDB
	updates map[string][]byte
	deletes map[string]struct{}
	size    int
	lock    sync.RWMutex
}

func (b *batchWriter) Replay(w ethdb.KeyValueWriter) error {
	for k, v := range b.updates {
		if err := w.Put([]byte(k), v); err != nil {
			return err
		}
	}
	for k := range b.deletes {
		if err := w.Delete([]byte(k)); err != nil {
			return err
		}
	}
	return nil
}

// NewBatch returns batchWriter with a badgerDB instance and an initialized mapping
func (db *BadgerDB) NewBatch() ethdb.Batch {
	return &batchWriter{
		db:      db,
		updates: make(map[string][]byte),
		deletes: make(map[string]struct{}),
	}
}

// Put puts the given key / value to the queue
func (db *BadgerDB) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	return db.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

// Has checks the given key exists already; returning true or false
func (db *BadgerDB) Has(key []byte) (exists bool, err error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	err = db.db.View(func(txn *badger.Txn) error {
		item, errr := txn.Get(key)
		if item != nil {
			exists = true
		}
		if errr == badger.ErrKeyNotFound {
			exists = false
			errr = nil
		}
		return errr
	})
	return exists, err
}

// Get returns the given key
func (db *BadgerDB) Get(key []byte) (data []byte, err error) {
	fmt.Println("145-----")
	db.lock.RLock()
	defer db.lock.RUnlock()
	err = db.db.View(func(txn *badger.Txn) error {
		item, e := txn.Get(key)
		if e != nil {
			return e
		}
		data, e = item.ValueCopy(nil)
		if e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Del removes the key from the queue and database
func (db *BadgerDB) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	return db.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		if err == badger.ErrKeyNotFound {
			err = nil
		}
		return err
	})
}

// Flush commits pending writes to disk
func (db *BadgerDB) Flush() error {
	return db.db.Sync()
}

// Close closes a DB
func (db *BadgerDB) Close() error {
	db.lock.Lock()
	defer db.lock.Unlock()
	if err := db.db.Close(); err != nil {
		return err
	}
	return nil
}

// ClearAll would delete all the data stored in DB.
func (db *BadgerDB) ClearAll() error {
	return db.db.DropAll()
}

// Subscribe to watch for changes for the given prefixes
func (db *BadgerDB) Subscribe(ctx context.Context, cb func(kv *KVList) error, prefixes []byte) error {
	return db.db.Subscribe(ctx, cb, prefixes)
}

// BadgerIterator struct contains a transaction, iterator and init.
type BadgerIterator struct {
	txn  *badger.Txn
	iter *badger.Iterator
	init bool
	lock sync.RWMutex
	err  error // new
}

func (i *BadgerIterator) Error() error {
	//TODO implement me
	panic("implement me Error")
}

// NewIterator returns a new iterator within the Iterator struct along with a new transaction
func (db *BadgerDB) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	txn := db.db.NewTransaction(false)
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	//iter := txn.NewIterator(opts)
	iter := txn.NewKeyIterator(start, opts)

	return &BadgerIterator{
		txn:  txn,
		iter: iter,
	}
}

// Release closes the iterator and discards the created transaction.
func (i *BadgerIterator) Release() {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.iter.Close()
	i.txn.Discard()
}

// Next rewinds the iterator to the zero-th position if uninitialized, and then will advance the iterator by one
// returns bool to ensure access to the item
func (i *BadgerIterator) Next() bool {
	i.lock.RLock()
	defer i.lock.RUnlock()

	if !i.init {
		i.iter.Rewind()
		i.init = true
		return i.iter.Valid()
	}

	if !i.iter.Valid() {
		return false
	}
	i.iter.Next()
	return i.iter.Valid()
}

// Seek will look for the provided key if present and go to that position. If
// absent, it would seek to the next smallest key
func (i *BadgerIterator) Seek(key []byte) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	i.iter.Seek(key)
}

// Key returns an item key
func (i *BadgerIterator) Key() []byte {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.iter.Item().Key()
}

// Value returns a copy of the value of the item
func (i *BadgerIterator) Value() []byte {
	i.lock.RLock()
	defer i.lock.RUnlock()
	val, err := i.iter.Item().ValueCopy(nil)
	if err != nil {
		i.err = err
	}
	return val
}

// Put encodes key-values and adds them to a mapping for batch writes, sets the size of item value
func (b *batchWriter) Put(key, value []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	strKey := string(key)
	b.updates[strKey] = value
	b.size += len(value)
	delete(b.deletes, strKey)
	return nil
}

// Del removes the key from the batch and database
func (b *batchWriter) Delete(key []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	strKey := string(key)
	b.deletes[strKey] = struct{}{}
	delete(b.updates, strKey)
	return nil
}

// Flush commits pending writes to disk
func (b *batchWriter) Write() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	wb := b.db.db.NewWriteBatch()
	defer wb.Cancel()

	for k, v := range b.updates {
		if err := wb.Set([]byte(k), v); err != nil {
			return err
		}
	}

	for k := range b.deletes {
		if err := wb.Delete([]byte(k)); err != nil {
			return err
		}
	}
	return wb.Flush()
}

// ValueSize returns the amount of data in the batch
func (b *batchWriter) ValueSize() int {
	return b.size
}

// Reset clears batch key-values and resets the size to zero
func (b *batchWriter) Reset() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.updates = make(map[string][]byte)
	b.deletes = make(map[string]struct{})
	b.size = 0
}