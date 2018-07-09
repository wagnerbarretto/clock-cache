# clock-cache

Package `clock` implements a fixed size cache using the clock algorithm for evictions. The cache is essentially a `map[interface{}]interface{}` with fixed size. Clock algorithm enables concurrent lookups by default while maintaining near LRU recency characteristics. No garbage collection avoidance tactics are utilized.

### Clock vs. LRU

LRU implementations commonly make use of linked lists to keep track of access recency. Every time an object is retrieved, its reference is moved to the head of the list. Moving objects in a linked list is not a thread safe operation and thus have to be serialized. LRU implementations traditionally employ techniques such as sharding and promotion buffering to improve concurrency.

Clock implementations utilize a ring buffer whose items don't need to be moved during lookups. This enables a simple implementation that can provide concurrent reads. Using a clock cache can make sense in applications that want to optimize read concurrency but sharding is overkill. An important drawback of clock is that worst case insert performance is lower than LRU, thus a high cache-hit ratio is fundamental.



## Basic Usage

```go
// create a new cache that can hold 128 entries
c := clock.New(128)

// add an entry
c.Put("k1", 1234)

// look up for a key
val, ok := c.Get("k1")
if ok {
  fmt.Println(val)
}

// remove an entry
c.Delete("k1")
```

## Optional Parameters

A ClockCache instance can receive two optional parameters: a *clock count* and an eviction callback.

* The clock count parameter refers to the *Generalized Clock Algorithm*. It configures how many times the clock hand needs to pass over an entry for it to be evicted. You can find more about the generalized clock algorithm in [this paper](http://www-inst.eecs.berkeley.edu/~cs266/sp10/readings/smith78.pdf) (page 11). Increasing this parameter makes the cache prioritize *frequently* accessed entries over *recently* accessed entries. Properly setting this parameter requires knowledge about the workload and experimentation. Be aware that setting this parameter over the default value of 1 can decrease insert performance since the clock may have to spin more times to find a victim entry for eviction.

* The eviction callback is a function that receives a pair of `interface{}` that represents the key and value of an evicted entry. The function is invoked in a new goroutine every time an eviction occurs.

```go
// callback function for evictions
cb := func(key, value interface{}) {
        // do something with key and value...
}

// create a new cache with clock count 3 and a callbak function for evictions
c := clock.New(128, clock.Count(3), clock.OnEvict(cb))
```