package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FeedbackCollector collects and analyzes context effectiveness feedback
type FeedbackCollector interface {
	// CollectImplicitFeedback collects feedback from task execution patterns
	CollectImplicitFeedback(task *Task, context *SelectedContext, executionData *TaskExecutionData) error
	
	// CollectExplicitFeedback collects direct user feedback
	CollectExplicitFeedback(feedback *ExplicitFeedback) error
	
	// AnalyzeFeedbackTrends analyzes patterns in collected feedback
	AnalyzeFeedbackTrends(timeWindow time.Duration) (*FeedbackAnalysis, error)
	
	// GetFeedbackSummary returns summary statistics
	GetFeedbackSummary() *FeedbackSummary
	
	// ExportFeedbackData exports feedback data for external analysis
	ExportFeedbackData(outputPath string) error
}

// TaskExecutionData contains implicit feedback from task execution
type TaskExecutionData struct {
	TaskID              string        `json:"task_id"`
	StartTime           time.Time     `json:"start_time"`
	EndTime             time.Time     `json:"end_time"`
	Duration            time.Duration `json:"duration"`
	TokensConsumed      int           `json:"tokens_consumed"`
	FilesAccessed       []string      `json:"files_accessed"`
	FilesModified       []string      `json:"files_modified"`
	ErrorsEncountered   []string      `json:"errors_encountered"`
	CompletionStatus    string        `json:"completion_status"` // "success", "partial", "failed"
	IterationCount      int           `json:"iteration_count"`
	UserInterventions   int           `json:"user_interventions"`
	MemoryUsage         int64         `json:"memory_usage"`
	CPUUsage            float64       `json:"cpu_usage"`
}

// ExplicitFeedback represents direct user feedback
type ExplicitFeedback struct {
	FeedbackID          string                 `json:"feedback_id"`
	TaskID              string                 `json:"task_id"`
	UserID              string                 `json:"user_id"`
	ContextQuality      int                    `json:"context_quality"`      // 1-5 rating
	RelevanceRating     int                    `json:"relevance_rating"`     // 1-5 rating
	CompletenessRating  int                    `json:"completeness_rating"`  // 1-5 rating
	EfficiencyRating    int                    `json:"efficiency_rating"`    // 1-5 rating
	MissingFiles        []string               `json:"missing_files"`
	IrrelevantFiles     []string               `json:"irrelevant_files"`
	SuggestedFiles      []string               `json:"suggested_files"`
	Comments            string                 `json:"comments"`
	PreferredStrategy   string                 `json:"preferred_strategy"`
	Timestamp           time.Time              `json:"timestamp"`
	AdditionalMetadata  map[string]interface{} `json:"additional_metadata"`
}

// FeedbackAnalysis provides insights from feedback analysis
type FeedbackAnalysis struct {
	TimeWindow          time.Duration          `json:"time_window"`
	TotalSamples        int                    `json:"total_samples"`
	AvgContextQuality   float64                `json:"avg_context_quality"`
	AvgTaskDuration     time.Duration          `json:"avg_task_duration"`
	SuccessRate         float64                `json:"success_rate"`
	TopMissingFiles     []FileRelevanceInfo    `json:"top_missing_files"`
	TopIrrelevantFiles  []FileRelevanceInfo    `json:"top_irrelevant_files"`
	StrategyEffectiveness map[string]float64   `json:"strategy_effectiveness"`
	TaskTypeInsights    map[string]*TaskTypeInsight `json:"task_type_insights"`
	QualityTrends       []QualityDataPoint     `json:"quality_trends"`
	Recommendations     []string               `json:"recommendations"`
}

// FileRelevanceInfo tracks file relevance patterns
type FileRelevanceInfo struct {
	FilePath        string  `json:"file_path"`
	MentionCount    int     `json:"mention_count"`
	AvgRelevance    float64 `json:"avg_relevance"`
	FileType        string  `json:"file_type"`
	Language        string  `json:"language"`
}

// TaskTypeInsight provides insights for specific task types
type TaskTypeInsight struct {
	TaskType            TaskType      `json:"task_type"`
	SampleCount         int           `json:"sample_count"`
	AvgQuality          float64       `json:"avg_quality"`
	AvgDuration         time.Duration `json:"avg_duration"`
	SuccessRate         float64       `json:"success_rate"`
	OptimalTokenBudget  int           `json:"optimal_token_budget"`
	PreferredStrategy   string        `json:"preferred_strategy"`
	CommonMissingFiles  []string      `json:"common_missing_files"`
	PatternObservations []string      `json:"pattern_observations"`
}

