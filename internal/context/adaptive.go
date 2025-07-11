package context

import (
	"context"
	"fmt"
	"time"
)

// AdaptiveContextManager provides intelligent, task-aware context management
type AdaptiveContextManager interface {
	// AdaptOptimalContext adapts context selection based on task characteristics and feedback
	AdaptOptimalContext(ctx context.Context, project *ProjectContext, task *Task, budget int) (*AdaptedContext, error)
	
	// LearnFromFeedback incorporates feedback to improve future selections
	LearnFromFeedback(feedback *ContextFeedback) error
	
	// GetAdaptiveConstraints returns task-optimized constraints
	GetAdaptiveConstraints(task *Task, budget int, projectCtx *ProjectContext) *ContextConstraints
	
	// PredictOptimalBudget suggests optimal token budget for a task
	PredictOptimalBudget(task *Task, projectCtx *ProjectContext) int
}

// AdaptedContext extends SelectedContext with adaptive features
type AdaptedContext struct {
	*SelectedContext
	AdaptationReasons []string              `json:"adaptation_reasons"`
	BudgetAdjustment  int                   `json:"budget_adjustment"`
	StrategyOverride  *SelectionStrategy    `json:"strategy_override,omitempty"`
	QualityPrediction float64               `json:"quality_prediction"`
	AdaptiveMetadata  map[string]interface{} `json:"adaptive_metadata"`
}

// ContextFeedback provides learning data for the adaptive system
type ContextFeedback struct {
	TaskID           string                `json:"task_id"`
	Task             *Task                 `json:"task"`
	SelectedContext  *SelectedContext      `json:"selected_context"`
	TaskSuccess      bool                  `json:"task_success"`
	QualityScore     float64               `json:"quality_score"`     // 0-1 rating of result quality
	CompletionTime   time.Duration         `json:"completion_time"`
	TokensUsed       int                   `json:"tokens_used"`
	MissingFiles     []string              `json:"missing_files"`     // Files that should have been included
	UnnecessaryFiles []string              `json:"unnecessary_files"` // Files that weren't needed
	UserRating       float64               `json:"user_rating"`       // Optional user feedback
	Timestamp        time.Time             `json:"timestamp"`
}

// TaskProfile represents learned characteristics for different task types
type TaskProfile struct {
	TaskType              TaskType               `json:"task_type"`
	OptimalTokenBudget    int                    `json:"optimal_token_budget"`
	PreferredStrategy     SelectionStrategy      `json:"preferred_strategy"`
	ImportantFileTypes    []string               `json:"important_file_types"`
	TypicalFileCount      int                    `json:"typical_file_count"`
	AvgQualityScore       float64                `json:"avg_quality_score"`
	SuccessRate           float64                `json:"success_rate"`
	AdaptationFactors     map[string]float64     `json:"adaptation_factors"`
	LastUpdated           time.Time              `json:"last_updated"`
	SampleCount           int                    `json:"sample_count"`
}

// DefaultAdaptiveManager implements adaptive context management
type DefaultAdaptiveManager struct {
	optimizer     ContextOptimizer
	analyzer      ContextAnalyzer
	cache         ContextCache
	profiles      map[TaskType]*TaskProfile
	feedbackLog   []ContextFeedback
	config        *AdaptiveConfig
}

// AdaptiveConfig configures the adaptive context manager
type AdaptiveConfig struct {
	LearningRate          float64       `json:"learning_rate"`
	MinSamplesForAdaptation int         `json:"min_samples_for_adaptation"`
	FeedbackRetentionDays   int         `json:"feedback_retention_days"`
	EnableBudgetAdaptation  bool        `json:"enable_budget_adaptation"`
	EnableStrategyAdaptation bool       `json:"enable_strategy_adaptation"`
	QualityThreshold        float64     `json:"quality_threshold"`
	MaxBudgetAdjustment     int         `json:"max_budget_adjustment"`
	AdaptationAggressiveness float64    `json:"adaptation_aggressiveness"`
}

// NewDefaultAdaptiveManager creates a new adaptive context manager
func NewDefaultAdaptiveManager(optimizer ContextOptimizer, analyzer ContextAnalyzer, cache ContextCache, config *AdaptiveConfig) *DefaultAdaptiveManager {
	if config == nil {
		config = &AdaptiveConfig{
			LearningRate:             0.1,
			MinSamplesForAdaptation:  5,
			FeedbackRetentionDays:    30,
			EnableBudgetAdaptation:   true,
			EnableStrategyAdaptation: true,
			QualityThreshold:         0.7,
			MaxBudgetAdjustment:      4000,
			AdaptationAggressiveness: 0.5,
		}
	}

	return &DefaultAdaptiveManager{
		optimizer:   optimizer,
		analyzer:    analyzer,
		cache:       cache,
		profiles:    make(map[TaskType]*TaskProfile),
		feedbackLog: []ContextFeedback{},
		config:      config,
	}
}

