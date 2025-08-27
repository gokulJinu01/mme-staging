"""
Database service for direct MongoDB access
Handles tag retrieval for rebalancing without going through mme-tagging-service

Implements singleton pattern to ensure single database service instance across all contexts.
Provides async-compatible methods for FastAPI environment with proper connection management.
"""

import os
import asyncio
import threading
from datetime import datetime
from typing import List, Dict, Optional, Any, Union
from pymongo import MongoClient
from pymongo.collection import Collection
from pymongo.database import Database
from pymongo.errors import ServerSelectionTimeoutError, ConnectionFailure, AutoReconnect, OperationFailure
from loguru import logger
from app.config import settings


class DatabaseService:
    """Singleton database service for MongoDB operations.
    
    Ensures single instance across all application contexts with proper
    connection management, pooling, and async compatibility.
    """
    
    _instance = None
    _lock = threading.Lock()
    _initialized = False
    
    def __new__(cls):
        """Singleton pattern implementation using thread-safe lazy initialization."""
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super(DatabaseService, cls).__new__(cls)
        return cls._instance
    
    def __init__(self):
        """Initialize database service (only once due to singleton pattern)."""
        # Prevent re-initialization
        if self._initialized:
            return
            
        with self._lock:
            if self._initialized:
                return
                
            # Connection objects
            self.client: Optional[MongoClient] = None
            self.database: Optional[Database] = None
            self.collection: Optional[Collection] = None
            
            # Connection state tracking
            self._connection_state = "disconnected"
            self._last_error: Optional[str] = None
            self._reconnect_attempts = 0
            self._max_reconnect_attempts = 3
            self._connection_health_score = 0.0
            
            # Connection pooling settings
            self._connection_pool_settings = {
                'maxPoolSize': 10,
                'minPoolSize': 1,
                'maxIdleTimeMS': 30000,  # 30 seconds
                'waitQueueTimeoutMS': 5000,  # 5 seconds
                'serverSelectionTimeoutMS': 5000,  # 5 seconds
                'connectTimeoutMS': 10000,  # 10 seconds
                'socketTimeoutMS': 20000,  # 20 seconds
                'retryWrites': True,
                'retryReads': True,
            }
            
            # Performance metrics
            self._operation_count = 0
            self._error_count = 0
            self._last_operation_time = None
            
            # Initialize connection
            self._connect()
            self._initialized = True
    
    def _connect(self) -> bool:
        """Establish connection to MongoDB with enhanced error handling and pooling."""
        try:
            self._connection_state = "connecting"
            
            # Parse MongoDB URI to get database name
            uri = settings.mongodb_uri
            db_name = settings.mongodb_database
            
            logger.info(f"Connecting to MongoDB: {db_name}.{settings.mongodb_collection}")
            
            # Create MongoDB client with connection pooling
            self.client = MongoClient(
                uri,
                **self._connection_pool_settings
            )
            
            # Test connection with ping
            self.client.admin.command('ping')
            
            # Get database and collection
            self.database = self.client[db_name]
            self.collection = self.database[settings.mongodb_collection]
            
            # Verify collection access
            self.collection.count_documents({}, limit=1)
            
            # Update connection state
            self._connection_state = "connected"
            self._last_error = None
            self._reconnect_attempts = 0
            self._connection_health_score = 1.0
            
            logger.info(f"✅ Successfully connected to MongoDB: {db_name}.{settings.mongodb_collection}")
            return True
            
        except (ServerSelectionTimeoutError, ConnectionFailure) as e:
            error_msg = f"MongoDB connection failed - server unreachable: {str(e)}"
            self._handle_connection_error(error_msg)
            return False
            
        except OperationFailure as e:
            error_msg = f"MongoDB connection failed - authentication or permission error: {str(e)}"
            self._handle_connection_error(error_msg)
            return False
            
        except Exception as e:
            error_msg = f"MongoDB connection failed - unexpected error: {str(e)}"
            self._handle_connection_error(error_msg)
            return False
    
    def _handle_connection_error(self, error_msg: str):
        """Handle connection errors with proper state management."""
        logger.error(error_msg)
        self._connection_state = "error"
        self._last_error = error_msg
        self._connection_health_score = 0.0
        
        # Clean up connection objects
        if self.client:
            try:
                self.client.close()
            except:
                pass
        
        self.client = None
        self.database = None
        self.collection = None
    
    def _reconnect(self) -> bool:
        """Attempt to reconnect with exponential backoff."""
        if self._reconnect_attempts >= self._max_reconnect_attempts:
            logger.warning(f"Maximum reconnection attempts ({self._max_reconnect_attempts}) reached")
            return False
            
        self._reconnect_attempts += 1
        logger.info(f"Attempting to reconnect to MongoDB (attempt {self._reconnect_attempts}/{self._max_reconnect_attempts})")
        
        return self._connect()
    
    def is_connected(self) -> bool:
        """Check if database connection is active with comprehensive health check."""
        if not self.client or self._connection_state != "connected":
            return False
        
        try:
            # Perform ping test
            self.client.admin.command('ping')
            
            # Verify collection access (use proper None comparison for PyMongo collections)
            if self.collection is not None:
                self.collection.count_documents({}, limit=1)
            
            # Update health metrics
            self._connection_health_score = min(1.0, self._connection_health_score + 0.1)
            return True
            
        except (ServerSelectionTimeoutError, ConnectionFailure, AutoReconnect):
            logger.warning("Database connection lost - attempting reconnection")
            self._connection_health_score = max(0.0, self._connection_health_score - 0.3)
            
            # Attempt reconnection
            if self._connection_health_score <= 0.3:
                return self._reconnect()
            return False
            
        except Exception as e:
            logger.error(f"Database health check failed: {str(e)}")
            self._connection_health_score = max(0.0, self._connection_health_score - 0.2)
            return False
    
    async def is_connected_async(self) -> bool:
        """Async wrapper for connection check."""
        def _check():
            return self.is_connected()
        
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, _check)
    
    def _execute_operation(self, operation_func, operation_name: str, *args, **kwargs):
        """Execute database operation with error handling and metrics tracking."""
        if not self.is_connected():
            logger.error(f"Cannot execute {operation_name}: database not connected")
            self._error_count += 1
            return None
        
        try:
            self._operation_count += 1
            self._last_operation_time = datetime.utcnow()
            
            result = operation_func(*args, **kwargs)
            
            # Update health score on successful operation
            self._connection_health_score = min(1.0, self._connection_health_score + 0.05)
            return result
            
        except (ServerSelectionTimeoutError, ConnectionFailure, AutoReconnect) as e:
            logger.error(f"Database operation {operation_name} failed due to connection issue: {str(e)}")
            self._error_count += 1
            self._connection_health_score = max(0.0, self._connection_health_score - 0.2)
            
            # Attempt reconnection for next operation
            self._reconnect()
            return None
            
        except Exception as e:
            logger.error(f"Database operation {operation_name} failed: {str(e)}")
            self._error_count += 1
            self._connection_health_score = max(0.0, self._connection_health_score - 0.1)
            return None
    
    async def _execute_operation_async(self, operation_func, operation_name: str, *args, **kwargs):
        """Async wrapper for database operations."""
        def _execute():
            return self._execute_operation(operation_func, operation_name, *args, **kwargs)
        
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, _execute)
    
    def get_tags_for_rebalancing(self, page: int = 0, limit: int = 100) -> List[Dict]:
        """
        Fetch tags from MongoDB for rebalancing
        Returns list of tag documents with metrics for hotness calculation
        """
        def _operation():
            # Build aggregation pipeline to extract tag metrics
            pipeline = [
                # Group by tags to aggregate metrics
                {
                    "$unwind": "$tags"
                },
                {
                    "$group": {
                        "_id": "$tags",
                        "tag": {"$first": "$tags"},
                        "useCount": {"$sum": 1},
                        "createdAt": {"$min": "$createdAt"},
                        "lastUsedAt": {"$max": "$createdAt"},
                        "confidence": {"$avg": "$confidence"},
                        "section": {"$first": "$section"},
                        "status": {"$first": "$status"},
                        "source": {"$first": "$source"}
                    }
                },
                # Add computed fields to match expected structure
                {
                    "$addFields": {
                        "meta": {
                            "tier": 2,  # Default tier
                            "confidence": "$confidence",
                            "section": "$section",
                            "status": "$status",
                            "source": "$source"
                        },
                        "metrics": {
                            "useCount": "$useCount",
                            "createdAt": "$createdAt",
                            "lastUsedAt": "$lastUsedAt",
                            "lastPromotedAt": None
                        }
                    }
                },
                # Sort by last used date (most recent first)
                {
                    "$sort": {"metrics.lastUsedAt": -1}
                },
                # Pagination
                {
                    "$skip": page * limit
                },
                {
                    "$limit": limit
                }
            ]
            
            # Execute aggregation
            cursor = self.collection.aggregate(pipeline)
            results = list(cursor)
            
            logger.info(f"Fetched {len(results)} tags for rebalancing (page={page}, limit={limit})")
            return results
        
        result = self._execute_operation(_operation, "get_tags_for_rebalancing")
        return result if result is not None else []
    
    async def get_tags_for_rebalancing_async(self, page: int = 0, limit: int = 100) -> List[Dict]:
        """Async version of get_tags_for_rebalancing."""
        return await self._execute_operation_async(
            lambda: self.get_tags_for_rebalancing(page, limit),
            "get_tags_for_rebalancing_async"
        ) or []
    
    def update_tag_tier(self, tag: str, new_tier: int, user_id: str = "system") -> bool:
        """
        Update tag tier in the database
        Updates all memory blocks containing this tag
        """
        def _operation():
            # Update all memory blocks containing this tag
            filter_query = {"tags": tag}
            update_query = {
                "$set": {
                    "meta.tier": new_tier,
                    "meta.lastUpdated": datetime.utcnow()
                },
                "$inc": {"meta.rebalanceCount": 1}
            }
            
            result = self.collection.update_many(filter_query, update_query)
            
            logger.info(f"Updated tier for tag '{tag}': {result.modified_count} documents modified")
            return result.modified_count > 0
        
        result = self._execute_operation(_operation, "update_tag_tier")
        return result if result is not None else False
    
    async def update_tag_tier_async(self, tag: str, new_tier: int, user_id: str = "system") -> bool:
        """Async version of update_tag_tier."""
        result = await self._execute_operation_async(
            lambda: self.update_tag_tier(tag, new_tier, user_id),
            "update_tag_tier_async"
        )
        return result if result is not None else False
    
    def get_tag_statistics(self) -> Dict:
        """
        Get overall tag statistics for monitoring
        """
        def _operation():
            # Get total number of unique tags
            unique_tags = self.collection.distinct("tags")
            
            # Get total number of memory blocks
            total_blocks = self.collection.count_documents({})
            
            # Get recent activity (last 24 hours)
            yesterday = datetime.utcnow().replace(hour=0, minute=0, second=0, microsecond=0)
            recent_blocks = self.collection.count_documents({
                "createdAt": {"$gte": yesterday}
            })
            
            return {
                "total_tags": len(unique_tags),
                "total_blocks": total_blocks,
                "recent_blocks_24h": recent_blocks,
                "database": settings.mongodb_database,
                "collection": settings.mongodb_collection
            }
        
        result = self._execute_operation(_operation, "get_tag_statistics")
        return result if result is not None else {"error": "Database operation failed"}
    
    async def get_tag_statistics_async(self) -> Dict:
        """Async version of get_tag_statistics."""
        result = await self._execute_operation_async(
            lambda: self.get_tag_statistics(),
            "get_tag_statistics_async"
        )
        return result if result is not None else {"error": "Database operation failed"}
    
    def get_connection_status(self) -> Dict[str, Any]:
        """Get detailed connection status and metrics."""
        return {
            "connection_state": self._connection_state,
            "is_connected": self.is_connected(),
            "health_score": self._connection_health_score,
            "last_error": self._last_error,
            "reconnect_attempts": self._reconnect_attempts,
            "max_reconnect_attempts": self._max_reconnect_attempts,
            "operation_count": self._operation_count,
            "error_count": self._error_count,
            "error_rate": self._error_count / max(1, self._operation_count),
            "last_operation_time": self._last_operation_time.isoformat() if self._last_operation_time else None,
            "pool_settings": self._connection_pool_settings,
            "singleton_id": id(self),
            "initialized": self._initialized
        }
    
    async def get_connection_status_async(self) -> Dict[str, Any]:
        """Async version of get_connection_status."""
        def _get_status():
            return self.get_connection_status()
        
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, _get_status)
    
    def close(self):
        """Close database connection and reset singleton state."""
        logger.info("Closing database connection...")
        
        if self.client:
            try:
                self.client.close()
                logger.info("MongoDB connection closed successfully")
            except Exception as e:
                logger.error(f"Error closing MongoDB connection: {str(e)}")
        
        # Reset connection objects
        self.client = None
        self.database = None
        self.collection = None
        
        # Reset state
        self._connection_state = "disconnected"
        self._connection_health_score = 0.0
        
        logger.info("Database service shutdown complete")


# Global singleton database service instance
# This ensures the same instance is used throughout the application
db_service = DatabaseService()


def get_database_service() -> DatabaseService:
    """
    Get the singleton database service instance.
    
    This function provides a clean interface for dependency injection
    and ensures the same instance is always returned.
    """
    return db_service


async def get_database_service_async() -> DatabaseService:
    """
    Async version of get_database_service for FastAPI dependency injection.
    """
    return get_database_service()