// QualityDataPoint represents quality over time
type QualityDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Quality   float64   `json:"quality"`
	Strategy  string    `json:"strategy"`
	TaskType  string    `json:"task_type"`
}

// FeedbackSummary provides high-level feedback statistics
type FeedbackSummary struct {
	TotalFeedbackCount    int                    `json:"total_feedback_count"`
	ImplicitFeedbackCount int                    `json:"implicit_feedback_count"`
	ExplicitFeedbackCount int                    `json:"explicit_feedback_count"`
	AvgUserSatisfaction   float64                `json:"avg_user_satisfaction"`
	MostCommonIssues      []string               `json:"most_common_issues"`
	BestPerformingStrategy string                `json:"best_performing_strategy"`
	WorstPerformingStrategy string               `json:"worst_performing_strategy"`
	RecentTrends          string                 `json:"recent_trends"`
	LastUpdated           time.Time              `json:"last_updated"`
}

// DefaultFeedbackCollector implements feedback collection and analysis
type DefaultFeedbackCollector struct {
	feedbackStore     FeedbackStore
	adaptiveManager   AdaptiveContextManager
	config           *FeedbackConfig
	mutex            sync.RWMutex
	analysisCache    map[string]*FeedbackAnalysis
	cacheExpiry      map[string]time.Time
}

// FeedbackConfig configures feedback collection behavior
type FeedbackConfig struct {
	EnableImplicitCollection bool          `json:"enable_implicit_collection"`
	EnableExplicitCollection bool          `json:"enable_explicit_collection"`
	RetentionDays           int           `json:"retention_days"`
	AnalysisCacheMinutes    int           `json:"analysis_cache_minutes"`
	AutoLearningEnabled     bool          `json:"auto_learning_enabled"`
	FeedbackStorePath       string        `json:"feedback_store_path"`
	MinSamplesForInsights   int           `json:"min_samples_for_insights"`
	QualityThresholds       QualityThresholds `json:"quality_thresholds"`
}

// QualityThresholds define what constitutes good/bad quality
type QualityThresholds struct {
	Excellent float64 `json:"excellent"` // >= 4.5
	Good      float64 `json:"good"`      // >= 3.5
	Fair      float64 `json:"fair"`      // >= 2.5
	Poor      float64 `json:"poor"`      // < 2.5
}

// FeedbackStore handles persistent storage of feedback data
type FeedbackStore interface {
	StoreFeedback(feedback interface{}) error
	GetFeedback(timeWindow time.Duration) ([]interface{}, error)
	GetFeedbackByType(feedbackType string, timeWindow time.Duration) ([]interface{}, error)
	CleanOldFeedback(retentionDays int) error
}

// NewDefaultFeedbackCollector creates a new feedback collector
func NewDefaultFeedbackCollector(store FeedbackStore, adaptiveManager AdaptiveContextManager, config *FeedbackConfig) *DefaultFeedbackCollector {
	if config == nil {
		config = &FeedbackConfig{
			EnableImplicitCollection: true,
			EnableExplicitCollection: true,
			RetentionDays:           90,
			AnalysisCacheMinutes:    15,
			AutoLearningEnabled:     true,
			FeedbackStorePath:       "./feedback_data",
			MinSamplesForInsights:   10,
			QualityThresholds: QualityThresholds{
				Excellent: 4.5,
				Good:      3.5,
				Fair:      2.5,
				Poor:      2.5,
			},
		}
	}

	return &DefaultFeedbackCollector{
		feedbackStore:   store,
		adaptiveManager: adaptiveManager,
		config:         config,
		analysisCache:  make(map[string]*FeedbackAnalysis),
		cacheExpiry:    make(map[string]time.Time),
	}
}

