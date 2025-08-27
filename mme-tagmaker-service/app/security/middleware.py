"""
Security middleware for MME Tagmaker Service (FastAPI)
Provides comprehensive security features including rate limiting, threat detection,
audit logging, and role-based access control.
"""

import time
import asyncio
import logging
from collections import defaultdict, Counter
from typing import Dict, Optional, Set, List
from datetime import datetime, timedelta
from fastapi import Request, HTTPException, status
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.responses import JSONResponse
import re
import json

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class SecurityConfig:
    """Security configuration for MME Tagmaker Service"""
    
    def __init__(self):
        self.rate_limit_enabled = True
        self.threat_detection = True
        self.audit_logging = True
        self.security_headers = True
        self.input_validation = True
        self.max_request_size = 10 * 1024 * 1024  # 10MB
        self.rate_limit_per_minute = 100
        self.rate_limit_burst = 20

class RateLimiter:
    """Advanced rate limiter with sliding window and burst protection"""
    
    def __init__(self, config: SecurityConfig):
        self.config = config
        self.requests: Dict[str, List[float]] = defaultdict(list)
        self.blocked: Dict[str, float] = {}
        self.cleanup_interval = 300  # 5 minutes
        
        # Start cleanup task
        asyncio.create_task(self._cleanup_task())
    
    async def _cleanup_task(self):
        """Periodically clean up old entries"""
        while True:
            await asyncio.sleep(self.cleanup_interval)
            await self._cleanup()
    
    async def _cleanup(self):
        """Remove expired entries"""
        current_time = time.time()
        window_start = current_time - 60  # 1 minute window
        
        # Clean up request history
        for key in list(self.requests.keys()):
            self.requests[key] = [t for t in self.requests[key] if t > window_start]
            if not self.requests[key]:
                del self.requests[key]
        
        # Clean up blocked IPs
        for key in list(self.blocked.keys()):
            if current_time - self.blocked[key] > 300:  # 5 minute block
                del self.blocked[key]
    
    def is_allowed(self, user_id: str, client_ip: str, endpoint: str) -> tuple[bool, int]:
        """Check if request is allowed based on rate limits"""
        current_time = time.time()
        key = f"{user_id}:{client_ip}:{endpoint}"
        
        # Check if currently blocked
        if key in self.blocked:
            if current_time - self.blocked[key] < 300:  # 5 minute block
                return False, 300
            else:
                del self.blocked[key]
        
        # Clean old requests (sliding window)
        window_start = current_time - 60
        self.requests[key] = [t for t in self.requests[key] if t > window_start]
        
        # Check rate limit
        request_count = len(self.requests[key])
        limit = self._get_endpoint_limit(endpoint)
        
        if request_count >= limit:
            self.blocked[key] = current_time
            return False, 300
        
        # Add current request
        self.requests[key].append(current_time)
        return True, 0
    
    def _get_endpoint_limit(self, endpoint: str) -> int:
        """Get rate limit based on endpoint"""
        if "/extract" in endpoint or "/tag" in endpoint:
            return 50  # AI endpoints have lower limits
        elif "/health" in endpoint or "/metrics" in endpoint:
            return 200  # Health checks have higher limits
        return self.config.rate_limit_per_minute

class ThreatDetector:
    """Advanced threat detection system"""
    
    def __init__(self):
        self.sql_patterns = [
            r"union.*select", r"insert.*into", r"delete.*from", r"update.*set",
            r"drop.*table", r"create.*table", r"alter.*table", r"exec.*sp_",
            r"xp_cmdshell", r"sp_executesql", r"openrowset", r"opendatasource"
        ]
        
        self.xss_patterns = [
            r"<script.*?>", r"javascript:", r"onload\s*=", r"onerror\s*=",
            r"onclick\s*=", r"alert\s*\(", r"document\.cookie", r"window\.location",
            r"eval\s*\(", r"expression\s*\(", r"vbscript:", r"data:text/html"
        ]
        
        self.injection_patterns = [
            r"\.\./", r"\.\.\\", r"~.*?/", r"/etc/passwd", r"/etc/shadow",
            r"cmd\.exe", r"powershell", r"/bin/sh", r"/bin/bash"
        ]
        
        self.bot_patterns = [
            r"nmap", r"sqlmap", r"nikto", r"burp", r"metasploit", r"scanner",
            r"crawler", r"spider", r"scraper", r"harvest", r"extract", r"bot"
        ]
    
    def analyze_request(self, request: Request) -> Optional[Dict]:
        """Analyze request for security threats"""
        threats = []
        
        # Analyze URL and query parameters
        url_str = str(request.url).lower()
        query_str = str(request.url.query).lower() if request.url.query else ""
        
        # Check for SQL injection
        if self._check_patterns(url_str + query_str, self.sql_patterns):
            threats.append({
                "type": "SQL_INJECTION",
                "severity": "HIGH",
                "description": "SQL injection attempt detected in URL/query"
            })
        
        # Check for XSS
        if self._check_patterns(url_str + query_str, self.xss_patterns):
            threats.append({
                "type": "XSS_ATTEMPT",
                "severity": "MEDIUM",
                "description": "Cross-site scripting attempt detected"
            })
        
        # Check for path traversal
        if self._check_patterns(url_str, self.injection_patterns):
            threats.append({
                "type": "PATH_TRAVERSAL",
                "severity": "HIGH",
                "description": "Path traversal attempt detected"
            })
        
        # Check user agent
        user_agent = request.headers.get("user-agent", "").lower()
        if self._check_patterns(user_agent, self.bot_patterns):
            threats.append({
                "type": "SUSPICIOUS_BOT",
                "severity": "LOW",
                "description": "Suspicious bot or scanner detected"
            })
        
        return threats[0] if threats else None
    
    def _check_patterns(self, text: str, patterns: List[str]) -> bool:
        """Check text against regex patterns"""
        for pattern in patterns:
            if re.search(pattern, text, re.IGNORECASE):
                return True
        return False

