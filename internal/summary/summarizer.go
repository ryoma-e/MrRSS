// Package summary provides text summarization using local algorithms.
// It implements TF-IDF and TextRank-based sentence scoring for extractive summarization.
package summary

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/go-ego/gse"
)

// SummaryLength represents the desired length of the summary
type SummaryLength string

const (
	// Short summary with fewer sentences
	Short SummaryLength = "short"
	// Medium summary with moderate sentences
	Medium SummaryLength = "medium"
	// Long summary with more sentences
	Long SummaryLength = "long"
)

// MinContentLength is the minimum text length required for meaningful summarization
const MinContentLength = 200

// MinSentenceCount is the minimum number of sentences required for summarization
const MinSentenceCount = 3

// Target word counts for different summary lengths
// For Chinese text, each Chinese character is roughly equivalent to one English word
const (
	ShortTargetWords  = 50  // ~50 words or Chinese characters
	MediumTargetWords = 100 // ~100 words or Chinese characters
	LongTargetWords   = 150 // ~150 words or Chinese characters
)

// Global segmenter instance with lazy initialization
var (
	segmenter     gse.Segmenter
	segmenterOnce sync.Once
)

// getSegmenter returns the global segmenter, initializing it if necessary
func getSegmenter() *gse.Segmenter {
	segmenterOnce.Do(func() {
		// Load default dictionary for Chinese segmentation
		segmenter.LoadDict()
	})
	return &segmenter
}

// Summarizer provides text summarization capabilities
type Summarizer struct{}

// NewSummarizer creates a new Summarizer instance
func NewSummarizer() *Summarizer {
	return &Summarizer{}
}

// SummaryResult contains the generated summary and metadata
type SummaryResult struct {
	Summary       string `json:"summary"`
	SentenceCount int    `json:"sentence_count"`
	IsTooShort    bool   `json:"is_too_short"`
}

// Summarize generates a summary of the given text using combined TF-IDF and TextRank scoring
func (s *Summarizer) Summarize(text string, length SummaryLength) SummaryResult {
	// Clean the text
	cleanedText := cleanText(text)

	// Check if text is too short
	if len(cleanedText) < MinContentLength {
		return SummaryResult{
			Summary:    cleanedText,
			IsTooShort: true,
		}
	}

	// Split into sentences
	sentences := splitSentences(cleanedText)

	// Check if we have enough sentences
	if len(sentences) < MinSentenceCount {
		return SummaryResult{
			Summary:       cleanedText,
			SentenceCount: len(sentences),
			IsTooShort:    true,
		}
	}

	// Check if text is primarily Chinese
	isChinese := isChineseText(cleanedText)

	// Get target word/character count based on length setting
	targetCount := getTargetWordCount(length)

	// Score sentences using combined TF-IDF and TextRank
	scoredSentences := s.scoreSentences(sentences)

	// Sort by score (descending)
	sort.Slice(scoredSentences, func(i, j int) bool {
		return scoredSentences[i].score > scoredSentences[j].score
	})

	// Select sentences until we reach the target word/character count
	var selectedSentences []scoredSentence
	currentCount := 0

	for _, sent := range scoredSentences {
		sentCount := countWordsOrChars(sent.text, isChinese)
		if currentCount+sentCount <= targetCount || len(selectedSentences) == 0 {
			selectedSentences = append(selectedSentences, sent)
			currentCount += sentCount
		}
		// Stop if we've reached the target
		if currentCount >= targetCount {
			break
		}
	}

	// Sort by original position to maintain narrative flow
	sort.Slice(selectedSentences, func(i, j int) bool {
		return selectedSentences[i].position < selectedSentences[j].position
	})

	// Build summary
	var summaryParts []string
	for _, sent := range selectedSentences {
		summaryParts = append(summaryParts, sent.text)
	}

	return SummaryResult{
		Summary:       strings.Join(summaryParts, " "),
		SentenceCount: len(selectedSentences),
		IsTooShort:    false,
	}
}

