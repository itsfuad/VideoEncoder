package main

import (
	"io"
	"os"
	"log"
	"flag"
	"bytes"
	"compress/flate"
)

func main() {
	var width, height int
	flag.IntVar(&width, "width", 384, "video width")
	flag.IntVar(&height, "height", 216, "video height")
	flag.Parse()

	frames := make([][]byte, 0)

	for {

		frame := make([]byte, width*height*3)

		// read the frame from stdin
		if _, err := io.ReadFull(os.Stdin, frame); err != nil {
			break
		}

		frames = append(frames, frame)
	}

	rawSize := size(frames)
	log.Printf("Raw size: %d bytes", rawSize)

	for i, frame := range frames {

		Y := make([]byte, width*height)
		U := make([]float64, width*height)
		V := make([]float64, width*height)
		for j := 0; j < width*height; j++ {
			r, g, b := float64(frame[3*j]), float64(frame[3*j+1]), float64(frame[3*j+2])

			y := +0.299*r + 0.587*g + 0.114*b
			u := -0.169*r - 0.331*g + 0.449*b + 128
			v := 0.499*r - 0.418*g - 0.0813*b + 128

			Y[j] = uint8(y)
			U[j] = u
			V[j] = v
		}

		uDownsampled := make([]byte, width*height/4)
		vDownsampled := make([]byte, width*height/4)
		for x := 0; x < height; x += 2 {
			for y := 0; y < width; y += 2 {

				u := (U[x*width+y] + U[x*width+y+1] + U[(x+1)*width+y] + U[(x+1)*width+y+1]) / 4
				v := (V[x*width+y] + V[x*width+y+1] + V[(x+1)*width+y] + V[(x+1)*width+y+1]) / 4

				uDownsampled[x/2*width/2+y/2] = uint8(u)
				vDownsampled[x/2*width/2+y/2] = uint8(v)
			}
		}

		yuvFrame := make([]byte, len(Y)+len(uDownsampled)+len(vDownsampled))

		copy(yuvFrame, Y)
		copy(yuvFrame[len(Y):], uDownsampled)
		copy(yuvFrame[len(Y)+len(uDownsampled):], vDownsampled)

		frames[i] = yuvFrame
	}

	// Now we have our YUV-encoded video, which takes half the space!

	yuvSize := size(frames)
	log.Printf("YUV420P size: %d bytes (%0.2f%% original size)", yuvSize, 100*float32(yuvSize)/float32(rawSize))

	if err := os.WriteFile("encoded.yuv", bytes.Join(frames, nil), 0644); err != nil {
		log.Fatal(err)
	}

	encoded := make([][]byte, len(frames))
	for i := range frames {

		if i == 0 {
			encoded[i] = frames[i]
			continue
		}

		delta := make([]byte, len(frames[i]))
		for j := 0; j < len(delta); j++ {
			delta[j] = frames[i][j] - frames[i-1][j]
		}

		var rle []byte
		for j := 0; j < len(delta); {
			// Count the number of times the current value repeats.
			var count byte

			for count = 0; count < 255 && j+int(count) < len(delta); count++ {
				if delta[j+int(count)] != delta[j] {
					break
				}
			}

			// Store the count and value.
			rle = append(rle, count)
			rle = append(rle, delta[j])

			j += int(count)
		}

		// Save the RLE frame.
		encoded[i] = rle
	}

	rleSize := size(encoded)
	log.Printf("RLE size: %d bytes (%0.2f%% original size)", rleSize, 100*float32(rleSize)/float32(rawSize))

	var deflated bytes.Buffer
	w, err := flate.NewWriter(&deflated, flate.BestCompression)
	if err != nil {
		log.Fatal(err)
	}
	for i := range frames {
		if i == 0 {
			// This is the keyframe, write the raw frame.
			if _, err := w.Write(frames[i]); err != nil {
				log.Fatal(err)
			}
			continue
		}

		delta := make([]byte, len(frames[i]))
		for j := 0; j < len(delta); j++ {
			delta[j] = frames[i][j] - frames[i-1][j]
		}
		if _, err := w.Write(delta); err != nil {
			log.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}

	deflatedSize := deflated.Len()
	log.Printf("DEFLATE size: %d bytes (%0.2f%% original size)", deflatedSize, 100*float32(deflatedSize)/float32(rawSize))

	var inflated bytes.Buffer
	r := flate.NewReader(&deflated)
	if _, err := io.Copy(&inflated, r); err != nil {
		log.Fatal(err)
	}
	if err := r.Close(); err != nil {
		log.Fatal(err)
	}

	// Split the inflated stream into frames.
	decodedFrames := make([][]byte, 0)
	for {
		frame := make([]byte, width*height*3/2)
		if _, err := io.ReadFull(&inflated, frame); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		decodedFrames = append(decodedFrames, frame)
	}

	for i := range decodedFrames {
		if i == 0 {
			continue
		}

		for j := 0; j < len(decodedFrames[i]); j++ {
			decodedFrames[i][j] += decodedFrames[i-1][j]
		}
	}

	if err := os.WriteFile("decoded.yuv", bytes.Join(decodedFrames, nil), 0644); err != nil {
		log.Fatal(err)
	}

	for i, frame := range decodedFrames {
		Y := frame[:width*height]
		U := frame[width*height : width*height+width*height/4]
		V := frame[width*height+width*height/4:]

		rgb := make([]byte, 0, width*height*3)
		for j := 0; j < height; j++ {
			for k := 0; k < width; k++ {
				y := float64(Y[j*width+k])
				u := float64(U[(j/2)*(width/2)+(k/2)]) - 128
				v := float64(V[(j/2)*(width/2)+(k/2)]) - 128

				r := clamp(y+1.402*v, 0, 255)
				g := clamp(y-0.344*u-0.714*v, 0, 255)
				b := clamp(y+1.772*u, 0, 255)

				rgb = append(rgb, uint8(r), uint8(g), uint8(b))
			}
		}
		decodedFrames[i] = rgb
	}

	out, err := os.Create("decoded.rgb24")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	for i := range decodedFrames {
		if _, err := out.Write(decodedFrames[i]); err != nil {
			log.Fatal(err)
		}
	}
}

func size(frames [][]byte) int {
	var size int
	for _, frame := range frames {
		size += len(frame)
	}
	return size
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}