class SecurityAuditor:
    """Security event audit logger"""
    
    def __init__(self):
        self.security_logger = logging.getLogger("security_audit")
        handler = logging.StreamHandler()
        formatter = logging.Formatter(
            '%(asctime)s - [SECURITY EVENT] - %(levelname)s - %(message)s'
        )
        handler.setFormatter(formatter)
        self.security_logger.addHandler(handler)
        self.security_logger.setLevel(logging.WARNING)
    
    def log_security_event(self, event_type: str, user_id: str, client_ip: str, 
                          description: str, endpoint: str = ""):
        """Log security event with structured data"""
        event_data = {
            "event_type": event_type,
            "user_id": user_id,
            "client_ip": client_ip,
            "description": description,
            "endpoint": endpoint,
            "service": "mme-tagmaker-service",
            "timestamp": datetime.utcnow().isoformat()
        }
        
        self.security_logger.warning(json.dumps(event_data))

class SecurityMetrics:
    """Security metrics collector"""
    
    def __init__(self):
        self.rate_limit_violations = Counter()
        self.threat_detections = Counter()
        self.unauthorized_access = Counter()
        self.threats_by_type = Counter()
        self.violations_by_endpoint = Counter()
        self.last_updated = datetime.utcnow()
    
    def record_rate_limit_violation(self, endpoint: str, user_id: str):
        """Record rate limit violation"""
        self.rate_limit_violations[f"{user_id}:{endpoint}"] += 1
        self.violations_by_endpoint[endpoint] += 1
        self.last_updated = datetime.utcnow()
    
    def record_threat_detection(self, threat_type: str, user_id: str):
        """Record threat detection"""
        self.threat_detections[f"{user_id}:{threat_type}"] += 1
        self.threats_by_type[threat_type] += 1
        self.last_updated = datetime.utcnow()
    
    def record_unauthorized_access(self, user_id: str):
        """Record unauthorized access attempt"""
        self.unauthorized_access[user_id] += 1
        self.last_updated = datetime.utcnow()
    
    def get_metrics(self) -> Dict:
        """Get current security metrics"""
        return {
            "rate_limit_violations": sum(self.rate_limit_violations.values()),
            "threat_detections": sum(self.threat_detections.values()),
            "unauthorized_access": sum(self.unauthorized_access.values()),
            "threats_by_type": dict(self.threats_by_type),
            "violations_by_endpoint": dict(self.violations_by_endpoint),
            "last_updated": self.last_updated.isoformat()
        }

