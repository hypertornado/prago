package pragosearch

import "math"

func calculateBM25Score(termFrequencyInDoc int64, docLength int64, avgDocLength float64, idf float64) float64 {
	//elasticsearch defaults
	var k1 float64 = 1.2
	var b float64 = 0.75

	nominator := float64(termFrequencyInDoc) * (k1 + 1)
	denominator := float64(docLength) / avgDocLength
	denominator = denominator * b
	denominator = denominator + 1 - b
	denominator = k1 * denominator
	denominator = denominator + float64(termFrequencyInDoc)

	return idf * (nominator / denominator)
}

func calculateIDF(documentsContainingQ, totalDocuments int64) float64 {
	ret := float64(totalDocuments-documentsContainingQ) + 0.5
	ret = ret / (float64(documentsContainingQ) + 0.5)
	ret += 1
	return math.Log(ret)

}