// CollectImplicitFeedback collects feedback from task execution patterns
func (f *DefaultFeedbackCollector) CollectImplicitFeedback(task *Task, context *SelectedContext, executionData *TaskExecutionData) error {
	if !f.config.EnableImplicitCollection {
		return nil
	}

	// Convert execution data to ContextFeedback for the adaptive manager
	feedback := &ContextFeedback{
		TaskID:          executionData.TaskID,
		Task:            task,
		SelectedContext: context,
		TaskSuccess:     executionData.CompletionStatus == "success",
		QualityScore:    f.inferQualityFromExecution(executionData),
		CompletionTime:  executionData.Duration,
		TokensUsed:      executionData.TokensConsumed,
		MissingFiles:    f.inferMissingFiles(executionData, context),
		UnnecessaryFiles: f.inferUnnecessaryFiles(executionData, context),
		UserRating:      0, // No explicit user rating for implicit feedback
		Timestamp:       time.Now(),
	}

	// Store the feedback
	if err := f.feedbackStore.StoreFeedback(feedback); err != nil {
		return fmt.Errorf("failed to store implicit feedback: %w", err)
	}

	// Send to adaptive manager for learning
	if f.config.AutoLearningEnabled && f.adaptiveManager != nil {
		if err := f.adaptiveManager.LearnFromFeedback(feedback); err != nil {
			return fmt.Errorf("failed to send feedback to adaptive manager: %w", err)
		}
	}

	return nil
}

// CollectExplicitFeedback collects direct user feedback
func (f *DefaultFeedbackCollector) CollectExplicitFeedback(feedback *ExplicitFeedback) error {
	if !f.config.EnableExplicitCollection {
		return nil
	}

	// Store the explicit feedback
	if err := f.feedbackStore.StoreFeedback(feedback); err != nil {
		return fmt.Errorf("failed to store explicit feedback: %w", err)
	}

	// Convert to ContextFeedback format for adaptive learning
	if f.config.AutoLearningEnabled && f.adaptiveManager != nil {
		contextFeedback := f.convertExplicitToContextFeedback(feedback)
		if err := f.adaptiveManager.LearnFromFeedback(contextFeedback); err != nil {
			return fmt.Errorf("failed to send explicit feedback to adaptive manager: %w", err)
		}
	}

	return nil
}

// AnalyzeFeedbackTrends analyzes patterns in collected feedback
func (f *DefaultFeedbackCollector) AnalyzeFeedbackTrends(timeWindow time.Duration) (*FeedbackAnalysis, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Check cache first
	cacheKey := fmt.Sprintf("trends_%v", timeWindow)
	if cached, exists := f.analysisCache[cacheKey]; exists {
		if expiry, exists := f.cacheExpiry[cacheKey]; exists && time.Now().Before(expiry) {
			return cached, nil
		}
	}

	// Fetch feedback data
	feedbackData, err := f.feedbackStore.GetFeedback(timeWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feedback data: %w", err)
	}

	// Perform analysis
	analysis := &FeedbackAnalysis{
		TimeWindow:          timeWindow,
		TotalSamples:        len(feedbackData),
		TopMissingFiles:     []FileRelevanceInfo{},
		TopIrrelevantFiles:  []FileRelevanceInfo{},
		StrategyEffectiveness: make(map[string]float64),
		TaskTypeInsights:    make(map[string]*TaskTypeInsight),
		QualityTrends:       []QualityDataPoint{},
		Recommendations:     []string{},
	}

	// Analyze feedback data
	f.analyzeFeedbackData(analysis, feedbackData)

	// Generate recommendations
	f.generateRecommendations(analysis)

	// Cache the analysis
	f.analysisCache[cacheKey] = analysis
	f.cacheExpiry[cacheKey] = time.Now().Add(time.Duration(f.config.AnalysisCacheMinutes) * time.Minute)

	return analysis, nil
}

// GetFeedbackSummary returns summary statistics
func (f *DefaultFeedbackCollector) GetFeedbackSummary() *FeedbackSummary {
	// Get recent feedback for summary
	recentFeedback, _ := f.feedbackStore.GetFeedback(7 * 24 * time.Hour) // Last 7 days

	summary := &FeedbackSummary{
		TotalFeedbackCount:    len(recentFeedback),
		LastUpdated:           time.Now(),
		MostCommonIssues:      []string{},
	}

	// Analyze recent feedback for summary statistics
	f.calculateSummaryStats(summary, recentFeedback)

	return summary
}