// AdaptOptimalContext provides adaptive context selection
func (m *DefaultAdaptiveManager) AdaptOptimalContext(ctx context.Context, project *ProjectContext, task *Task, budget int) (*AdaptedContext, error) {
	adaptedBudget := budget
	adaptationReasons := []string{}
	var strategyOverride *SelectionStrategy

	// Get or create task profile
	profile := m.getOrCreateTaskProfile(task.Type)

	// Adapt budget based on learning
	if m.config.EnableBudgetAdaptation && profile.SampleCount >= m.config.MinSamplesForAdaptation {
		if profile.OptimalTokenBudget > 0 {
			budgetAdjustment := int(float64(profile.OptimalTokenBudget-budget) * m.config.AdaptationAggressiveness)
			
			// Limit adjustment magnitude
			if budgetAdjustment > m.config.MaxBudgetAdjustment {
				budgetAdjustment = m.config.MaxBudgetAdjustment
			} else if budgetAdjustment < -m.config.MaxBudgetAdjustment {
				budgetAdjustment = -m.config.MaxBudgetAdjustment
			}

			if budgetAdjustment != 0 {
				adaptedBudget = budget + budgetAdjustment
				adaptationReasons = append(adaptationReasons, 
					fmt.Sprintf("Budget adjusted by %d based on learned optimal budget of %d", 
						budgetAdjustment, profile.OptimalTokenBudget))
			}
		}
	}

	// Adapt strategy based on success patterns
	if m.config.EnableStrategyAdaptation && profile.SampleCount >= m.config.MinSamplesForAdaptation {
		if profile.SuccessRate > m.config.QualityThreshold && profile.PreferredStrategy != "" {
			strategyOverride = &profile.PreferredStrategy
			adaptationReasons = append(adaptationReasons, 
				fmt.Sprintf("Strategy overridden to '%s' (%.1f%% success rate)", 
					profile.PreferredStrategy, profile.SuccessRate*100))
		}
	}

	// Get adaptive constraints
	constraints := m.GetAdaptiveConstraints(task, adaptedBudget, project)
	if strategyOverride != nil {
		constraints.Strategy = *strategyOverride
	}

	// Apply task-specific adaptations
	m.applyTaskSpecificAdaptations(constraints, task, profile, project)
	
	// Perform context selection
	selectedContext, err := m.optimizer.SelectOptimalContext(ctx, project, task, constraints)
	if err != nil {
		return nil, err
	}

	// Predict quality based on historical data
	qualityPrediction := m.predictQuality(selectedContext, task, profile)

	// Create adapted context
	adapted := &AdaptedContext{
		SelectedContext:   selectedContext,
		AdaptationReasons: adaptationReasons,
		BudgetAdjustment:  adaptedBudget - budget,
		StrategyOverride:  strategyOverride,
		QualityPrediction: qualityPrediction,
		AdaptiveMetadata: map[string]interface{}{
			"profile_samples":    profile.SampleCount,
			"profile_success":    profile.SuccessRate,
			"optimal_budget":     profile.OptimalTokenBudget,
			"preferred_strategy": profile.PreferredStrategy,
		},
	}

	return adapted, nil
}

