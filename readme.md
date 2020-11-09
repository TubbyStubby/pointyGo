## About
RestAPI written in Go. Provides snippets of stackoverflow questions tagged with Go from a set of 100 questions.


## Endpoints
```
GET /articles                           - lists all the articles

GET /articles?limit=<int>&offset=<int>  - Used for paging. Where limit is articles per page and offset/limit is page number

POST /articles                          - using a json body post an article

GET /articles/<id>                      - fetch a particular article

GET /articles/search?q=<query terms>    - searches terms in title, subtitle, content
```
