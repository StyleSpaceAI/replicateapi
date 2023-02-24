# ReplicateAPI

A dead simple API wrapper around the replicate API.

## Quick start

```bash
go get github.com/StyleSpaceAI/replicateapi@latest
```

```go
import "github.com/StyleSpaceAI/replicateapi"

const (
    token = "getYourTokenFromReplicateProfile"
    MODELNAME = "stability-ai/stable-diffusion"
)

func main() {
    // Initialize a new API client
	cli, err := replicateapi.NewClient(token, MODELNAME, "")
	if err != nil {
		log.Fatal("init client", err)
	}

	// Fetch all the available versions for this model
	vers, err := cli.GetModelVersions(context.Background())
	if err != nil {
		log.Fatal("fetch versions", err)
	}

	// Picking the latest version of the model
	cli.Version = vers[0].ID

	// Register an asynchronous prediction task
	result, err := cli.CreatePrediction(context.Background(), map[string]interface{}{
		"prompt": "putin sucks huge cock, 4k",
	})
	if err != nil {
		log.Fatal("create prediction", err)
	}

	// The response of the API is async, so we need to wait for the response
	for keepChecking := true; keepChecking; {
		time.Sleep(time.Second * 3)

		// Fetch status and results of existnig prediction
		result, err = cli.GetResult(context.Background(), result.ID)
		if err != nil {
			log.Fatal("fetch prediction result", err)
		}

		switch result.Status {
		case replicateapi.PredictionStatusSucceeded, replicateapi.PredictionStatusCanceled, replicateapi.PredictionStatusFailed:
			// Final statuses
			keepChecking = false
		case replicateapi.PredictionStatusProcessing, replicateapi.PredictionStatusStarting:
			// Still processing
		}
	}
	fmt.Printf("%+v\n", result)
}
```
