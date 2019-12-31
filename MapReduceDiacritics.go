package main

import (
	"fmt"
	"sync"
)

func mapper(in <-chan string, out chan<- map[string]int) {

	countWords := map[string]int{}

	manyDiacritics := "Two or more diacritics"

	for word := range in {

		runes := []rune(word)

		var diacriticsInword int
		diacriticsInword = 0

		for j := 0; j < len(runes); j++ {
			switch runes[j] {
			case 'ă', 'ș', 'ț', 'î', 'â':
				diacriticsInword++
			}
		}

		if diacriticsInword >= 2 {
			countWords[manyDiacritics]++
		}

	}

	out <- countWords
	close(out)
}

func reducer(in <-chan int, out chan<- float32) {
	sum, count := 0, 0
	for n := range in {
		sum += n
		count++
	}
	out <- float32(sum) / float32(count)
	close(out)
}

func inputReader(out [3]chan<- string) {
	input := [][]string{
		{"aeiou", "ăăstăț", "abcedfg", "țțîî", "â"},
		{"ăasdî", "țșșsssd", "ltsf", "hjtyc", "mmnbop", "ââ"},
		{"trs", "btw", "hhfj", "abs"},
	}

	for i := range out {
		go func(channel chan<- string, word []string) {
			for _, letter := range word {
				channel <- letter
			}
			close(channel)
		}(out[i], input[i])
	}
}

func shuffler(in []<-chan map[string]int, out chan<- int) {
	var wg sync.WaitGroup
	wg.Add(len(in))

	for _, ch := range in {
		go func(channel <-chan map[string]int) {
			for m := range channel {
				manyDiacritics, ok := m["Two or more diacritics"]
				if ok {
					out <- manyDiacritics
				}
			}
			wg.Done()
		}(ch)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
}

func outputWriter(in []<-chan float32) {
	var wg sync.WaitGroup
	wg.Add(len(in))

	name := []string{"words with many diacritics"}

	for i := 0; i < len(in); i++ {
		go func(n int, channel <-chan float32) {
			for avg := range channel {
				fmt.Printf("Average number of %ss per input text: %f\n", name[n], avg)
			}
			wg.Done()
		}(i, in[i])
	}

	wg.Wait()
}

func main() {
	size := 10
	text1 := make(chan string, size)
	text2 := make(chan string, size)
	text3 := make(chan string, size)
	map1 := make(chan map[string]int, size)
	map2 := make(chan map[string]int, size)
	map3 := make(chan map[string]int, size)
	reduce := make(chan int, size)
	avg := make(chan float32, size)

	go inputReader([3]chan<- string{text1, text2, text3})

	go mapper(text1, map1)

	go mapper(text2, map2)

	go mapper(text3, map3)

	go shuffler([]<-chan map[string]int{map1, map2, map3}, chan<- int(reduce))

	go reducer(reduce, avg)

	outputWriter([]<-chan float32{avg})
}
