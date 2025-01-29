package main

import (
	"flag"
	"log"
	"os"
	"video-encoder/encoder"
)

func main() {
	var width, height int
	var videoPath string
	flag.Usage = func() {
		log.Printf("Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.IntVar(&width, "w", 384, "video width")
	flag.IntVar(&height, "h", 216, "video height")
	flag.StringVar(&videoPath, "v", "", "path to video file")
	flag.Parse()

	if videoPath != "" {
		file, err := os.Open(videoPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		os.Stdin = file
	} else {
		log.Println("Reading video from standard in")
	}

	frames := encoder.ReadFrames(width, height)

	rawSize := encoder.Size(frames)
	log.Printf("Raw size: %d bytes", rawSize)

	encoder.ConvertToYUV420P(frames, width, height)
	yuvSize := encoder.Size(frames)
	log.Printf("YUV420P size: %d bytes (%0.2f%% original size)", yuvSize, 100*float32(yuvSize)/float32(rawSize))

	encoder.SaveToFile("encoded.yuv", frames)

	rleFrames := encoder.ApplyRLE(frames)
	rleSize := encoder.Size(rleFrames)
	log.Printf("RLE size: %d bytes (%0.2f%% original size)", rleSize, 100*float32(rleSize)/float32(rawSize))

	deflatedSize := encoder.ApplyDeflate(frames)
	log.Printf("DEFLATE size: %d bytes (%0.2f%% original size)", deflatedSize, 100*float32(deflatedSize)/float32(rawSize))

	decodedFrames := encoder.DecodeFrames(width, height)

	encoder.SaveRGB(decodedFrames, width, height)
}
