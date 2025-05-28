# HTML2IMG

<b>html2img</b> is a Go module and Dockerized CLI tool for converting HTML files to image formats (e.g., PNG) using headless Chromium.


## Usage

### Build the Docker image

```bash
$ make build
```

### Run the container to render HTML to an image
```bash
$ docker run --rm -v $(pwd):/work html2img <html_file_path> <output.png>
```