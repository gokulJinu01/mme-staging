from pydantic import BaseModel, Field
from typing import List, Dict, Optional, Literal
from datetime import datetime

class Tag(BaseModel):
    label: str
    section: Optional[str] = None
    origin: Literal["agent", "system", "user", "unknown"] = "agent"
    scope: Literal["local", "shared", "global"] = "shared"
    type: str = "concept"  # Allow any string for domain types
    confidence: Optional[float] = 0.8
    links: List[str] = []
    usageCount: int = 0
    lastUsed: Optional[datetime] = None

# Legacy models for backward compatibility
class TagMeta(BaseModel):
    section: str
    status: str = "active"
    tier: int = 2
    scope: str = "user"
    confidence: float
    source: str
    isFrozen: bool = False

class TagMetrics(BaseModel):
    useCount: int = 1
    createdAt: datetime
    lastUsedAt: datetime
    lastPromotedAt: Optional[datetime] = None
    lastMergedAt: datetime

class TagContext(BaseModel):
    originMemoryIds: List[str] = []
    cues: List[str] = Field(default_factory=list, max_items=200)
    relatedTags: Dict[str, int] = {}
    cueHashes: List[str] = []

class LegacyTag(BaseModel): # Old Tag model for backward compatibility
    tag: str
    meta: TagMeta
    metrics: TagMetrics
    context: TagContext
