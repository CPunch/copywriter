package replicate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ReplicateClient struct {
	APIKey string
	ID     string
}

func NewClient(apiKey string) *ReplicateClient {
	return &ReplicateClient{
		APIKey: apiKey,
	}
}

type ReplicatePredictionInput struct {
	Prompt      string `json:"prompt"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Num_Outputs int    `json:"num_outputs"`
}

type ReplicatePredictionBody struct {
	Version string                   `json:"version"`
	Input   ReplicatePredictionInput `json:"input"`
}

type ReplicatePredictionResponse struct {
	Error string `json:"error"`
	ID    string `json:"id"`
}

// using this model to generate Images: https://replicate.com/stability-ai/stable-diffusion/api
func (c *ReplicateClient) sendImagePrompt(prompt string) error {
	/*
		curl -s -X POST \
		-d '{"version": "ac732df83cea7fff18b8472768c88ad041fa750ff7682a21affe81863cbe77e4", "input": {"prompt": "a vision of paradise. unreal engine"}}' \
		-H "Authorization: Token $REPLICATE_API_TOKEN" \
		"https://api.replicate.com/v1/predictions"
	*/

	// create body
	body := ReplicatePredictionBody{
		Version: "ac732df83cea7fff18b8472768c88ad041fa750ff7682a21affe81863cbe77e4",
		Input: ReplicatePredictionInput{
			Prompt:      prompt,
			Width:       960,
			Height:      640,
			Num_Outputs: 1,
		},
	}

	// marshal body
	payloadBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// create request
	req, err := http.NewRequest("POST", "https://api.replicate.com/v1/predictions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Token "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to send prompt: %d", resp.StatusCode)
	}

	// decode response
	var predictionResponse ReplicatePredictionResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&predictionResponse)
	if err != nil {
		return err
	}

	// check for error
	if predictionResponse.Error != "" {
		return fmt.Errorf("replicate error: %s", predictionResponse.Error)
	}

	// set id
	c.ID = predictionResponse.ID
	return nil
}

/*
	{
	  "id": "j6t4en2gxjbnvnmxim7ylcyihu",
	  "input": {"prompt": "a vision of paradise. unreal engine"},
	  "output": "...",
	  "status": "succeeded"
	}
*/
type ReplicatePredictionStatusResponse struct {
	ID     string                   `json:"id"`
	Input  ReplicatePredictionInput `json:"input"`
	Output []string                 `json:"output"`
	Status string                   `json:"status"`
}

const (
	MAX_POLL_ATTEMPTS = 20
)

// returns url of generated image
func (c *ReplicateClient) waitForPredictionFinished() (string, error) {
	/* curl -s -H "Authorization: Token $REPLICATE_API_TOKEN" \
	"https://api.replicate.com/v1/predictions/j6t4en2gxjbnvnmxim7ylcyihu" */
	req, err := http.NewRequest("GET", "https://api.replicate.com/v1/predictions/"+c.ID, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Token "+c.APIKey)

	for i := 0; i < MAX_POLL_ATTEMPTS; i++ {
		// make request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}

		if resp.StatusCode != 200 {
			continue
		}

		// decode response
		var predictionResponse ReplicatePredictionStatusResponse
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&predictionResponse)
		if err != nil {
			return "", err
		}

		// check if output succeeded
		if predictionResponse.Status == "succeeded" {
			if len(predictionResponse.Output) == 0 {
				return "", fmt.Errorf("prediction succeeded but no output found")
			}

			return predictionResponse.Output[0], nil
		}

		// sleep for 2 seconds
		time.Sleep(time.Second * 2)
	}

	return "", fmt.Errorf("prediction timed out!!")
}

func (c *ReplicateClient) MakePrediction(prompt string) (string, error) {
	err := c.sendImagePrompt(prompt)
	if err != nil {
		return "", err
	}

	return c.waitForPredictionFinished()
}
