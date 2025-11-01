package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

func initRepo(repoPath string) {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		dirPath := path.Join(repoPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			must(fmt.Errorf("error creating directory: %s", err))
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		must(fmt.Errorf("error writing file: %s", err))
	}
	fmt.Println("Initialized git directory")
}

func catFile(hash string) {
	s := readObject(hash)
	parts := strings.Split(string(s), "\x00")
	fmt.Print(parts[1])
}

func hashObject(path string) {
	file, _ := os.ReadFile(path)
	stats, _ := os.Stat(path)
	content := string(file)
	contentAndHeader := fmt.Sprintf("blob %d\x00%s", stats.Size(), content)
	sha := (sha1.Sum([]byte(contentAndHeader)))
	hash := hex.EncodeToString(sha[:])

	dir := fmt.Sprintf(".git/objects/%s", hash[:2])
	blobPath := fmt.Sprintf("%s/%s", dir, hash[2:])

	if err := os.MkdirAll(string(dir), 0755); err != nil {
		must(fmt.Errorf("mkdir %s got err=%w", string(dir), err))
	}

	var buffer bytes.Buffer
	z := zlib.NewWriter(&buffer)
	z.Write([]byte(contentAndHeader))
	z.Close()
	os.WriteFile(blobPath, buffer.Bytes(), 0755)
	fmt.Print(hash)
}

func lsTree(hash string) {
	data := readObject(hash)
	reader := bytes.NewReader(data)

	if header, err := readUntil(reader, 0); err != nil || !bytes.HasPrefix(header, []byte("tree")) {
		must(fmt.Errorf("invalid object, not has prefix tree"))
	}

	for {
		if _, err := readUntil(reader, ' '); err != nil {
			break
		}

		if filename, err := readUntil(reader, 0); err != nil {
			break
		} else {
			reader.Seek(20, io.SeekCurrent)
			fmt.Println(string(filename))
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: git <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		initRepo(".")
	case "cat-file": // git cat-file -p <blob_sha>
		if len(os.Args) < 4 {
			must(errors.New("usage: mygit cat-file -p [<args>...]"))
		}
		if os.Args[2] != "-p" {
			must(errors.New("usage: mygit cat-file -p [<args>...]"))
		}
		catFile(os.Args[3])
	case "hash-object":
		if len(os.Args) < 4 {
			must(errors.New("usage: mygit hash-object -w <path-file>"))
		}
		if os.Args[2] != "-w" {
			must(errors.New("usage: mygit hash-object -w <path-file>"))
		}
		hashObject(os.Args[3])
	case "ls-tree":
		if len(os.Args) < 4 {
			must(errors.New("usage: mygit ls-tree --name-only <tree_sha>"))
		}
		if os.Args[2] != "--name-only" {
			must(errors.New("usage: mygit ls-tree --name-only <tree_sha>"))
		}
		lsTree(os.Args[3])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func readObject(hash string) []byte {
	path := fmt.Sprintf(".git/objects/%v/%v", hash[0:2], hash[2:])
	file, _ := os.Open(path)
	r, _ := zlib.NewReader(io.Reader(file))
	s, _ := io.ReadAll(r)
	r.Close()
	return s
}

func readUntil(reader *bytes.Reader, separator byte) ([]byte, error) {
	var result []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == separator {
			return result, nil
		}
		result = append(result, b)
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
