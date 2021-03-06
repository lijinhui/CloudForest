package CloudForest

import (
	"bufio"
	"encoding/csv"
	"io"
	"log"
	"strings"
)

func ParseARFF(input io.Reader) *FeatureMatrix {

	reader := bufio.NewReader(input)

	data := make([]Feature, 0, 100)
	lookup := make(map[string]int, 0)
	labels := make([]string, 0, 0)

	i := 0
	for {

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Print("Error:", err)
			return nil
		}
		norm := strings.ToLower(line)

		if strings.HasPrefix(norm, "@data") {
			break
		}

		if strings.HasPrefix(norm, "@attribute") {
			vals := strings.Fields(line)

			if strings.ToLower(vals[2]) == "numeric" || strings.ToLower(vals[2]) == "real" {
				data = append(data, &DenseNumFeature{
					make([]float64, 0, 0),
					make([]bool, 0, 0),
					vals[1],
					false})
			} else {
				data = append(data, &DenseCatFeature{
					&CatMap{make(map[string]int, 0),
						make([]string, 0, 0)},
					make([]int, 0, 0),
					make([]bool, 0, 0),
					vals[1],
					false,
					false})
			}

			lookup[vals[1]] = i
			labels = append(labels, vals[1])
			i++
		}

	}

	fm := &FeatureMatrix{data, lookup, labels}

	csvdata := csv.NewReader(reader)
	csvdata.Comment = '%'
	//csvdata.Comma = ','

	fm.LoadCases(csvdata, false)
	return fm

}
