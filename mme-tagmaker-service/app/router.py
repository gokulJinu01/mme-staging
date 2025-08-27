from fastapi import APIRouter, HTTPException, Header
from typing import Optional
from app.models.request import TagRequest
from app.services.llm_tagger import extract_cues
from app.services.merge import build_delta
from app.services.client import post_delta
from app.services.database import db_service
from app.config import settings

router = APIRouter(
    tags=["Tag Management"],
    responses={
        502: {"description": "Tagging service unavailable"},
        500: {"description": "Internal server error"}
    }
)

@router.get("/", 
           summary="Service Status",
           description="Check if the mme-tagmaker-service is running and operational")
async def root():
    """
    Simple health check endpoint that returns service status.
    Used for container orchestration and monitoring.
    """
    return {"message": "mme-tagmaker-service is running"}

@router.get("/queue-status",
           summary="Queue Status",
           description="Check failed delta queue status")
async def queue_status():
    """
    Check the status of the failed delta queue.
    Returns count of pending retries and queue file info.
    """
    import os
    queue_file = "/tmp/tagmaker_retry.jsonl"
    
    if not os.path.exists(queue_file):
        return {"queue_count": 0, "status": "empty"}
    
    try:
        with open(queue_file, "r") as f:
            lines = [line.strip() for line in f if line.strip()]
            queue_count = len(lines)
        
        file_size = os.path.getsize(queue_file)
        return {
            "queue_count": queue_count,
            "file_size_bytes": file_size,
            "status": "active" if queue_count > 0 else "empty"
        }
    except Exception as e:
        return {"error": str(e), "status": "error"}

@router.get("/database-status",
           summary="Database Status",
           description="Check MongoDB connection and tag statistics")
async def database_status():
    """
    Check MongoDB connection status and get tag statistics.
    Returns database connectivity and tag metrics.
    """
    try:
        # Check database connection
        is_connected = db_service.is_connected()
        
        # Get tag statistics if connected
        stats = db_service.get_tag_statistics() if is_connected else {"error": "Database not connected"}
        
        return {
            "database_connected": is_connected,
            "statistics": stats,
            "status": "healthy" if is_connected else "unhealthy"
        }
    except Exception as e:
        return {
            "database_connected": False,
            "error": str(e),
            "status": "error"
        }

@router.get("/database-debug",
           summary="Database Debug",
           description="Get detailed database service debug information")
async def database_debug():
    """
    Get detailed debug information about the database service instance.
    Returns connection status, metrics, and singleton information.
    """
    try:
        # Get detailed connection status
        status = db_service.get_connection_status()
        
        return {
            "debug_info": status,
            "service_instance_id": id(db_service),
            "service_type": type(db_service).__name__
        }
    except Exception as e:
        return {
            "error": str(e),
            "service_instance_id": id(db_service) if db_service else None,
            "service_type": type(db_service).__name__ if db_service else None
        }

@router.get("/database-direct-test",
           summary="Database Direct Test",
           description="Direct test of MongoDB connection")
async def database_direct_test():
    """
    Direct test of MongoDB connection bypassing is_connected() method.
    """
    try:
        # Test direct connection
        if not db_service.client:
            return {"status": "error", "message": "No MongoDB client"}
        
        # Test ping
        ping_result = db_service.client.admin.command('ping')
        
        # Test collection access
        if db_service.collection is None:
            return {"status": "error", "message": "No collection object"}
        
        # Test count
        count = db_service.collection.count_documents({}, limit=1)
        
        return {
            "status": "success",
            "ping_result": ping_result,
            "collection_accessible": True,
            "client_exists": db_service.client is not None,
            "database_exists": db_service.database is not None,
            "collection_exists": db_service.collection is not None,
            "connection_state": db_service._connection_state
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "client_exists": db_service.client is not None if db_service else False,
            "database_exists": db_service.database is not None if db_service else False,
            "collection_exists": db_service.collection is not None if db_service else False,
            "connection_state": getattr(db_service, '_connection_state', 'unknown') if db_service else 'unknown'
        }

@router.post("/manual-rebalance",
            summary="Manual Rebalance",
            description="Trigger manual tag rebalancing")
async def manual_rebalance():
    """
    Manually trigger tag rebalancing process.
    Useful for testing or immediate tier updates.
    """
    try:
        from app.services.tiering import rebalance_all_tags
        # Run in background to avoid timeout
        import threading
        thread = threading.Thread(target=rebalance_all_tags)
        thread.start()
        
        return {"message": "Rebalancing started", "status": "triggered"}
    except Exception as e:
        raise HTTPException(500, f"Failed to trigger rebalancing: {str(e)}")

