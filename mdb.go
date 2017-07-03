package mdb

import (
	"errors"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/hashicorp/go-immutable-radix"
)

// MDB is our memory database
type MDB struct {
	root unsafe.Pointer

	// There can only be a single writer at once
	writer sync.Mutex
}

// Txn is a transaction
type Txn struct {
	db      *MDB
	write   bool
	rootTxn *iradix.Txn
}

// New Database
func New() *MDB {
	return &MDB{
		root: unsafe.Pointer(iradix.New()),
	}
}

func (db *MDB) getRoot() *iradix.Tree {
	root := (*iradix.Tree)(atomic.LoadPointer(&db.root))
	return root
}

// Txn creates a transaction
func (db *MDB) Txn(write bool) *Txn {
	if write {
		db.writer.Lock()
	}

	return &Txn{
		db:      db,
		write:   write,
		rootTxn: db.getRoot().Txn(),
	}
}

// Put a value into our database
func (txn *Txn) Put(k string, v interface{}) error {
	if !txn.write {
		return errors.New("mdb: cannot put in read-only transaction")
	}

	txn.rootTxn.Insert([]byte(k), v)
	return nil
}

// Get a value from our database
func (txn *Txn) Get(k string) interface{} {
	v, _ := txn.rootTxn.Get([]byte(k))
	return v
}

// Delete a value from our database
func (txn *Txn) Delete(k string) error {
	if !txn.write {
		return errors.New("mdb: cannot delete in read-only transaction")
	}

	_, found := txn.rootTxn.Delete([]byte(k))
	if !found {
		return errors.New(`mdb: "` + k + `" wasn't found`)
	}

	return nil
}

// All fn
func (txn *Txn) All(prefix string) *iradix.Iterator {
	it := txn.rootTxn.Root().Iterator()
	it.SeekPrefix([]byte(prefix))
	return it
}

// Commit a transaction
func (txn *Txn) Commit() {
	// Noop for a read transaction
	if !txn.write {
		return
	}

	// Check if already aborted or committed
	if txn.rootTxn == nil {
		return
	}

	newRoot := txn.rootTxn.Commit()
	atomic.StorePointer(&txn.db.root, unsafe.Pointer(newRoot))

	// Clear the txn
	txn.rootTxn = nil

	// Release the writer lock
	txn.db.writer.Unlock()
}

// Abort a transaction
func (txn *Txn) Abort() {
	// Noop for a read transaction
	if !txn.write {
		return
	}

	// Check if already aborted or committed
	if txn.rootTxn == nil {
		return
	}

	// Clear the txn
	txn.rootTxn = nil

	// Release the writer lock
	txn.db.writer.Unlock()
}
