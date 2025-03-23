package languagedetector

import (
	"errors"
	"github.com/pemistahl/lingua-go"
	"go.uber.org/zap"
	"indexer/internal/pkg/logger"
	"indexer/internal/pkg/metrics"
)

// Detects the language of a given text and returns the ISO 639-1 code.
func DetectLanguage(languageDetector lingua.LanguageDetector, text string) (string, error) {
    const minTextLength = 20
    if len(text) < minTextLength {
        return "unknown", nil
    }

    // Detect language and calculate confidence values
    detectedLang, exists := languageDetector.DetectLanguageOf(text)
    if !exists {
        metrics.LanguageDetectionFailures.Inc()
        return "", errors.New("language detection failed")
    }

    // Get confidence values for all languages
    confidenceValues := languageDetector.ComputeLanguageConfidenceValues(text)
    var englishConfidence float64

    // Find English confidence value
    for _, conf := range confidenceValues {
        if conf.Language() == lingua.English {
            englishConfidence = conf.Value()
            break
        }
    }

    logger.Log.Debug("Language detection result", 
        zap.String("detected_language", detectedLang.String()),
        zap.Float64("english_confidence", englishConfidence))

	if detectedLang == lingua.English {
		return "en", nil
	} else if englishConfidence > 0.33 {
		return detectedLang.IsoCode639_1().String(), nil
	}

    // If not English or low confidence, skip this document
    metrics.NonEnglishPagesSkipped.Inc()
    return detectedLang.IsoCode639_1().String(), errors.New("not an English page, skipping")
}