@router.post("/extract-tags",
            summary="Extract Tags Only",
            description="Extract structured semantic tags from content using LLM without saving to tagging service",
            response_description="Extraction results with structured tags and confidence score",
            responses={
                200: {
                    "description": "Successfully extracted structured tags",
                    "content": {
                        "application/json": {
                            "example": {
                                "tags": [
                                    {
                                        "label": "IRAP submission timeline",
                                        "section": "funding-proposal",
                                        "origin": "agent",
                                        "scope": "shared",
                                        "type": "action",
                                        "confidence": 0.92,
                                        "links": ["CDAP form", "IRAP budget"],
                                        "usageCount": 1,
                                        "lastUsed": "2025-07-08T15:44:00Z"
                                    }
                                ],
                                "confidence": 0.95,
                                "primary_tag": "IRAP submission timeline"
                            }
                        }
                    }
                }
            })
async def extract_tags_only(req: TagRequest):
    """
    Extract structured semantic tags from content using LLM without saving to tagging service.
    
    **Process Flow:**
    1. Analyze content with LLM to extract key factual sentences
    2. Convert sentences to structured Tag objects with metadata
    3. Return structured tags with full metadata
    
    **Tag Structure:**
    Each tag includes label, section, origin, scope, type, confidence, links, usageCount, and lastUsed
    """
    try:
        tags, conf, primary_tag = extract_cues(req.content)
        
        if not tags:
            raise HTTPException(400, "No extractable content found")
            
        return {
            "tags": tags, 
            "confidence": conf,
            "primary_tag": primary_tag
        }
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, f"Tag extraction failed: {str(e)}")

@router.post("/generate-and-save",
            summary="Extract Tags and Save",
            description="Extract semantic tags from content using LLM and save to tagging service (Optional feature)",
            response_description="Extraction results with cues and confidence score",
            responses={
                200: {
                    "description": "Successfully extracted and saved tags",
                    "content": {
                        "application/json": {
                            "example": {
                                "saved": True,
                                "cues": ["project:final submission completed", "deadline:met successfully"],
                                "confidence": 0.95
                            }
                        }
                    }
                },
                502: {
                    "description": "Tagging service unavailable",
                    "content": {
                        "application/json": {
                            "example": {"detail": "tagging-service unavailable"}
                        }
                    }
                }
            })
async def generate_and_save(req: TagRequest, authorization: Optional[str] = Header(None)):
    """
    Extract semantic cues from agent output content and save as tags.
    
    **Note**: This feature requires the tagging service to be enabled.
    If tagging service is disabled, only tag extraction will be performed.
    
    **Process Flow:**
    1. Analyze content with LLM to extract key factual sentences
    2. Convert sentences to structured cues (head:detail format)
    3. Generate SHA256 hashes for deduplication
    4. Build delta operations for tag database
    5. Post to tagging-service with JWT authentication (if enabled)
    
    **Cue Format:**
    Each cue follows the pattern `head:detail` where:
    - `head`: Primary concept or entity
    - `detail`: Contextual information or action
    
    **Error Handling:**
    - Failed requests are queued to disk for automatic retry
    - Service returns 502 if tagging-service is unavailable
    - Confidence scores reflect extraction quality
    """
    try:
        tags, conf, primary_tag = extract_cues(req.content)
        
        if not tags:
            raise HTTPException(400, "No extractable content found")
            
        # Convert structured tags to legacy format for delta building
        cues = [f"{tag.label}:{tag.links[0] if tag.links else ''}" for tag in tags]
        hashes = [tag.label for tag in tags]  # Use label as hash for now
        
        # primary_tag is now semantically selected by extract_cues
        delta = build_delta(primary_tag, cues, hashes, {})
        
        # Extract JWT token from Authorization header
        jwt_token = None
        if authorization and authorization.startswith("Bearer "):
            jwt_token = authorization[7:]  # Remove "Bearer " prefix
        
        # Check if tagging service is enabled
        if not settings.enable_tagging_service:
            return {
                "saved": False,
                "tags": tags, 
                "confidence": conf,
                "primary_tag": primary_tag,
                "message": "Tagging service is disabled, only extraction performed"
            }
        
        ok = post_delta(delta, req.userId, req.orgId, jwt_token)
        if not ok:
            raise HTTPException(502, "tagging-service unavailable")
            
        return {
            "saved": True, 
            "tags": tags, 
            "confidence": conf,
            "primary_tag": primary_tag
        }
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, f"Tag extraction failed: {str(e)}")
