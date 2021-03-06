package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"github.com/ryanbressler/CloudForest"
	"io"
	"log"
	"os"
)

func main() {
	fm := flag.String("fm",
		"featurematrix.afm", "AFM formated feature matrix containing data.")

	targetname := flag.String("target",
		"", "The row header of the target in the feature matrix.")
	train := flag.String("train",
		"train_%v.fm", "Format string for training fms.")
	test := flag.String("test",
		"test_%v.fm", "Format string for testing fms.")

	var zipoutput bool
	flag.BoolVar(&zipoutput, "zip", false, "Output ziped files.")
	var writelibsvm bool
	flag.BoolVar(&writelibsvm, "writelibsvm", false, "Output libsvm.")

	var folds int
	flag.IntVar(&folds, "folds", 5, "Number of folds to generate.")

	flag.Parse()

	//Parse Data
	data, err := CloudForest.LoadAFM(*fm)
	if err != nil {
		log.Fatal(err)
	}

	//find the target feature
	fmt.Printf("Target : %v\n", *targetname)
	targeti, ok := data.Map[*targetname]
	if !ok {
		log.Fatal("Target not found in data.")
	}

	targetf := data.Data[targeti]

	foldis := make([][]int, 0, folds)

	foldsize := len(data.CaseLabels) / folds
	fmt.Printf("%v cases, foldsize %v\n", len(data.CaseLabels), foldsize)
	for i := 0; i < folds; i++ {
		foldis = append(foldis, make([]int, 0, foldsize))
	}

	//sample folds stratified by case
	fmt.Printf("Stratifying by %v classes.\n", targetf.(*CloudForest.DenseCatFeature).NCats())
	bSampler := CloudForest.NewBalancedSampler(targetf.(*CloudForest.DenseCatFeature))

	fmt.Printf("Stratifying by %v classes.\n", len(bSampler.Cases))
	var samples []int
	for i := 0; i < len(bSampler.Cases); i++ {
		fmt.Printf("%v cases in class %v.\n", len(bSampler.Cases[i]), i)
		//shuffle in place
		CloudForest.SampleFirstN(&bSampler.Cases[i], &samples, len(bSampler.Cases[i]), 0)
		stratFoldSize := len(bSampler.Cases[i]) / folds
		for j := 0; j < folds; j++ {
			for k := j * stratFoldSize; k < (j+1)*stratFoldSize; k++ {
				foldis[j] = append(foldis[j], bSampler.Cases[i][k])

			}
		}

	}

	//TODO: unstratified sampeling.

	trainis := make([]int, 0, foldsize*(folds-1))
	//Write training and testing matrixes
	for i := 0; i < folds; i++ {
		trainfn := fmt.Sprintf(*train, i)
		testfn := fmt.Sprintf(*test, i)
		trainfo, err := os.Create(trainfn)
		if err != nil {
			log.Fatal(err)
		}
		testfo, err := os.Create(testfn)
		if err != nil {
			log.Fatal(err)
		}
		var trainW io.Writer = trainfo
		var testW io.Writer = testfo
		if zipoutput {
			trainz := zip.NewWriter(trainfo)
			trainW, err = trainz.Create(trainfn)
			if err != nil {
				log.Fatal(err)
			}
			defer trainz.Close()
			testz := zip.NewWriter(testfo)
			testW, err = testz.Create(testfn)
			if err != nil {
				log.Fatal(err)
			}
			defer testz.Close()
		}

		trainis = trainis[0:0]
		for j := 0; j < folds; j++ {
			if i != j {
				trainis = append(trainis, foldis[j]...)
			}
		}

		if writelibsvm {
			CloudForest.WriteLibSvmCases(data, foldis[i], *targetname, testW)
			CloudForest.WriteLibSvmCases(data, trainis, *targetname, trainW)
		} else {
			data.WriteCases(testW, foldis[i])
			data.WriteCases(trainW, trainis)
		}

		fmt.Printf("Wrote fold %v. %v testing cases and %v training cases.\n", i, len(foldis[i]), len(trainis))
	}

}