// isChineseText checks if the text is primarily Chinese
func isChineseText(text string) bool {
	chineseCount := 0
	totalCount := 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			totalCount++
			if unicode.Is(unicode.Han, r) {
				chineseCount++
			}
		}
	}
	if totalCount == 0 {
		return false
	}
	return float64(chineseCount)/float64(totalCount) > 0.3
}

// getTargetWordCount returns the target word count based on length setting
func getTargetWordCount(length SummaryLength) int {
	switch length {
	case Short:
		return ShortTargetWords
	case Long:
		return LongTargetWords
	default:
		return MediumTargetWords
	}
}

// countWordsOrChars counts words for English or characters for Chinese
func countWordsOrChars(text string, isChinese bool) int {
	if isChinese {
		// Count Chinese characters
		count := 0
		for _, r := range text {
			if unicode.Is(unicode.Han, r) {
				count++
			}
		}
		// Also count English words in mixed text
		englishWords := 0
		inWord := false
		for _, r := range text {
			if unicode.IsLetter(r) && !unicode.Is(unicode.Han, r) {
				if !inWord {
					englishWords++
					inWord = true
				}
			} else {
				inWord = false
			}
		}
		return count + englishWords
	}
	// Count English words
	words := strings.Fields(text)
	return len(words)
}

// scoredSentence holds a sentence with its calculated score and position
type scoredSentence struct {
	text     string
	score    float64
	position int
}

// scoreSentences calculates scores for each sentence using combined TF-IDF and TextRank
func (s *Summarizer) scoreSentences(sentences []string) []scoredSentence {
	// Calculate TF-IDF scores
	tfidfScores := calculateTFIDF(sentences)

	// Calculate TextRank scores
	textRankScores := calculateTextRank(sentences)

	// Calculate average sentence length for penalty calculation
	totalLen := 0
	for _, sent := range sentences {
		totalLen += len(sent)
	}
	avgLen := float64(totalLen) / float64(len(sentences))

	// Combine scores
	result := make([]scoredSentence, len(sentences))
	for i, sentence := range sentences {
		// Weight TF-IDF at 0.5 and TextRank at 0.5
		combinedScore := 0.5*tfidfScores[i] + 0.5*textRankScores[i]

		// Boost first sentence slightly (often contains key info)
		if i == 0 {
			combinedScore *= 1.15
		}

		// Penalize very long sentences (more than 2x average length)
		// This prevents selecting overly verbose sentences
		sentLen := float64(len(sentence))
		if sentLen > avgLen*2 {
			penalty := avgLen * 2 / sentLen
			combinedScore *= penalty
		}

		// Slight penalty for very short sentences (less than 0.3x average)
		// They often lack sufficient information
		if sentLen < avgLen*0.3 {
			combinedScore *= 0.8
		}

		result[i] = scoredSentence{
			text:     sentence,
			score:    combinedScore,
			position: i,
		}
	}

	return result
}

// calculateTFIDF computes TF-IDF scores for each sentence
func calculateTFIDF(sentences []string) []float64 {
	// Build document frequency map
	docFreq := make(map[string]int)
	allTerms := make([]map[string]int, len(sentences))

	for i, sentence := range sentences {
		terms := tokenize(sentence)
		termFreq := make(map[string]int)
		seenTerms := make(map[string]bool)

		for _, term := range terms {
			termFreq[term]++
			if !seenTerms[term] {
				docFreq[term]++
				seenTerms[term] = true
			}
		}
		allTerms[i] = termFreq
	}

	numDocs := float64(len(sentences))
	scores := make([]float64, len(sentences))

	for i, termFreq := range allTerms {
		var score float64
		totalTerms := 0
		for _, count := range termFreq {
			totalTerms += count
		}

		for term, count := range termFreq {
			// TF: normalized term frequency
			tf := float64(count) / float64(totalTerms)

			// IDF: inverse document frequency
			idf := math.Log(numDocs / float64(docFreq[term]))

			score += tf * idf
		}
		scores[i] = score
	}

	// Normalize scores
	maxScore := 0.0
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}
	if maxScore > 0 {
		for i := range scores {
			scores[i] /= maxScore
		}
	}

	return scores
}

