import requests
import math
from datetime import datetime, timedelta
from typing import Dict, List, Optional
from loguru import logger
from app.config import settings
from app.services.database import db_service

def compute_hotness_score(use_count: int, 
                         created_at: datetime, 
                         last_used_at: datetime, 
                         last_promoted_at: Optional[datetime] = None) -> float:
    """
    Compute hotness score for tag tiering
    Formula: (use_count * recency_factor) / age_penalty
    """
    now = datetime.utcnow()
    
    # Age penalty: older tags get lower scores
    age_days = (now - created_at).days + 1  # +1 to avoid division by zero
    age_penalty = math.log(age_days + 1)  # Logarithmic penalty
    
    # Recency factor: recently used tags get higher scores
    days_since_use = (now - last_used_at).days
    recency_factor = 1.0 / (days_since_use + 1)  # Inverse relationship
    
    # Promotion boost: recently promoted tags get temporary boost
    promotion_boost = 1.0
    if last_promoted_at:
        days_since_promotion = (now - last_promoted_at).days
        if days_since_promotion < 7:  # Boost for 7 days
            promotion_boost = 1.5
    
    hotness = (use_count * recency_factor * promotion_boost) / age_penalty
    return round(hotness, 3)

def determine_tier(hotness_score: float) -> int:
    """Determine tier based on hotness score"""
    if hotness_score >= 5.0:
        return 1  # Hot tier
    elif hotness_score >= 1.0:
        return 2  # Warm tier  
    else:
        return 3  # Cold tier

def get_tags_for_rebalancing(page: int = 0, limit: int = 100) -> List[Dict]:
    """
    Fetch tags from MongoDB for rebalancing using direct database access
    """
    logger.info(f"Fetching tags for rebalancing (page={page}, limit={limit})")
    
    # Use the database service to fetch real tags
    tags = db_service.get_tags_for_rebalancing(page=page, limit=limit)
    
    if not tags:
        logger.warning(f"No tags found for rebalancing (page={page})")
        return []
    
    logger.info(f"Successfully fetched {len(tags)} tags for rebalancing")
    return tags

def post_tier_update(tag: str, new_tier: int, user_id: str = "system") -> bool:
    """Update tag tier directly in the database"""
    logger.info(f"Updating tier for tag '{tag}' to tier {new_tier}")
    
    # Use the database service to update the tier directly
    success = db_service.update_tag_tier(tag, new_tier, user_id)
    
    if success:
        logger.info(f"Successfully updated tier for tag '{tag}' to {new_tier}")
    else:
        logger.error(f"Failed to update tier for tag '{tag}'")
    
    return success

def safe_parse_datetime(date_value) -> Optional[datetime]:
    """Safely parse datetime values from MongoDB"""
    if not date_value:
        return None
    
    try:
        if isinstance(date_value, str):
            return datetime.fromisoformat(date_value.replace("Z", "")).replace(tzinfo=None)
        elif isinstance(date_value, datetime):
            return date_value.replace(tzinfo=None)
        else:
            logger.warning(f"Unexpected date type: {type(date_value)} for value: {date_value}")
            return None
    except Exception as e:
        logger.error(f"Error parsing datetime {date_value}: {str(e)}")
        return None

def safe_get_int(value, default: int = 0) -> int:
    """Safely convert value to integer"""
    try:
        if isinstance(value, int):
            return value
        elif isinstance(value, str):
            return int(value)
        elif isinstance(value, float):
            return int(value)
        else:
            return default
    except (ValueError, TypeError):
        return default

def rebalance_all_tags():
    """
    Rebalance all tags based on hotness scores
    Paginates through all tags and updates tiers as needed
    """
    logger.info("Starting tag rebalancing process")
    
    # Check database connection
    if not db_service.is_connected():
        logger.error("Database not connected, cannot perform rebalancing")
        return
    
    page = 0
    limit = 100
    total_processed = 0
    total_rebalanced = 0
    
    try:
        while True:
            tags = get_tags_for_rebalancing(page=page, limit=limit)
            if not tags:
                break
                
            for tag_data in tags:
                try:
                    # Safely extract and parse data
                    tag_name = tag_data.get('tag', 'unknown')
                    metrics = tag_data.get('metrics', {})
                    meta = tag_data.get('meta', {})
                    
                    # Safely parse datetime strings
                    created_at = safe_parse_datetime(metrics.get("createdAt"))
                    last_used_at = safe_parse_datetime(metrics.get("lastUsedAt"))
                    last_promoted_at = safe_parse_datetime(metrics.get("lastPromotedAt"))
                    
                    # Safely get use count
                    use_count = safe_get_int(metrics.get("useCount"), 1)
                    
                    # Validate required data
                    if not created_at or not last_used_at:
                        logger.warning(f"Skipping tag '{tag_name}' - missing required date fields")
                        continue
                    
                    # Calculate hotness score
                    hotness = compute_hotness_score(
                        use_count=use_count,
                        created_at=created_at,
                        last_used_at=last_used_at,
                        last_promoted_at=last_promoted_at
                    )
                    
                    # Determine new tier
                    new_tier = determine_tier(hotness)
                    current_tier = safe_get_int(meta.get("tier"), 2)
                    
                    logger.debug(f"Tag '{tag_name}': hotness={hotness}, current_tier={current_tier}, new_tier={new_tier}")
                    
                    # Update tier if changed
                    if new_tier != current_tier:
                        success = post_tier_update(tag_name, new_tier)
                        if success:
                            logger.info(f"Rebalanced tag '{tag_name}': tier {current_tier} -> {new_tier} (hotness={hotness})")
                            total_rebalanced += 1
                        else:
                            logger.error(f"Failed to rebalance tag '{tag_name}'")
                    
                    total_processed += 1
                    
                except Exception as e:
                    logger.error(f"Error processing tag '{tag_data.get('tag', 'unknown')}': {str(e)}")
            
            page += 1
            
        logger.info(f"Tag rebalancing completed: {total_processed} processed, {total_rebalanced} rebalanced")
        
    except Exception as e:
        logger.error(f"Error during tag rebalancing: {str(e)}")
    
    # Update Prometheus metrics for observability
    try:
        # Import Prometheus client if available
        from prometheus_client import Counter, Histogram
        
        # Define metrics
        tag_rebalanced_total = Counter('mme_tag_rebalanced_total', 'Total number of tags rebalanced')
        tag_rebalancing_duration = Histogram('mme_tag_rebalancing_duration_seconds', 'Time spent rebalancing tags')
        
        # Update metrics
        tag_rebalanced_total.inc(total_rebalanced)
        
        # Note: Duration would be measured if we had timing information
        # tag_rebalancing_duration.observe(duration_seconds)
        
        logger.info(f"Updated Prometheus metrics: {total_rebalanced} tags rebalanced")
        
    except ImportError:
        logger.warning("Prometheus client not available, skipping metrics update")
    except Exception as e:
        logger.error(f"Failed to update Prometheus metrics: {str(e)}")