class SecurityMiddleware(BaseHTTPMiddleware):
    """Comprehensive security middleware for FastAPI"""
    
    def __init__(self, app, config: SecurityConfig = None):
        super().__init__(app)
        self.config = config or SecurityConfig()
        self.rate_limiter = RateLimiter(self.config)
        self.threat_detector = ThreatDetector()
        self.auditor = SecurityAuditor()
        self.metrics = SecurityMetrics()
    
    async def dispatch(self, request: Request, call_next):
        """Main security middleware dispatch"""
        start_time = time.time()
        
        # Extract user information from gateway headers
        user_id = request.headers.get("x-user-id", "anonymous")
        user_roles = request.headers.get("x-user-roles", "")
        client_ip = self._get_client_ip(request)
        endpoint = request.url.path
        
        try:
            # Security headers
            if self.config.security_headers:
                response = await self._add_security_headers(request, call_next)
            else:
                response = await call_next(request)
            
            # Input validation
            if self.config.input_validation and not await self._validate_input(request, user_id, client_ip):
                return JSONResponse(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    content={"error": "Invalid request", "code": "INVALID_REQUEST"}
                )
            
            # Rate limiting
            if self.config.rate_limit_enabled:
                allowed, retry_after = self.rate_limiter.is_allowed(user_id, client_ip, endpoint)
                if not allowed:
                    self.auditor.log_security_event(
                        "RATE_LIMIT_VIOLATION", user_id, client_ip,
                        f"Rate limit exceeded for endpoint {endpoint}", endpoint
                    )
                    self.metrics.record_rate_limit_violation(endpoint, user_id)
                    
                    return JSONResponse(
                        status_code=status.HTTP_429_TOO_MANY_REQUESTS,
                        content={
                            "error": "Rate limit exceeded",
                            "code": "RATE_LIMIT_EXCEEDED",
                            "retry_after": retry_after
                        }
                    )
            
            # Threat detection
            if self.config.threat_detection:
                threat = self.threat_detector.analyze_request(request)
                if threat:
                    self.auditor.log_security_event(
                        "THREAT_DETECTED", user_id, client_ip,
                        f"Threat: {threat['type']} - {threat['description']}", endpoint
                    )
                    self.metrics.record_threat_detection(threat["type"], user_id)
                    
                    return JSONResponse(
                        status_code=status.HTTP_403_FORBIDDEN,
                        content={
                            "error": "Request blocked by security policy",
                            "code": "SECURITY_VIOLATION"
                        }
                    )
            
            # Process time for monitoring
            process_time = time.time() - start_time
            response.headers["X-Process-Time"] = str(process_time)
            
            return response
            
        except Exception as e:
            logger.error(f"Security middleware error: {e}")
            return JSONResponse(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                content={"error": "Internal security error", "code": "SECURITY_ERROR"}
            )
    
    async def _add_security_headers(self, request: Request, call_next):
        """Add security headers to response"""
        response = await call_next(request)
        
        # Security headers
        response.headers["X-Content-Type-Options"] = "nosniff"
        response.headers["X-Frame-Options"] = "DENY"
        response.headers["X-XSS-Protection"] = "1; mode=block"
        response.headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
        response.headers["Content-Security-Policy"] = "default-src 'self'"
        response.headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
        
        return response
    
    async def _validate_input(self, request: Request, user_id: str, client_ip: str) -> bool:
        """Validate request input"""
        # Check request size
        content_length = request.headers.get("content-length")
        if content_length and int(content_length) > self.config.max_request_size:
            self.auditor.log_security_event(
                "OVERSIZED_REQUEST", user_id, client_ip,
                f"Request size: {content_length} bytes", request.url.path
            )
            return False
        
        # Check for path traversal
        path = str(request.url.path)
        if ".." in path or "~" in path:
            self.auditor.log_security_event(
                "PATH_TRAVERSAL_ATTEMPT", user_id, client_ip,
                f"Suspicious path: {path}", path
            )
            return False
        
        return True
    
    def _get_client_ip(self, request: Request) -> str:
        """Extract client IP from request headers"""
        # Check for forwarded headers (from reverse proxy)
        forwarded_for = request.headers.get("x-forwarded-for")
        if forwarded_for:
            return forwarded_for.split(",")[0].strip()
        
        real_ip = request.headers.get("x-real-ip")
        if real_ip:
            return real_ip
        
        return request.client.host if request.client else "unknown"

def has_required_role(user_roles: str, required_role: str) -> bool:
    """Check if user has the required role"""
    if not user_roles:
        return False
    
    # Admin role has all permissions
    if "ADMIN" in user_roles:
        return True
    
    # Check for specific role
    return required_role in user_roles

# Global instance for access from handlers
security_middleware_instance = None

def require_role(required_role: str):
    """Decorator to require specific role for endpoint access"""
    def decorator(func):
        async def wrapper(*args, **kwargs):
            # Extract request from args/kwargs
            request = None
            for arg in args:
                if isinstance(arg, Request):
                    request = arg
                    break
            
            if not request:
                raise HTTPException(
                    status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                    detail="Internal error: Request object not found"
                )
            
            user_roles = request.headers.get("x-user-roles", "")
            if not has_required_role(user_roles, required_role):
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail={
                        "error": "Access denied - insufficient privileges",
                        "code": "INSUFFICIENT_PRIVILEGES",
                        "required_role": required_role
                    }
                )
            
            return await func(*args, **kwargs)
        return wrapper
    return decorator