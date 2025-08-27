"""
Edge Learning Admin Routes

Provides admin endpoints for edge learning operations.
"""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import Optional
from app.jobs.edge_learning import run_edge_learning_for_org

router = APIRouter(prefix="/admin", tags=["edge-learning"])

class EdgeLearningRequest(BaseModel):
    orgId: str
    hours: Optional[int] = None

class EdgeLearningResponse(BaseModel):
    orgId: str
    updated: int
    pruned: int
    duration_seconds: float
    status: str

@router.post("/edge-learn/replay", response_model=EdgeLearningResponse)
async def replay_edge_learning(request: EdgeLearningRequest):
    """
    Trigger edge learning for a specific organization.
    
    Args:
        request: EdgeLearningRequest with orgId and optional hours
        
    Returns:
        EdgeLearningResponse with results summary
    """
    try:
        # Run edge learning for the specified org
        result = run_edge_learning_for_org(
            org_id=request.orgId,
            hours=request.hours
        )
        
        return EdgeLearningResponse(
            orgId=request.orgId,
            updated=result['updated'],
            pruned=result['pruned'],
            duration_seconds=result['duration_seconds'],
            status="completed"
        )
        
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Edge learning failed: {str(e)}"
        )
