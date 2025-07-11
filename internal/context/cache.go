package context

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

// InMemoryContextCache provides in-memory caching of context selections
type InMemoryContextCache struct {
	cache    map[string]*CacheEntry
	mutex    sync.RWMutex
	config   *CacheConfig
	stats    *CacheStatistics
}

// CacheEntry represents a cached context selection
type CacheEntry struct {
	Key               string           `json:"key"`
	SelectedContext   *SelectedContext `json:"selected_context"`
	ProjectFingerprint string          `json:"project_fingerprint"`
	CreatedAt         time.Time        `json:"created_at"`
	LastAccessed      time.Time        `json:"last_accessed"`
	AccessCount       int              `json:"access_count"`
	TTL               time.Duration    `json:"ttl"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// CacheConfig configures caching behavior
type CacheConfig struct {
	MaxEntries          int           `json:"max_entries"`
	DefaultTTL          time.Duration `json:"default_ttl"`
	EnableCompression   bool          `json:"enable_compression"`
	EnableInvalidation  bool          `json:"enable_invalidation"`
	EnableStats         bool          `json:"enable_stats"`
	CleanupInterval     time.Duration `json:"cleanup_interval"`
	MaxMemoryMB         int           `json:"max_memory_mb"`
	InvalidateOnChange  bool          `json:"invalidate_on_change"`
}

// CacheStatistics tracks cache performance
type CacheStatistics struct {
	Hits              int64   `json:"hits"`
	Misses            int64   `json:"misses"`
	Evictions         int64   `json:"evictions"`
	Invalidations     int64   `json:"invalidations"`
	TotalRequests     int64   `json:"total_requests"`
	HitRatio          float64 `json:"hit_ratio"`
	AvgLookupTime     float64 `json:"avg_lookup_time_ms"`
	MemoryUsageBytes  int64   `json:"memory_usage_bytes"`
	LastCleanup       time.Time `json:"last_cleanup"`
}

// ContextReuseManager manages context reuse across similar tasks
type ContextReuseManager interface {
	// FindReusableContext finds existing context that can be reused
	FindReusableContext(task *Task, projectCtx *ProjectContext, budget int) (*SelectedContext, float64, error)
	
	// AdaptReusedContext adapts a reused context for the new task
	AdaptReusedContext(reusedContext *SelectedContext, newTask *Task, similarity float64) (*SelectedContext, error)
	
	// StoreContextForReuse stores context for future reuse
	StoreContextForReuse(context *SelectedContext, task *Task, projectCtx *ProjectContext) error
	
	// CalculateTaskSimilarity calculates similarity between tasks
	CalculateTaskSimilarity(task1, task2 *Task) float64
}

// NewInMemoryContextCache creates a new in-memory context cache
func NewInMemoryContextCache(config *CacheConfig) *InMemoryContextCache {
	if config == nil {
		config = &CacheConfig{
			MaxEntries:          1000,
			DefaultTTL:          30 * time.Minute,
			EnableCompression:   false,
			EnableInvalidation:  true,
			EnableStats:         true,
			CleanupInterval:     5 * time.Minute,
			MaxMemoryMB:         100,
			InvalidateOnChange:  true,
		}
	}

	cache := &InMemoryContextCache{
		cache:  make(map[string]*CacheEntry),
		config: config,
		stats: &CacheStatistics{
			LastCleanup: time.Now(),
		},
	}

	// Start cleanup goroutine
	if config.CleanupInterval > 0 {
		go cache.startCleanupRoutine()
	}

	return cache
}

// Set stores a context selection in the cache
func (c *InMemoryContextCache) Set(key string, context *SelectedContext, expiry time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if expiry == 0 {
		expiry = c.config.DefaultTTL
	}

	// Generate project fingerprint for invalidation
	fingerprint := c.generateProjectFingerprint(context)

	entry := &CacheEntry{
		Key:               key,
		SelectedContext:   context,
		ProjectFingerprint: fingerprint,
		CreatedAt:         time.Now(),
		LastAccessed:      time.Now(),
		AccessCount:       0,
		TTL:               expiry,
		Metadata:          make(map[string]interface{}),
	}

	// Check if we need to evict entries
	if len(c.cache) >= c.config.MaxEntries {
		c.evictLRU()
	}

	c.cache[key] = entry

	if c.config.EnableStats {
		c.updateMemoryUsage()
	}

	return nil
}

// Get retrieves a context selection from the cache
func (c *InMemoryContextCache) Get(key string) (*SelectedContext, bool) {
	startTime := time.Now()
	
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.config.EnableStats {
		c.stats.TotalRequests++
		defer func() {
			lookupTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
			c.stats.AvgLookupTime = (c.stats.AvgLookupTime*float64(c.stats.TotalRequests-1) + lookupTime) / float64(c.stats.TotalRequests)
		}()
	}

	entry, exists := c.cache[key]
	if !exists {
		if c.config.EnableStats {
			c.stats.Misses++
			c.updateHitRatio()
		}
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.CreatedAt) > entry.TTL {
		delete(c.cache, key)
		if c.config.EnableStats {
			c.stats.Misses++
			c.stats.Evictions++
			c.updateHitRatio()
		}
		return nil, false
	}

	// Update access information
	entry.LastAccessed = time.Now()
	entry.AccessCount++

	if c.config.EnableStats {
		c.stats.Hits++
		c.updateHitRatio()
	}

	return entry.SelectedContext, true
}

// Delete removes an entry from the cache
func (c *InMemoryContextCache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.cache[key]; exists {
		delete(c.cache, key)
		if c.config.EnableStats {
			c.stats.Invalidations++
			c.updateMemoryUsage()
		}
	}

	return nil
}

// Clear removes all entries from the cache
func (c *InMemoryContextCache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*CacheEntry)
	
	if c.config.EnableStats {
		c.stats.Invalidations += int64(len(c.cache))
		c.updateMemoryUsage()
	}

	return nil
}

// InvalidateByProjectChange invalidates cache entries when project changes
func (c *InMemoryContextCache) InvalidateByProjectChange(projectCtx *ProjectContext) int {
	if !c.config.EnableInvalidation {
		return 0
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	currentFingerprint := c.generateProjectFingerprintFromContext(projectCtx)
	invalidated := 0

	for key, entry := range c.cache {
		if entry.ProjectFingerprint != currentFingerprint {
			delete(c.cache, key)
			invalidated++
		}
	}

	if c.config.EnableStats {
		c.stats.Invalidations += int64(invalidated)
		c.updateMemoryUsage()
	}

	return invalidated
}

// GetStatistics returns cache performance statistics
func (c *InMemoryContextCache) GetStatistics() *CacheStatistics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.config.EnableStats {
		return nil
	}

	// Return a copy
	statsCopy := *c.stats
	return &statsCopy
}

// DefaultContextReuseManager implements context reuse logic
type DefaultContextReuseManager struct {
	cache    ContextCache
	analyzer ContextAnalyzer
	config   *ReuseConfig
}

// ReuseConfig configures context reuse behavior
type ReuseConfig struct {
	MinSimilarityThreshold   float64 `json:"min_similarity_threshold"`
	MaxAdaptationDistance    float64 `json:"max_adaptation_distance"`
	EnableSemanticSimilarity bool    `json:"enable_semantic_similarity"`
	EnableFileOverlapReuse   bool    `json:"enable_file_overlap_reuse"`
	MinFileOverlapRatio      float64 `json:"min_file_overlap_ratio"`
}

// NewDefaultContextReuseManager creates a new context reuse manager
func NewDefaultContextReuseManager(cache ContextCache, analyzer ContextAnalyzer, config *ReuseConfig) *DefaultContextReuseManager {
	if config == nil {
		config = &ReuseConfig{
			MinSimilarityThreshold:   0.7,
			MaxAdaptationDistance:    0.5,
			EnableSemanticSimilarity: true,
			EnableFileOverlapReuse:   true,
			MinFileOverlapRatio:      0.6,
		}
	}

	return &DefaultContextReuseManager{
		cache:    cache,
		analyzer: analyzer,
		config:   config,
	}
}

// FindReusableContext finds existing context that can be reused
func (r *DefaultContextReuseManager) FindReusableContext(task *Task, projectCtx *ProjectContext, budget int) (*SelectedContext, float64, error) {
	// For this implementation, we'll search through recent cache entries
	// In a production system, you might maintain a separate index
	
	// Generate a search key pattern based on task characteristics
	searchPattern := r.generateSearchPattern(task, budget)
	
	// Try exact match first
	if reusedContext, found := r.cache.Get(searchPattern); found {
		return reusedContext, 1.0, nil // Perfect match
	}
	
	// For this simplified implementation, return nil
	// In a full implementation, you would:
	// 1. Search through cached contexts
	// 2. Calculate similarity scores
	// 3. Return the best match above threshold
	
	return nil, 0.0, nil
}

// AdaptReusedContext adapts a reused context for the new task
func (r *DefaultContextReuseManager) AdaptReusedContext(reusedContext *SelectedContext, newTask *Task, similarity float64) (*SelectedContext, error) {
	// Create adapted context based on reused context
	adapted := &SelectedContext{
		Task:            newTask,
		Files:           make([]ContextFile, len(reusedContext.Files)),
		TotalTokens:     reusedContext.TotalTokens,
		TotalFiles:      reusedContext.TotalFiles,
		SelectionScore:  reusedContext.SelectionScore * similarity, // Adjust score based on similarity
		Strategy:        reusedContext.Strategy,
		Constraints:     reusedContext.Constraints,
		Metadata:        make(map[string]interface{}),
		CreatedAt:       time.Now(),
		SelectionTime:   0, // Reuse is instant
	}

	// Copy files with potential re-scoring
	copy(adapted.Files, reusedContext.Files)

	// Add reuse metadata
	adapted.Metadata["reused_from"] = reusedContext.Task.Description
	adapted.Metadata["similarity_score"] = similarity
	adapted.Metadata["adaptation_applied"] = true

	return adapted, nil
}

// StoreContextForReuse stores context for future reuse
func (r *DefaultContextReuseManager) StoreContextForReuse(context *SelectedContext, task *Task, projectCtx *ProjectContext) error {
	key := r.generateCacheKey(task, context.Constraints.MaxTokens, projectCtx)
	return r.cache.Set(key, context, 30*time.Minute) // 30 minute TTL for reuse
}

// CalculateTaskSimilarity calculates similarity between tasks
func (r *DefaultContextReuseManager) CalculateTaskSimilarity(task1, task2 *Task) float64 {
	if task1.Type != task2.Type {
		return 0.0 // Different task types are not similar
	}

	// Simple keyword-based similarity for now
	// In production, you might use more sophisticated NLP techniques
	
	words1 := extractWords(task1.Description)
	words2 := extractWords(task2.Description)
	
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Calculate Jaccard similarity (intersection over union)
	intersection := 0
	union := make(map[string]bool)
	
	// Add all words to union
	for _, word := range words1 {
		union[word] = true
	}
	for _, word := range words2 {
		union[word] = true
	}
	
	// Count intersection
	words1Set := make(map[string]bool)
	for _, word := range words1 {
		words1Set[word] = true
	}
	
	for _, word := range words2 {
		if words1Set[word] {
			intersection++
		}
	}
	
	if len(union) == 0 {
		return 0.0
	}
	
	return float64(intersection) / float64(len(union))
}

// Helper methods for cache implementation

func (c *InMemoryContextCache) generateProjectFingerprint(context *SelectedContext) string {
	// Generate a fingerprint based on selected files and their modification times
	fingerprint := ""
	for _, file := range context.Files {
		fingerprint += fmt.Sprintf("%s:%d:", file.FileInfo.Path, file.FileInfo.LastModified.Unix())
	}
	
	hash := md5.Sum([]byte(fingerprint))
	return fmt.Sprintf("%x", hash)
}

func (c *InMemoryContextCache) generateProjectFingerprintFromContext(projectCtx *ProjectContext) string {
	// Generate fingerprint from project context
	fingerprint := ""
	for _, file := range projectCtx.Files {
		fingerprint += fmt.Sprintf("%s:%d:", file.Path, file.LastModified.Unix())
	}
	
	hash := md5.Sum([]byte(fingerprint))
	return fmt.Sprintf("%x", hash)
}

func (c *InMemoryContextCache) evictLRU() {
	if len(c.cache) == 0 {
		return
	}

	// Find least recently used entry
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, entry := range c.cache {
		if entry.LastAccessed.Before(oldestTime) {
			oldestTime = entry.LastAccessed
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
		if c.config.EnableStats {
			c.stats.Evictions++
		}
	}
}

func (c *InMemoryContextCache) updateHitRatio() {
	total := c.stats.Hits + c.stats.Misses
	if total > 0 {
		c.stats.HitRatio = float64(c.stats.Hits) / float64(total)
	}
}

func (c *InMemoryContextCache) updateMemoryUsage() {
	// Rough estimate of memory usage
	c.stats.MemoryUsageBytes = int64(len(c.cache)) * 1024 // Rough estimate: 1KB per entry
}

func (c *InMemoryContextCache) startCleanupRoutine() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

func (c *InMemoryContextCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expired := []string{}

	for key, entry := range c.cache {
		if now.Sub(entry.CreatedAt) > entry.TTL {
			expired = append(expired, key)
		}
	}

	for _, key := range expired {
		delete(c.cache, key)
	}

	if c.config.EnableStats {
		c.stats.Evictions += int64(len(expired))
		c.stats.LastCleanup = now
		c.updateMemoryUsage()
	}
}

// Helper methods for reuse manager

func (r *DefaultContextReuseManager) generateSearchPattern(task *Task, budget int) string {
	return fmt.Sprintf("task:%s:budget:%d", string(task.Type), budget)
}

func (r *DefaultContextReuseManager) generateCacheKey(task *Task, budget int, projectCtx *ProjectContext) string {
	// Generate a unique key for caching
	taskHash := md5.Sum([]byte(fmt.Sprintf("%s:%s", task.Type, task.Description)))
	projectHash := md5.Sum([]byte(projectCtx.RootPath))
	
	return fmt.Sprintf("ctx:%x:%x:%d", taskHash, projectHash, budget)
}

func extractWords(text string) []string {
	// Simple word extraction - in production you'd use proper tokenization
	words := []string{}
	current := ""
	
	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			current += string(char)
		} else {
			if len(current) > 2 { // Only include words longer than 2 chars
				words = append(words, current)
			}
			current = ""
		}
	}
	
	if len(current) > 2 {
		words = append(words, current)
	}
	
	return words
}