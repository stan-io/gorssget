/* main.go -- gorssget
 *
 * Copyright (C) 2017 Stan
 *
 * This software may be modified and distributed under the terms
 * of the MIT license.  See the LICENSE file for details.
 */

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"regexp"

	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v2"
)

type CommandLineArgs struct {
	configFilePath string
}

type Task struct {
	Title    string
	Rss      string
	Cookies  string
	Quality  string
	Download string
	Shows    []string
}

type Config struct {
	DB    string
	Tasks map[string]Task
}

func fetchRssItem(item *gofeed.Item, jar http.CookieJar, destination string) {
	client := &http.Client{
		Jar: jar,
	}
	log.Printf("fetching %s", item.Title)
	for _, enclosure := range item.Enclosures {
		response, getErr := client.Get(enclosure.URL)
		if getErr != nil {
			log.Printf("could not get enclosure %s", enclosure.URL)
			continue
		}
		defer response.Body.Close()

		_, fileParams, headerParseErr := mime.ParseMediaType(response.Header.Get("Content-Disposition"))
		if headerParseErr != nil {
			log.Printf("could not get real file name due to %s", headerParseErr)
			continue
		}
		outFilePath := path.Join(destination, path.Base(fileParams["filename"]))
		if _, outFileErr := os.Stat(outFilePath); outFileErr == nil {
			log.Printf("already exists, skipping...")
			continue
		}
		log.Printf("storing %s", outFilePath)
		outFile, fileCreateErr := os.Create(outFilePath)
		if fileCreateErr != nil {
			log.Println("failed to create file")
			continue
		}
		body, readBodyErr := ioutil.ReadAll(response.Body)
		if readBodyErr != nil {
			log.Println("failed to get content")
		}

		outFile.Write(body)
		defer outFile.Close()
	}
}

func main() {
	args := CommandLineArgs{}
	flag.StringVar(&args.configFilePath, "config", ".gorssget.yaml", "Path to config file")
	flag.Parse()

	config := Config{}
	{
		configData, configFileErr := ioutil.ReadFile(args.configFilePath)
		if configFileErr != nil {
			log.Fatalf("could not read config %s due to %s", args.configFilePath, configFileErr)
		}
		configParseErr := yaml.Unmarshal(configData, &config)
		if configParseErr != nil {
			log.Fatalf("could not parse config due to %s", configParseErr)
		}
	}

	log.Printf("loaded config: %s", args.configFilePath)
	log.Printf("tasks count: %d", len(config.Tasks))

	feedParser := gofeed.NewParser()
	for taskName, task := range config.Tasks {
		jar, _ := cookiejar.New(nil)
		if len(task.Cookies) > 0 {
			header := http.Header{}
			header.Add("Cookie", task.Cookies)
			request := http.Request{Header: header}
			parsedRssURL, _ := url.Parse(task.Rss)
			jar.SetCookies(parsedRssURL, request.Cookies())
		}
		log.Printf("executing task: %s", taskName)
		feed, feedParserErr := feedParser.ParseURL(task.Rss)
		if feedParserErr != nil {
			log.Printf("could not parse feed :( due to %s", feedParserErr)
			continue
		}
		for _, item := range feed.Items {
			for _, showToMatch := range task.Shows {
				reString := fmt.Sprintf("\\b(%s)\\b(.*)\\b(%s)\\b", showToMatch, task.Quality)
				re, reCompileErr := regexp.Compile(reString)
				if reCompileErr != nil {
					log.Printf("could not compile regexp expression due to %s", reCompileErr)
					continue
				}
				if re.MatchString(item.Title) {
					fetchRssItem(item, jar, task.Download)
					break
				}
			}
		}
	}
}
