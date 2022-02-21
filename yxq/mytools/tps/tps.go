package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

const filename = "../../../dev/oec.txt"
const TIME_LAYOUT = "2006-01-02 15:04:05.000"


func parseTime(s string) int64 {
	t, err := time.Parse(TIME_LAYOUT, s)
	if err != nil {
		panic(err)
	}
	return t.UnixNano()
}

func main() {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sep := ">, Tx<"
	var beginTime, endTime int64
	var heightCount, totalTxCount int64
	br := bufio.NewReader(f)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		index := bytes.Index(a, []byte(sep))
		if index == -1 {
			continue
		}
		start := index+6
		end := start
		for a[end] != '>' {
			end++
		}
		if start == end {
			continue
		}
		txCount, err := strconv.ParseInt(string(a[start:end]), 10, 64)
		if err != nil {
			panic(err)
		}
		if txCount < 10 {
			continue
		}
		if beginTime == 0 {
			beginTime = parseTime(string(a[2:12]) + " " + string(a[13:25]))
			continue
		}
		totalTxCount += txCount
		heightCount++
		endTime = parseTime(string(a[2:12]) + " " + string(a[13:25]))
		fmt.Printf("height:%d, txCount:%d, avg block time:%0.2f, avg tps:%d\n", heightCount, txCount, float64(endTime-beginTime)/1e9/float64(heightCount), totalTxCount*1e9/(endTime-beginTime))
		}
}
