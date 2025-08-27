"""
Security handlers for MME Tagmaker Service
Provides security endpoints for health checks and metrics.
"""

from fastapi import APIRouter, Request, HTTPException, status, Depends
from typing import Dict, Any
from datetime import datetime
import asyncio

from .middleware import SecurityMiddleware, require_role

security_router = APIRouter(prefix="/security", tags=["security"])

# Global security middleware instance (will be set during app initialization)
_security_middleware: SecurityMiddleware = None

def set_security_middleware(middleware: SecurityMiddleware):
    """Set the global security middleware instance"""
    global _security_middleware
    _security_middleware = middleware

def get_security_middleware() -> SecurityMiddleware:
    """Get the security middleware instance"""
    if _security_middleware is None:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Security middleware not initialized"
        )
    return _security_middleware

@security_router.get("/health")
async def get_security_health(request: Request):
    """Get security health status - accessible to all authenticated users"""
    try:
        middleware = get_security_middleware()
        
        # Basic health check
        health_status = {
            "status": "UP",
            "service": "mme-tagmaker-security",
            "components": {
                "rate_limiter": "UP",
                "threat_detector": "UP", 
                "audit_logger": "UP",
                "security_monitor": "UP"
            },
            "config": {
                "rate_limit_enabled": middleware.config.rate_limit_enabled,
                "threat_detection": middleware.config.threat_detection,
                "audit_logging": middleware.config.audit_logging,
                "security_headers": middleware.config.security_headers,
                "input_validation": middleware.config.input_validation
            },
            "timestamp": datetime.utcnow().isoformat()
        }
        
        return health_status
        
    except Exception as e:
        return {
            "status": "DOWN",
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat()
        }

@security_router.get("/metrics")
@require_role("ADMIN")
async def get_security_metrics(request: Request):
    """Get detailed security metrics - admin only"""
    try:
        middleware = get_security_middleware()
        user_id = request.headers.get("x-user-id", "admin")
        client_ip = middleware._get_client_ip(request)
        
        # Log metrics access
        middleware.auditor.log_security_event(
            "SECURITY_METRICS_ACCESS", user_id, client_ip,
            "Security metrics accessed by admin user", "/security/metrics"
        )
        
        # Get current metrics
        metrics = middleware.metrics.get_metrics()
        
        # Add additional runtime information
        metrics.update({
            "rate_limiter_stats": {
                "active_limits": len(middleware.rate_limiter.requests),
                "blocked_clients": len(middleware.rate_limiter.blocked)
            },
            "system_info": {
                "uptime": datetime.utcnow().isoformat(),
                "service": "mme-tagmaker-service",
                "security_version": "1.0.0"
            }
        })
        
        return {
            "status": "success",
            "data": metrics,
            "timestamp": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to retrieve security metrics: {str(e)}"
        )

@security_router.post("/test/rate-limit")
@require_role("ADMIN")
async def test_rate_limit(request: Request):
    """Test rate limiting functionality - admin only"""
    try:
        middleware = get_security_middleware()
        user_id = request.headers.get("x-user-id", "test-user")
        client_ip = middleware._get_client_ip(request)
        
        # Test rate limiting
        allowed, retry_after = middleware.rate_limiter.is_allowed(
            user_id, client_ip, "/test/rate-limit"
        )
        
        return {
            "status": "success",
            "allowed": allowed,
            "retry_after": retry_after,
            "test_endpoint": "/test/rate-limit",
            "timestamp": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Rate limit test failed: {str(e)}"
        )

@security_router.post("/test/threat-detection")
@require_role("ADMIN") 
async def test_threat_detection(request: Request):
    """Test threat detection functionality - admin only"""
    try:
        middleware = get_security_middleware()
        
        # Test threat detection with the current request
        threat = middleware.threat_detector.analyze_request(request)
        
        return {
            "status": "success",
            "threat_detected": threat is not None,
            "threat_details": threat if threat else None,
            "timestamp": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Threat detection test failed: {str(e)}"
        )

@security_router.get("/config")
@require_role("ADMIN")
async def get_security_config(request: Request):
    """Get current security configuration - admin only"""
    try:
        middleware = get_security_middleware()
        
        config_data = {
            "rate_limit_enabled": middleware.config.rate_limit_enabled,
            "threat_detection": middleware.config.threat_detection,
            "audit_logging": middleware.config.audit_logging,
            "security_headers": middleware.config.security_headers,
            "input_validation": middleware.config.input_validation,
            "max_request_size": middleware.config.max_request_size,
            "rate_limit_per_minute": middleware.config.rate_limit_per_minute,
            "rate_limit_burst": middleware.config.rate_limit_burst
        }
        
        return {
            "status": "success",
            "config": config_data,
            "timestamp": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to retrieve security config: {str(e)}"
        )

@security_router.get("/audit/recent")
@require_role("ADMIN")
async def get_recent_audit_events(request: Request, limit: int = 100):
    """Get recent audit events - admin only"""
    try:
        # Get security metrics from middleware
        from app.security.middleware import security_middleware_instance
        
        if security_middleware_instance:
            metrics = security_middleware_instance.metrics.get_metrics()
            
            # Generate recent events summary from metrics
            events = []
            
            # Rate limit violations
            if metrics.get("rate_limit_violations", 0) > 0:
                events.append({
                    "event_type": "RATE_LIMIT_VIOLATION",
                    "count": metrics["rate_limit_violations"],
                    "severity": "MEDIUM",
                    "description": f"Total rate limit violations: {metrics['rate_limit_violations']}"
                })
            
            # Threat detections
            if metrics.get("threat_detections", 0) > 0:
                events.append({
                    "event_type": "THREAT_DETECTION", 
                    "count": metrics["threat_detections"],
                    "severity": "HIGH",
                    "description": f"Total threats detected: {metrics['threat_detections']}"
                })
            
            # Unauthorized access attempts
            if metrics.get("unauthorized_access", 0) > 0:
                events.append({
                    "event_type": "UNAUTHORIZED_ACCESS",
                    "count": metrics["unauthorized_access"], 
                    "severity": "HIGH",
                    "description": f"Total unauthorized access attempts: {metrics['unauthorized_access']}"
                })
            
            # Add threat breakdown
            threats_by_type = metrics.get("threats_by_type", {})
            for threat_type, count in threats_by_type.items():
                events.append({
                    "event_type": f"THREAT_{threat_type}",
                    "count": count,
                    "severity": "MEDIUM",
                    "description": f"{threat_type} threats detected: {count}"
                })
            
            return {
                "status": "success",
                "events": events[:limit],
                "total_events": len(events),
                "metrics_last_updated": metrics.get("last_updated"),
                "limit": limit,
                "timestamp": datetime.utcnow().isoformat(),
                "note": "These are aggregated security metrics. Check application logs for detailed events."
            }
        else:
            return {
                "status": "success",
                "events": [],
                "message": "Security middleware not initialized",
                "limit": limit,
                "timestamp": datetime.utcnow().isoformat()
            }
        
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to retrieve audit events: {str(e)}"
        )