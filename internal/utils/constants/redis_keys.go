package constants

import "time"

// Redis key prefixes
const (
    // Idempotency keys
    IdempotencyKeyPrefix   = "idempotency:payment:"
    IdempotencyLockPrefix  = "lock:idempotency:payment:"
    
    // Cache keys
    PaymentMethodsKey      = "cache:payment_methods"
    PaymentKeyPrefix       = "cache:payment:"
    BookingKeyPrefix       = "cache:booking:"
    UserKeyPrefix          = "cache:user:"
    
    // Rate limiting keys
    RateLimitPrefix        = "ratelimit:"
    RateLimitByIPPrefix    = "ratelimit:ip:"
    RateLimitByUserPrefix  = "ratelimit:user:"
    
    // Queue keys
    PaymentQueuePrefix     = "queue:payment:"
    WebhookQueuePrefix     = "queue:webhook:"
    
    // Session keys
    SessionPrefix          = "session:"
    OTPSessionPrefix       = "otp:"
)

// TTL constants (in seconds)
const (
    // Idempotency TTLs
    IdempotencyTTL         = 24 * 60 * 60 // 24 hours
    IdempotencyLockTTL     = 30           // 30 seconds
    
    // Cache TTLs
    PaymentCacheTTL        = 5 * 60       // 5 minutes
    PaymentMethodsTTL      = 60 * 60      // 1 hour
    BookingCacheTTL        = 10 * 60      // 10 minutes
    UserCacheTTL           = 30 * 60      // 30 minutes
    
    // Rate limit TTLs
    RateLimitTTL           = 60           // 1 minute
    RateLimitStrictTTL     = 24 * 60 * 60 // 24 hours
    
    // Session TTLs
    SessionTTL             = 7 * 24 * 60 * 60 // 7 days
    OTPSessionTTL          = 5 * 60           // 5 minutes
    
    // Queue TTLs
    QueueTTL               = 30 * 60      // 30 minutes
    WebhookQueueTTL        = 24 * 60 * 60 // 24 hours
)

// Time duration helpers
func (k KeyPrefix) WithDuration(ttlSeconds int) time.Duration {
    return time.Duration(ttlSeconds) * time.Second
}

// KeyPrefix type for method binding
type KeyPrefix string

// BuildKey builds a Redis key with prefix and parts
func (k KeyPrefix) BuildKey(parts ...string) string {
    key := string(k)
    for _, part := range parts {
        key += part
    }
    return key
}

// Pre-defined key builders
var (
    // Idempotency keys
    IdempotencyKey = KeyPrefix(IdempotencyKeyPrefix)
    IdempotencyLock = KeyPrefix(IdempotencyLockPrefix)
    
    // Cache keys
    PaymentKey = KeyPrefix(PaymentKeyPrefix)
    BookingKey = KeyPrefix(BookingKeyPrefix)
    UserKey = KeyPrefix(UserKeyPrefix)
    
    // Rate limit keys
    RateLimitByIP = KeyPrefix(RateLimitByIPPrefix)
    RateLimitByUser = KeyPrefix(RateLimitByUserPrefix)
    
    // Queue keys
    PaymentQueue = KeyPrefix(PaymentQueuePrefix)
    WebhookQueue = KeyPrefix(WebhookQueuePrefix)
    
    // Session keys
    Session = KeyPrefix(SessionPrefix)
    OTPSession = KeyPrefix(OTPSessionPrefix)
)