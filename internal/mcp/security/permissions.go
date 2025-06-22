package security

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// Permission represents a security permission
type Permission string

const (
	// File system permissions
	PermissionReadFile     Permission = "fs:read"
	PermissionWriteFile    Permission = "fs:write"
	PermissionListDir      Permission = "fs:list"
	PermissionDeleteFile   Permission = "fs:delete"
	
	// Command execution permissions
	PermissionExecCommand  Permission = "cmd:exec"
	PermissionExecSystem   Permission = "cmd:system"
	
	// Network permissions
	PermissionNetworkRead  Permission = "net:read"
	PermissionNetworkWrite Permission = "net:write"
	
	// Resource permissions
	PermissionResourceRead Permission = "resource:read"
)

// SecurityPolicy defines what operations are allowed
type SecurityPolicy struct {
	AllowedPermissions []Permission          `json:"allowed_permissions"`
	DeniedPermissions  []Permission          `json:"denied_permissions"`
	PathRestrictions   PathRestrictions      `json:"path_restrictions"`
	CommandWhitelist   []string              `json:"command_whitelist"`
	ResourceLimits     ResourceLimits        `json:"resource_limits"`
	AuditLog          bool                  `json:"audit_log"`
}

// PathRestrictions define file system access restrictions
type PathRestrictions struct {
	AllowedPaths    []string `json:"allowed_paths"`
	DeniedPaths     []string `json:"denied_paths"`
	RequireBasePath string   `json:"require_base_path"`
}

// ResourceLimits define resource usage limits
type ResourceLimits struct {
	MaxMemoryMB     int `json:"max_memory_mb"`
	MaxCPUPercent   int `json:"max_cpu_percent"`
	MaxExecutionSec int `json:"max_execution_sec"`
	MaxFileSize     int `json:"max_file_size"`
}

// SecurityContext holds the current security state
type SecurityContext struct {
	Policy      *SecurityPolicy
	UserID      string
	SessionID   string
	AuditTrail  []AuditEntry
}

// AuditEntry records security-relevant operations
type AuditEntry struct {
	Timestamp   string     `json:"timestamp"`
	Operation   string     `json:"operation"`
	Permission  Permission `json:"permission"`
	Resource    string     `json:"resource"`
	Result      string     `json:"result"`
	Error       string     `json:"error,omitempty"`
}

