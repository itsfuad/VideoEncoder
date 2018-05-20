# Video Processing and Compression Tool

This project is a command-line tool for processing and compressing video frames. The tool supports converting RGB frames to the YUV420P format, applying RLE (Run Length Encoding) compression, and using DEFLATE for further compression. It also supports decoding the frames back to their original format.

## Features

1. **Frame Conversion**:
   - Converts RGB frames to the YUV420P format to reduce size while preserving color information.
   
2. **Compression**:
   - Applies RLE compression to reduce redundancy between consecutive frames.
   - Uses DEFLATE to compress frames further.

3. **Decoding**:
   - Decodes the compressed frames back to the YUV420P and RGB formats.

4. **File Output**:
   - Writes the processed YUV and RGB frames to output files for further use.

## Prerequisites

- Go (1.16 or later)

## Installation

1. Clone the repository:
   ```bash
   git clone <repository_url>
   cd <repository_directory>
   ```

2. Build the executable:
   ```bash
   go build -o video_tool main.go
   ```

## Usage

Run the tool with the following command-line arguments:

```bash
"video.rgb24" | ./video_tool --width <video_width> --height <video_height>
```

### Arguments:

- `--width`: Specifies the width of the video frames (default: 384).
- `--height`: Specifies the height of the video frames (default: 216).

## Workflow

1. **Reading Frames**:
   - Reads RGB video frames from `stdin`.

2. **Converting to YUV420P**:
   - Converts each frame from RGB to YUV420P format, which is more storage-efficient.

3. **Compressing Frames**:
   - Applies RLE compression on the frame deltas.
   - Uses DEFLATE for additional compression.

4. **Writing to File**:
   - Writes the compressed frames to an output file.

5. **Decoding Frames**:
   - Decodes the compressed frames back into YUV420P and RGB formats.

## File Outputs

1. `encoded.yuv`: Contains the YUV420P-encoded frames.
2. `decoded.yuv`: Contains the decoded YUV420P frames.
3. `decoded.rgb24`: Contains the final decoded RGB frames.

## Key Functions

### Frame Processing
- `convertRGBToYUV420P`: Converts RGB frames to YUV420P.
- `compressFrames`: Applies RLE and DEFLATE compression.

### Frame Decoding
- `decodeFrames`: Decodes the compressed frames back into RGB format.

### Utility
- `size`: Computes the total size of a slice of frames.
- `clamp`: Clamps a value between a minimum and maximum range.

## Notes

- The tool assumes video frames are fed via `stdin` and outputs files in the current working directory.
- Ensure the input file matches the specified width and height; otherwise, processing may fail.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## Contributions

Contributions are welcome! Feel free to open an issue or submit a pull request for enhancements or bug fixes.