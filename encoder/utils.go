package encoder

import (
	"bytes"
	"compress/flate"
	"io"
	"log"
	"os"
)

func ReadFrames(width, height int) [][]byte {
	frames := make([][]byte, 0)
	for {
		frame := make([]byte, width*height*3)
		if _, err := io.ReadFull(os.Stdin, frame); err != nil {
			break
		}
		frames = append(frames, frame)
	}
	return frames
}

func Size(frames [][]byte) int {
	var size int
	for _, frame := range frames {
		size += len(frame)
	}
	return size
}

func ConvertToYUV420P(frames [][]byte, width, height int) {
	for i, frame := range frames {
		Y, U, V := splitRGBToYUV(frame, width, height)
		uDownsampled, vDownsampled := downsampleUV(U, V, width, height)
		frames[i] = mergeYUV(Y, uDownsampled, vDownsampled)
	}
}

func splitRGBToYUV(frame []byte, width, height int) ([]byte, []float64, []float64) {
	Y := make([]byte, width*height)
	U := make([]float64, width*height)
	V := make([]float64, width*height)
	for j := 0; j < width*height; j++ {
		r, g, b := float64(frame[3*j]), float64(frame[3*j+1]), float64(frame[3*j+2])
		Y[j] = uint8(0.299*r + 0.587*g + 0.114*b)
		U[j] = -0.169*r - 0.331*g + 0.449*b + 128
		V[j] = 0.499*r - 0.418*g - 0.0813*b + 128
	}
	return Y, U, V
}

func downsampleUV(U, V []float64, width, height int) ([]byte, []byte) {
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
	return uDownsampled, vDownsampled
}

func mergeYUV(Y, U, V []byte) []byte {
	yuvFrame := make([]byte, len(Y)+len(U)+len(V))
	copy(yuvFrame, Y)
	copy(yuvFrame[len(Y):], U)
	copy(yuvFrame[len(Y)+len(U):], V)
	return yuvFrame
}

func SaveToFile(filename string, frames [][]byte) {
	if err := os.WriteFile(filename, bytes.Join(frames, nil), 0644); err != nil {
		log.Fatal(err)
	}
}

func ApplyRLE(frames [][]byte) [][]byte {
	rleFrames := make([][]byte, len(frames))
	for i := range frames {
		if i == 0 {
			rleFrames[i] = frames[i]
			continue
		}
		rleFrames[i] = EncodeRLE(frames[i], frames[i-1])
	}
	return rleFrames
}

func EncodeRLE(current, previous []byte) []byte {
	delta := make([]byte, len(current))
	for j := 0; j < len(delta); j++ {
		delta[j] = current[j] - previous[j]
	}
	var rle []byte
	for j := 0; j < len(delta); {
		var count byte
		for count = 0; count < 255 && j+int(count) < len(delta); count++ {
			if delta[j+int(count)] != delta[j] {
				break
			}
		}
		rle = append(rle, count, delta[j])
		j += int(count)
	}
	return rle
}

func ApplyDeflate(frames [][]byte) int {
	var deflated bytes.Buffer
	w, err := flate.NewWriter(&deflated, flate.BestCompression)
	if err != nil {
		log.Fatal(err)
	}
	for _, frame := range frames {
		if _, err := w.Write(frame); err != nil {
			log.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}
	return deflated.Len()
}

func DecodeFrames(width, height int) [][]byte {
	var deflated bytes.Buffer
	r := flate.NewReader(&deflated)
	var inflated bytes.Buffer
	if _, err := io.Copy(&inflated, r); err != nil {
		log.Fatal(err)
	}
	if err := r.Close(); err != nil {
		log.Fatal(err)
	}
	return SplitInflatedFrames(&inflated, width, height)
}

func SplitInflatedFrames(inflated *bytes.Buffer, width, height int) [][]byte {
	decodedFrames := make([][]byte, 0)
	for {
		frame := make([]byte, width*height*3/2)
		if _, err := io.ReadFull(inflated, frame); err != nil {
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
	return decodedFrames
}

func SaveRGB(decodedFrames [][]byte, width, height int) {
	for i, frame := range decodedFrames {
		decodedFrames[i] = ConvertYUVToRGB(frame, width, height)
	}
	out, err := os.Create("decoded.rgb24")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	for _, frame := range decodedFrames {
		if _, err := out.Write(frame); err != nil {
			log.Fatal(err)
		}
	}
}

func ConvertYUVToRGB(frame []byte, width, height int) []byte {
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
	return rgb
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
