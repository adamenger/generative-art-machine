package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
  "runtime"
	"sync"
)

// ColorFunc defines a function that computes RGB values based on x, y.
type ColorFunc interface {
	Eval(x, y float64) (r, g, b float64)
	String() string
}

// Constant represents a constant RGB color.
type Constant struct {
	r, g, b float64
}

func (c *Constant) Eval(x, y float64) (float64, float64, float64) {
	return c.r, c.g, c.b
}

func (c *Constant) String() string {
	return fmt.Sprintf("Constant(%.2f, %.2f, %.2f)", c.r, c.g, c.b)
}

// VariableX represents the X coordinate as a color.
type VariableX struct{}

func (v *VariableX) Eval(x, y float64) (float64, float64, float64) {
	return x, x, x
}

func (v *VariableX) String() string {
	return "VariableX"
}

// VariableY represents the Y coordinate as a color.
type VariableY struct{}

func (v *VariableY) Eval(x, y float64) (float64, float64, float64) {
	return y, y, y
}

func (v *VariableY) String() string {
	return "VariableY"
}

// Sin represents a sine wave operation on a subexpression.
type Sin struct {
	phase, freq float64
	sub         ColorFunc
}

func (s *Sin) Eval(x, y float64) (float64, float64, float64) {
	r, g, b := s.sub.Eval(x, y)
	return math.Sin(s.phase + s.freq*r), math.Sin(s.phase + s.freq*g), math.Sin(s.phase + s.freq*b)
}

func (s *Sin) String() string {
	return fmt.Sprintf("Sin(phase=%.2f, freq=%.2f, sub=%s)", s.phase, s.freq, s.sub.String())
}

// Product multiplies the results of two subexpressions.
type Product struct {
	left, right ColorFunc
}

func (p *Product) Eval(x, y float64) (float64, float64, float64) {
	r1, g1, b1 := p.left.Eval(x, y)
	r2, g2, b2 := p.right.Eval(x, y)
	return r1 * r2, g1 * g2, b1 * b2
}

func (p *Product) String() string {
	return fmt.Sprintf("Product(left=%s, right=%s)", p.left.String(), p.right.String())
}

// Mix blends two subexpressions based on a fixed weight.
type Mix struct {
	w     float64
	left  ColorFunc
	right ColorFunc
}

func (m *Mix) Eval(x, y float64) (float64, float64, float64) {
	r1, g1, b1 := m.left.Eval(x, y)
	r2, g2, b2 := m.right.Eval(x, y)
	return m.w*r1 + (1-m.w)*r2, m.w*g1 + (1-m.w)*g2, m.w*b1 + (1-m.w)*b2
}

func (m *Mix) String() string {
	return fmt.Sprintf("Mix(w=%.2f, left=%s, right=%s)", m.w, m.left.String(), m.right.String())
}

// Well creates a well-like pattern for the subexpression.
type Well struct {
	sub ColorFunc
}

func (w *Well) Eval(x, y float64) (float64, float64, float64) {
	r, g, b := w.sub.Eval(x, y)
	well := func(v float64) float64 {
		return 1 - 2/(1+math.Pow(v, 2))
	}
	return well(r), well(g), well(b)
}

func (w *Well) String() string {
	return fmt.Sprintf("Well(sub=%s)", w.sub.String())
}

// FractalNoise generates a fractal noise pattern.
type FractalNoise struct {
	scale float64
	sub   ColorFunc
}

func (f *FractalNoise) Eval(x, y float64) (float64, float64, float64) {
	r, g, b := f.sub.Eval(x*f.scale, y*f.scale)
	smooth := func(v float64) float64 {
		return 0.5 * (math.Sin(5*v) + math.Cos(5*v))
	}
	return smooth(r), smooth(g), smooth(b)
}

func (f *FractalNoise) String() string {
	return fmt.Sprintf("FractalNoise(scale=%.2f, sub=%s)", f.scale, f.sub.String())
}

