# storage
Storage interface by golang

## AppendOnlyStore
Default implement is MemoryAppendOnlyStore, you can also implement it by mongodb etc.

### Usage

    import (
        "github.com/berkaroad/storage"
    )
    
    var store = &storage.MemoryAppendOnlyStore{}
    store.InitFunc()()
    // append the first record, so expected version is zero.
    store.Append("ProductCatalog_0001", []byte("{\"Id\":\"0001\", \"Name\":\"Pear\"}"), 0)
    // read by stream name
    store.ReadRecords("ProductCatalog_0001", 0, 100000)

You can also use it with ioc

    import (
        "github.com/berkaroad/ioc"
        "github.com/berkaroad/storage"
    )

    var container = ioc.NewContainer()

    func main() {
        container.RegisterTo(&storage.MemoryAppendOnlyStore{}, (*storage.AppendOnlyStore)(nil), ioc.Singleton)
        container.Invoke(function(store storage.AppendOnlyStore){
            // append the first record, so expected version is zero.
            store.Append("ProductCatalog_0001", []byte("{\"Id\":\"0001\", \"Name\":\"Pear\"}"), 0)
            // read by stream name
            store.ReadRecords("ProductCatalog_0001", 0, 100000)
        })
    }