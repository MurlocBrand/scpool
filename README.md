scpool: soundcloud client_id key pool server.

 register a key

    
    curl -X POST localhost:8000/new \
        -d '{ "key" : "your-key" }' \
        -H "Content-Type: application/json"
    

 get a key

    curl localhost:8000/get

 set key reset time

    
    curl -X POST localhost:8000/new \
        -d '{ "key" : "your-key", "time" : "2033/03/01 15:33:22 0000" }' \
        -H "Content-Type: application/json"
    
