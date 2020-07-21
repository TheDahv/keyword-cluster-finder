# Keyword Cluster Finder

This project is a proof of concept to demonstrate the programmatic
computation of related keywords with respect to search engine results.

## Packages

### data

Manages interacting with the product database to fetch SERP data. If you
don't have access to the product database I'm using, this isn't very
interesting to you.

### graph

Computes the [Markov cluster](https://micans.org/mcl/) from a graph of
keywords with edges weighted by the similarity scores among them.

### rankings

Logic for parsing rankings data -- either from stored JSON files or from a
database -- and combining them into a SERP containing prominent results for a search
as well as the keyword that yielded those results.

### rbo

A Go port of a Python implementation of the rank-biased overlap algorithm
designed for SERP structures defined in the rankings package.

Credit to [dlukes/rbo](https://github.com/dlukes/rbo) for the original
implementation.

## Programs

### build-from-disk

Computes keyword clusters based on SERP data stored in JSON files. It accepts
a single path to a directory containing SERP data. See the program comments
for the required data schema or for sample data to use.

### build-from-db

**Requires access and credentials to the product database.** You probably
don't want to try this if we don't work together.

Computes keyword clustes based on SERP data obtained from the product
database.

It accepts as arguments parameters for both the RBO algorithm for computing
similarity as well as the Markov cluster algorithm to determine neighbor
selection.

For more information about RBO parameters, read
[documentation of an implementation in R](https://rdrr.io/bioc/gespeR/man/rbo.html)
or in the [original paper](http://codalism.com/research/papers/wmz10_tois.pdf).

For more information about Markov cluster parameters, read
["Demystifying Markov Clustering"](https://medium.com/analytics-vidhya/demystifying-markov-clustering-aeb6cdabbfc7#0179).


```
Usage of build-from-db:
  -config string
    	app JSON config
  -domainID int
    	Domain ID
  -inf int
    	Cluster inflation (default 2)
  -p float
    	RBO p value (default 0.9)
  -pow int
    	Cluster power (default 5)
```