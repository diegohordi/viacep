# ViaCEP Go client

Go client for [ViaCEP](http://viacep.com.br) API, used for search Brazilian Postal Codes.

## Usage

### Install

`go get github.com/diegohordi/viacep`

### Creating a client

In order to use this client, you need to create your own http.Client, with your set of configurations, plus the base 
URL of the ViaCEP API,`http://viacep.com.br/ws`:

```
import "github.com/diegohordi/viacep"
...
httpClient := &http.Client{Timeout: time.Second * 5}
apiURL := "http://viacep.com.br/ws"
client := viacep.NewClient(apiURL, httpClient)
...
cep, err := client.Consultar(context.TODO(), "01001000")
...
```

#### Timeouts

If you need a different timeout from the base client that you created, you can create a context.WithTimeout and 
pass it as parameter to the functions, as they are able to deal with context signalling too.

## Tests

The coverage so far is greater than 90%, covering also failure scenarios. Also, as the handlers are dealing with 
context timeout, there are no race conditions detected in the -race tests.

You can run the short test and the race condition test from Makefile, as below:

### Short
`make test_short`

### Race
`make test_race`

### Integration
`make test`
