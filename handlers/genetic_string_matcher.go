package handlers

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"time"

	ga "github.com/tomcraven/goga"
)

type stringMaterSimulator struct {
}

func (sms *stringMaterSimulator) OnBeginSimulation() {
}
func (sms *stringMaterSimulator) OnEndSimulation() {
}
func (sms *stringMaterSimulator) Simulate(g *ga.IGenome) {
	bits := (*g).GetBits()
	for i, character := range targetString {
		for j := 0; j < 8; j++ {
			targetBit := character & (1 << uint(j))
			bit := bits.Get((i * 8) + j)
			if targetBit != 0 && bit == 1 {
				(*g).SetFitness((*g).GetFitness() + 1)
			} else if targetBit == 0 && bit == 0 {
				(*g).SetFitness((*g).GetFitness() + 1)
			}
		}
	}
}
func (sms *stringMaterSimulator) ExitFunc(g *ga.IGenome) bool {
	totalFitness = (*g).GetFitness()
	if !limitHit {
		return (*g).GetFitness() == targetLength
	} else {
		return true
	}
}

type myBitsetCreate struct {
}

func (bc *myBitsetCreate) Go() ga.Bitset {
	b := ga.Bitset{}
	b.Create(targetLength)
	for i := 0; i < targetLength; i++ {
		b.Set(i, rand.Intn(2))
	}
	return b
}

type myEliteConsumer struct {
	currentIter int
}

func (ec *myEliteConsumer) OnElite(g *ga.IGenome) {
	gBits := (*g).GetBits()
	ec.currentIter++
	var genomeString string
	var genomeBitSet []byte
	for i := 0; i < gBits.GetSize(); i += 8 {
		c := int(0)
		for j := 0; j < 8; j++ {
			bit := gBits.Get(i + j)
			if bit != 0 {
				c |= 1 << uint(j)
			}
		}
		genomeString += string(c)
		genomeBitSet = append(genomeBitSet, byte(c))
	}

	fmt.Println(ec.currentIter, "\t", genomeString, "\t", genomeBitSet, "\t", (*g).GetFitness())
	totalIterations = ec.currentIter
	if ec.currentIter >= iterLimit {
		limitHit = true
	}
}

var (
	populationSize  = 1
	populationCap   = 2000
	iterLimit       = 5000
	totalFitness    = 0
	totalIterations = 0
)

var (
	targetString = "some random string"
	targetLength int
	limitHit     = false
)

func getMaxStringLength(maxStringLength int) int {
	eightyPercent := int(math.Ceil(float64(maxStringLength) * 0.8))
	numArr := rand.Perm(maxStringLength)
	stringLength := 0
	for _, number := range numArr {
		if stringLength == 0 {
			if number >= eightyPercent {
				stringLength = number
				return stringLength
			}
		}
	}
	return stringLength
}

const corpus = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRandomString(maxStringLength int) string {
	stringLength := getMaxStringLength(maxStringLength)
	charList := make([]byte, stringLength)
	for i := range charList {
		charList[i] = corpus[rand.Intn(len(charList))]
	}
	return string(charList)
}

func hackSimulate(seed int64, popSize int, iterationLimit int, maxStringLength int) (float64, int) {
	// make sure caps are respected
	if popSize > populationCap {
		popSize = populationCap
	}
	populationSize = popSize
	if iterationLimit > iterLimit {
		iterationLimit = iterLimit
	}
	iterLimit = iterationLimit
	// set the rand seed
	rand.Seed(seed)
	// may need to impose hard limit on inputString length
	targetString = generateRandomString(maxStringLength)
	// multiply length of string by 8 (b/c we're finding length of all bytes in the string)
	targetLength = len(targetString) * 8
	numThreads := 4
	runtime.GOMAXPROCS(numThreads)

	genAlgo := ga.NewGeneticAlgorithm()

	genAlgo.Simulator = &stringMaterSimulator{}
	genAlgo.BitsetCreate = &myBitsetCreate{}
	genAlgo.EliteConsumer = &myEliteConsumer{}
	genAlgo.Mater = ga.NewMater(
		[]ga.MaterFunctionProbability{
			{P: 1.0, F: ga.TwoPointCrossover},
			{P: 1.0, F: ga.Mutate},
			{P: 1.0, F: ga.UniformCrossover, UseElite: true},
		},
	)
	genAlgo.Selector = ga.NewSelector(
		[]ga.SelectorFunctionProbability{
			{P: 1.0, F: ga.Roulette},
		},
	)

	genAlgo.Init(populationSize, numThreads)

	startTime := time.Now()
	genAlgo.Simulate()
	// reset the seed
	rand.Seed(time.Now().UnixNano())
	fmt.Println(time.Since(startTime))
	return (float64(totalFitness) / float64(targetLength)), totalIterations
}