// ExportFeedbackData exports feedback data for external analysis
func (f *DefaultFeedbackCollector) ExportFeedbackData(outputPath string) error {
	// Get all feedback data
	allFeedback, err := f.feedbackStore.GetFeedback(365 * 24 * time.Hour) // Last year
	if err != nil {
		return fmt.Errorf("failed to fetch feedback for export: %w", err)
	}

	// Create export directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Export as JSON
	data, err := json.MarshalIndent(allFeedback, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal feedback data: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// Helper methods for feedback analysis

func (f *DefaultFeedbackCollector) inferQualityFromExecution(data *TaskExecutionData) float64 {
	baseQuality := 0.5

	// Success/failure impact
	switch data.CompletionStatus {
	case "success":
		baseQuality = 0.8
	case "partial":
		baseQuality = 0.5
	case "failed":
		baseQuality = 0.2
	}

	// Duration impact (shorter is generally better, but not always)
	if data.Duration < 5*time.Minute {
		baseQuality += 0.1 // Quick completion is good
	} else if data.Duration > 30*time.Minute {
		baseQuality -= 0.2 // Very long tasks suggest context issues
	}

	// Error count impact
	if len(data.ErrorsEncountered) == 0 {
		baseQuality += 0.1
	} else {
		baseQuality -= float64(len(data.ErrorsEncountered)) * 0.05
	}

	// Iteration count impact
	if data.IterationCount > 5 {
		baseQuality -= 0.1 // Too many iterations suggest poor initial context
	}

	// User intervention impact
	if data.UserInterventions > 3 {
		baseQuality -= 0.15 // Frequent interventions suggest context gaps
	}

	// Ensure quality is between 0 and 1
	if baseQuality < 0 {
		baseQuality = 0
	}
	if baseQuality > 1 {
		baseQuality = 1
	}

	return baseQuality
}

func (f *DefaultFeedbackCollector) inferMissingFiles(data *TaskExecutionData, context *SelectedContext) []string {
	// Files that were accessed but not included in context
	contextFiles := make(map[string]bool)
	for _, file := range context.Files {
		contextFiles[file.FileInfo.Path] = true
	}

	missingFiles := []string{}
	for _, accessedFile := range data.FilesAccessed {
		if !contextFiles[accessedFile] {
			missingFiles = append(missingFiles, accessedFile)
		}
	}

	return missingFiles
}

func (f *DefaultFeedbackCollector) inferUnnecessaryFiles(data *TaskExecutionData, context *SelectedContext) []string {
	// Files that were included in context but never accessed
	accessedFiles := make(map[string]bool)
	for _, file := range data.FilesAccessed {
		accessedFiles[file] = true
	}

	unnecessaryFiles := []string{}
	for _, contextFile := range context.Files {
		if !accessedFiles[contextFile.FileInfo.Path] {
			unnecessaryFiles = append(unnecessaryFiles, contextFile.FileInfo.Path)
		}
	}

	return unnecessaryFiles
}

func (f *DefaultFeedbackCollector) convertExplicitToContextFeedback(explicit *ExplicitFeedback) *ContextFeedback {
	// Convert 1-5 rating to 0-1 quality score
	qualityScore := float64(explicit.ContextQuality-1) / 4.0

	return &ContextFeedback{
		TaskID:           explicit.TaskID,
		TaskSuccess:      explicit.ContextQuality >= 3, // 3+ out of 5 is success
		QualityScore:     qualityScore,
		MissingFiles:     explicit.MissingFiles,
		UnnecessaryFiles: explicit.IrrelevantFiles,
		UserRating:       qualityScore,
		Timestamp:        explicit.Timestamp,
	}
}

func (f *DefaultFeedbackCollector) analyzeFeedbackData(analysis *FeedbackAnalysis, feedbackData []interface{}) {
	qualitySum := 0.0
	successCount := 0
	
	for _, data := range feedbackData {
		switch feedback := data.(type) {
		case *ContextFeedback:
			qualitySum += feedback.QualityScore
			if feedback.TaskSuccess {
				successCount++
			}
			
			// Track quality trends
			analysis.QualityTrends = append(analysis.QualityTrends, QualityDataPoint{
				Timestamp: feedback.Timestamp,
				Quality:   feedback.QualityScore,
				Strategy:  string(feedback.SelectedContext.Strategy),
				TaskType:  string(feedback.Task.Type),
			})
		}
	}

	if len(feedbackData) > 0 {
		analysis.AvgContextQuality = qualitySum / float64(len(feedbackData))
		analysis.SuccessRate = float64(successCount) / float64(len(feedbackData))
	}
}

func (f *DefaultFeedbackCollector) generateRecommendations(analysis *FeedbackAnalysis) {
	recommendations := []string{}

	// Quality-based recommendations
	if analysis.AvgContextQuality < f.config.QualityThresholds.Fair {
		recommendations = append(recommendations, "Context quality is below threshold - consider adjusting selection strategies")
	}

	// Success rate recommendations
	if analysis.SuccessRate < 0.7 {
		recommendations = append(recommendations, "Low success rate detected - review task-specific context patterns")
	}

	// Sample size recommendations
	if analysis.TotalSamples < f.config.MinSamplesForInsights {
		recommendations = append(recommendations, "Insufficient feedback samples for reliable insights - continue collecting data")
	}

	analysis.Recommendations = recommendations
}

func (f *DefaultFeedbackCollector) calculateSummaryStats(summary *FeedbackSummary, feedbackData []interface{}) {
	qualitySum := 0.0
	qualityCount := 0

	for _, data := range feedbackData {
		switch feedback := data.(type) {
		case *ContextFeedback:
			summary.ImplicitFeedbackCount++
			qualitySum += feedback.QualityScore
			qualityCount++
		case *ExplicitFeedback:
			summary.ExplicitFeedbackCount++
			qualitySum += float64(feedback.ContextQuality-1) / 4.0
			qualityCount++
		}
	}

	if qualityCount > 0 {
		summary.AvgUserSatisfaction = qualitySum / float64(qualityCount)
	}

	// Determine recent trends
	if summary.AvgUserSatisfaction >= 0.8 {
		summary.RecentTrends = "Improving"
	} else if summary.AvgUserSatisfaction >= 0.6 {
		summary.RecentTrends = "Stable"
	} else {
		summary.RecentTrends = "Declining"
	}
}

// SimpleFeedbackStore provides basic file-based feedback storage
type SimpleFeedbackStore struct {
	storePath string
	mutex     sync.RWMutex
}

// NewSimpleFeedbackStore creates a new file-based feedback store
func NewSimpleFeedbackStore(storePath string) *SimpleFeedbackStore {
	return &SimpleFeedbackStore{
		storePath: storePath,
	}
}

// StoreFeedback stores feedback to a JSON file
func (s *SimpleFeedbackStore) StoreFeedback(feedback interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create store directory if it doesn't exist
	if err := os.MkdirAll(s.storePath, 0755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("feedback_%s_%d.json", timestamp, time.Now().UnixNano())
	filepath := filepath.Join(s.storePath, filename)

	// Marshal and write feedback
	data, err := json.MarshalIndent(feedback, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal feedback: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write feedback file: %w", err)
	}

	return nil
}

// GetFeedback retrieves feedback within a time window
func (s *SimpleFeedbackStore) GetFeedback(timeWindow time.Duration) ([]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cutoff := time.Now().Add(-timeWindow)
	feedback := []interface{}{}

	// Read all feedback files
	files, err := filepath.Glob(filepath.Join(s.storePath, "feedback_*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list feedback files: %w", err)
	}

	for _, file := range files {
		// Check file modification time
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			continue
		}

		// Read and unmarshal feedback
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var feedbackItem interface{}
		if err := json.Unmarshal(data, &feedbackItem); err != nil {
			continue
		}

		feedback = append(feedback, feedbackItem)
	}

	return feedback, nil
}

// GetFeedbackByType retrieves feedback of a specific type
func (s *SimpleFeedbackStore) GetFeedbackByType(feedbackType string, timeWindow time.Duration) ([]interface{}, error) {
	// For this simple implementation, return all feedback
	// A more sophisticated implementation would filter by type
	return s.GetFeedback(timeWindow)
}

// CleanOldFeedback removes feedback older than retention days
func (s *SimpleFeedbackStore) CleanOldFeedback(retentionDays int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	files, err := filepath.Glob(filepath.Join(s.storePath, "feedback_*.json"))
	if err != nil {
		return fmt.Errorf("failed to list feedback files for cleanup: %w", err)
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(file) // Ignore errors for cleanup
		}
	}

	return nil
}