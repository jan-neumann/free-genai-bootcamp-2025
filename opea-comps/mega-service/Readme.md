## Download (Pull) a model

```
curl http://localhost:8000/v1/example-service/pull -d '{"model": "qwen3:4b"}'
```

## Ollama API (Generate a Request)

Once the ollama-server is running, you can use the following curl command to test it:
```
curl --noproxy "*" http://localhost:8000/v1/example-service/generate -d '{
  "model": "qwen3:4b",
  "prompt":"Why is the sky blue?"
}'
```
