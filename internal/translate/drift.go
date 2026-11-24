package translate

import (
	"math"
	"unicode/utf8"
)

// A correct translation's cue lengths track the original's. Two signs that they no longer do:
// lines slipped by one cue track a neighbour better than their own (caught a shifted WebRTC span);
// lines merged or dropped track nothing, which only means something when the original lengths vary.
// The second sign caught Claude Sonnet 5.5 drifting cues 95-120 of a lecture, and flags none of 58 correct files.
const (
	driftWindow                = 30
	driftStep                  = 10
	driftMinPairs              = driftWindow / 2
	driftMinAdvantage          = 0.2
	driftMinAlignedCorrelation = 0.4
	driftMinLengthVariation    = 0.4
)

type span struct {
	from, to int
}

func driftSpans(original, translated []string) []span {
	sourceLengths, translatedLengths := runeLengths(original), runeLengths(translated)
	var spans []span
	for _, start := range windowStarts(len(sourceLengths)) {
		if isDrifting(sourceLengths, translatedLengths, start) {
			spans = append(spans, span{from: start, to: min(start+driftWindow, len(sourceLengths))})
		}
	}
	return spans
}

// windowStarts steps through the cues and adds a last window ending at the final cue, so the tail is always checked.
func windowStarts(cues int) []int {
	var starts []int
	for start := 0; start < max(1, cues-driftWindow+1); start += driftStep {
		starts = append(starts, start)
	}
	if tail := cues - driftWindow; tail > starts[len(starts)-1] {
		starts = append(starts, tail)
	}
	return starts
}

func isDrifting(source, translated []float64, start int) bool {
	aligned, isDefined := windowCorrelation(source, translated, start, 0)
	previous, _ := windowCorrelation(source, translated, start, -1)
	next, _ := windowCorrelation(source, translated, start, 1)
	if max(previous, next)-aligned > driftMinAdvantage {
		return true
	}
	window := source[start:min(start+driftWindow, len(source))]
	return isDefined && aligned < driftMinAlignedCorrelation && variation(window) >= driftMinLengthVariation
}

// windowCorrelation is 0 and not defined when the window is too short or either side is constant.
func windowCorrelation(source, translated []float64, start, shift int) (float64, bool) {
	var xs, ys []float64
	for position := start; position < min(start+driftWindow, len(source)); position++ {
		if neighbour := position + shift; neighbour >= 0 && neighbour < len(translated) {
			xs, ys = append(xs, source[position]), append(ys, translated[neighbour])
		}
	}
	if len(xs) < driftMinPairs {
		return 0, false
	}
	return pearson(xs, ys)
}

func pearson(xs, ys []float64) (float64, bool) {
	meanX, meanY := mean(xs), mean(ys)
	var covariance, varianceX, varianceY float64
	for position := range xs {
		dx, dy := xs[position]-meanX, ys[position]-meanY
		covariance += dx * dy
		varianceX += dx * dx
		varianceY += dy * dy
	}
	if varianceX == 0 || varianceY == 0 {
		return 0, false
	}
	return covariance / math.Sqrt(varianceX*varianceY), true
}

// variation is the coefficient of variation: how much lengths differ, relative to their mean.
func variation(values []float64) float64 {
	average := mean(values)
	if average == 0 {
		return 0
	}
	var squares float64
	for _, value := range values {
		squares += (value - average) * (value - average)
	}
	return math.Sqrt(squares/float64(len(values))) / average
}

func mean(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func runeLengths(texts []string) []float64 {
	lengths := make([]float64, len(texts))
	for position, text := range texts {
		lengths[position] = float64(utf8.RuneCountInString(singleLine(text)))
	}
	return lengths
}
