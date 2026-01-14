package audit

import (
	"encoding/json"
	"sync"
	"time"

	"backend/pkg/logger"
)

// EventType represents the type of audit event
type EventType string

const (
	// Authentication events
	EventLogin              EventType = "LOGIN"
	EventLoginFailed        EventType = "LOGIN_FAILED"
	EventLogout             EventType = "LOGOUT"
	EventTokenRefresh       EventType = "TOKEN_REFRESH"
	EventTokenRefreshFailed EventType = "TOKEN_REFRESH_FAILED"

	// Password events
	EventPasswordChange       EventType = "PASSWORD_CHANGE"
	EventPasswordChangeFailed EventType = "PASSWORD_CHANGE_FAILED"
	EventPasswordReset        EventType = "PASSWORD_RESET"
	EventPasswordResetRequest EventType = "PASSWORD_RESET_REQUEST"

	// Session events
	EventSwitchTenant  EventType = "SWITCH_TENANT"
	EventSwitchCompany EventType = "SWITCH_COMPANY"

	// Security events
	EventCSRFValidationFailed EventType = "CSRF_VALIDATION_FAILED"
	EventBruteForceDetected   EventType = "BRUTE_FORCE_DETECTED"
	EventAccountLocked        EventType = "ACCOUNT_LOCKED"
	EventSuspiciousActivity   EventType = "SUSPICIOUS_ACTIVITY"
)