// SecurityValidator validates operations against security policies
type SecurityValidator struct {
	context *SecurityContext
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator(policy *SecurityPolicy, userID, sessionID string) *SecurityValidator {
	return &SecurityValidator{
		context: &SecurityContext{
			Policy:     policy,
			UserID:     userID,
			SessionID:  sessionID,
			AuditTrail: make([]AuditEntry, 0),
		},
	}
}

// ValidateFileOperation validates file system operations
func (sv *SecurityValidator) ValidateFileOperation(ctx context.Context, operation string, path string) error {
	// Determine required permission
	var requiredPerm Permission
	switch operation {
	case "read":
		requiredPerm = PermissionReadFile
	case "write":
		requiredPerm = PermissionWriteFile
	case "list":
		requiredPerm = PermissionListDir
	case "delete":
		requiredPerm = PermissionDeleteFile
	default:
		return fmt.Errorf("unknown file operation: %s", operation)
	}
	
	// Check permission
	if !sv.hasPermission(requiredPerm) {
		sv.auditDenied(operation, requiredPerm, path, "permission denied")
		return fmt.Errorf("permission denied: %s on %s", operation, path)
	}
	
	// Check path restrictions
	if err := sv.validatePath(path); err != nil {
		sv.auditDenied(operation, requiredPerm, path, err.Error())
		return fmt.Errorf("path restriction: %w", err)
	}
	
	// Audit success
	sv.auditAllowed(operation, requiredPerm, path)
	return nil
}

// ValidateCommandExecution validates command execution
func (sv *SecurityValidator) ValidateCommandExecution(ctx context.Context, command string, args []string) error {
	// Check basic execution permission
	if !sv.hasPermission(PermissionExecCommand) {
		sv.auditDenied("exec", PermissionExecCommand, command, "permission denied")
		return fmt.Errorf("command execution permission denied")
	}
	
	// Check command whitelist
	if !sv.isCommandAllowed(command) {
		sv.auditDenied("exec", PermissionExecCommand, command, "command not in whitelist")
		return fmt.Errorf("command not allowed: %s", command)
	}
	
	// Check for dangerous system commands
	if sv.isDangerousCommand(command, args) {
		if !sv.hasPermission(PermissionExecSystem) {
			sv.auditDenied("exec", PermissionExecSystem, command, "system command permission denied")
			return fmt.Errorf("system command permission denied: %s", command)
		}
	}
	
	// Audit success
	sv.auditAllowed("exec", PermissionExecCommand, command)
	return nil
}

// ValidateResourceAccess validates resource access
func (sv *SecurityValidator) ValidateResourceAccess(ctx context.Context, resourceURI string) error {
	if !sv.hasPermission(PermissionResourceRead) {
		sv.auditDenied("resource", PermissionResourceRead, resourceURI, "permission denied")
		return fmt.Errorf("resource access permission denied")
	}
	
	sv.auditAllowed("resource", PermissionResourceRead, resourceURI)
	return nil
}

// hasPermission checks if a permission is granted
func (sv *SecurityValidator) hasPermission(perm Permission) bool {
	// Check denied permissions first
	for _, denied := range sv.context.Policy.DeniedPermissions {
		if denied == perm {
			return false
		}
	}
	
	// Check allowed permissions
	for _, allowed := range sv.context.Policy.AllowedPermissions {
		if allowed == perm {
			return true
		}
	}
	
	return false
}

// validatePath checks path restrictions
func (sv *SecurityValidator) validatePath(path string) error {
	// Clean and resolve path
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	
	restrictions := sv.context.Policy.PathRestrictions
	
	// Check base path requirement
	if restrictions.RequireBasePath != "" {
		basePath, err := filepath.Abs(restrictions.RequireBasePath)
		if err != nil {
			return fmt.Errorf("invalid base path: %w", err)
		}
		
		if !strings.HasPrefix(cleanPath, basePath) {
			return fmt.Errorf("path outside allowed base: %s", cleanPath)
		}
	}
	
	// Check denied paths
	for _, denied := range restrictions.DeniedPaths {
		deniedAbs, err := filepath.Abs(denied)
		if err != nil {
			continue
		}
		
		if strings.HasPrefix(cleanPath, deniedAbs) {
			return fmt.Errorf("path explicitly denied: %s", cleanPath)
		}
	}
	
	// Check allowed paths (if specified)
	if len(restrictions.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range restrictions.AllowedPaths {
			allowedAbs, err := filepath.Abs(allowedPath)
			if err != nil {
				continue
			}
			
			if strings.HasPrefix(cleanPath, allowedAbs) {
				allowed = true
				break
			}
		}
		
		if !allowed {
			return fmt.Errorf("path not in allowed list: %s", cleanPath)
		}
	}
	
	return nil
}

// isCommandAllowed checks if command is in whitelist
func (sv *SecurityValidator) isCommandAllowed(command string) bool {
	if len(sv.context.Policy.CommandWhitelist) == 0 {
		return true // No whitelist means all commands allowed
	}
	
	for _, allowed := range sv.context.Policy.CommandWhitelist {
		if allowed == command {
			return true
		}
	}
	
	return false
}

// isDangerousCommand checks if command is considered dangerous
func (sv *SecurityValidator) isDangerousCommand(command string, args []string) bool {
	dangerousCommands := []string{
		"rm", "rmdir", "del", "sudo", "su", "chmod", "chown",
		"curl", "wget", "nc", "netcat", "telnet", "ssh",
		"bash", "sh", "cmd", "powershell", "python", "node",
	}
	
	for _, dangerous := range dangerousCommands {
		if command == dangerous {
			return true
		}
	}
	
	// Check for suspicious arguments
	for _, arg := range args {
		if strings.Contains(arg, "..") || 
		   strings.Contains(arg, "/etc/") ||
		   strings.Contains(arg, "/var/") ||
		   strings.Contains(arg, "C:\\Windows") {
			return true
		}
	}
	
	return false
}

// auditAllowed records successful operation
func (sv *SecurityValidator) auditAllowed(operation string, permission Permission, resource string) {
	if sv.context.Policy.AuditLog {
		entry := AuditEntry{
			Timestamp:  "2025-06-22T08:00:00Z", // Simplified for testing
			Operation:  operation,
			Permission: permission,
			Resource:   resource,
			Result:     "allowed",
		}
		sv.context.AuditTrail = append(sv.context.AuditTrail, entry)
	}
}

// auditDenied records denied operation
func (sv *SecurityValidator) auditDenied(operation string, permission Permission, resource string, reason string) {
	if sv.context.Policy.AuditLog {
		entry := AuditEntry{
			Timestamp:  "2025-06-22T08:00:00Z", // Simplified for testing
			Operation:  operation,
			Permission: permission,
			Resource:   resource,
			Result:     "denied",
			Error:      reason,
		}
		sv.context.AuditTrail = append(sv.context.AuditTrail, entry)
	}
}

// GetAuditTrail returns the current audit trail
func (sv *SecurityValidator) GetAuditTrail() []AuditEntry {
	return sv.context.AuditTrail
}

// GetSecurityContext returns the current security context
func (sv *SecurityValidator) GetSecurityContext() *SecurityContext {
	return sv.context
}

// DefaultRestrictivePolicy creates a restrictive security policy
func DefaultRestrictivePolicy(basePath string) *SecurityPolicy {
	return &SecurityPolicy{
		AllowedPermissions: []Permission{
			PermissionReadFile,
			PermissionListDir,
		},
		DeniedPermissions: []Permission{
			PermissionDeleteFile,
			PermissionExecSystem,
		},
		PathRestrictions: PathRestrictions{
			RequireBasePath: basePath,
			DeniedPaths: []string{
				"/etc",
				"/var",
				"/usr",
				"/bin",
				"/sbin",
			},
		},
		CommandWhitelist: []string{
			"echo",
			"pwd",
			"ls",
			"date",
		},
		ResourceLimits: ResourceLimits{
			MaxMemoryMB:     100,
			MaxCPUPercent:   50,
			MaxExecutionSec: 30,
			MaxFileSize:     10 * 1024 * 1024, // 10MB
		},
		AuditLog: true,
	}
}

// DefaultPermissivePolicy creates a permissive security policy for testing
func DefaultPermissivePolicy() *SecurityPolicy {
	return &SecurityPolicy{
		AllowedPermissions: []Permission{
			PermissionReadFile,
			PermissionWriteFile,
			PermissionListDir,
			PermissionExecCommand,
			PermissionResourceRead,
		},
		DeniedPermissions: []Permission{
			PermissionDeleteFile,
			PermissionExecSystem,
		},
		PathRestrictions: PathRestrictions{
			// No restrictions for testing
		},
		CommandWhitelist: []string{
			"echo", "pwd", "ls", "date", "whoami", "cat",
		},
		ResourceLimits: ResourceLimits{
			MaxMemoryMB:     500,
			MaxCPUPercent:   80,
			MaxExecutionSec: 60,
			MaxFileSize:     50 * 1024 * 1024, // 50MB
		},
		AuditLog: true,
	}
}