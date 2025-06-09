### Build:

```bash
  docker build -t nlp-phone-book .
```

### Run:

You need **`OPENAI_API_KEY`** env set in order to run

```bash
    docker run -it --rm \
      -e OPENAI_API_KEY \
      -v $(pwd)/data:/app/data \
      -p 8080:8080 \
      nlp-phone-book
```