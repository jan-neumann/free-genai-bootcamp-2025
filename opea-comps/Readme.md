# Running Ollama Third-Party Service

## Required
You can get the model ID from the Ollama model catalog: https://ollama.com/models
LLM_MODEL_ID=qwen3:14b  # or another model you want to use
host_ip=host.docker.internal  # or your actual host IP

## Optional - only if you need to set up proxy
 http_proxy=http://your-proxy:port
 https_proxy=http://your-proxy:port
 no_proxy=localhost,127.0.0.1

## Download (Pull) a model

```
curl http://localhost:8008/api/pull -d '{"model": "qwen3:4b"}'
```

## Ollama API

Once the ollama-server is running, you can use the following curl command to test it:
```
curl --noproxy "*" http://localhost:8008/api/generate -d '{
  "model": "qwen3:4b",
  "prompt":"Why is the sky blue?"
}'
```

## Technical Uncertainties

Q: Does bridge mode mean we can only access th Ollama API with another model in the docker compose?

A: No, we can access the Ollama API with the host machine as well.

Q: Which port is being mapped 8008->12345?

A: 8008 is the port that the ollama-server is running on, 12345 is the port that the ollama-server is listening on.

Q: If we pass the LLM_MODEL_ID to the ollama server, will it download the model when we start the docker compose?

A: It does not appear to download the model when we start the docker compose. We need to use the curl command to download the model.

Q: Will the model be downloaded in the container?
Does that mean the ml model will be deleted when the container stops running?

The model will download into the container, and vanish when the container stops running. You need to mount a local drive and there is probably more work to be done.