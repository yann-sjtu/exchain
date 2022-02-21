package file

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	writeFile(t)
	readFile(t)
}

func writeFile(t *testing.T) {
	f, err := os.OpenFile("./tx.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	msg := "test1"
	w := bufio.NewWriter(f)
	for i:=0;;i++ {
		_, err := w.Write([]byte(msg))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println([]byte("\n"), len([]byte("\n")))
		_, err = w.Write([]byte("\n"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T) {
	f, err := os.OpenFile("./tx.txt", os.O_RDONLY, 0666)
	if err != nil {
		t.Fatal(err)
	}
	r := bufio.NewReader(f)
	var line int
	for {
		data, _, _ := r.ReadLine()
		if len(data) == 0 {
			return
		}
		fmt.Println(line, string(data))
		line++
	}

}