// Generate creates a random expression tree with a minimum depth.
func Generate(minDepth, maxDepth int) ColorFunc {
	if maxDepth == 0 || (minDepth <= 0 && rand.Float64() < 0.2) {
		switch rand.Intn(3) {
		case 0:
			return &VariableX{}
		case 1:
			return &VariableY{}
		default:
			return &Constant{
				r: rand.Float64()*2 - 1,
				g: rand.Float64()*2 - 1,
				b: rand.Float64()*2 - 1,
			}
		}
	}
	switch rand.Intn(4) { // Removed Kaleidoscope and Spiral
	case 0:
		return &Sin{
			phase: rand.Float64() * 2 * math.Pi,
			freq:  0.5 + rand.Float64()*3.0,
			sub:   Generate(minDepth-1, maxDepth-1),
		}
	case 1:
		return &Mix{
			w:     rand.Float64(),
			left:  Generate(minDepth-1, maxDepth-1),
			right: Generate(minDepth-1, maxDepth-1),
		}
	case 2:
		return &Product{
			left:  Generate(minDepth-1, maxDepth-1),
			right: Generate(minDepth-1, maxDepth-1),
		}
	default:
		return &FractalNoise{
			scale: 0.5 + rand.Float64()*2.0,
			sub:   Generate(minDepth-1, maxDepth-1),
		}
	}
}

// CreateImage generates the random art as an image using multithreading.
func CreateImage(seed string, size int) image.Image {
	rand.Seed(int64(hash(seed)))
	var img *image.RGBA
	var variance float64

	// Generate images until a threshold is met
	for attempts := 0; attempts < 3; attempts++ { // Retry up to 5 times
		img = image.NewRGBA(image.Rect(0, 0, size, size))
		art := Generate(10, 30)

		fmt.Printf("Expression tree for seed '%s':\n%s\n", seed, art.String())

		var wg sync.WaitGroup
  	numWorkers := runtime.GOMAXPROCS(0)
		rowsPerWorker := size / numWorkers

		for worker := 0; worker < numWorkers; worker++ {
			wg.Add(1)
			go func(worker int) {
				defer wg.Done()
				startRow := worker * rowsPerWorker
				endRow := startRow + rowsPerWorker
				if worker == numWorkers-1 {
					endRow = size
				}

				for py := startRow; py < endRow; py++ {
					for px := 0; px < size; px++ {
						x := 2*float64(px)/float64(size) - 1
						y := 2*float64(py)/float64(size) - 1
						r, g, b := art.Eval(x, y)
						r, g, b = normalize(r), normalize(g), normalize(b)
						color := color.RGBA{
							R: uint8(128 + r*127),
							G: uint8(128 + g*127),
							B: uint8(128 + b*127),
							A: 255,
						}
						img.Set(px, py, color)
					}
				}
			}(worker)
		}

		wg.Wait()

		// Check variance
		variance = calculateColorVariance(img, size)
		if variance > 30.0 { // Example threshold for diversity
			break
		}

		fmt.Println("Low variance detected. Retrying with new expression tree...")
	}

	if variance <= 50.0 {
		fmt.Println("Warning: Generated image still has low variance.")
	}

	return img
}

// Normalize ensures RGB values stay in the range [-1, 1].
func normalize(value float64) float64 {
	if value < -1 {
		return -1
	}
	if value > 1 {
		return 1
	}
	return value
}

// Polar transforms Cartesian to polar coordinates.
func polar(x, y float64) (r, theta float64) {
	r = math.Sqrt(x*x + y*y)
	theta = math.Atan2(y, x)
	return
}

// Calculate color variance to detect overly uniform images.
func calculateColorVariance(img *image.RGBA, size int) float64 {
	var rTotal, gTotal, bTotal, rSq, gSq, bSq float64
	pixelCount := float64(size * size)

	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			c := img.RGBAAt(px, py)
			r := float64(c.R)
			g := float64(c.G)
			b := float64(c.B)
			rTotal += r
			gTotal += g
			bTotal += b
			rSq += r * r
			gSq += g * g
			bSq += b * b
		}
	}

	rMean, gMean, bMean := rTotal/pixelCount, gTotal/pixelCount, bTotal/pixelCount
	return math.Sqrt((rSq/pixelCount-rMean*rMean) + (gSq/pixelCount-gMean*gMean) + (bSq/pixelCount-bMean*bMean))
}

// Hash the seed to get a consistent random seed.
func hash(seed string) uint32 {
	h := uint32(2166136261)
	for _, c := range seed {
		h = (h * 16777619) ^ uint32(c)
	}
	return h
}
