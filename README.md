# Simple High-Performance Local Cache

Implementation of various local Cache and test it benchmark performance

following [Tour](https://github.com/go-programming-tour-book/cache-example).

Based on them, reimplement a high-performance Cache according to BigCache, Considering  sharding Lock and GC optimization





### Intro

**SCache**

Simple Cache. It implement three kind of cache strategies, `LRU` ; `LFU` ; `FIFO`. However, it use only one Lock to ensure the thread-safe. Geting Lock and releasing Lock will cost many performance overhead 



**FCache**

Fast Cache. It optimize the performance through sharding records. Thus, two records from different shards won't share one lock. It improves the performance in Parallel Operation.



**AmeCache**

copy the design pattern of bigCache. Considering the GC optimization





### Benchmark Test

```shell
$ cd test/benchmark
$ go test -bench=. -benchmem ./... -timeout 30m
```



