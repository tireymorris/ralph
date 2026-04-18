package tui

const (
	minLogLines     = 4
	minMainLines    = 6
	scrollChrome    = 10
	logBiasMaxLines = 24
)

func computePaneHeights(termHeight, logLineCount, logBias int) (mainH, logH int) {
	if termHeight < 10 {
		h := max(4, termHeight/2)
		return h, max(3, termHeight-h)
	}
	avail := termHeight - scrollChrome
	if avail < 12 {
		half := max(4, avail/2)
		return half, avail - half
	}
	logCap := min(max(6, termHeight/3), max(6, avail-minMainLines))
	bias := min(logBiasMaxLines, max(-logBiasMaxLines, logBias))

	var natural int
	if logLineCount <= 0 {
		natural = minLogLines
	} else {
		natural = max(minLogLines, min(logCap, logLineCount+1))
	}

	logH = natural + bias
	if logH < minLogLines {
		logH = minLogLines
	}
	if logH > logCap {
		logH = logCap
	}

	mainH = avail - logH
	if mainH < minMainLines {
		mainH = minMainLines
		logH = avail - mainH
		if logH > logCap {
			logH = logCap
			mainH = avail - logH
		}
		if logH < minLogLines {
			logH = minLogLines
			mainH = max(minMainLines, avail-logH)
		}
	}
	return mainH, logH
}