// calculateTextRank computes TextRank scores using sentence similarity
func calculateTextRank(sentences []string) []float64 {
	n := len(sentences)
	if n == 0 {
		return []float64{}
	}

	// Build similarity matrix
	similarity := make([][]float64, n)
	for i := range similarity {
		similarity[i] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			sim := sentenceSimilarity(sentences[i], sentences[j])
			similarity[i][j] = sim
			similarity[j][i] = sim
		}
	}

	// Initialize scores
	scores := make([]float64, n)
	for i := range scores {
		scores[i] = 1.0
	}

	// Damping factor
	d := 0.85
	iterations := 30

	// Iterate to convergence
	for iter := 0; iter < iterations; iter++ {
		newScores := make([]float64, n)
		for i := 0; i < n; i++ {
			sum := 0.0
			for j := 0; j < n; j++ {
				if i != j {
					// Calculate sum of similarities for sentence j
					sumSim := 0.0
					for k := 0; k < n; k++ {
						if k != j {
							sumSim += similarity[j][k]
						}
					}
					if sumSim > 0 {
						sum += similarity[i][j] / sumSim * scores[j]
					}
				}
			}
			newScores[i] = (1 - d) + d*sum
		}
		scores = newScores
	}

	// Normalize scores
	maxScore := 0.0
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}
	if maxScore > 0 {
		for i := range scores {
			scores[i] /= maxScore
		}
	}

	return scores
}

// sentenceSimilarity calculates similarity between two sentences using word overlap
func sentenceSimilarity(s1, s2 string) float64 {
	words1 := tokenize(s1)
	words2 := tokenize(s2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}

	// Create word sets
	set1 := make(map[string]bool)
	for _, w := range words1 {
		set1[w] = true
	}

	set2 := make(map[string]bool)
	for _, w := range words2 {
		set2[w] = true
	}

	// Guard against empty sets to avoid math.Log(0) which returns -Inf
	if len(set1) == 0 || len(set2) == 0 {
		return 0
	}

	// Count common words
	common := 0
	for w := range set1 {
		if set2[w] {
			common++
		}
	}

	// Normalized overlap
	denom := math.Log(float64(len(set1))) + math.Log(float64(len(set2)))
	if denom == 0 {
		return 0
	}

	return float64(common) / denom
}

// cleanText removes HTML tags and normalizes whitespace
func cleanText(text string) string {
	// Remove HTML tags
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	text = htmlRegex.ReplaceAllString(text, " ")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	// Normalize whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// splitSentences splits text into sentences
func splitSentences(text string) []string {
	// Simple sentence splitting with common abbreviations handling
	// Split on sentence-ending punctuation followed by space (or end of text)
	sentenceRegex := regexp.MustCompile(`([.!?。！？]+)(\s+|$)`)

	// Split the text
	parts := sentenceRegex.Split(text, -1)

	// Get the delimiters
	matches := sentenceRegex.FindAllStringSubmatch(text, -1)

	var sentences []string
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Add back the punctuation if available
		if i < len(matches) && len(matches[i]) > 1 {
			part += matches[i][1]
		}

		// Filter out very short sentences (likely fragments)
		// Use a lower threshold to support various languages
		if len(part) > 10 {
			sentences = append(sentences, part)
		}
	}

	return sentences
}

