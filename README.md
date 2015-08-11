scpool: soundcloud client_id key pool server.

includes a (useful) js script for initializing your pool with some keys.

 register a key
    
    curl -X POST localhost:8000/new \
        -d '{ "key" : "your-key" }' \
        -H "Content-Type: application/json"
    

 get a key

    curl localhost:8000/get

 set key reset time
    
    curl -X POST localhost:8000/set \
        -d '{ "key" : "your-key", "time" : "2033/03/01 15:33:22 0000" }' \
        -H "Content-Type: application/json"

 list all keys

    curl localhost:8000/keys

package: https://gobuilder.me/github.com/murlocbrand/scpool
  license: MIT (LICENSE)