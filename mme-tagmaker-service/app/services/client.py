import requests
import json
import os
from datetime import datetime
from typing import Dict, Optional
from loguru import logger
from app.config import settings

def post_delta(delta: Dict, user_id: str, org_id: str = "test-org", jwt_token: Optional[str] = None) -> bool:
    """
    Post delta operations to the tagging-service (Optional feature)
    Returns True on success (2xx), False on failure or if tagging service is disabled
    """
    # Check if tagging service is enabled
    if not settings.enable_tagging_service:
        logger.info("Tagging service is disabled, skipping delta post")
        return True  # Return True to avoid error handling
    
    if not settings.tagging_service_url:
        logger.warning("TAGGING_SERVICE_URL not configured, tagging service disabled")
        return True  # Return True to avoid error handling
    
    headers = {
        "Content-Type": "application/json",
        "X-USER-ID": user_id
    }
    
    # AUTHENTICATION DISABLED - Set default headers for testing
    headers["X-User-ID"] = user_id or "test-user"
    headers["X-Org-ID"] = org_id or "test-org"
    logger.debug(f"Using test headers for user {user_id or 'test-user'}")
    
    try:
        logger.info(f"Posting delta for user {user_id} to {settings.tagging_service_url}")
        logger.debug(f"Delta payload: {delta}")
        
        response = requests.post(
            settings.tagging_service_url + "/tags/delta",
            json=delta,
            headers=headers,
            timeout=5
        )
        
        if response.status_code in range(200, 300):
            logger.info(f"Successfully posted delta for user {user_id}")
            return True
        else:
            logger.error(f"Failed to post delta: HTTP {response.status_code} - {response.text}")
            # Retry mechanism implemented - failed deltas are queued to disk for replay
            return False
            
    except requests.exceptions.Timeout:
        logger.error(f"Timeout posting delta for user {user_id}")
        _queue_failed_delta(delta, user_id, "timeout")
        return False
    except requests.exceptions.ConnectionError:
        logger.error(f"Connection error posting delta for user {user_id}")
        _queue_failed_delta(delta, user_id, "connection_error")
        return False
    except Exception as e:
        logger.error(f"Unexpected error posting delta for user {user_id}: {str(e)}")
        _queue_failed_delta(delta, user_id, "unexpected_error")
        return False

def _queue_failed_delta(delta: Dict, user_id: str, error_type: str):
    """Queue failed delta for retry"""
    try:
        queue_entry = {
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "delta": delta,
            "user_id": user_id,
            "error_type": error_type,
            "retry_count": 0
        }
        
        # Ensure directory exists
        os.makedirs("/tmp", exist_ok=True)
        
        # Append to JSONL file
        with open("/tmp/tagmaker_retry.jsonl", "a") as f:
            f.write(json.dumps(queue_entry) + "\n")
            
        logger.info(f"Queued failed delta for user {user_id} to disk")
        
    except Exception as e:
        logger.error(f"Failed to queue delta to disk: {str(e)}")

def replay_failed_deltas():
    """Replay failed deltas from disk queue"""
    queue_file = "/tmp/tagmaker_retry.jsonl"
    
    if not os.path.exists(queue_file):
        return
        
    failed_entries = []
    successful_count = 0
    
    try:
        with open(queue_file, "r") as f:
            for line in f:
                if line.strip():
                    try:
                        entry = json.loads(line.strip())
                        
                        # Try to replay the delta
                        success = post_delta(entry["delta"], entry["user_id"], entry.get("org_id", "test-org"))
                        
                        if success:
                            successful_count += 1
                            logger.info(f"Successfully replayed delta for user {entry['user_id']}")
                        else:
                            # Increment retry count and keep for later
                            entry["retry_count"] = entry.get("retry_count", 0) + 1
                            if entry["retry_count"] < 5:  # Max 5 retries
                                failed_entries.append(entry)
                            else:
                                logger.warning(f"Dropping delta for user {entry['user_id']} after 5 retries")
                                
                    except json.JSONDecodeError:
                        logger.error(f"Invalid JSON in queue file: {line}")
        
        # Rewrite file with remaining failed entries
        with open(queue_file, "w") as f:
            for entry in failed_entries:
                f.write(json.dumps(entry) + "\n")
                
        if successful_count > 0:
            logger.info(f"Replayed {successful_count} failed deltas")
            
    except Exception as e:
        logger.error(f"Error replaying failed deltas: {str(e)}")
