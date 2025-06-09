### Dependencies:

`Go 1.24.2`

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

### Usage:

Example commands:

`Add John's number 111222333`

`Remove John's number`

`What's John's number`

`Show all contacts`