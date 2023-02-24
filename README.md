# ReplicateAPI

A dead simple API wrapper around the replicate API.

## Quick start

```bash
go get github.com/StyleSpaceAI/replicateapi@latest
```

```go
import "github.com/StyleSpaceAI/replicateapi"

const (
    token = "totalyvalidtoken"
    MODELNAME = "stability-ai/stable-diffusion"
    MODELVERSION = "db21e45d3f7023abc2a46ee38a23973f6dce16bb082a930b0c49861f96d1e5bf"
)

func main() {
    // Initialize a new API client
    cli, err := replicateapi.NewClient(token, MODELNAME, MODELVERSION)
    if err != nil {
	    log.Fatal("init client", err)
    }

    // Register an asynchronous prediction task
    result, err := cli.CreatePrediction(context.Background(), map[string]interface{}{
	    "prompt": "putin sucks huge cock, 4k",
    })
    if err != nil {
	    log.Fatal("create prediction", err)
    }

    // Fetch status and results of existing prediction
    result, err = cli.GetResult(context.Background(), result.ID)
    if err != nil {
	    log.Fatal("fetch prediction result", err)
    }
}
```
