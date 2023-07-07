stock-api
---------

Full disclosure, ChatGPT helped me write this. I basically laid out the design and then
the AI just whipped it up. It wasn't totally usable out of the box, for example it
was inserting the entire json string into the redis cache as a value needlessly
but all I had to do was use value.Price from the redis Get call and it worked
pretty much flawlessly after that...

I also asked ChatGPT to write the helm chart and again it mostly did the work
even after I asked it to add and Environment variable for `REDIS_HOST`.

For the stock ticker data, I asked it where it recomended I get this data for free
and somehow after I decided we'd use Yahoo Finance it came up with this RapidAPI
endpoint that I had to create a free account with and it gave me a token after
setting up the apidojo Yahoo Finance API with RapidAPI, again all for free...

Changelog
---------
7-7-2023 Switched from rapidapi to chromedp and browser scraping
