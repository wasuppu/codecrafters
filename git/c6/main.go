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
	"path/filepath"
	"sort"
	"strings"
	"time"
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
	data, _ := os.ReadFile(path)
	hash := writeObject("blob", data)
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

type TreeEntry struct {
	Name string
	Mode string
	Hash string
}

func writeTree(path string) string {
	items, err := os.ReadDir(path)
	must(err)

	entries := []TreeEntry{}
	for _, item := range items {
		if item.Name() == ".git" {
			continue
		}
		if item.IsDir() {
			hash := writeTree(filepath.Join(path, item.Name()))
			entries = append(entries, TreeEntry{Name: item.Name(), Mode: "40000", Hash: hash})
		} else {
			content, _ := os.ReadFile(filepath.Join(path, item.Name()))
			hash := writeObject("blob", content)
			entries = append(entries, TreeEntry{Name: item.Name(), Mode: "100644", Hash: hash})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	var buffer bytes.Buffer
	for _, entry := range entries {
		hash, _ := hex.DecodeString(entry.Hash)
		row := fmt.Sprintf("%s %s\x00%s", entry.Mode, entry.Name, hash)
		buffer.WriteString(row)
	}
	return writeObject("tree", buffer.Bytes())
}

func commitTree(treeHash, parentHash, msg string) string {
	var builder strings.Builder
	builder.WriteString("tree " + treeHash + "\n")
	builder.WriteString("parent " + parentHash + "\n")

	authorName := os.Getenv("GIT_AUTHOR_NAME")
	authorEmail := os.Getenv("GIT_AUTHOR_EMAIL")
	author := fmt.Sprintf("%s <%s>", authorName, authorEmail)

	timestamp := time.Now().Unix()
	_, offset := time.Now().Zone()
	utcOffset := offset / 3600
	offsetMinutes := (offset % 3600) / 60

	sign := "+"
	if utcOffset < 0 {
		sign = "-"
		utcOffset = -utcOffset
	}

	authorTime := fmt.Sprintf("%d %s%02d%02d", timestamp, sign, utcOffset, offsetMinutes)

	builder.WriteString(fmt.Sprintf("author %s %s\n", author, authorTime))
	builder.WriteString(fmt.Sprintf("committer %s %s\n\n", author, authorTime))
	builder.WriteString(msg + "\n")
	sha := writeObject("commit", []byte(builder.String()))
	return sha
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
	case "write-tree":
		fmt.Println(writeTree("."))
	case "commit-tree":
		if len(os.Args) < 7 {
			must(fmt.Errorf("usage: mygit commit-tree <tree_sha> -p <commit_sha> -m <message>"))
		}
		fmt.Println(commitTree(os.Args[2], os.Args[4], os.Args[6]))
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func writeObject(typ string, data []byte) string {
	contentAndHeader := fmt.Sprintf("%s %d\x00%s", typ, len(data), string(data))

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

	return hash
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
