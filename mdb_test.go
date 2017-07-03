package mdb_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/matthewmueller/mdb"
	"github.com/stretchr/testify/assert"
)

func TestPutAbort(t *testing.T) {
	db := mdb.New()
	txn := db.Txn(true)

	v := txn.Get("a")
	assert.Nil(t, v)

	if e := txn.Put("a", "b"); e != nil {
		t.Fatal(e)
	}
	assert.Equal(t, "b", txn.Get("a").(string))

	txn.Abort()
	txn = db.Txn(false)
	assert.Nil(t, txn.Get("a"))
}

func TestPutCommit(t *testing.T) {
	db := mdb.New()
	txn := db.Txn(true)

	v := txn.Get("a")
	assert.Nil(t, v)

	if e := txn.Put("a", "b"); e != nil {
		t.Fatal(e)
	}
	assert.Equal(t, "b", txn.Get("a").(string))

	txn.Commit()
	txn = db.Txn(false)
	assert.Equal(t, "b", txn.Get("a").(string))
}

func TestDeleteAbort(t *testing.T) {
	db := mdb.New()
	txn := db.Txn(true)
	txn.Put("a", "a")
	txn.Commit()
	txn = db.Txn(true)
	assert.Equal(t, "a", txn.Get("a").(string))
	txn.Delete("a")
	assert.Nil(t, txn.Get("a"))
	txn.Abort()
	txn = db.Txn(false)
	assert.Equal(t, "a", txn.Get("a").(string))
}

func TestDeleteCommit(t *testing.T) {
	db := mdb.New()
	txn := db.Txn(true)
	txn.Put("a", "a")
	txn.Commit()
	txn = db.Txn(true)
	assert.Equal(t, "a", txn.Get("a").(string))
	txn.Delete("a")
	assert.Nil(t, txn.Get("a"))
	txn.Commit()
	txn = db.Txn(false)
	assert.Nil(t, txn.Get("a"))
}

func TestWriteSafety(t *testing.T) {
	db := mdb.New()
	txn := db.Txn(true)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		t := db.Txn(true)
		defer t.Commit()
		i := t.Get("lease").(int)
		t.Put("lease", i+1)
		wg.Done()
	}()

	go func() {
		t := db.Txn(true)
		defer t.Commit()
		i := t.Get("lease").(int)
		t.Put("lease", i+1)
		wg.Done()
	}()

	time.Sleep(100 * time.Millisecond)
	txn.Put("lease", 1)
	txn.Commit()
	wg.Wait()
	txn = db.Txn(false)
	assert.Equal(t, 3, txn.Get("lease").(int))
}

func TestReadConcurrency(t *testing.T) {
	db := mdb.New()
	txn := db.Txn(true)
	txn.Put("lease", 1)
	txn.Commit()

	txn = db.Txn(true)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		tx := db.Txn(false)
		i := tx.Get("lease").(int)
		assert.Equal(t, 1, i)
		wg.Done()
	}()

	go func() {
		tx := db.Txn(false)
		i := tx.Get("lease").(int)
		assert.Equal(t, 1, i)
		wg.Done()
	}()

	time.Sleep(100 * time.Millisecond)
	i := txn.Get("lease").(int)
	txn.Put("lease", i+1)
	// this would freeze up if reads
	// couldn't progress during writes
	wg.Wait()
	txn.Commit()

	txn = db.Txn(false)
	assert.Equal(t, 2, txn.Get("lease").(int))
}

func TestIterator(t *testing.T) {
	db := mdb.New()
	txn := db.Txn(true)
	txn.Put("a.1", "a.1")
	txn.Put("a.2", "a.2")
	txn.Put("c.1", "c.1")
	txn.Commit()

	txn = db.Txn(false)
	matches := txn.All("")
	vals := []string{}

	for _, match := range matches {
		vals = append(vals, match.(string))
	}

	assert.Equal(t, "a.1,a.2,c.1", strings.Join(vals, ","))
}

func TestIteratorWithPrefix(t *testing.T) {
	db := mdb.New()
	txn := db.Txn(true)
	txn.Put("a.1", "a.1")
	txn.Put("a.2", "a.2")
	txn.Put("c.1", "c.1")
	txn.Commit()

	txn = db.Txn(false)
	matches := txn.All("a")
	vals := []string{}

	for _, match := range matches {
		vals = append(vals, match.(string))
	}

	assert.Equal(t, "a.1,a.2", strings.Join(vals, ","))
}
