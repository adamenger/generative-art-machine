package main

import (
	"bytes"
	"html/template"
	"image/png"
	"net/http"
	"strconv"
)

// Templates
var templates = template.Must(template.New("main").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Random Art Generator</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css">
    <style>
        body { text-align: center; margin-top: 50px; }
        #progress { display: none; }
        img { cursor: pointer; margin: auto; width: 512px; height: 512px;}
    </style>
</head>
<body>
    <section class="section">
        <div class="container">
            <h1 class="title is-4">Random Art Generator</h1>
            <p>Enter a seed to generate unique random art:</p>
            <form id="form" class="field is-grouped is-justify-content-center">
                <div class="control">
                    <input class="input" type="text" name="seed" placeholder="Enter seed" required>
                </div>
                <div class="control">
                    <button class="button is-primary" type="submit">Generate</button>
                </div>
                <div class="select is-primary">
                      <select id="size">
                        <option>256</option>
                        <option>512</option>
                        <option>1024</option>
                        <option>2048</option>
                        <option>4096</option>
                      </select>
               </div>
            </form>
            <progress id="progress" class="progress is-small is-primary mt-3" max="100">Loading...</progress>
            <figure class="image is-1by1 mt-5" style=" margin: auto;">
                <img id="art" src="" style="display:none;">
            </figure>
        </div>
    </section>

    <div id="modal" class="modal">
        <div class="modal-background"></div>
        <div class="modal-content">
            <figure class="image">
                <img id="modal-art" src="">
            </figure>
        </div>
        <button class="modal-close is-large" aria-label="close"></button>
    </div>

    <script>
        const form = document.getElementById('form');
        const img = document.getElementById('art');
        const size = document.getElementById('size');
        const progress = document.getElementById('progress');
        const modal = document.getElementById('modal');
        const modalArt = document.getElementById('modal-art');
        const closeModal = document.querySelector('.modal-close');

        form.onsubmit = async (e) => {
            e.preventDefault();
            const seed = e.target.seed.value;
            const sizeValue = size.value; // Get the selected size
            img.style.display = 'none';
            progress.style.display = 'block';

            img.src = "/generate?seed=" + encodeURIComponent(seed) + "&size=" + encodeURIComponent(sizeValue); // Include size in the query
            img.onload = () => {
                progress.style.display = 'none';
                img.style.display = 'block';
            };
        };

        img.onclick = () => {
            modalArt.src = img.src;
            modal.classList.add('is-active');
        };

        closeModal.onclick = () => {
            modal.classList.remove('is-active');
        };

        document.querySelector('.modal-background').onclick = () => {
            modal.classList.remove('is-active');
        };
    </script>
</body>
</html>
`))

// Serve the main page
func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	templates.Execute(w, nil)
}

// Generate the image dynamically
func generateImage(w http.ResponseWriter, r *http.Request) {
	seed := r.URL.Query().Get("seed")
	if seed == "" {
		http.Error(w, "Seed is required", http.StatusBadRequest)
		return
	}

	sizeParam := r.URL.Query().Get("size")
	size := 512
	if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 {
		size = s
	}

	img := CreateImage(seed, size)

	var buf bytes.Buffer
	png.Encode(&buf, img)

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf.Bytes())
}
