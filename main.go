package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"time"
)

const (
	MassCut = 125.0
)

var (
	g_train = flag.String("train", "training.csv", "training data file")
	g_test  = flag.String("test", "test.csv", "input data file")

	// somewhat arbitrary value, should be optimised
	g_cutoff = flag.Float64("cut", -22.0, "cut-off value")
)

type Score struct {
	Id    int
	Score float64
}

type Scores []Score

func (p Scores) Len() int           { return len(p) }
func (p Scores) Less(i, j int) bool { return p[i].Score < p[j].Score }
func (p Scores) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// AMS returns the Approximate Median Significance for sig and bkg,
// where sig and bkg are the sum of signal and background weights in the
// selection region.
func AMS(sig, bkg float64) float64 {
	if bkg == 0 {
		return 0
	}
	return math.Sqrt(2 * ((sig+bkg+10)*math.Log(1+sig/(bkg+10)) - sig))
}

func main() {
	flag.Parse()

	start := time.Now()

	fmt.Printf("::: read training file [%s]\n", *g_train)

	ftrain, err := os.Open(*g_train)
	if err != nil {
		panic(err)
	}
	defer ftrain.Close()

	fmt.Printf("::: loop on training dataset and compute the score\n")
	evts := make([]Event, 0, 1024)
	dec := NewDecoder(ftrain)
	for {
		i := len(evts)
		evts = append(evts, Event{})
		var evt *Event = &evts[i]
		err = dec.Decode(evt)
		if err != nil {
			evts = evts[:i]
			break
		}
		//fmt.Printf("evt=%d weight=%12.10f label=%s\n", evt.EventId, evt.Weight, evt.Label)
		// this is a simple discriminating variable.
		// Signal should be closer to zero.
		// minus sign so that signal has the highest values
		// so we will be making a simple window cut on the Higgs mass estimator
		// 125 GeV is the middle of the window
		evt.Score = -math.Abs(evt.DER_mass_MMC - MassCut)
	}

	if err != nil && err != io.EOF {
		panic(err)
	}

	cutoff := *g_cutoff

	fmt.Printf("::: loop again to determine the AMS, using threshold=%v\n", cutoff)
	sumsig := 0.0
	sumbkg := 0.0
	for i := range evts {
		evt := &evts[i]
		// sum event weight passing the selection.
		// of course, in real life the threshold should be optimised
		if evt.Score <= cutoff {
			//fmt.Printf(">>> discard evt=%d (score=%v)\n", evt.EventId, evt.Score)
			continue
		}
		switch evt.Label {
		case "s":
			sumsig += evt.Weight
		case "b":
			sumbkg += evt.Weight
		}
	}

	ams := AMS(sumsig, sumbkg)
	fmt.Printf("::: AMS computed from training file=%v (sig=%v, bkg=%v)\n",
		ams,
		sumsig,
		sumbkg,
	)

	fmt.Printf("::: compute the score for the test file entries [%s]\n", *g_test)
	ftest, err := os.Open(*g_test)
	if err != nil {
		panic(err)
	}
	defer ftest.Close()

	dec = NewDecoder(ftest)
	tests := make([]Event, 0, 1024)
	for {
		i := len(tests)
		tests = append(tests, Event{})
		evt := &tests[i]
		err = dec.Decode(evt)
		if err != nil {
			tests = tests[:i]
			break
		}

		evt.Score = -math.Abs(evt.DER_mass_MMC - MassCut)
	}

	if err != nil && err != io.EOF {
		panic(err)
	}

	fmt.Printf("::: loop again on test file to load BDT score pairs\n")
	testpairs := make([]Score, len(tests))
	for i := range tests {
		evt := &tests[i]
		testpairs[i] = Score{evt.EventId, evt.Score}
	}

	fmt.Printf("::: sort on the score\n")
	sort.Sort(Scores(testpairs))

	fmt.Printf("::: build a map key=id, value=rank\n")
	dict := make(map[int]int, len(testpairs))
	for rank, bdt := range testpairs {
		dict[bdt.Id] = rank + 1 // kaggle asks to start at 1
	}

	out, err := os.Create("go.submission_simplest.csv")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// write header
	fmt.Fprintf(out, "EventId,RankOrder,Class\n")

	for i := range tests {
		evt := &tests[i]
		rank, ok := dict[evt.EventId]
		if !ok {
			fmt.Printf("*** evt-id=%d not in map\n", evt.EventId)
			os.Exit(1)
		}
		if rank > len(tests) {
			fmt.Printf("*** large rank=%d for event #%d (id=%d)\n",
				rank, i, evt.EventId,
			)
			break
		}

		// compute label
		label := "b"
		if evt.Score > cutoff {
			label = "s"
		}

		fmt.Fprintf(out, "%d,%d,%s\n", evt.EventId, rank, label)
	}
	err = out.Sync()
	if err != nil {
		panic(err)
	}
	fmt.Printf("::: timing: %v\n", time.Since(start))
	fmt.Printf("::: you can now submit [%s] to Kaggle website\n", out.Name())
	fmt.Printf("::: bye.\n")
}