// GetAdaptiveConstraints returns task-optimized constraints
func (m *DefaultAdaptiveManager) GetAdaptiveConstraints(task *Task, budget int, projectCtx *ProjectContext) *ContextConstraints {
	profile := m.getOrCreateTaskProfile(task.Type)
	
	constraints := &ContextConstraints{
		MaxTokens:         budget,
		MaxFiles:          50,
		MinRelevanceScore: 0.1,
		PreferredTypes:    []string{"source"},
		IncludeTests:      false,
		IncludeDocs:       false,
		FreshnessBias:     0.2,
		DependencyDepth:   2,
		Strategy:          StrategyBalanced, // Default
	}

	// Task-specific constraint adaptations
	switch task.Type {
	case TaskTypeFeature:
		constraints.PreferredTypes = []string{"source", "configuration"}
		constraints.IncludeTests = false
		constraints.IncludeDocs = false
		constraints.FreshnessBias = 0.3
		constraints.Strategy = StrategyRelevance
		
	case TaskTypeDebug:
		constraints.PreferredTypes = []string{"source"}
		constraints.IncludeTests = true
		constraints.IncludeDocs = false
		constraints.FreshnessBias = 0.4 // Recent changes more important for debugging
		constraints.DependencyDepth = 3 // Deeper dependency analysis
		constraints.Strategy = StrategyDependency
		
	case TaskTypeRefactor:
		constraints.PreferredTypes = []string{"source"}
		constraints.IncludeTests = true
		constraints.IncludeDocs = false
		constraints.FreshnessBias = 0.1 // Less bias toward recent files
		constraints.DependencyDepth = 4 // Maximum dependency analysis
		constraints.Strategy = StrategyDependency
		
	case TaskTypeTest:
		constraints.PreferredTypes = []string{"source", "test"}
		constraints.IncludeTests = true
		constraints.IncludeDocs = false
		constraints.FreshnessBias = 0.2
		constraints.Strategy = StrategyRelevance
		
	case TaskTypeDocumentation:
		constraints.PreferredTypes = []string{"source", "documentation"}
		constraints.IncludeTests = false
		constraints.IncludeDocs = true
		constraints.FreshnessBias = 0.2
		constraints.Strategy = StrategyRelevance
	}

	// Apply learned preferences from profile
	if profile.SampleCount >= m.config.MinSamplesForAdaptation {
		if len(profile.ImportantFileTypes) > 0 {
			constraints.PreferredTypes = profile.ImportantFileTypes
		}
		if profile.PreferredStrategy != "" {
			constraints.Strategy = profile.PreferredStrategy
		}
	}

	return constraints
}

// PredictOptimalBudget suggests optimal token budget for a task
func (m *DefaultAdaptiveManager) PredictOptimalBudget(task *Task, projectCtx *ProjectContext) int {
	profile := m.getOrCreateTaskProfile(task.Type)
	
	// Base prediction on project size
	baseBudget := 8000 // Default budget
	
	// Adjust based on project size
	if projectCtx.TotalTokens > 200000 {
		baseBudget = 12000 // Large projects need more context
	} else if projectCtx.TotalTokens < 50000 {
		baseBudget = 4000 // Small projects need less
	}
	
	// Apply learned optimal budget if available
	if profile.SampleCount >= m.config.MinSamplesForAdaptation && profile.OptimalTokenBudget > 0 {
		// Weighted average of base prediction and learned optimal
		weight := min(1.0, float64(profile.SampleCount)/20.0) // Increase confidence with more samples
		baseBudget = int(float64(baseBudget)*(1-weight) + float64(profile.OptimalTokenBudget)*weight)
	}
	
	return baseBudget
}

// LearnFromFeedback incorporates feedback to improve future selections
func (m *DefaultAdaptiveManager) LearnFromFeedback(feedback *ContextFeedback) error {
	// Add to feedback log
	m.feedbackLog = append(m.feedbackLog, *feedback)
	
	// Clean old feedback
	m.cleanOldFeedback()
	
	// Update task profile
	profile := m.getOrCreateTaskProfile(feedback.Task.Type)
	m.updateTaskProfile(profile, feedback)
	
	return nil
}

// applyTaskSpecificAdaptations applies learned adaptations
func (m *DefaultAdaptiveManager) applyTaskSpecificAdaptations(constraints *ContextConstraints, task *Task, profile *TaskProfile, project *ProjectContext) {
	// Adjust max files based on learned patterns
	if profile.TypicalFileCount > 0 && profile.SampleCount >= m.config.MinSamplesForAdaptation {
		// Use learned typical file count with some buffer
		constraints.MaxFiles = int(float64(profile.TypicalFileCount) * 1.2)
		if constraints.MaxFiles < 10 {
			constraints.MaxFiles = 10
		}
		if constraints.MaxFiles > 100 {
			constraints.MaxFiles = 100
		}
	}
	
	// Adjust relevance threshold based on quality patterns
	if profile.AvgQualityScore > 0 && profile.SampleCount >= m.config.MinSamplesForAdaptation {
		if profile.AvgQualityScore < m.config.QualityThreshold {
			// Lower threshold to include more files if quality is low
			constraints.MinRelevanceScore *= 0.8
		} else if profile.AvgQualityScore > 0.9 {
			// Raise threshold to be more selective if quality is high
			constraints.MinRelevanceScore *= 1.2
		}
	}
}

