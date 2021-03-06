go-higgsml
==========

`go-higgsml` is a simple starting-kit for the HiggsML challenge:
  http://higgsml.lal.in2p3.fr/

(this `go-higgsml` kit is a transcription of the `python` one from
[here](http://higgsml.lal.in2p3.fr/software/simplest-python-kit/))

## Installation

```sh
$ go get github.com/sbinet/go-higgsml
```

## Example

Once you've downloaded the `test.csv` and `training.csv` files from
[Kaggle](https://www.kaggle.com/c/connectomics/data):

```sh
$ go-higgsml -train training.csv trained.dat
::: read training file [training.csv]
::: loop on training dataset and compute the score
::: loop again to determine the AMS, using threshold=-22
::: AMS computed from training file=1.5289067550142865 (sig=457.2791382866634, bkg=89291.91212537605)
::: timing: 8.482504175s
::: bye.

$ go-higgsml test.csv trained.dat scores_test.csv
::: compute the score for the test file entries [test.csv]
::: loop again on test file to load BDT score pairs
::: sort on the score
::: build a map key=id, value=rank
::: you can now submit [scores_test.csv] to Kaggle website
::: timing: 22.022769973s
::: bye.
```
