package spamdetector

import (
    "strings"
    "github.com/cloudflare/ahocorasick"  // Efficient Aho-Corasick implementation
    "go.uber.org/zap"
    "indexer/internal/pkg/logger"
)

// Detects spam content using Aho-Corasick algorithm
type SpamDetector struct {
    matcher       *ahocorasick.Matcher
    spamPhrases   []string
    phraseScores  map[string]int  // Different phrases can have different weights
    blockThreshold int            // Pages with scores above this are rejected
}

// Contains spam detection results
type SpamResult struct {
    Score       int            // Overall spam score
    IsHighSpam  bool           // Whether content exceeds block threshold
}

// Creates a new detector with the given spam phrases
func NewSpamDetector(blockThreshold int) *SpamDetector {
    // Convert phrases to byte slices for the Aho-Corasick matcher
    patterns := make([][]byte, len(spamPhrases))
    for i, phrase := range spamPhrases {
        patterns[i] = []byte(strings.ToLower(phrase))
    }
    
    // Set default weights for phrases without explicit weights
    phraseScores := make(map[string]int)
    for _, phrase := range spamPhrases {
        if weight, exists := weights[phrase]; exists {
            phraseScores[phrase] = weight
        } else {
            phraseScores[phrase] = 1 // Default weight
        }
    }
    
    logger.Log.Info("Initializing spam detector", 
        zap.Int("phrase_count", len(spamPhrases)), 
        zap.Int("block_threshold", blockThreshold))
    
    return &SpamDetector{
        matcher:       	ahocorasick.NewMatcher(patterns),
        spamPhrases:   	spamPhrases,
        phraseScores:  	phraseScores,
        blockThreshold: blockThreshold,
    }
}

// Analyzes text for spam content
func (sd *SpamDetector) DetectSpam(text string) SpamResult {
    if text == "" {
        return SpamResult{
            Score:      0,
            IsHighSpam: false,
        }
    }
    
    // Convert to lowercase for case-insensitive matching
    lowerText := strings.ToLower(text)
    textBytes := []byte(lowerText)
    
    // Calculate text length for density calculations
    textLength := len([]rune(text))
    
    // Find all matches using Aho-Corasick
    hits := sd.matcher.Match(textBytes)
    
    // Calculate spam score and match counts
    totalScore := 0
	
	// Calculate spam score based on matched phrases
    for _, hit := range hits {
        totalScore += sd.phraseScores[sd.spamPhrases[hit]]
    }
    
    // Adjust score based on text length (longer legitimate content dilutes spam)
    if textLength > 0 && len(hits) > 0 {
        // Apply a small normalization factor for very long content
        if textLength > 5000 {
            totalScore = (totalScore * 5000) / textLength
        }
    }
    
    // Check if this is high-spam content that should be blocked
    isHighSpam := totalScore >= sd.blockThreshold
    
    return SpamResult{
        Score:      totalScore,
        IsHighSpam: isHighSpam,
    }
}