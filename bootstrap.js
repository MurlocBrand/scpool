var http = require('http'),
    fs = require('fs')

if (!fs.existsSync('./keys')) {
    console.error('No key file (./keys): exiting...')
    process.exit(1)
}
var keys = fs.readFileSync('./keys').toString().split('\n')

var scopt = {
    host: 'api.soundcloud.com'
}
var poolOptNew = {
    host: 'localhost',
    port: '8000',
    method: 'POST',
    path: '/new'
}
var poolOptSet = {
    host: poolOptNew.host,
    port: poolOptNew.port,
    method: 'PUT',
    path: '/set'
}

var echoResponse = function(response) {
    var str = ''
    response.on('data', function (chunk) {
        str += chunk
    })
    response.on('end', function () {
        console.log(str)
    })
}
function addKey (key) {
    var req = http.request(poolOptNew, echoResponse)
    req.write(JSON.stringify({ key: key }))
    req.end()  

    scopt.path = '/tracks/13158665/stream?client_id=' + key
    http.request(scopt, function (res) {
        var str = ''
        res.on('data', function (chunk) {
            str += chunk
        })

        res.on('end', function () {
            if (res.statusCode === 429) {
                var obj = JSON.parse(str)
                var req = http.request(poolOptSet, outputFn)
                req.write(JSON.stringify({ 
                    key: key, 
                    time : obj['reset_time']
                }))
                req.end()  
            }
        })
    }).end()
}

for (var i = keys.length - 1; i >= 0; i--) {
    addKey(keys[i])
}