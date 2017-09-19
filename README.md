# numbers
This Go package is a solution to the Go challenge (as provided in go-challenge.txt).

The package is divided in 3 files (barring tests). There are two primary design requirements in the challenge:

1. Query the requested URLs.
2. Collect, sort, and return the data back as response.

Point 1 above is a more general requirement. Hence it is implemented by a standalone package leven function `ProcessURLs`
found in the file *numbers.go*.

The job of this function is to query the input URLs and return their response over a channel, which is returned to the caller.
Caller of `ProcessURLs` should be able to range over the returned channel to receive the response in form of `[]int`s.
The channel relays `[]int`s instead of `int`s since a slice can be sent over a channel in a constant time, independent
of its size.

The caller does not have to close the received channel.

By giving `ProcessURLs` a single responsibility, the function can be used more independently when required.

`ProcessURLs` takes a configuration object as input. This object controls the following parameters:

* Response Timeout: This is the time duration by which `ProcessURLs` must return, i.e. it must have processed the
input URLs by this time. If some URLs are still being queried and/or there are more unprocessed URLs remaining after
this duration has expired, these URLs are not processed. A `context.Context` object has been plumbed through relevant
calls to enforce this timeout.

* Request Timeout: This is the time duration by which each individual URL that needs to be processed must complete
its network operation. This value is kept as a property of the `http.Client` which makes all the `GET` requests.

* Number of goroutines: This value controls the concurrency factor. Querying a URL is an I/O bound process, and thus
goroutines are ideal to query multiple URLs in parallel.

* `URLGetter` Interface: This is an embedded interface type that allows `ProcessURLs` to actually perform network calls.
The reason this interface is included is so that the function can have the ability to make complex network calls. This also
allows for better function testing.

With respect to the given challenge, the two timeouts defined above control the requirement of returning the response within
500ms. This time limit has been interpreted as the time duration starting from the execution of HTTP handler till its
return. Note that this time does not include request/response (to the main handler, not related to URLs that must be queried)
in-flight times. This also does not include the time to write the final result over the network connection writer.

Setting a Response timeout of 500ms would ensure that `ProcessURLs` does not perform any additional work after 500ms and
returns. However, once `ProcessURLs` returns, the numbers fetched by it must be collected and sorted. Some cursory
benchmarking (see *mainloop_bench_test.go*) reveals that output count of upto 10000 numbers take < 10ms to process and
an output count of up to 100000 numbers take < 50ms to process.

This would suggest a global response timeout passed to `ProcessURLs` to be set to around 450ms. If responses returned
by input URLs are not very large, this value can be kept close to 500ms.

The individual URL timeouts can be set to any value less than above timeout, or it can not be set at all, since the
global response timeout will cancel stragglers in any case. However, a value less than global response timeout will
allow long running requests to return early freeing up goroutines and giving other URLs a chance to be processed.

These timeout values however are difficult to determine without a better understanding of the nature of URLs that will be
queried. Better timeouts can be set by extensive benchmarking and reviewing of historical data. Timoeouts can be made
dynamic based on these methodologies. While the package is designed so that it can be updated to contain these new features,
current implementation only uses static timeouts which can be configured using command line parameters as well.

The lack of details about the URLs to be queries also does not permit caching by `ProcessURLs`. If the response returned by
the URLs changes frequently there is no point in caching it (for example, test servers */rand* handler).

The other files in the package are *numbergetter.go* and *server.go*.

*numbergetter.go* contains the definition of `defaultGetter`, which is the default implementation of `URLGetter`
interface used by `ProcessURLs`.

*server.go* contains an `http.ServeHTTP` type called `NumbersGetter` that handles the incoming requests, sends the input
URLs off to `ProcessURLs` along with a relevant configuration, and ranges over the returned channel to collect and sort
the data. It is also responsible for encoding the JSON response and writing it to the network. `NumbersGetter` could
easily be given a global timeout that would simply return an empty response if time is about to expire, however functionally
it does not appear to add much value. Again, perhaps this design decision can be made with greater accuracy if more
information about the quried URLs is available.

Apart from this, *numbers_test.go* contains test cases for `ProcessURLs`. `NumbersGetter` is not tested since the logic
contained within it is trivial.

The server has been tested for multiple subsequent requests along with Go's in-built race detector. No failures were
detected.
