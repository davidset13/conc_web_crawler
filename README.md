# Concurrent Web Crawler

* This project is being used to web-crawl data for my [Agnetic Evaluation Server](https://github.com/davidset13/intelligence_eval) project.

* This a lightweight web-crawler that is mostly intended for NLP training purposes, it just returns pure raw text data. This is not a complex web-scraper, it is meant for the purpose of acquiring large swaths of data in a short time.

# Instructions

```bash
cd conc_web_crawler
go run ./src
```

* That's it! Your data will save to a compressed jsonl file, so make sure you have a means of decompression.


