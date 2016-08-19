package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"sync"

	G "gopkg.in/gilmour-libs/gilmour-e-go.v5"
	"gopkg.in/gilmour-libs/gilmour-e-go.v5/backends/redis"
)

func echoEngine() *G.Gilmour {
	r := redis.MakeRedis("127.0.0.1:6379", "")
	engine := G.Get(r)
	return engine
}

func getInput(req *G.Request) (out string) {
	if err := req.Data(&out); err != nil {
		panic(err)
	}
	return
}

//Fetch a remote file from the URL received in Request.
func fetchReply(g *G.Gilmour) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		url := getInput(req)

		line := fmt.Sprintf("Fetch %v", url)
		g.Signal("example.log", G.NewMessage().SetData(line))

		response, err := http.Get(url)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			panic(err)
		}

		//Send back the contents.
		resp.SetData(string(contents))
	}
}

func wordsReply(g *G.Gilmour) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		line := fmt.Sprintf("WordCount")
		g.Signal("example.log", G.NewMessage().SetData(line))

		input := getInput(req)
		wordRe := regexp.MustCompile("\\w+")
		words := wordRe.FindAllString(input, -1)
		resp.SetData(words)
	}
}

func countReply(g *G.Gilmour) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		line := fmt.Sprintf("WordCount")
		g.Signal("example.log", G.NewMessage().SetData(line))

		input := []string{}
		if err := req.Data(&input); err != nil {
			panic(err)
		}

		word_counts := make(map[string]int)
		for _, word := range input {
			word_counts[word]++
		}

		resp.SetData(word_counts)
	}
}

func popularReply(g *G.Gilmour, wordLen int) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		line := fmt.Sprintf("Popular %v", wordLen)
		g.Signal("example.log", G.NewMessage().SetData(line))

		input := map[string]int{}
		if err := req.Data(&input); err != nil {
			panic(err)
		}

		popular := map[int][]string{}
		score := []int{}

		for k, v := range input {
			if len(k) != wordLen {
				continue
			}
			popular[v] = append(popular[v], k)
		}

		for k := range popular {
			score = append(score, k)
		}

		sort.Sort(sort.Reverse(sort.IntSlice(score)))

		output := struct {
			WordLength int
			Score      int
			Words      []string
		}{WordLength: wordLen}

		if len(score) > 0 {
			output.Score = score[0]
			output.Words = popular[score[0]]
		}
		resp.SetData(output)
	}
}

func findReply(g *G.Gilmour) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		line := fmt.Sprintf("Find")
		g.Signal("example.log", G.NewMessage().SetData(line))

		input := getInput(req)

		resp.SetData(input)
	}
}

func saveReply(g *G.Gilmour) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		line := fmt.Sprintf("Save")
		g.Signal("example.log", G.NewMessage().SetData(line))

		resp.SetData("Hello from Save")
	}
}

func stopWordsReply(g *G.Gilmour, stopList *[]string) func(req *G.Request, resp *G.Message) {
	return func(req *G.Request, resp *G.Message) {
		line := fmt.Sprintf("Filtering out stop words")
		g.Signal("example.log", G.NewMessage().SetData(line))

		words := []string{}
		if err := req.Data(&words); err != nil {
			panic(words)
		}

		stopMap := map[string]bool{}
		for _, w := range *stopList {
			stopMap[w] = true
		}

		filtered := []string{}
		for _, w := range words {
			if _, ok := stopMap[w]; !ok {
				filtered = append(filtered, w)
			}
		}

		resp.SetData(filtered)

	}
}

//Bind all service endpoints to their topics.
func bindListeners(g *G.Gilmour) {
	//Third argument is for Opts, which in this case is nil
	g.ReplyTo("example.fetch", fetchReply(g), nil)
	g.ReplyTo("example.words", wordsReply(g), nil)
	g.ReplyTo("example.count", countReply(g), nil)
	g.ReplyTo("example.stopfilter", stopWordsReply(g, &stopWords), nil)
	g.ReplyTo("example.popular3", popularReply(g, 3), nil)
	g.ReplyTo("example.popular4", popularReply(g, 4), nil)
	g.ReplyTo("example.popular5", popularReply(g, 5), nil)
	g.ReplyTo("example.find", findReply(g), nil)
	g.ReplyTo("example.save", saveReply(g), nil)
}

func main() {
	engine := echoEngine()
	bindListeners(engine)

	engine.Start()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
