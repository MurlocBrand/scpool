package main

// Get an available key:
//  client -- GET /get --> server
//  client <-- key -- server (plain text)
//
// Register a key:
//  client -- POST /new {key:k} --> server (json)
//
// Tell server that key is fucked:
// client -- PUT /set {key:k,time:t} --> server (json)

import (
    "net/http"
    "container/list"
    "encoding/JSON"
    "errors"
    "fmt"
    "flag"
    "time"
)

var (
    Keys = list.New()
    currentKey *list.Element

    TimeLayout = "2006/01/02 15:04:05 0000"
)

var (
    ErrNoKeys = errors.New("Found no registered keys")
    ErrNoAvailable = errors.New("All keys are rate limited")
    ErrKeyNotFound = errors.New("Key wasn't found in pool")
    ErrKeyAlreadyExists = errors.New("Key has already been registered")
)

type KeyEntry struct {
    // The client_id of an app
    Key string

    // Some time in the past: available
    // Some time in the future: unavailable
    ResetDate time.Time
}
func (e *KeyEntry) Available() bool {
    return time.Now().After(e.ResetDate)
}

type apiError struct {
     Message string
     Code int
     Error error
 } 

type handler func (http.ResponseWriter, *http.Request) *apiError

func (fn handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := fn(w, r); err != nil {
        fmt.Printf("%s %s %d (%s)\n", r.Method, r.URL, err.Code, err.Error.Error())
        http.Error(w, err.Message, err.Code)
    } else {
        fmt.Printf("%s %s %d\n", r.Method, r.URL, 200)
    }
}

// curl -X POST http://localhost:8000/new -D key=wopp
func registerKey(w http.ResponseWriter, r *http.Request) *apiError {
    decoder := json.NewDecoder(r.Body)
    defer r.Body.Close()

    var msg struct {
        Key string
    }

    if err := decoder.Decode(&msg); err != nil {
        return &apiError{
            Message: "You fucked now, son",
            Code: 400,
            Error: err,
        }
    }

    for e := Keys.Front(); e != nil; e = e.Next() {
        ke := e.Value.(*KeyEntry)
        if ke.Key == msg.Key {
            return &apiError{
                Message: "Key already exists",
                Code: 400,
                Error: ErrKeyAlreadyExists,
            }
        }
    }

    Keys.PushFront(&KeyEntry{
        Key: msg.Key,
        ResetDate: time.Now(),
    })

    if currentKey == nil {
        currentKey = Keys.Front()
    }
    
    w.Write([]byte(msg.Key))
    return nil
}
// curl -X PUT http://localhost:8000/set -D key=wopp,time=t
func updateKey(w http.ResponseWriter, r *http.Request) *apiError {
    // parse body
    decoder := json.NewDecoder(r.Body)
    defer r.Body.Close()

    var msg struct {
        Key string
        Time string
    }

    if err := decoder.Decode(&msg); err != nil {
        return &apiError{
            Message: "You fucked now, son",
            Code: 400,
            Error: err,
        }
    }

    fmt.Printf("%v\n", msg)

    t, e := time.Parse(TimeLayout, msg.Time)
    if e != nil {
        return &apiError{
            Message: "Couldn't parse provided time!",
            Code: 400,
            Error: e,
        }
    }

    el := Keys.Front()
    if el == nil {
        return &apiError{
            Message: ErrNoKeys.Error(),
            Code: 500,
            Error: ErrNoKeys,
        }
    }

    ke := el.Value.(*KeyEntry)
    for ke.Key != msg.Key {
        el = el.Next()
        if el == nil {
            return &apiError{
                Message: ErrKeyNotFound.Error(),
                Code: 400,
                Error: ErrKeyNotFound,
            }
        }
        ke = el.Value.(*KeyEntry)
    }
    ke.ResetDate = t

    w.Write([]byte("Ok"))
    return nil
}
// curl http://localhost:8000/get
func acquireKey(w http.ResponseWriter, r *http.Request) *apiError {
    if currentKey == nil {
        return &apiError{
            Message: ErrNoKeys.Error(),
            Code: 500,
            Error: ErrNoKeys,
        }
    }

    // start on currentKey.Next
    el := currentKey.Next()
    if el == nil {
        el = Keys.Front()
    }
    ke := el.Value.(*KeyEntry)

    // stop when we've gone full circle or have found a valid key
    // we only care about getting a functional key (not rate limited),
    // the round-robin is just for simplistic balancing
    for !ke.Available() && el != currentKey {
        el = el.Next()
        if el == nil {
            el = Keys.Front()
        }
        ke = el.Value.(*KeyEntry)
    }

    // so here we don't check if el == currentKey, because we don't care...
    currentKey = el
    if !ke.Available() {
        return &apiError{
            Message: ErrNoAvailable.Error(),
            Code: 500,
            Error: ErrNoAvailable,
        }
    }

    w.Write([]byte(ke.Key))
    return nil
}

func b2str(b bool) string {
    if b {
        return "available"
    }
    return "unavailable"
}
func dumpKeys(w http.ResponseWriter, r *http.Request) *apiError {
    for e := Keys.Front(); e != nil; e = e.Next() {
        ke := e.Value.(*KeyEntry)
        dur := ke.ResetDate.Sub(time.Now())
        fmt.Fprintf(w, "%s %s %s\n", 
            ke.Key, 
            b2str(dur.Seconds() <= 0), 
            ke.ResetDate.Format(TimeLayout))
    }
    return nil
}

func main() {
    serverAddr := flag.String("-addr", ":8000", "server address")
    flag.Parse()

    http.Handle("/new", handler(registerKey))
    http.Handle("/get", handler(acquireKey))
    http.Handle("/set", handler(updateKey))
    http.Handle("/keys", handler(dumpKeys))

    fmt.Printf("Starting scpool server at %s\n", *serverAddr)
    err := http.ListenAndServe(*serverAddr, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}