// predictQuality predicts task completion quality based on context selection
func (m *DefaultAdaptiveManager) predictQuality(selectedContext *SelectedContext, task *Task, profile *TaskProfile) float64 {
	if profile.SampleCount < m.config.MinSamplesForAdaptation {
		return 0.75 // Default prediction
	}
	
	// Base prediction on historical average
	basePrediction := profile.AvgQualityScore
	
	// Adjust based on context characteristics
	tokenRatio := float64(selectedContext.TotalTokens) / float64(selectedContext.Constraints.MaxTokens)
	fileRatio := float64(selectedContext.TotalFiles) / float64(selectedContext.Constraints.MaxFiles)
	
	// Optimal ranges for good quality
	var qualityAdjustment float64 = 0
	
	// Token usage - sweet spot around 70-90%
	if tokenRatio >= 0.7 && tokenRatio <= 0.9 {
		qualityAdjustment += 0.05
	} else if tokenRatio < 0.3 || tokenRatio > 0.95 {
		qualityAdjustment -= 0.1
	}
	
	// File count - should use reasonable portion of available files
	if fileRatio >= 0.3 && fileRatio <= 0.8 {
		qualityAdjustment += 0.05
	}
	
	// Relevance score factor
	if selectedContext.SelectionScore > 0.8 {
		qualityAdjustment += 0.1
	} else if selectedContext.SelectionScore < 0.4 {
		qualityAdjustment -= 0.15
	}
	
	prediction := basePrediction + qualityAdjustment
	return max(0.0, min(1.0, prediction))
}

// getOrCreateTaskProfile gets or creates a profile for a task type
func (m *DefaultAdaptiveManager) getOrCreateTaskProfile(taskType TaskType) *TaskProfile {
	if profile, exists := m.profiles[taskType]; exists {
		return profile
	}
	
	profile := &TaskProfile{
		TaskType:           taskType,
		OptimalTokenBudget: 0,
		PreferredStrategy:  "",
		ImportantFileTypes: []string{},
		TypicalFileCount:   0,
		AvgQualityScore:    0.0,
		SuccessRate:        0.0,
		AdaptationFactors:  make(map[string]float64),
		LastUpdated:        time.Now(),
		SampleCount:        0,
	}
	
	m.profiles[taskType] = profile
	return profile
}

// updateTaskProfile updates a task profile with new feedback
func (m *DefaultAdaptiveManager) updateTaskProfile(profile *TaskProfile, feedback *ContextFeedback) {
	profile.SampleCount++
	profile.LastUpdated = time.Now()
	
	// Update running averages using exponential moving average
	alpha := m.config.LearningRate
	
	// Update quality score
	if profile.AvgQualityScore == 0 {
		profile.AvgQualityScore = feedback.QualityScore
	} else {
		profile.AvgQualityScore = alpha*feedback.QualityScore + (1-alpha)*profile.AvgQualityScore
	}
	
	// Update success rate
	successValue := 0.0
	if feedback.TaskSuccess {
		successValue = 1.0
	}
	if profile.SuccessRate == 0 {
		profile.SuccessRate = successValue
	} else {
		profile.SuccessRate = alpha*successValue + (1-alpha)*profile.SuccessRate
	}
	
	// Update optimal token budget
	if feedback.TaskSuccess && feedback.QualityScore > m.config.QualityThreshold {
		if profile.OptimalTokenBudget == 0 {
			profile.OptimalTokenBudget = feedback.SelectedContext.TotalTokens
		} else {
			profile.OptimalTokenBudget = int(alpha*float64(feedback.SelectedContext.TotalTokens) + (1-alpha)*float64(profile.OptimalTokenBudget))
		}
	}
	
	// Update typical file count
	if profile.TypicalFileCount == 0 {
		profile.TypicalFileCount = feedback.SelectedContext.TotalFiles
	} else {
		profile.TypicalFileCount = int(alpha*float64(feedback.SelectedContext.TotalFiles) + (1-alpha)*float64(profile.TypicalFileCount))
	}
	
	// Update preferred strategy if this one was successful
	if feedback.TaskSuccess && feedback.QualityScore > profile.AvgQualityScore {
		profile.PreferredStrategy = feedback.SelectedContext.Strategy
	}
}

// cleanOldFeedback removes feedback older than retention period
func (m *DefaultAdaptiveManager) cleanOldFeedback() {
	cutoff := time.Now().AddDate(0, 0, -m.config.FeedbackRetentionDays)
	
	filtered := []ContextFeedback{}
	for _, feedback := range m.feedbackLog {
		if feedback.Timestamp.After(cutoff) {
			filtered = append(filtered, feedback)
		}
	}
	
	m.feedbackLog = filtered
}

// GetProfileStatistics returns statistics about learned profiles
func (m *DefaultAdaptiveManager) GetProfileStatistics() map[TaskType]*TaskProfile {
	result := make(map[TaskType]*TaskProfile)
	for taskType, profile := range m.profiles {
		// Return a copy to prevent external modification
		profileCopy := *profile
		result[taskType] = &profileCopy
	}
	return result
}

// Helper functions
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}