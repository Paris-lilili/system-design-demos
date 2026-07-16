package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const randomFile = "random.bin"
const rebuiltFile = "rebuiltFile.bin"
const chunkFolderPath = "chunks"

// manifest is the metadata of the big file, about the chunks, the order, the file name
type Manifest struct {
	FileName string   `json:"file_name"`
	Size     int64    `json:"size"` // total bytes, not total chunks
	Chunks   []string `json:"chunks"`
}

func main() {
	// 1. create a random file
	if _, err := os.Stat(randomFile); errors.Is(err, os.ErrNotExist) {
		createFile()
	} else if err != nil {
		panic(err)
	}

	checkErr(os.MkdirAll(chunkFolderPath, 0755)) //idempotency, if exist do nothing

	// 2. stream the file via a fixed-size buffer
	f, err := os.Open(randomFile) // ⚠️ return file for reading only
	checkErr(err)
	defer f.Close()

	// Test 1. replace one byte for original file
	// expect only the first chunk hash value change
	if len(os.Args) > 1 && os.Args[1] == "oneByte" {
		replaceOneByte(3)
	}

	buf := make([]byte, 64*1024) // 64kb
	var newChunks, reusedChunks int
	var totalByteSize int64
	var manifest Manifest
	manifest.FileName = randomFile
	for {
		// number of bytes from read, not byte slice length
		n, err := f.Read(buf)
		totalByteSize += int64(n)
		// check n > 0 before EOF, because the last time read will return data and EOF together, then lost the last byte
		if n > 0 {
			// hash buffer + dedup + write to /chunks folder
			// chunk := sha256.Sum256(buf) ❌ buf len is fixed
			chunkData := buf[:n]
			digest := sha256.Sum256(chunkData)
			s := hex.EncodeToString(digest[:])

			// no matter this chunk exist or not, manifest should record the chunk string.
			// e.g same blank chunk can be reused, only one chunk, but manifest two records
			manifest.Chunks = append(manifest.Chunks, s)

			chunkPath := filepath.Join(chunkFolderPath, s)
			if _, err := os.Stat(chunkPath); errors.Is(err, os.ErrNotExist) { // dedup
				err := os.WriteFile(chunkPath, chunkData, 0666)
				checkErr(err)
				newChunks++
			} else if err == nil {
				reusedChunks++
			} else {
				panic(err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
	}

	manifest.Size = totalByteSize

	// 3. manifest JSON serialization
	m, err := json.MarshalIndent(manifest, "", "  ") // no prefix, two space indent
	checkErr(err)
	err = os.WriteFile("random.bin.manifest.json", append(m, '\n'), 0666) // append new line to avoid zsh add % after last character
	checkErr(err)

	// 4. check the total number of new chunks and reused chunks
	fmt.Printf("New chunks number: %d, reused chunks number: %d, total chunks number: %d, they should be same\n", newChunks, reusedChunks, len(manifest.Chunks))

	// 5. reassemble file according to manifest and check if new one is same as old one
	rebuiltBytes, err := reassembleFile(&manifest)
	checkErr(err)

	if compareTwoFiles() && rebuiltBytes == int(manifest.Size) {
		fmt.Printf("two files are same, total bytes: %d\n", rebuiltBytes)
	} else {
		fmt.Printf("Not equal after rebuilding, original file size is %d, rebuild file bytes size is: %d\n", int(manifest.Size), rebuiltBytes)
	}
}

// create a random file if not exist
func createFile() {
	f, err := os.Create(randomFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	written, err := io.CopyN(f, rand.Reader, 1*1024*1024+100) // 1mb + 100byte, so read bytes != len(buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully wrote %d random bytes to random.bin file.\n", written)
}

// reassemble chunks according to manifest
// read chunk file with manifest order, one time only one chunk in memory, write them to the file, it returns the number of bytes.
// check to total write number if it equals to manifest total bytes
func reassembleFile(m *Manifest) (totalBytesNumber int, err error) {
	rebuiltF, err := os.Create(rebuiltFile)
	if err != nil {
		return 0, err
	}
	defer rebuiltF.Close()

	// stream chunks to the output file
	for _, chunkName := range m.Chunks {
		bytesData, err := os.ReadFile(filepath.Join(chunkFolderPath, chunkName))
		if err != nil {
			return 0, err
		}

		bytesNumber, err := rebuiltF.Write(bytesData)
		if err != nil {
			return 0, err
		}
		totalBytesNumber += bytesNumber
	}
	return totalBytesNumber, nil
}

// compare two files via hash
func compareTwoFiles() bool {
	f1, err := os.Open(randomFile)
	checkErr(err)
	defer f1.Close()

	f2, err := os.Open(rebuiltFile)
	checkErr(err)
	defer f2.Close()

	hasher1 := sha256.New()
	_, err = io.Copy(hasher1, f1)
	checkErr(err)

	hasher2 := sha256.New()
	_, err = io.Copy(hasher2, f2)
	checkErr(err)

	// Sum appends the current hash(digest) to b(nil) and returns the resulting slice. only compare digest hash itself.
	return bytes.Equal(hasher1.Sum(nil), hasher2.Sum(nil))
}	

func replaceOneByte(index int64) {
	// Open file here rather than using same file handler to avoid the offset staying at index
	// f, err := os.Open(randomFile) ⚠️ return file for reading only
	 // why 0? Because if the file does not exist, and the O_CREATE flag is passed, it is created with mode perm. Now no O_CREATE flag, doesn't need perm 
	f, err := os.OpenFile(randomFile, os.O_RDWR, 0)
	checkErr(err)
	defer f.Close()

	// Seek sets the offset for the Write on file to offset
	// SeekStart=0 means relative for the original file
	_, err = f.Seek(index, io.SeekStart)
	checkErr(err)

	// write 170 to replace index position byte
	b := []byte{0xAA}
	_, err = f.Write(b)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