// Event represents a security audit event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
	Email     string                 `json:"email,omitempty"`
	TenantID  string                 `json:"tenant_id,omitempty"`
	CompanyID string                 `json:"company_id,omitempty"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// AuditLogger handles security event logging
type AuditLogger struct {
	logger *logger.Logger
	mu     sync.Mutex
}

var (
	defaultAuditLogger *AuditLogger
	once               sync.Once
)

// Init initializes the default audit logger
func Init() {
	once.Do(func() {
		defaultAuditLogger = &AuditLogger{
			logger: logger.GetDefault(),
		}
	})
}

// GetAuditLogger returns the default audit logger instance
func GetAuditLogger() *AuditLogger {
	if defaultAuditLogger == nil {
		Init()
	}
	return defaultAuditLogger
}

// Log logs a security audit event
func (a *AuditLogger) Log(event Event) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = generateEventID()
	}

	// Convert event to JSON for structured logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		a.logger.Errorf("[AUDIT] Failed to marshal event: %v", err)
		return
	}

	// Log based on event type and success
	logPrefix := "[AUDIT]"
	if event.Success {
		a.logger.Infof("%s %s", logPrefix, string(eventJSON))
	} else {
		a.logger.Warnf("%s %s", logPrefix, string(eventJSON))
	}
}

// LogLogin logs a successful login event
func (a *AuditLogger) LogLogin(userID, email, tenantID, ipAddress, userAgent string) {
	a.Log(Event{
		Type:      EventLogin,
		UserID:    userID,
		Email:     email,
		TenantID:  tenantID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
		Message:   "User logged in successfully",
	})
}

// LogLoginFailed logs a failed login attempt
func (a *AuditLogger) LogLoginFailed(email, ipAddress, userAgent, reason string) {
	a.Log(Event{
		Type:      EventLoginFailed,
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   false,
		Message:   "Login failed",
		Details: map[string]interface{}{
			"reason": reason,
		},
	})
}

// LogLogout logs a logout event
func (a *AuditLogger) LogLogout(userID, ipAddress, userAgent string) {
	a.Log(Event{
		Type:      EventLogout,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
		Message:   "User logged out",
	})
}

// LogTokenRefresh logs a token refresh event
func (a *AuditLogger) LogTokenRefresh(userID, ipAddress, userAgent string, success bool, reason string) {
	event := Event{
		Type:      EventTokenRefresh,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
		Message:   "Token refresh",
	}

	if !success {
		event.Type = EventTokenRefreshFailed
		event.Message = "Token refresh failed"
		event.Details = map[string]interface{}{
			"reason": reason,
		}
	}

	a.Log(event)
}

// LogPasswordChange logs a password change event
func (a *AuditLogger) LogPasswordChange(userID, ipAddress, userAgent string, success bool, reason string) {
	event := Event{
		Type:      EventPasswordChange,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
		Message:   "Password changed",
	}

	if !success {
		event.Type = EventPasswordChangeFailed
		event.Message = "Password change failed"
		event.Details = map[string]interface{}{
			"reason": reason,
		}
	}

	a.Log(event)
}

// LogPasswordReset logs a password reset event
func (a *AuditLogger) LogPasswordReset(email, ipAddress, userAgent string) {
	a.Log(Event{
		Type:      EventPasswordReset,
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
		Message:   "Password reset completed",
	})
}

// LogPasswordResetRequest logs a password reset request
func (a *AuditLogger) LogPasswordResetRequest(email, ipAddress, userAgent string) {
	a.Log(Event{
		Type:      EventPasswordResetRequest,
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
		Message:   "Password reset requested",
	})
}

// LogSwitchTenant logs a tenant switch event
func (a *AuditLogger) LogSwitchTenant(userID, fromTenantID, toTenantID, ipAddress, userAgent string) {
	a.Log(Event{
		Type:      EventSwitchTenant,
		UserID:    userID,
		TenantID:  toTenantID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
		Message:   "Switched tenant",
		Details: map[string]interface{}{
			"from_tenant_id": fromTenantID,
			"to_tenant_id":   toTenantID,
		},
	})
}

// LogSwitchCompany logs a company switch event
func (a *AuditLogger) LogSwitchCompany(userID, tenantID, fromCompanyID, toCompanyID, ipAddress, userAgent string) {
	a.Log(Event{
		Type:      EventSwitchCompany,
		UserID:    userID,
		TenantID:  tenantID,
		CompanyID: toCompanyID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
		Message:   "Switched company",
		Details: map[string]interface{}{
			"from_company_id": fromCompanyID,
			"to_company_id":   toCompanyID,
		},
	})
}

// LogBruteForceDetected logs a brute force attack detection
func (a *AuditLogger) LogBruteForceDetected(email, ipAddress, userAgent string, attemptCount int, tier int) {
	a.Log(Event{
		Type:      EventBruteForceDetected,
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   false,
		Message:   "Brute force attack detected",
		Details: map[string]interface{}{
			"attempt_count": attemptCount,
			"lockout_tier":  tier,
		},
	})
}

// LogAccountLocked logs an account lockout event
func (a *AuditLogger) LogAccountLocked(email, ipAddress, userAgent string, lockoutDuration time.Duration, tier int) {
	a.Log(Event{
		Type:      EventAccountLocked,
		Email:     email,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   false,
		Message:   "Account locked due to failed login attempts",
		Details: map[string]interface{}{
			"lockout_duration_seconds": lockoutDuration.Seconds(),
			"lockout_tier":             tier,
		},
	})
}

// LogCSRFValidationFailed logs a CSRF validation failure
func (a *AuditLogger) LogCSRFValidationFailed(userID, ipAddress, userAgent, endpoint string) {
	a.Log(Event{
		Type:      EventCSRFValidationFailed,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   false,
		Message:   "CSRF validation failed",
		Details: map[string]interface{}{
			"endpoint": endpoint,
		},
	})
}

// LogSuspiciousActivity logs suspicious activity
func (a *AuditLogger) LogSuspiciousActivity(userID, ipAddress, userAgent, activityType, description string) {
	a.Log(Event{
		Type:      EventSuspiciousActivity,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   false,
		Message:   description,
		Details: map[string]interface{}{
			"activity_type": activityType,
		},
	})
}

// Package-level convenience functions

// LogLogin logs a successful login using the default audit logger
func LogLogin(userID, email, tenantID, ipAddress, userAgent string) {
	GetAuditLogger().LogLogin(userID, email, tenantID, ipAddress, userAgent)
}

// LogLoginFailed logs a failed login using the default audit logger
func LogLoginFailed(email, ipAddress, userAgent, reason string) {
	GetAuditLogger().LogLoginFailed(email, ipAddress, userAgent, reason)
}

// LogLogout logs a logout using the default audit logger
func LogLogout(userID, ipAddress, userAgent string) {
	GetAuditLogger().LogLogout(userID, ipAddress, userAgent)
}

// LogTokenRefresh logs a token refresh using the default audit logger
func LogTokenRefresh(userID, ipAddress, userAgent string, success bool, reason string) {
	GetAuditLogger().LogTokenRefresh(userID, ipAddress, userAgent, success, reason)
}

// LogPasswordChange logs a password change using the default audit logger
func LogPasswordChange(userID, ipAddress, userAgent string, success bool, reason string) {
	GetAuditLogger().LogPasswordChange(userID, ipAddress, userAgent, success, reason)
}

// LogPasswordReset logs a password reset using the default audit logger
func LogPasswordReset(email, ipAddress, userAgent string) {
	GetAuditLogger().LogPasswordReset(email, ipAddress, userAgent)
}

// LogPasswordResetRequest logs a password reset request using the default audit logger
func LogPasswordResetRequest(email, ipAddress, userAgent string) {
	GetAuditLogger().LogPasswordResetRequest(email, ipAddress, userAgent)
}

// LogSwitchTenant logs a tenant switch using the default audit logger
func LogSwitchTenant(userID, fromTenantID, toTenantID, ipAddress, userAgent string) {
	GetAuditLogger().LogSwitchTenant(userID, fromTenantID, toTenantID, ipAddress, userAgent)
}

// LogSwitchCompany logs a company switch using the default audit logger
func LogSwitchCompany(userID, tenantID, fromCompanyID, toCompanyID, ipAddress, userAgent string) {
	GetAuditLogger().LogSwitchCompany(userID, tenantID, fromCompanyID, toCompanyID, ipAddress, userAgent)
}

// LogBruteForceDetected logs a brute force detection using the default audit logger
func LogBruteForceDetected(email, ipAddress, userAgent string, attemptCount int, tier int) {
	GetAuditLogger().LogBruteForceDetected(email, ipAddress, userAgent, attemptCount, tier)
}

// LogAccountLocked logs an account lockout using the default audit logger
func LogAccountLocked(email, ipAddress, userAgent string, lockoutDuration time.Duration, tier int) {
	GetAuditLogger().LogAccountLocked(email, ipAddress, userAgent, lockoutDuration, tier)
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return time.Now().Format("20060102150405.000000")
}