// tokenize splits text into lowercase tokens, removing stopwords
// Uses gse for Chinese word segmentation for better accuracy
func tokenize(text string) []string {
	// Convert to lowercase for English
	text = strings.ToLower(text)

	// Check if text contains Chinese characters
	hasChinese := false
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			hasChinese = true
			break
		}
	}

	var tokens []string

	if hasChinese {
		// Use gse for Chinese text segmentation
		seg := getSegmenter()
		segments := seg.Cut(text, true) // true = search mode for better recall

		for _, word := range segments {
			word = strings.TrimSpace(word)
			// Skip empty strings, stopwords, and very short words
			if len(word) > 0 && !isStopWord(word) {
				// For Chinese, single characters can be meaningful
				// For English, require at least 2 characters
				isChinese := false
				for _, r := range word {
					if unicode.Is(unicode.Han, r) {
						isChinese = true
						break
					}
				}
				if isChinese || len(word) > 2 {
					tokens = append(tokens, word)
				}
			}
		}
	} else {
		// Use simple tokenization for non-Chinese text
		var currentWord strings.Builder

		for _, r := range text {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				currentWord.WriteRune(r)
			} else {
				if currentWord.Len() > 0 {
					word := currentWord.String()
					// Skip stopwords and very short words
					if len(word) > 2 && !isStopWord(word) {
						tokens = append(tokens, word)
					}
					currentWord.Reset()
				}
			}
		}

		// Don't forget the last word
		if currentWord.Len() > 0 {
			word := currentWord.String()
			if len(word) > 2 && !isStopWord(word) {
				tokens = append(tokens, word)
			}
		}
	}

	return tokens
}

// isStopWord checks if a word is a common stopword (English and Chinese)
func isStopWord(word string) bool {
	stopWords := map[string]bool{
		// English stopwords
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
		"are": true, "were": true, "been": true, "be": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true,
		"shall": true, "can": true, "this": true, "that": true, "these": true,
		"those": true, "it": true, "its": true, "they": true, "them": true,
		"their": true, "what": true, "which": true, "who": true, "whom": true,
		"whose": true, "where": true, "when": true, "why": true, "how": true,
		"all": true, "each": true, "every": true, "both": true, "few": true,
		"more": true, "most": true, "other": true, "some": true, "such": true,
		"than": true, "too": true, "very": true, "just": true, "only": true,
		"own": true, "same": true, "so": true, "not": true, "also": true,
		"into": true, "about": true, "your": true, "you": true, "our": true,
		"his": true, "her": true, "my": true, "we": true, "he": true, "she": true,
		"over": true, "out": true, "up": true, "down": true, "then": true, "now": true,
		// Extended Chinese stopwords for better accuracy
		"的": true, "了": true, "和": true, "是": true, "在": true, "有": true,
		"这": true, "个": true, "我": true, "不": true, "人": true, "都": true,
		"一": true, "他": true, "就": true, "们": true, "上": true, "也": true,
		"你": true, "说": true, "着": true, "对": true, "为": true, "与": true,
		"而": true, "等": true, "被": true, "把": true, "让": true, "给": true,
		"向": true, "从": true, "到": true, "之": true, "于": true, "或": true,
		"因": true, "但": true, "却": true, "即": true, "若": true, "虽": true,
		"所": true, "以": true, "如": true, "则": true, "其": true, "它": true,
		"她": true, "这个": true, "那个": true, "什么": true, "怎么": true,
		"为什么": true, "哪个": true, "哪些": true, "这些": true, "那些": true,
		"可以": true, "能够": true, "已经": true, "正在": true, "将要": true,
		"可能": true, "应该": true, "必须": true, "需要": true, "没有": true,
		"因为": true, "所以": true, "但是": true, "而且": true, "或者": true,
		"如果": true, "虽然": true, "然": true, "此": true, "彼": true,
		"自己": true, "我们": true, "你们": true, "他们": true, "它们": true,
		"这里": true, "那里": true, "哪里": true, "任何": true, "某些": true,
		"每个": true, "很": true, "非常": true, "十分": true, "比较": true,
		"更": true, "最": true, "太": true, "又": true, "再": true, "还": true,
	}
	return stopWords[word]
}
