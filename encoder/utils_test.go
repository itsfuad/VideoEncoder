package encoder

import (
	"bytes"
	"io"
	"os"
	"testing"
)

const expStr = "frame %d: expected %v, got %v"

func TestReadFrames(t *testing.T) {
	width, height := 2, 2
	input := []byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
		12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23,
	}
	expectedFrames := [][]byte{
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		{12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
	}

	// Backup original os.Stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a temporary file with the input data
	tmpfile, err := os.CreateTemp("", "testinput")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(input); err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	// Set os.Stdin to the temporary file
	os.Stdin = tmpfile

	frames := ReadFrames(width, height)

	if len(frames) != len(expectedFrames) {
		t.Fatalf("expected %d frames, got %d", len(expectedFrames), len(frames))
	}

	for i, frame := range frames {
		if !bytes.Equal(frame, expectedFrames[i]) {
			t.Errorf(expStr, i, expectedFrames[i], frame)
		}
	}
}

func TestSize(t *testing.T) {
	frames := [][]byte{
		{0, 1, 2, 3},
		{4, 5, 6, 7},
	}
	expectedSize := 8

	size := Size(frames)
	if size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, size)
	}
}

func TestApplyRLE(t *testing.T) {
	frames := [][]byte{
		{0, 0, 0, 0},
		{0, 0, 0, 0},
		{1, 1, 1, 1},
	}
	expectedRLE := [][]byte{
		{0, 0, 0, 0},
		{4, 0},
		{4, 1},
	}

	rleFrames := ApplyRLE(frames)

	for i, frame := range rleFrames {
		if !bytes.Equal(frame, expectedRLE[i]) {
			t.Errorf(expStr, i, expectedRLE[i], frame)
		}
	}
}