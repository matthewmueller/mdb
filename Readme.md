# MDB

In-memory key/value store designed for concurrent use. A simpler, but less feature-rich version of [go-memdb](https://github.com/hashicorp/go-memdb).

## Features (as compared to go-memdb)

✅ Multi-Version Concurrency Control (MVCC) - By leveraging immutable radix trees the database is able to support any number of concurrent readers without locking, and allows a writer to make progress.

✅ Transaction Support - The database allows for rich transactions, in which multiple objects are inserted, updated or deleted. The transactions can span multiple tables, and are applied atomically. The database provides atomicity and isolation in ACID terminology, such that until commit the updates are not visible.

🚫 Rich Indexing - Tables can support any number of indexes, which can be simple like a single field index, or more advanced compound field indexes. Certain types like UUID can be efficiently compressed from strings into byte indexes for reduced storage requirements.

🚫 Watches - Callers can populate a watch set as part of a query, which can be used to detect when a modification has been made to the database which affects the query results. This lets callers easily watch for changes in the database in a very general way.

## For the curious: How the MVCC part works

The benefit of an MVCC system is that reads don't block writes and writes don't block reads. This is unlike [mutex.RWLock](), where read locks block write locks and write locks block read locks.

The way this works is by using an immutable data structure underneath. In the case of the library and go-memdb, it's using an [immutable radix tree](https://github.com/hashicorp/go-immutable-radix). This allows readers to use a snapshot, while a writer's transaction is taking place. The compromise here is that a reader may not have the most up-to-date data.

## Example

```go
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
```

## License

MIT
